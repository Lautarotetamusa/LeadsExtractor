package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"leadsextractor/models"
	"leadsextractor/store"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

type PropertyPublishPayload struct {
    PropertyID  int16   `json:"property_id"`
    Portal      string  `json:"portal"`
}

type PublishPayload struct {
    Properties []PropertyPublishPayload `json:"properties" validate:"required"`
}

type PublishedPropertyHandler struct {
	storer          store.PublishedPropertyStorer
    propertyStorer  store.PropertyPortalStore
    appHost string
    logger  *slog.Logger

	mu              sync.Mutex
    queue           []*PropertyPublishPayload
	current         *PropertyPublishPayload
}

func NewPublishedPropertyHandler(storer store.PublishedPropertyStorer, propertyStorer store.PropertyPortalStore, appHost string) *PublishedPropertyHandler {
	return &PublishedPropertyHandler{
        storer: storer,
        propertyStorer: propertyStorer,
        appHost: appHost,
        queue: make([]*PropertyPublishPayload, 0),
        logger: slog.Default(),
    }
}

func (h *PublishedPropertyHandler) WithLogger(logger *slog.Logger) {
    h.logger = logger
}

func (h *PublishedPropertyHandler) RegisterRoutes(router *mux.Router) {
	r := router.PathPrefix("/property/{propId}/publications").Subrouter()
    r.Methods(http.MethodOptions)

    // Property publications
	router.HandleFunc("/publish", HandleErrors(h.Publish)).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("", HandleErrors(h.GetPublications)).Methods(http.MethodGet)
	r.HandleFunc("/{portal}", HandleErrors(h.GetPublication)).Methods(http.MethodGet)
	r.HandleFunc("/{portal}", HandleErrors(h.Update)).Methods(http.MethodPut)
}

// CreateHandler handles property creation
func (h *PublishedPropertyHandler) Publish(w http.ResponseWriter, r *http.Request) error {
	var payload PublishPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return jsonErr(err)
	}

    if err := validate.Struct(payload); err != nil {
        return ErrBadRequest(err.Error())
    }

    if len(payload.Properties) == 0 {
        return ErrBadRequest("must publish at least one property")
    }

    // Create or update all the published property status to "in_queue"
    // Add all the properties to the queue
    // Consume the queue, get the first prop, if the status its not 'finished' or 'failed' call the publication app
    // When the status its 'finished' or 'failed' drop the first prop and go to the next

    h.mu.Lock()
	defer h.mu.Unlock()
    for _, pp := range payload.Properties {
        prop, _ := h.storer.GetOne(pp.Portal, int64(pp.PropertyID))
        if prop != nil {
            // If the property its already publicated on this portal 
            //   and the status its not "published" then dont add it to the queue
            if prop.Status == store.StatusInQueue || prop.Status == store.StatusInProgress {
                continue
                // return ErrBadRequest(fmt.Sprintf("the property %d has a publication in progress, wait until the end", pp.PropertyID))
            }

            // The property exists but status its "published", then republish
            err := h.storer.UpdateStatus(pp.Portal, int64(pp.PropertyID), store.StatusInQueue)
            if err != nil {
                return err
            }
        } else {
            // If the property does not have a publication on the portal, create it
            err := h.storer.Create(&store.PublishedProperty{
                PropertyID: models.NullInt16{Int16: pp.PropertyID, Valid: true},
                Portal: models.NullString{String: pp.Portal,Valid: true},
                Status: store.StatusInProgress,
            })

            if err != nil {
                return err
            }
        }

        // Copy the prop to avoid loop variable issues
		copyProp := pp
        h.queue = append(h.queue, &copyProp)
    }

    // Start processing if not already running
	if h.current == nil {
		go h.processNextItem()
	}

    messageResponse(w, "publishing process has started")
    return nil
}

func (h *PublishedPropertyHandler) Update(w http.ResponseWriter, r *http.Request) error {
	portal := mux.Vars(r)["portal"]
	propertyIDStr := mux.Vars(r)["propId"]
	
	propId, err := strconv.Atoi(propertyIDStr)
	if err != nil {
		return ErrBadRequest("Invalid property ID")
	}

	var statusUpdate struct {
		Status store.PublishedStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&statusUpdate); err != nil {
		return jsonErr(err)
	}

	if !validStatus(statusUpdate.Status) {
		return ErrBadRequest("Invalid status value")
	}

	if err := h.storer.UpdateStatus(portal, int64(propId), statusUpdate.Status); err != nil {
		return err
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// Check if we're updating the current processing item
	if h.current != nil &&
		h.current.Portal == portal &&
		h.current.PropertyID == int16(propId) {
		// Clear current item
		h.current = nil
		
		go h.processNextItem()
	}

    messageResponse(w, "status updated successfully")
    return nil
}


func (h *PublishedPropertyHandler) processNextItem() {
	h.mu.Lock()
	defer h.mu.Unlock()

    // if the queue has a element
	if len(h.queue) == 0 || h.current != nil {
		return
	}

	// Get next item from queue
	h.current = h.queue[0]
	h.queue = h.queue[1:] // drop the first element

    // update the status of the current item to 'in_progress'
    err := h.storer.UpdateStatus(h.current.Portal, int64(h.current.PropertyID), store.StatusInProgress)
    if err != nil {
        h.logger.Error("error updating the status", 
            "portal", h.current.Portal,
            "id", h.current.PropertyID,
        )
        return
    }

	// Start processing in background
	go h.publish(h.current)
}

// At this point we know that the item exists, otherwise its not added to the queue
func (h *PublishedPropertyHandler) publish(item *PropertyPublishPayload) error {
    h.logger.Info("processing item", "property", item)

    // Get the property and the images
    property, err := h.propertyStorer.GetOne(int64(item.PropertyID))
    if err != nil {
        return err
    }
    if err := h.propertyStorer.GetImages(property); err != nil {
        return err
    }

    // Run the scraper to publish the property
    h.logger.Info("calling publicator app", "property", property)
    if err = runPublicatorApp(h.appHost, item.Portal, *property); err != nil {
        return err
    }

    return nil
}

func (h *PublishedPropertyHandler) handleProcessingError(item *PropertyPublishPayload, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.storer.UpdateStatus(item.Portal, int64(item.PropertyID), store.StatusFailed)
	
	// Clear current and process next
	h.current = nil
	go h.processNextItem()

	h.logger.Error("Publication failed",
		"property_id", item.PropertyID,
		"portal", item.Portal,
		"error", err,
	)
}

func (h *PublishedPropertyHandler) GetPublications(w http.ResponseWriter, r *http.Request) error {
    strId := mux.Vars(r)["propId"]
    id, err := strconv.ParseInt(strId, 10, 16)
    if err != nil {
        return InvalidPropID
    }

    props, err := h.storer.GetAllByProp(id)
    if err != nil {
        return err
    }

    dataResponse(w, props)
    return nil
}

func (h *PublishedPropertyHandler) GetPublication(w http.ResponseWriter, r *http.Request) error {
	portal := mux.Vars(r)["portal"]
	propertyIDStr := mux.Vars(r)["propId"]
	
	propId, err := strconv.Atoi(propertyIDStr)
	if err != nil {
		return ErrBadRequest("Invalid property ID")
	}

	pp, err := h.storer.GetOne(portal, int64(propId))
	if err != nil {
		return err
	}

    dataResponse(w, pp)
    return nil
}

func runPublicatorApp(appHost string, portal string, property store.PortalProp) error {
    url := fmt.Sprintf("%s/publish/%s", appHost, portal)
    slog.Info("publication post", "url", url)

    // pass the pointer to the Marshaller to correctly unmarshall the NullString fields
	jsonBody, err := json.Marshal(&property)
    if err != nil {
        return err
    }
	bodyReader := bytes.NewReader(jsonBody)

    client := http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest(http.MethodPost, url, bodyReader)
	if err != nil {
		return nil
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
    if err != nil {
        return err
    }
    if res.StatusCode != http.StatusCreated {
        return ErrBadRequest("error publishing the property")
    }

    return nil
}


func validStatus(s store.PublishedStatus) bool {
	switch s {
	case store.StatusInProgress, store.StatusPublished, store.StatusFailed:
		return true
	default:
		return false
	}
}
