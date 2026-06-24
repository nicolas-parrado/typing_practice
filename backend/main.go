package main

import (
	"log"
	"net/http"
	"os"

	"github.com/nparrado/typing-exercises/backend/db"
	"github.com/nparrado/typing-exercises/backend/handlers"
)

func main() {
	// 1. Initialize SQLite Database
	database := db.InitDB()
	defer database.Close()

	// 2. Seed Exercises (runs check to populate up to 1000+ texts)
	db.SeedExercises(database)

	// 3. Set up enrouter using standard library multiplexer
	mux := http.NewServeMux()

	// Profiles
	mux.HandleFunc("GET /api/profiles", handlers.GetProfiles)
	mux.HandleFunc("POST /api/profiles", handlers.CreateProfile)
	mux.HandleFunc("GET /api/profiles/{id}", handlers.GetProfile)
	mux.HandleFunc("DELETE /api/profiles/{id}", handlers.DeleteProfile)

	// Exercises & Sessions
	mux.HandleFunc("GET /api/exercises", handlers.GetExercises)
	mux.HandleFunc("POST /api/sessions", handlers.SaveSession)

	// Stats, Retries & Medals
	mux.HandleFunc("GET /api/profiles/{id}/stats", handlers.GetStats)
	mux.HandleFunc("GET /api/profiles/{id}/retries", handlers.GetFailedAttempts)
	mux.HandleFunc("GET /api/profiles/{id}/achievements", handlers.GetAchievements)

	// 4. CORS and Request Logging Middlewares
	handler := loggingMiddleware(corsMiddleware(mux))

	// 5. Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
