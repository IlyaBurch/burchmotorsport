package main

import (
	"net/http"
	"time"
)

type NextRace struct {
	Key      int    `json:"session_key"`
	Name     string `json:"session_name"`
	Location string `json:"location"`
	Country  string `json:"country_name"`
	Start    string `json:"date_start"`
}

// GET /api/next -> the next Race session on the calendar, cached for an hour
func nextHandler(w http.ResponseWriter, _ *http.Request) {
	v, err := store.get("next", func() (any, time.Duration, error) {
		var s []NextRace
		now := time.Now().UTC().Format("2006-01-02T15:04:05")
		if err := get("sessions?session_name=Race&date_start>="+now, &s); err != nil {
			return nil, 0, err
		}
		if len(s) == 0 {
			return nil, time.Hour, nil // off-season: null
		}
		return s[0], time.Hour, nil
	})
	writeJSON(w, v, err)
}
