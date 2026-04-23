package main

import (
	"fmt"
	"net/http"

	"proyecto1-STW-backend/handlers"
	"proyecto1-STW-backend/middleware"
)

func main() {
	db := InitDB()
	defer db.Close()

	h := &handlers.SeriesHandler{DB: db}

	mux := http.NewServeMux()
	mux.Handle("/series", h)
	mux.Handle("/series/", h)

	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", middleware.CORS(mux))
}
