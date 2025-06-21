package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/faiakak/block-phone-number/config"
	"github.com/faiakak/block-phone-number/handlers"
	"github.com/faiakak/block-phone-number/routes"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var db *sql.DB

func main() {

	// Load .env file only for local development
	if os.Getenv("ENV") != "production" {
		err := godotenv.Load()
		if err != nil {
			log.Println("No .env file found")
		}
	}

	// Initialize database connection
	var err error

	connStr := config.GetDBConnectionString()

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Connected to PostgreSQL database")

	handlers.SetDB(db) // inject DB into handlers

	config.RunMigrations(db)

	// Get APP_PORT from env or fallback to 8080
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	r := routes.InitRoutes()

	log.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
