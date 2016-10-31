package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/BestPrice/backend/sql"
	"github.com/julienschmidt/httprouter"
)

type Handler struct {
	*httprouter.Router

	// Service *bp.Service
	client *sql.Client
}

func NewHandler(c *sql.Client) *Handler {
	h := &Handler{
		Router: httprouter.New(),
		client: c,
	}
	h.GET("/", h.handleIndex)
	h.GET("/chainstores", h.handleGetChainstores)
	h.GET("/products", h.handleGetProducts)
	h.GET("/categories", h.handleGetCategories)
	h.GET("/stores", h.handleGetStores)
	return h
}

func (h *Handler) handleIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Write([]byte(
		`- GET /categories
returns all categories

- GET /chainstores
return all chainstores

- GET /products?query=string
return all products matching string, maximum 100 products

- GET /stores?chainstore=string&district=string&region=string
return all stores by given query`))
}

func (h *Handler) handleGetChainstores(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	s := h.client.Connect().Service()
	v, err := s.Chainstores()
	if err != nil {
		log.Println(err)
	}

	err = json.NewEncoder(w).Encode(v)
	if err != nil {
		log.Println(err)
	}
}

func (h *Handler) handleGetCategories(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	s := h.client.Connect().Service()
	v, err := s.Categories()
	if err != nil {
		log.Println(err)
	}

	err = json.NewEncoder(w).Encode(v)
	if err != nil {
		log.Println(err)
	}
}

func (h *Handler) handleGetProducts(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	query := r.URL.Query().Get("query")
	if query == "" {
		// TODO: handle empty query
	}

	s := h.client.Connect().Service()
	v, err := s.Products(query)
	if err != nil {
		log.Println(err)
	}

	err = json.NewEncoder(w).Encode(v)
	if err != nil {
		log.Println(err)
	}
}

func (h *Handler) handleGetStores(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	s := h.client.Connect().Service()
	v, err := s.Stores("", "", "")
	if err != nil {
		log.Println(err)
	}

	err = json.NewEncoder(w).Encode(v)
	if err != nil {
		log.Println(err)
	}
}
