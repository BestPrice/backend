package http

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"

	"github.com/BestPrice/backend/bp"
	"github.com/julienschmidt/httprouter"
)

type BestpriceHandler struct {
	r *httprouter.Router

	Service bp.Service
}

func NewHandler(service bp.Service) *BestpriceHandler {
	h := &BestpriceHandler{
		r:       httprouter.New(),
		Service: service,
	}
	h.r.GET("/categories", h.categories)
	h.r.GET("/chainstores", h.chainstores)
	h.r.GET("/products", h.products)
	h.r.GET("/stores", h.stores)
	h.r.POST("/shop", h.shop)
	return h
}

func (h *BestpriceHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
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

	h.r.ServeHTTP(rw, req)
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

func (h *BestpriceHandler) products(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	query, err := url.QueryUnescape(r.URL.Query().Get("search"))
	log.Println(r.URL, query, err)

	if err != nil {
		log.Println(err)
		return
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
