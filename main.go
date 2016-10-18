package main

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func GetChainstores(w http.ResponseWriter, r *http.Request) {
}

func PostStore(w http.ResponseWriter, r *http.Request) {
}

func GetCategories(w http.ResponseWriter, r *http.Request) {
}

func PostShop(w http.ResponseWriter, r *http.Request) {
}

type Server struct {
	Port string
}

func (s *Server) Start() error {
	r := mux.NewRouter()

	r.HandleFunc("/chainstores", GetChainstores).Methods(http.MethodGet)
	r.HandleFunc("/store", PostStore).Methods(http.MethodPost)
	r.HandleFunc("/categories", GetCategories).Methods(http.MethodGet)
	r.HandleFunc("/shop", PostShop).Methods(http.MethodPost)

	return http.ListenAndServe(s.Port, r)
}

func main() {
	var (
		port = os.Getenv("PORT")
	)

	if port == "" {
		port = "8080"
	}

	server := Server{
		Port: ":" + port,
	}
	server.Start()
}
