package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// RADIO_RELAY=http://relay-host:8080 makes this instance fetch through another
// instance abroad. F1's CDN geo-blocks Russia, the site is hosted there.
var radioRelay = os.Getenv("RADIO_RELAY")

// GET /api/radio?url=https://livetiming.formula1.com/static/.../x.mp3
// ponytail: plain pass-through so the browser never talks to F1's CDN directly.
// Adds nothing but Cache-Control; if the CDN geo-blocks the server too, a relay
// outside that region is the next step, not more code here.
func radioHandler(w http.ResponseWriter, r *http.Request) {
	u := r.URL.Query().Get("url")
	if !strings.HasPrefix(u, "https://livetiming.formula1.com/static/") || !strings.HasSuffix(u, ".mp3") {
		http.Error(w, "bad url", http.StatusBadRequest)
		return
	}
	target := u
	if radioRelay != "" {
		target = radioRelay + "/api/radio?url=" + url.QueryEscape(u)
	}
	req, _ := http.NewRequestWithContext(r.Context(), "GET", target, nil)
	req.Header.Set("Range", r.Header.Get("Range")) // let <audio> seek
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		log.Printf("radio %s: %s", u, resp.Status)
		http.Error(w, "upstream "+resp.Status, http.StatusBadGateway)
		return
	}
	for _, h := range []string{"Content-Type", "Content-Length", "Content-Range", "Accept-Ranges"} {
		if v := resp.Header.Get(h); v != "" {
			w.Header().Set(h, v)
		}
	}
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
