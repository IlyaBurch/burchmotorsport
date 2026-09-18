package main

import (
	"log"
	"net/http"
	"os"
)

func handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("GET /api/live", liveHandler)
	mux.HandleFunc("GET /api/meetings", meetingsHandler)
	mux.HandleFunc("GET /api/sessions", sessionsHandler)
	mux.HandleFunc("GET /api/track", trackHandler)
	mux.HandleFunc("GET /api/radio", radioHandler)
	mux.HandleFunc("GET /api/next", nextHandler)
	return mux
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("api listening on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, handler()))
}
