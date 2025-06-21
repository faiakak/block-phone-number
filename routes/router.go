package routes

import (
	"net/http"

	"github.com/gorilla/mux"

	"blocked-phone-numbers/handlers"
)

// InitRoutes sets up all API and static routes with middleware
func InitRoutes() *mux.Router {
	// Create a new main router
	r := mux.NewRouter()

	// Subrouter for API endpoints
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/blocked-phones", handlers.GetBlockedPhones).Methods("GET")
	api.HandleFunc("/blocked-phones", handlers.AddBlockedPhone).Methods("POST")
	api.HandleFunc("/blocked-phones/{id}", handlers.RemoveBlockedPhone).Methods("DELETE")
	api.HandleFunc("/check-phone", handlers.CheckPhone).Methods("POST")

	// Serve static files (frontend HTML/JS/CSS)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	// Apply CORS middleware to all routes
	r.Use(handlers.CORSMiddleware)

	return r
}
