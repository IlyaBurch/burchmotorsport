package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// OpenF1 sponsor auth (openf1.org/auth.html): POST username+password to /token,
// bearer token good for an hour. Without credentials the api runs on the free
// tier: archive only, 30 req/min.
var (
	openf1User = os.Getenv("OPENF1_USERNAME")
	openf1Pass = os.Getenv("OPENF1_PASSWORD")
	tokenURL   = "https://api.openf1.org/token"

	tokMu  sync.Mutex
	token  string
	tokExp time.Time
)

func sponsored() bool { return openf1User != "" && openf1Pass != "" }

// bearer returns a valid token, refreshing it a few minutes before expiry.
func bearer() (string, error) {
	tokMu.Lock()
	defer tokMu.Unlock()
	if token != "" && time.Now().Before(tokExp) {
		return token, nil
	}
	form := url.Values{"username": {openf1User}, "password": {openf1Pass}}
	resp, err := client.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("openf1 token: %s", resp.Status)
	}
	// expires_in arrives as the string "3600", not a number as the docs show
	var t struct {
		AccessToken string          `json:"access_token"`
		ExpiresIn   json.RawMessage `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil || t.AccessToken == "" {
		return "", fmt.Errorf("openf1 token: bad response")
	}
	ttl, _ := strconv.Atoi(strings.Trim(string(t.ExpiresIn), `"`))
	if ttl < 600 {
		ttl = 3600
	}
	token = t.AccessToken
	tokExp = time.Now().Add(time.Duration(ttl)*time.Second - 5*time.Minute)
	log.Printf("openf1 token refreshed, valid %ds", ttl)
	return token, nil
}

// authorize adds the bearer header when credentials are configured.
func authorize(req *http.Request) error {
	if !sponsored() {
		return nil
	}
	tok, err := bearer()
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	return nil
}
