package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

type Server struct {
}

func New() *Server {
	return &Server{}
}

func (s *Server) Start() error {
	http.HandleFunc("/health", indexHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("defaulting to port %s", port)
	}

	log.Printf("listening on port: %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
	return nil
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/health" {
		http.NotFound(w, r)
		return
	}
	fmt.Fprint(w, "Healthy")
}
