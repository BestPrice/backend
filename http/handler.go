package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/BestPrice/backend/bp"
	"github.com/julienschmidt/httprouter"
)

type BestpriceHandler struct {
	*httprouter.Router

	Service bp.Service
}

func NewHandler(service bp.Service) *BestpriceHandler {
	h := &BestpriceHandler{
		Router:  httprouter.New(),
		Service: service,
	}
	h.GET("/chainstores", h.chainstores)
	h.GET("/products", h.products)
	h.GET("/categories", h.categories)
	h.GET("/stores", h.stores)
	// h.POST("/shop", h.shop)
	return h
}

func (h *BestpriceHandler) handleIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

func (h *BestpriceHandler) chainstores(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	v, err := h.Service.Chainstores()
	if err != nil {
		log.Println(err)
	}

	err = json.NewEncoder(w).Encode(v)
	if err != nil {
		log.Println(err)
	}
}

func (h *BestpriceHandler) categories(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	v, err := h.Service.Categories()
	if err != nil {
		log.Println(err)
	}

	err = json.NewEncoder(w).Encode(v)
	if err != nil {
		log.Println(err)
	}
}

func (h *BestpriceHandler) products(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	query := r.URL.Query().Get("query")
	if query == "" {
		// TODO: handle empty query
	}

	v, err := h.Service.Products(query)
	if err != nil {
		log.Println(err)
	}

	err = json.NewEncoder(w).Encode(v)
	if err != nil {
		log.Println(err)
	}
}

func (h *BestpriceHandler) stores(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	v, err := h.Service.Stores("", "", "")
	if err != nil {
		log.Println(err)
	}

	err = json.NewEncoder(w).Encode(v)
	if err != nil {
		log.Println(err)
	}
}

func (h *BestpriceHandler) shop(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

}
