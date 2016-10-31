package http

import (
	"log"
	"net/http"
)

type Server struct {
	Port    string
	Handler http.Handler
}

func (s *Server) Run() error {
	log.Println("http.Server: running on port " + s.Port)
	return http.ListenAndServe(s.Port, s.Handler)
}
