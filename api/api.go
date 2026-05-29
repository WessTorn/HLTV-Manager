package api

import (
	_ "HLTV-Manager/docs"
	"HLTV-Manager/service"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
)

type Server struct {
	service service.Service
}

func NewServer(svc service.Service) *Server {
	return &Server{service: svc}
}

func (s *Server) InitAPI() {
	http.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	http.HandleFunc("/", withCORSFunc(s.rootHandler))
	http.HandleFunc("/api/v1/health", withCORSFunc(s.healthHandler))

	http.HandleFunc("/api/v1/hltv", withCORSFunc(s.hltvListHandler))

	http.HandleFunc("/api/v1/hltv/", withCORSFunc(s.hltvHandlers))
}
