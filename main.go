package main

import (
	"log"
	"net/http"

	"go-http/handlers"
	"go-http/store"
)

func main() {
	s, err := store.New()
	if err != nil {
		log.Fatalf("failed to initialize store: %v", err)
	}

	h := handlers.New(s)
	mux := http.NewServeMux()

	// Collection
	mux.HandleFunc("GET /api/winners", h.ListWinners)
	mux.HandleFunc("POST /api/winners", h.CreateWinner)

	// Item (path parameter)
	mux.HandleFunc("GET /api/winners/{id}", h.GetWinner)
	mux.HandleFunc("PUT /api/winners/{id}", h.ReplaceWinner)
	mux.HandleFunc("PATCH /api/winners/{id}", h.PatchWinner)
	mux.HandleFunc("DELETE /api/winners/{id}", h.DeleteWinner)

	// Health check
	mux.HandleFunc("GET /api/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	port := ":24725"
	log.Printf("Balón de Oro API running on port %s", port)
	log.Fatal(http.ListenAndServe(port, mux))
}
