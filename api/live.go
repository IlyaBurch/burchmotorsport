package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"
)

type Session struct {
	Key     int    `json:"session_key"`
	Meeting int    `json:"meeting_key"`
	Name    string `json:"session_name"`
	Type    string `json:"session_type"`
	Circuit string `json:"circuit_short_name"`
	Country string `json:"country_name"`
	Start   string `json:"date_start"`
	End     string `json:"date_end"`
}

type Driver struct {
	Position int      `json:"position"`
	Number   int      `json:"number"`
	Acronym  string   `json:"acronym"`
	Name     string   `json:"name"`
	Team     string   `json:"team"`
	Gap      any      `json:"gap"`      // seconds (float) or text like "+1 LAP"
	Interval any      `json:"interval"` // same
	LastLap  *float64 `json:"lastLap"`
	BestLap  *float64 `json:"bestLap"`
	Lap      int      `json:"lap"`
}

type Live struct {
	Session   Session  `json:"session"`
	Drivers   []Driver `json:"drivers"`
	UpdatedAt string   `json:"updatedAt"`
}

const forever = 100 * 365 * 24 * time.Hour

func writeJSON(w http.ResponseWriter, v any, err error) {
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

// GET /api/live?session_key=latest|<int>
func liveHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("session_key")
	if key == "" {
		key = "latest"
	}
	v, err := store.get("live:"+key, func() (any, time.Duration, error) {
		l, err := fetchLive(key)
		if err != nil {
			return nil, 0, err
		}
		if end, e := time.Parse(time.RFC3339, l.Session.End); e == nil && end.Add(time.Hour).Before(time.Now()) {
			return l, forever, nil
		}
		return l, 5 * time.Second, nil
	})
	writeJSON(w, v, err)
}

// GET /api/meetings?year=2025 -> openf1 meetings, cached for a day
func meetingsHandler(w http.ResponseWriter, r *http.Request) {
	year := r.URL.Query().Get("year")
	v, err := store.get("meetings:"+year, func() (any, time.Duration, error) {
		var m []json.RawMessage
		return m, 24 * time.Hour, get("meetings?year="+year, &m)
	})
	writeJSON(w, v, err)
}

// GET /api/sessions?meeting_key=1294 -> openf1 sessions of a meeting
func sessionsHandler(w http.ResponseWriter, r *http.Request) {
	mk := r.URL.Query().Get("meeting_key")
	v, err := store.get("sessions:"+mk, func() (any, time.Duration, error) {
		var s []json.RawMessage
		return s, time.Hour, get("sessions?meeting_key="+mk, &s)
	})
	writeJSON(w, v, err)
}

var openf1 = "https://api.openf1.org/v1/"

func get(path string, v any) error {
	resp, err := http.Get(openf1 + path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("openf1 %s: %s", path, resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

func fetchLive(sessionKey string) (Live, error) {
	var sessions []Session
	if err := get("sessions?session_key="+sessionKey, &sessions); err != nil {
		return Live{}, err
	}
	if len(sessions) == 0 {
		return Live{}, fmt.Errorf("openf1: no session")
	}
	key := fmt.Sprint(sessions[0].Key)

	var drivers []struct {
		Number  int    `json:"driver_number"`
		Acronym string `json:"name_acronym"`
		Name    string `json:"full_name"`
		Team    string `json:"team_name"`
	}
	var positions []struct {
		Number   int `json:"driver_number"`
		Position int `json:"position"`
	}
	var intervals []struct {
		Number   int `json:"driver_number"`
		Gap      any `json:"gap_to_leader"`
		Interval any `json:"interval"`
	}
	var laps []struct {
		Number   int      `json:"driver_number"`
		Lap      int      `json:"lap_number"`
		Duration *float64 `json:"lap_duration"`
	}
	// sequential with a pause on purpose: openf1 allows 3 req/s
	for _, q := range []struct {
		path string
		into any
	}{
		{"drivers?session_key=" + key, &drivers},
		{"position?session_key=" + key, &positions},
		{"intervals?session_key=" + key, &intervals},
		{"laps?session_key=" + key, &laps},
	} {
		time.Sleep(400 * time.Millisecond)
		if err := get(q.path, q.into); err != nil {
			return Live{}, err
		}
	}

	byNum := map[int]*Driver{}
	out := make([]Driver, 0, len(drivers))
	for _, d := range drivers {
		out = append(out, Driver{Number: d.Number, Acronym: d.Acronym, Name: d.Name, Team: d.Team})
	}
	for i := range out {
		byNum[out[i].Number] = &out[i]
	}
	// arrays are chronological: last write wins
	for _, p := range positions {
		if d := byNum[p.Number]; d != nil {
			d.Position = p.Position
		}
	}
	for _, iv := range intervals {
		if d := byNum[iv.Number]; d != nil {
			d.Gap, d.Interval = iv.Gap, iv.Interval
		}
	}
	for _, l := range laps {
		d := byNum[l.Number]
		if d == nil {
			continue
		}
		d.Lap = l.Lap
		if l.Duration != nil {
			d.LastLap = l.Duration
			if d.BestLap == nil || *l.Duration < *d.BestLap {
				d.BestLap = l.Duration
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i].Position, out[j].Position
		if a == 0 {
			a = 99
		}
		if b == 0 {
			b = 99
		}
		return a < b
	})
	return Live{Session: sessions[0], Drivers: out, UpdatedAt: time.Now().UTC().Format(time.RFC3339)}, nil
}
