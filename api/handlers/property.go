package handlers

import (
	"encoding/json"
	"leadsextractor/middleware"
	"leadsextractor/store"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type PropertyHandler struct {
    storer      store.PropertyPortalStore
    logger      *slog.Logger
}

var InvalidPropID = APIError{
    Status: http.StatusBadRequest,
    Msg:    "the property id must be an integer",
}

func NewPropertyHandler(s store.PropertyPortalStore, logger *slog.Logger) *PropertyHandler {
	return &PropertyHandler{
		storer: s,
        logger: logger.With("module", "propertyHandler"),
	}
}

func (h *PropertyHandler) RegisterRoutes(router *mux.Router) {
    var props []store.PortalProp
    csvHandler := middleware.NewCSVHandler("csv_file", "properties", props)
    // TODO: change the name to /property xD
    router.Handle("/properties/csv", csvHandler.CSVMiddleware(
        http.HandlerFunc(HandleErrors(h.CreateFromCSV)),
    )).Methods(http.MethodPost)

	r := router.PathPrefix("/property").Subrouter()

    r.Methods(http.MethodOptions)

	r.HandleFunc("", HandleErrors(h.GetAll)).Methods(http.MethodGet)
	r.HandleFunc("/{propId}", HandleErrors(h.GetOne)).Methods(http.MethodGet)
	r.HandleFunc("", HandleErrors(h.Insert)).Methods(http.MethodPost)
	r.HandleFunc("/{propId}", HandleErrors(h.Update)).Methods(http.MethodPut)
	r.HandleFunc("/{propId}", HandleErrors(h.Delete)).Methods(http.MethodDelete)

    // add an image to a property
	r.HandleFunc("/{propId}/image", HandleErrors(h.AddImages)).Methods(http.MethodPost)
	r.HandleFunc("/{propId}/image/{imageId}", HandleErrors(h.DeleteImage)).Methods(http.MethodDelete)
	r.HandleFunc("/{propId}/image", HandleErrors(h.GetImages)).Methods(http.MethodGet)
}

func (h *PropertyHandler) CreateFromCSV(w http.ResponseWriter, r *http.Request) error {
    props, ok := r.Context().Value("properties").([]store.PortalProp)
    if !ok {
        return ErrBadRequest("properties does not exists in the context")
    }
    // for debugging
    // j, _ := json.MarshalIndent(&props[0], "\t", "\t")
    // fmt.Println(string(j))

    // TODO: make only one request..
    for _, prop := range props {
        err := h.storer.Insert(&prop)
        if err != nil {
            h.logger.Error("error inserting property", "err", err)
            continue
        }
        h.logger.Info("property created succesfully")
    }

    messageResponse(w, "properties creation process has started")
    return nil
}

func (h *PropertyHandler) GetAll(w http.ResponseWriter, r *http.Request) error {
	props, err := h.storer.GetAll()
	if err != nil {
		return err
	}

	dataResponse(w, props)
	return nil
}

func (h *PropertyHandler) GetOne(w http.ResponseWriter, r *http.Request) error {
    strId := mux.Vars(r)["propId"]
    id, err := strconv.ParseInt(strId, 10, 16)
    if err != nil {
        return InvalidPropID
    }

	prop, err := h.storer.GetOne(id)
	if err != nil {
		return err
	}

    if err := h.storer.GetImages(prop); err != nil {
        return err
    }

	dataResponse(w, prop)
	return nil
}

func (h *PropertyHandler) Insert(w http.ResponseWriter, r *http.Request) error {
	var prop store.PortalProp
	err := json.NewDecoder(r.Body).Decode(&prop)
	if err != nil {
		return jsonErr(err)
	}

	if err = validate.Struct(prop); err != nil {
		return ErrBadRequest(err.Error())
	}

    if err = validateImages(prop.Images); err != nil {
        return err
    }

	if err := h.storer.Insert(&prop); err != nil {
		return err
	}

    // get the created_at, updated_at and id fields
    p, err := h.storer.GetOne(prop.ID)
    if err != nil {
        return err
    }
    p.Images = prop.Images

	createdResponse(w, "property created successfully", p)
	return nil
}

func (h *PropertyHandler) GetImages(w http.ResponseWriter, r *http.Request) error {
    strId := mux.Vars(r)["propId"]
    propId, err := strconv.ParseInt(strId, 10, 16)
    if err != nil {
        return InvalidPropID
    }

    prop, err := h.storer.GetOne(propId)
    if err != nil {
        return err
    }

    if err := h.storer.GetImages(prop); err != nil {
        return err
    }

	dataResponse(w, prop.Images)
    return nil
}

func (h *PropertyHandler) AddImages(w http.ResponseWriter, r *http.Request) error {
    strId := mux.Vars(r)["propId"]
    propId, err := strconv.ParseInt(strId, 10, 16)
    if err != nil {
        return InvalidPropID
    }

	var images []store.PropertyImage
	if err := json.NewDecoder(r.Body).Decode(&images); err != nil {
		return jsonErr(err)
	}

    if err = validateImages(images); err != nil {
        return err
    }

    if err := h.storer.InsertImages(propId, images); err != nil {
        return err
    }

    createdResponse(w, "images added successfully", images)
    return nil
}

func (h *PropertyHandler) Delete(w http.ResponseWriter, r *http.Request) error {
    strId := mux.Vars(r)["propId"]
    id, err := strconv.ParseInt(strId, 10, 16)
    if err != nil {
        return InvalidPropID
    }

    if err := h.storer.Delete(int64(id)); err != nil {
        return err
    }

	messageResponse(w, "property deleted successfully")
	return nil
}

func (h *PropertyHandler) DeleteImage(w http.ResponseWriter, r *http.Request) error {
    strId := mux.Vars(r)["propId"]
    propId, err := strconv.ParseInt(strId, 10, 16)
    if err != nil {
        return InvalidPropID
    }

    strId = mux.Vars(r)["imageId"]
    imageId, err := strconv.ParseInt(strId, 10, 16)
    if err != nil {
        return ErrBadRequest("the image id must be a integer")
    }

    if err = h.storer.DeleteImage(propId, imageId); err != nil {
        return err
    }

    messageResponse(w, "image deleted successfully")
    return nil
}

func (h *PropertyHandler) Update(w http.ResponseWriter, r *http.Request) error {
    strId := mux.Vars(r)["propId"]
    id, err := strconv.ParseInt(strId, 10, 16)
    if err != nil {
        return InvalidPropID
    }

	var prop store.PortalProp
	if err := json.NewDecoder(r.Body).Decode(&prop); err != nil {
		return jsonErr(err)
	}
    prop.ID = id

    if _, err := h.storer.GetOne(id); err != nil {
        return err
    }

	if err := validate.Struct(prop); err != nil {
		return ErrBadRequest(err.Error())
	}

	if err := h.storer.Update(&prop); err != nil {
		return err
	}

	createdResponse(w, "property updated successfully", &prop)
	return nil
}

func validateImages(images []store.PropertyImage) error {
    for _, image := range images {
        if err := validate.Var(image.Url, "required,url"); err != nil {
            return ErrBadRequest("'image must have a valid url")
        }
    } 
    return nil
}
