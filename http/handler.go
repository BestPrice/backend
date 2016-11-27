package http

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/BestPrice/backend/bp"
	"github.com/gorilla/mux"
)

type handlerFunc func(rw http.ResponseWriter, req *http.Request) error

type statusError struct {
	error
	status int
}

type errorHandler handlerFunc

func (h errorHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if err := h(rw, req); err != nil {
		log.Println(err)
		switch err.(type) {
		case statusError:
		default:
			code := http.StatusInternalServerError
			http.Error(rw, http.StatusText(code), code)
		}
	}
}

type accessControlHandler struct {
	http.Handler
}

func (h accessControlHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if origin := req.Header.Get("Origin"); origin != "" {
		rw.Header().Set("Access-Control-Allow-Origin", origin)
		rw.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		rw.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	}
	// Stop here if its Preflighted OPTIONS request
	if req.Method == "OPTIONS" {
		return
	}
	h.Handler.ServeHTTP(rw, req)
}

type Handler struct {
	*mux.Router
	Service bp.Service
}

func NewHandler(service bp.Service) http.Handler {
	h := &Handler{
		Router:  mux.NewRouter(),
		Service: service,
	}
	h.Handle("/categories", errorHandler(h.categories)).Methods(http.MethodGet)
	h.Handle("/chainstores", errorHandler(h.chainstores)).Methods(http.MethodGet)
	h.Handle("/products", errorHandler(h.products)).Methods(http.MethodGet)
	h.Handle("/stores", errorHandler(h.stores)).Methods(http.MethodGet)
	h.Handle("/shop", errorHandler(h.shop)).Methods(http.MethodPost)
	h.Handle("/api", errorHandler(h.api)).Methods(http.MethodGet)

	return &accessControlHandler{h}
}

func encodeJSON(w io.Writer, v interface{}) error {
	e := json.NewEncoder(w)
	e.SetIndent("", "\t")
	return e.Encode(v)
}

func (h Handler) categories(w http.ResponseWriter, r *http.Request) error {
	v, err := h.Service.Categories()
	if err != nil {
		return err
	}
	return encodeJSON(w, v)
}

func (h Handler) chainstores(w http.ResponseWriter, r *http.Request) error {
	v, err := h.Service.Chainstores()
	if err != nil {
		log.Println(err)
	}
	return encodeJSON(w, v)
}

func (h Handler) products(w http.ResponseWriter, r *http.Request) error {

	phrase, err := url.QueryUnescape(r.URL.Query().Get("search"))
	if err != nil {
		return err
	}

	category, _ := bp.NewID(r.URL.Query().Get("category"))

	v, err := h.Service.Products(category, phrase)
	if err != nil {
		return err
	}

	return encodeJSON(w, v)
}

func (h Handler) stores(w http.ResponseWriter, r *http.Request) error {
	v, err := h.Service.Stores("", "", "")
	if err != nil {
		return err
	}
	return encodeJSON(w, v)
}

func (h Handler) shop(w http.ResponseWriter, r *http.Request) error {
	var req bp.ShopRequest
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}

	if err := req.Valid(); err != nil {
		return encodeJSON(w, bp.Shop{Error: err.Error()})
	}

	shop, err := h.Service.Shop(&req)
	if err != nil {
		return encodeJSON(w, bp.Shop{Error: err.Error()})
	}

	return encodeJSON(w, shop)
}

func (h Handler) api(w http.ResponseWriter, r *http.Request) error {
	var (
		buf bytes.Buffer
		enc = json.NewEncoder(&buf)
	)
	enc.SetIndent("", "\t")

	buf.WriteString("GET /categories\n")
	enc.Encode([]bp.Category{
		bp.Category{
			Subcategories: []bp.Category{bp.Category{}},
		},
		bp.Category{},
	})

	buf.WriteString("\n\nGET /chainstores\n")
	enc.Encode([]bp.Chainstore{bp.Chainstore{}, bp.Chainstore{}})

	buf.WriteString("\n\nGET /products?category=uuid;search=string\n")
	enc.Encode([]bp.Product{bp.Product{}, bp.Product{}})

	buf.WriteString("\n\nGET /stores\n")
	enc.Encode([]bp.Store{bp.Store{}, bp.Store{}})

	buf.WriteString("\n\nPOST /shop\n")
	enc.Encode(bp.ShopRequest{
		Products: []bp.ShopRequestProduct{
			{ID: bp.RandID(), Count: 1},
			{ID: bp.RandID(), Count: 1},
		},

		UserPreference: bp.UserPreference{
			IDs: []bp.ID{bp.RandID(), bp.RandID()},
			Max: 3,
		},
	})

	_, err := buf.WriteTo(w)
	return err
}
