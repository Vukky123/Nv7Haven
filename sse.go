package main

import (
	"net/http"

	"github.com/r3labs/sse/v2"
)

type sseServer struct {
	sse *sse.Server
}

func (s *sseServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.sse.HTTPHandler(w, r)
}

func NewSseServer(server *sse.Server) *sseServer {
	return &sseServer{
		sse: server,
	}
}
