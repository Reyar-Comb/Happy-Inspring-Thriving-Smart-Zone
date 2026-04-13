package server

import (
	"fmt"
	"net/http"

	"github.com/Reyar-Comb/HITPlane/config"
)

func (s *Server) StartHTTP() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/login", s.HandleLogin)
	mux.HandleFunc("/api/register", s.HandleRegister)

	addr := config.GlobalConfig.HTTPPort

	fmt.Println("Server: HTTP server listening on", addr)
	return http.ListenAndServe(addr, mux)
}
