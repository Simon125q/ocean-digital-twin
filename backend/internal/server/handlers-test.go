package server

import (
	"fmt"
	"log/slog"
	"net/http"
)

func (s *Server) GetCountHandler(w http.ResponseWriter, r *http.Request) {
	count := s.db.GetCount()
	s.respondWithJSON(w, http.StatusOK, count)
}

func (s *Server) UpdateCountHandler(w http.ResponseWriter, r *http.Request) {
	count := s.db.GetCount()
	err := s.db.UpdateCount(count + 1)
	if err != nil {
		s.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("err: %v", err))
	}
	s.respondWithJSON(w, http.StatusOK, map[string]int{"count": count + 1})
}

func (s *Server) NewCountHandler(w http.ResponseWriter, r *http.Request) {
	_, err := s.db.NewCount()
	if err != nil {
		slog.Error("Error creating new count", "err", err)
		s.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Coudnt create new count: %v", err))
	}
	s.respondWithJSON(w, http.StatusOK, map[string]string{"message": ""})
}
