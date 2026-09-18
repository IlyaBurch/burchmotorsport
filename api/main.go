package main

import (
	"log"
	"net/http"
	"os"
)

func handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) {
		tier, auth := "free", "n/a"
		if sponsored() {
			tier, auth = "sponsor", "ok"
			if _, err := bearer(); err != nil {
				auth = err.Error()
			}
		}
		writeJSON(w, map[string]string{"status": "ok", "openf1": tier, "auth": auth}, nil)
	})
	mux.HandleFunc("GET /api/live", liveHandler)
	mux.HandleFunc("GET /api/meetings", meetingsHandler)
	mux.HandleFunc("GET /api/sessions", sessionsHandler)
	mux.HandleFunc("GET /api/track", trackHandler)
	mux.HandleFunc("GET /api/radio", radioHandler)
	mux.HandleFunc("GET /api/next", nextHandler)
	mux.HandleFunc("GET /api/resolve", resolveHandler)
	return mux
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("api listening on :%s, openf1 sponsored=%v", port, sponsored())
	log.Fatal(http.ListenAndServe(":"+port, handler()))
}
