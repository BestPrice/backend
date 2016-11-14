package http

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/url"

	"github.com/BestPrice/backend/bp"
	"github.com/julienschmidt/httprouter"
)

// type Handler func(rw http.ResponseWriter, req *http.Request) error

// type ErrorHandler struct {
// 	H Handler
// }

// func (h ErrorHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
// 	err := h.H(rw, req)
// 	switch err.(type) {
// 	default:
// 		log.Println(err)
// 	}
// }

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
	h.r.GET("/help", h.api)
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

	e := json.NewEncoder(w)
	e.SetIndent("", "\t")
	err = e.Encode(v)
	if err != nil {
		log.Println(err)
	}
}

func (h *BestpriceHandler) chainstores(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	v, err := h.Service.Chainstores()
	if err != nil {
		log.Println(err)
	}

	e := json.NewEncoder(w)
	e.SetIndent("", "\t")
	err = e.Encode(v)
	if err != nil {
		log.Println(err)
	}
}

func (h *BestpriceHandler) products(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.Println(r.URL)

	phrase, err := url.QueryUnescape(r.URL.Query().Get("search"))
	if err != nil {
		log.Println(err)
	}

	category, _ := bp.NewID(r.URL.Query().Get("category"))

	v, err := h.Service.Products(category, phrase)
	if err != nil {
		log.Println(err)
	}

	e := json.NewEncoder(w)
	e.SetIndent("", "\t")
	err = e.Encode(v)
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

func (h *BestpriceHandler) api(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

	buf.WriteString("\n\nTODO: POST /shop\n")
	// enc.Encode([]bp.Store{bp.Store{}, bp.Store{}})

	buf.WriteTo(w)
}
