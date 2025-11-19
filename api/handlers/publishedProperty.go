package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"leadsextractor/models"
	"leadsextractor/store"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

type PropertyPublishPayload struct {
    PropertyID  int16   `json:"property_id"`
    Portal      string  `json:"portal"`
    Plan        string  `json:"plan"`
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
    h := &PublishedPropertyHandler{
        storer: storer,
        propertyStorer: propertyStorer,
        appHost: appHost,
        queue: make([]*PropertyPublishPayload, 0),
        logger: slog.Default(),
    }

    h.logger.Info("getting the properties in queue")
    props, err := storer.GetAllByStatus(store.StatusInQueue)
    if err != nil {
        panic(fmt.Sprintf("cannot get the props in queue err: %w", err))
    }
    for _, prop := range props {
        h.queue = append(h.queue, &PropertyPublishPayload{
            PropertyID: prop.PropertyID.Int16,
            Portal: prop.Portal.String,
            Plan: prop.Plan.String,
        })
    }
    if len(props) > 0 {
        // h.current = h.queue[0]
		go h.processNextItem()
    }
    return h
}

func (h *PublishedPropertyHandler) WithLogger(logger *slog.Logger) {
    h.logger = logger
}

func (h *PublishedPropertyHandler) RegisterRoutes(router *mux.Router) {
	r := router.PathPrefix("/property/{propId}/publications").Subrouter()
    r.Methods(http.MethodOptions)

    // Property publications
	router.HandleFunc("/publish", HandleErrors(h.Publish)).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc("/publish-all", HandleErrors(h.RePublishAll)).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("", HandleErrors(h.GetPublications)).Methods(http.MethodGet)
	r.HandleFunc("/{portal}", HandleErrors(h.GetPublication)).Methods(http.MethodGet)
	r.HandleFunc("/{portal}", HandleErrors(h.Update)).Methods(http.MethodPut)
}


func (h *PublishedPropertyHandler) RePublishAll(w http.ResponseWriter, r *http.Request) error {
	publications, err := h.storer.GetAll()
	if err != nil {
		return err
	}

	var properties []PropertyPublishPayload
	for _, pub := range publications {
		properties = append(properties, PropertyPublishPayload{
			PropertyID: pub.PropertyID.Int16,
			Portal: pub.Portal.String,
			Plan: pub.Plan.String,
		})
	}

	payload := PublishPayload {
		Properties: properties,
	}
	fmt.Printf("%#v\n", payload)

	if err := h.addToQueue(payload); err != nil {
		return err
	}

    messageResponse(w, "publishing process has started")
    return nil
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

	if err := h.addToQueue(payload); err != nil {
		return err
	}

    messageResponse(w, "publishing process has started")
    return nil
}

// Create or update all the published property status to "in_queue"
// Add all the properties to the queue
// Consume the queue, get the first prop, if the status its not 'finished' or 'failed' call the publication app
// When the status its 'finished' or 'failed' drop the first prop and go to the next
func (h *PublishedPropertyHandler) addToQueue(payload PublishPayload) error {
    h.mu.Lock()
	defer h.mu.Unlock()
    for _, pp := range payload.Properties {
        prop, _ := h.storer.GetOne(pp.Portal, int64(pp.PropertyID))
        if prop != nil {
            // If the property its already publicated on this portal 
            //   and the status its not "published" then dont add it to the queue
            if prop.Status == store.StatusInQueue || prop.Status == store.StatusInProgress {
                // continue
                return ErrBadRequest(fmt.Sprintf("the property %d has a publication in progress, wait until the end", pp.PropertyID))
            }

            // The property exists but status its "published", then republish
            err := h.storer.Update(pp.Portal, int64(pp.PropertyID), &store.UpdatePublishedProperty{
                Status: store.StatusInQueue,
            })
            if err != nil {
                return err
            }
            pp.Plan = prop.Plan.String
        } else {
            // If the property does not have a publication on the portal, create it
            err := h.storer.Create(&store.PublishedProperty{
                PropertyID: models.NullInt16{Int16: pp.PropertyID, Valid: true},
                Portal: models.NullString{String: pp.Portal,Valid: true},
                Status: store.StatusInProgress,
                Plan: models.NullString{String: pp.Plan, Valid: true},
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

	return nil
}

func (h *PublishedPropertyHandler) Update(w http.ResponseWriter, r *http.Request) error {
	portal := mux.Vars(r)["portal"]
	propertyIDStr := mux.Vars(r)["propId"]
	
	propId, err := strconv.Atoi(propertyIDStr)
	if err != nil {
		return ErrBadRequest("Invalid property ID")
	}

	var update store.UpdatePublishedProperty
    if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		return jsonErr(err)
	}

    if err = h.updateStatus(portal, int64(propId), &update); err != nil {
        return err
    }

    messageResponse(w, "status updated successfully")
    return nil
}

func (h *PublishedPropertyHandler) updateStatus(portal string, propId int64, update *store.UpdatePublishedProperty) error {
	if !validStatus(update.Status) {
		return ErrBadRequest("Invalid status value")
	}

	if err := h.storer.Update(portal, propId, update); err != nil {
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
		
        // this maybe can create an infinity call of goroutines??
		go h.processNextItem()
	}
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
    err := h.storer.Update(h.current.Portal, int64(h.current.PropertyID), &store.UpdatePublishedProperty{
        Status: store.StatusInProgress,
    })
    if err != nil {
        h.logger.Error("error updating the status", 
            "portal", h.current.Portal,
            "id", h.current.PropertyID,
        )
        return
    }

	// Start processing in background
	go func(){
		publication, err := h.storer.GetOne(h.current.Portal, int64(h.current.PropertyID))
		if err != nil {
			h.logger.Error("error on getting publication", "err", err)
			return
		}

		// If publication is already publicated (have publication_id), unpublish first
		if publication.PublicationID.Valid {
			h.logger.Info("publication already exists", "publication id", publication.PublicationID.String)
			if err := h.unpublish(h.current.Portal, publication.PublicationID.String); err != nil {
				h.logger.Error("error unpublishing", "err", err)

				err = h.updateStatus(h.current.Portal, int64(h.current.PropertyID), &store.UpdatePublishedProperty{
					Status: store.StatusFailed,
				})
				return
			}
		}

        if err := h.publish(h.current); err != nil {
            h.logger.Error("error in publish", "err", err)
            // update status processing the next item
            // this maybe can create an infinity call of goroutines?? because updateStatus create another goroutine..
            err = h.updateStatus(h.current.Portal, int64(h.current.PropertyID), &store.UpdatePublishedProperty{
                Status: store.StatusFailed,
            })
        }     
    }()
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
    if err = runPublicatorApp(h.appHost, item, *property); err != nil {
        if errors.Is(err, syscall.ECONNREFUSED) {
            h.mu.Lock()
            defer h.mu.Unlock()
            h.logger.Debug("connection refused error, cleaning queue")
            h.queue = make([]*PropertyPublishPayload, 0)
        }         
        return err
    }

    return nil
}

func (h *PublishedPropertyHandler) unpublish(portal string, publicationId string) error {
    url := fmt.Sprintf("%s/unpublish/%s/%s", h.appHost, portal, publicationId)

    client := http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	res, err := client.Do(req)
    if err != nil {
        return err
    }
    defer res.Body.Close()
    if res.StatusCode != http.StatusCreated {
        var text []byte
        res.Body.Read(text)
        fmt.Println(string(text))
        return ErrBadRequest("error unpublishing the property")
    }

    return nil
}

func (h *PublishedPropertyHandler) handleProcessingError(item *PropertyPublishPayload, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.storer.Update(item.Portal, int64(item.PropertyID), &store.UpdatePublishedProperty{
        Status: store.StatusFailed,
    })
	
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
	fmt.Println(portal)
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

func runPublicatorApp(appHost string, p *PropertyPublishPayload, property store.PortalProp) error {
    url := fmt.Sprintf("%s/publish/%s", appHost, p.Portal)

    // pass the pointer to the Marshaller to correctly unmarshall the NullString fields
    type Payload struct {
        store.PortalProp
        Plan string `json:"plan"`
    }
    payload := Payload{
        Plan: string(p.Plan),
        PortalProp: property,
    }
	jsonBody, err := json.Marshal(&payload)
    if err != nil {
        return err
    }
	bodyReader := bytes.NewReader(jsonBody)

    client := http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest(http.MethodPost, url, bodyReader)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
    if err != nil {
        return err
    }
    defer res.Body.Close()
    if res.StatusCode != http.StatusCreated {
        var text []byte
        res.Body.Read(text)
        fmt.Println(string(text))
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
