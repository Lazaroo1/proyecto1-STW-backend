package main

import (
	"fmt"
	"net/http"
	"os"

	"proyecto1-STW-backend/handlers"
	"proyecto1-STW-backend/middleware"
)

func main() {
	db := InitDB()
	defer db.Close()

	sh := &handlers.SeriesHandler{DB: db}
	sw := &handlers.SwaggerHandler{}

	mux := http.NewServeMux()

	// API routes
	mux.Handle("/series", sh)
	mux.Handle("/series/", sh)

	// Swagger
	mux.Handle("/docs", sw)
	mux.Handle("/swagger.yaml", sw)

	// Health check
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","message":"Series Tracker API"}`))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Server running on http://localhost:" + port)
	http.ListenAndServe(":"+port, middleware.CORS(mux))
}
