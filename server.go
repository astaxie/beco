package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
)

const (
	// CONNECT HTTP method
	CONNECT = "CONNECT"
	// DELETE HTTP method
	DELETE = "DELETE"
	// GET HTTP method
	GET = "GET"
	// HEAD HTTP method
	HEAD = "HEAD"
	// OPTIONS HTTP method
	OPTIONS = "OPTIONS"
	// PATCH HTTP method
	PATCH = "PATCH"
	// POST HTTP method
	POST = "POST"
	// PUT HTTP method
	PUT = "PUT"
	// TRACE HTTP method
	TRACE = "TRACE"
)

var (
	methods = [...]string{
		CONNECT,
		DELETE,
		GET,
		HEAD,
		OPTIONS,
		PATCH,
		POST,
		PUT,
		TRACE,
	}
)

type Server struct {
	*http.Server
	pool   sync.Pool
	router *Router
	debug  bool
}

func (s *Server) Handle(pattern string, handler http.Handler) {
	for _, method := range methods {
		s.router.Add(method, pattern, handler)
	}
}

func (s *Server) Run(addr string) error {
	s.Server.Addr = addr
	err := s.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServeTLS: ", err, fmt.Sprintf("%d", os.Getpid()))
	}
	return err
}

func (s *Server) RunTLS(addr, certFile, keyFile string) error {
	s.Server.Addr = addr
	err := s.ListenAndServeTLS(certFile, keyFile)
	if err != nil {
		log.Fatal("ListenAndServeTLS: ", err, fmt.Sprintf("%d", os.Getpid()))
	}
	return err
}

// ServeHTTP implements `http.Handler` interface, which serves HTTP requests.
func (s *Server) ServerHTTP(w http.ResponseWriter, r *http.Request) {
	h, err := s.router.Find(r.Method, r.URL.Path)
	if err != nil {

	}
	h.ServerHTTP(w, r)
}

func NewServer() (*Server, error) {
	s := Server{
		Server: &http.Server{},
		pool:   sync.Pool{},
		router: &Route{},
	}
	return s, nil
}
