package server

import (
	"encoding/json"
	"net/http"

	"github.com/hranicka/qwen-foo/internal/db"
)

type response struct {
	Message string `json:"message"`
	Counter int64  `json:"counter"`
}

type Server struct {
	handlers map[string]http.HandlerFunc
	db       *db.Pool
}

func New(pool *db.Pool) *Server {
	s := &Server{
		handlers: make(map[string]http.HandlerFunc),
		db:       pool,
	}

	s.handlers["GET /hello"] = s.helloHandler
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	key := r.Method + " " + r.URL.Path

	h, ok := s.handlers[key]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	h(w, r)
}

func (s *Server) helloHandler(w http.ResponseWriter, r *http.Request) {
	counterVal, err := s.db.Incr(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := response{
		Message: "Hello",
		Counter: counterVal,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
