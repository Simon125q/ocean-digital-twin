package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"ocean-digital-twin/internal/database"
)

type Server struct {
	port int

	db database.Service
}

func NewServer() (*http.Server, database.Service) {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	dbService := database.New()

	apiServer := &Server{
		port: port,
		db:   dbService,
	}

	// migrate database up
	apiServer.db.Up()

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", apiServer.port),
		Handler:      apiServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server, dbService
}

func (s *Server) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	jsonResp, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(jsonResp)
	if err != nil {
		log.Printf("Error writing JSON response: %v", err)
	}
}

func (s *Server) respondWithError(w http.ResponseWriter, code int, message string) {
	s.respondWithJSON(w, code, map[string]string{"error": message})
}
