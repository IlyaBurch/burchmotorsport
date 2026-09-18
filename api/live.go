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
	Position int         `json:"position"`
	Number   int         `json:"number"`
	Acronym  string      `json:"acronym"`
	Name     string      `json:"name"`
	Team     string      `json:"team"`
	Colour   string      `json:"teamColour"` // hex without #, from openf1
	Gap      any         `json:"gap"`        // seconds (float) or text like "+1 LAP"
	Interval any         `json:"interval"`   // same
	LastLap  *float64    `json:"lastLap"`
	BestLap  *float64    `json:"bestLap"`
	Lap      int         `json:"lap"`
	Sectors  [3]*float64 `json:"sectors"`     // last lap
	BestSect [3]*float64 `json:"bestSectors"` // personal best per sector
	Speed    int         `json:"speedTrap"`   // km/h on last lap
	Compound string      `json:"compound"`    // SOFT/MEDIUM/HARD/INTERMEDIATE/WET
	TyreAge  int         `json:"tyreAge"`     // laps on current set
	Pits     int         `json:"pits"`
	X        int         `json:"x"` // track coords, 0,0 if unknown
	Y        int         `json:"y"`
}

type Weather struct {
	Air      float64 `json:"air_temperature"`
	Track    float64 `json:"track_temperature"`
	Humidity float64 `json:"humidity"`
	Rain     float64 `json:"rainfall"`
	Wind     float64 `json:"wind_speed"`
}

type RaceControl struct {
	Date     string `json:"date"`
	Category string `json:"category"`
	Flag     string `json:"flag"`
	Message  string `json:"message"`
	Lap      int    `json:"lap_number"`
}

type Live struct {
	Session     Session       `json:"session"`
	Drivers     []Driver      `json:"drivers"`
	Weather     *Weather      `json:"weather"`
	RaceControl []RaceControl `json:"raceControl"` // newest first, last 20
	UpdatedAt   string        `json:"updatedAt"`
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

// GET /api/track?session_key=latest|<int> -> [[x,y],...] outline of one lap
func trackHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("session_key")
	if key == "" {
		key = "latest"
	}
	v, err := store.get("track:"+key, func() (any, time.Duration, error) {
		pts, err := fetchTrack(key)
		return pts, forever, err
	})
	writeJSON(w, v, err)
}

// ponytail: outline = one full lap of whoever completed lap 3. Good enough for
// a map; swap for a static per-circuit SVG if openf1 ever drops location.
func fetchTrack(sessionKey string) ([][2]int, error) {
	var laps []struct {
		Number   int      `json:"driver_number"`
		Start    string   `json:"date_start"`
		Duration *float64 `json:"lap_duration"`
	}
	if err := get("laps?lap_number=3&session_key="+sessionKey, &laps); err != nil {
		return nil, err
	}
	for _, l := range laps {
		if l.Duration == nil || l.Start == "" {
			continue
		}
		start, err := time.Parse(time.RFC3339Nano, l.Start)
		if err != nil {
			continue
		}
		end := start.Add(time.Duration(*l.Duration * float64(time.Second)))
		time.Sleep(400 * time.Millisecond)
		var loc []struct{ X, Y int }
		q := fmt.Sprintf("location?session_key=%s&driver_number=%d&date>=%s&date<%s",
			sessionKey, l.Number, start.Format(time.RFC3339), end.Format(time.RFC3339))
		if err := get(q, &loc); err != nil {
			return nil, err
		}
		pts := make([][2]int, 0, len(loc))
		for _, p := range loc {
			if p.X != 0 || p.Y != 0 {
				pts = append(pts, [2]int{p.X, p.Y})
			}
		}
		if len(pts) > 50 {
			return pts, nil
		}
	}
	return nil, fmt.Errorf("openf1: no location data for lap 3")
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
		Colour  string `json:"team_colour"`
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
		S1       *float64 `json:"duration_sector_1"`
		S2       *float64 `json:"duration_sector_2"`
		S3       *float64 `json:"duration_sector_3"`
		Speed    int      `json:"st_speed"`
	}
	var stints []struct {
		Number   int    `json:"driver_number"`
		LapStart int    `json:"lap_start"`
		Compound string `json:"compound"`
		Age      int    `json:"tyre_age_at_start"`
	}
	var weather []Weather
	var rc []RaceControl

	// ponytail: sequential with a pause on purpose, openf1 allows 3 req/s.
	// ~3s per refresh; parallel batches if it ever matters.
	for _, q := range []struct {
		path string
		into any
	}{
		{"drivers?session_key=" + key, &drivers},
		{"position?session_key=" + key, &positions},
		{"intervals?session_key=" + key, &intervals},
		{"laps?session_key=" + key, &laps},
		{"stints?session_key=" + key, &stints},
		{"weather?session_key=" + key, &weather},
		{"race_control?session_key=" + key, &rc},
	} {
		time.Sleep(400 * time.Millisecond)
		if err := get(q.path, q.into); err != nil {
			return Live{}, err
		}
	}

	// car positions: last 10s for a live session, around the chequered flag otherwise
	at := time.Now().UTC().Add(-10 * time.Second)
	if end, e := time.Parse(time.RFC3339, sessions[0].End); e == nil && end.Add(time.Hour).Before(time.Now()) {
		at = end.Add(-10 * time.Second)
		for _, m := range rc {
			if m.Flag == "CHEQUERED" {
				if t, e := time.Parse(time.RFC3339Nano, m.Date); e == nil {
					at = t.Add(-5 * time.Second)
				}
				break
			}
		}
	}
	var loc []struct {
		Number int `json:"driver_number"`
		X, Y   int
	}
	time.Sleep(400 * time.Millisecond)
	if err := get(fmt.Sprintf("location?session_key=%s&date>=%s&date<%s", key,
		at.Format(time.RFC3339), at.Add(10*time.Second).Format(time.RFC3339)), &loc); err != nil {
		return Live{}, err
	}

	byNum := map[int]*Driver{}
	out := make([]Driver, 0, len(drivers))
	for _, d := range drivers {
		out = append(out, Driver{Number: d.Number, Acronym: d.Acronym, Name: d.Name, Team: d.Team, Colour: d.Colour})
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
		d.Sectors = [3]*float64{l.S1, l.S2, l.S3}
		if l.Speed > 0 {
			d.Speed = l.Speed
		}
		for i, s := range d.Sectors {
			if s != nil && (d.BestSect[i] == nil || *s < *d.BestSect[i]) {
				d.BestSect[i] = s
			}
		}
		if l.Duration != nil {
			d.LastLap = l.Duration
			if d.BestLap == nil || *l.Duration < *d.BestLap {
				d.BestLap = l.Duration
			}
		}
	}
	for _, s := range stints {
		if d := byNum[s.Number]; d != nil {
			d.Pits++ // counts stints; corrected below
			d.Compound = s.Compound
			d.TyreAge = s.Age + d.Lap - s.LapStart + 1
		}
	}
	for i := range out {
		if out[i].Pits > 0 {
			out[i].Pits--
		}
	}
	for _, p := range loc {
		if d := byNum[p.Number]; d != nil && (p.X != 0 || p.Y != 0) {
			d.X, d.Y = p.X, p.Y
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

	live := Live{Session: sessions[0], Drivers: out, UpdatedAt: time.Now().UTC().Format(time.RFC3339)}
	if len(weather) > 0 {
		live.Weather = &weather[len(weather)-1]
	}
	for i := len(rc) - 1; i >= 0 && len(live.RaceControl) < 20; i-- {
		live.RaceControl = append(live.RaceControl, rc[i])
	}
	return live, nil
}
