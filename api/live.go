package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
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
	Stints   []Stint     `json:"stints"`
	ByLap    []int       `json:"positions"`   // position after each lap, index 0 = lap 1, 0 = unknown
	ChampPos int         `json:"champPos"`    // standings before this session, 0 if unknown
	ChampPts float64     `json:"champPoints"` // points before this session
	X        int         `json:"x"`           // track coords, 0,0 if unknown
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

type Stint struct {
	Compound string `json:"compound"`
	From     int    `json:"from"` // lap
	To       int    `json:"to"`   // lap, inclusive
}

type Team struct {
	Name   string  `json:"name"`
	Colour string  `json:"colour"`
	Pos    int     `json:"champPos"`
	Points float64 `json:"champPoints"`
}

type Radio struct {
	Date   string `json:"date"`
	Number int    `json:"driver_number"`
	URL    string `json:"recording_url"`
}

// errNoData: openf1 knows the session but has no timing for it (yet, or at all)
var errNoData = errors.New("openf1: no data for session")

type Live struct {
	Session     Session       `json:"session"`
	Upcoming    bool          `json:"upcoming"` // session exists but has not started
	Drivers     []Driver      `json:"drivers"`
	Teams       []Team        `json:"teams"`
	Radio       []Radio       `json:"radio"` // newest first
	Weather     *Weather      `json:"weather"`
	RaceControl []RaceControl `json:"raceControl"` // newest first
	UpdatedAt   string        `json:"updatedAt"`
}

const forever = 100 * 365 * 24 * time.Hour

func writeJSON(w http.ResponseWriter, v any, err error) {
	if err != nil {
		if d := cooling(); d > 0 {
			w.Header().Set("Retry-After", fmt.Sprint(int(d.Seconds())+1))
			http.Error(w, "openf1 rate limit", http.StatusServiceUnavailable)
			return
		}
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
		if l.Upcoming {
			return l, time.Minute, nil
		}
		if end, e := time.Parse(time.RFC3339, l.Session.End); e == nil && end.Add(time.Hour).Before(time.Now()) {
			return l, forever, nil
		}
		return l, 5 * time.Second, nil // underlying endpoints have their own ttls
	})
	if errors.Is(err, errNoData) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
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
	if err := cached("laps?lap_number=3&session_key="+sessionKey, forever, &laps); err != nil {
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
		var loc []struct{ X, Y int }
		q := fmt.Sprintf("location?session_key=%s&driver_number=%d&date>=%s&date<%s",
			sessionKey, l.Number, start.Format(time.RFC3339), end.Format(time.RFC3339))
		if err := cached(q, forever, &loc); err != nil {
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

// ponytail: one global limiter for openf1 (3 req/s), shared by every fetch.
// A mutex + timestamp, not time.Tick: Tick buffers a tick and lets two calls
// through at once after idle, which is exactly what got us 429s.
var (
	limMu   sync.Mutex
	limLast time.Time
)

// free: 3 req/s, 30/min. sponsor: 6 req/s, 60/min.
func limInterval() time.Duration {
	if sponsored() {
		return 200 * time.Millisecond
	}
	return 400 * time.Millisecond
}

func throttle() {
	limMu.Lock()
	defer limMu.Unlock()
	if d := time.Until(limLast.Add(limInterval())); d > 0 {
		time.Sleep(d)
	}
	limLast = time.Now()
}

var client = &http.Client{Timeout: 20 * time.Second}

// openf1 answers 429 with Retry-After once the per-minute budget is spent.
// Until then every call fails fast instead of digging the hole deeper.
var (
	coolMu    sync.Mutex
	coolUntil time.Time
)

func cooling() time.Duration {
	coolMu.Lock()
	defer coolMu.Unlock()
	return time.Until(coolUntil)
}

// get fetches one openf1 endpoint, retrying once on 5xx or a network error.
func get(path string, v any) error {
	if d := cooling(); d > 0 {
		return fmt.Errorf("openf1 %s: rate limited, %s left", path, d.Round(time.Second))
	}
	var err error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			log.Printf("openf1 retry %s: %v", path, err)
			time.Sleep(2 * time.Second)
		}
		req, _ := http.NewRequest("GET", openf1+path, nil)
		if err = authorize(req); err != nil {
			return err
		}
		throttle()
		var resp *http.Response
		resp, err = client.Do(req)
		if err != nil {
			continue
		}
		if resp.StatusCode == 429 {
			resp.Body.Close()
			wait := 60 * time.Second
			if ra, e := time.ParseDuration(resp.Header.Get("Retry-After") + "s"); e == nil && ra > 0 {
				wait = ra
			}
			coolMu.Lock()
			coolUntil = time.Now().Add(wait)
			coolMu.Unlock()
			log.Printf("openf1 429 on %s, cooling down %s", path, wait)
			return fmt.Errorf("openf1 %s: 429, cooling down %s", path, wait)
		}
		if resp.StatusCode >= 500 {
			resp.Body.Close()
			err = fmt.Errorf("openf1 %s: %s", path, resp.Status)
			continue
		}
		if resp.StatusCode != 200 {
			resp.Body.Close()
			return fmt.Errorf("openf1 %s: %s", path, resp.Status)
		}
		err = json.NewDecoder(resp.Body).Decode(v)
		resp.Body.Close()
		return err
	}
	return err
}

// cached wraps get with a per-path cache so each openf1 endpoint is refreshed on
// its own schedule. Raw JSON is stored and decoded into v on every hit.
func cached(path string, ttl time.Duration, v any) error {
	raw, err := store.get("u:"+path, func() (any, time.Duration, error) {
		var r json.RawMessage
		return r, ttl, get(path, &r)
	})
	if err != nil {
		return err
	}
	return json.Unmarshal(raw.(json.RawMessage), v)
}

func fetchLive(sessionKey string) (Live, error) {
	var sessions []Session
	if err := cached("sessions?session_key="+sessionKey, time.Minute, &sessions); err != nil {
		if strings.Contains(err.Error(), "404") {
			return Live{}, errNoData
		}
		return Live{}, err
	}
	if len(sessions) == 0 {
		return Live{}, errNoData
	}
	key := fmt.Sprint(sessions[0].Key)

	// not started yet: nothing to fetch, openf1 answers 404s or empty lists at random
	if start, e := time.Parse(time.RFC3339, sessions[0].Start); e == nil && start.After(time.Now()) {
		return upcomingLive(sessions[0]), nil
	}

	// budget per live session: free 30/min → hot 4×6 + warm 4 ≈ 28;
	// sponsor 60/min → hot 4×12 + warm 4 ≈ 52. Finished sessions never change.
	hot, warm, cold := 10*time.Second, time.Minute, 10*time.Minute
	if sponsored() {
		hot = 5 * time.Second
	}
	if end, e := time.Parse(time.RFC3339, sessions[0].End); e == nil && end.Add(time.Hour).Before(time.Now()) {
		hot, warm, cold = forever, forever, forever
	}

	var drivers []struct {
		Number  int    `json:"driver_number"`
		Acronym string `json:"name_acronym"`
		Name    string `json:"full_name"`
		Team    string `json:"team_name"`
		Colour  string `json:"team_colour"`
	}
	var positions []struct {
		Number   int    `json:"driver_number"`
		Position int    `json:"position"`
		Date     string `json:"date"`
	}
	var intervals []struct {
		Number   int `json:"driver_number"`
		Gap      any `json:"gap_to_leader"`
		Interval any `json:"interval"`
	}
	var laps []struct {
		Number   int      `json:"driver_number"`
		Lap      int      `json:"lap_number"`
		Start    string   `json:"date_start"`
		Duration *float64 `json:"lap_duration"`
		S1       *float64 `json:"duration_sector_1"`
		S2       *float64 `json:"duration_sector_2"`
		S3       *float64 `json:"duration_sector_3"`
		Speed    int      `json:"st_speed"`
	}
	var stints []struct {
		Number   int    `json:"driver_number"`
		LapStart int    `json:"lap_start"`
		LapEnd   int    `json:"lap_end"`
		Compound string `json:"compound"`
		Age      int    `json:"tyre_age_at_start"`
	}
	var weather []Weather
	var rc []RaceControl
	var champ []struct {
		Number int     `json:"driver_number"`
		Pos    *int    `json:"position_start"`
		Pts    float64 `json:"points_start"`
	}
	var champTeams []struct {
		Name string  `json:"team_name"`
		Pos  *int    `json:"position_start"`
		Pts  float64 `json:"points_start"`
	}
	var radio []Radio

	// all in parallel, throttle() paces the ones that miss cache
	if err := getAll(
		query{"drivers?session_key=" + key, cold, &drivers},
		query{"position?session_key=" + key, hot, &positions},
		query{"intervals?session_key=" + key, hot, &intervals},
		query{"laps?session_key=" + key, hot, &laps},
		query{"stints?session_key=" + key, warm, &stints},
		query{"weather?session_key=" + key, warm, &weather},
		query{"race_control?session_key=" + key, warm, &rc},
		query{"championship_drivers?session_key=" + key, cold, &champ},
		query{"championship_teams?session_key=" + key, cold, &champTeams},
		query{"team_radio?session_key=" + key, warm, &radio},
	); err != nil {
		if !strings.Contains(err.Error(), "404") {
			return Live{}, err
		}
		// on the calendar, already started, but no timing rows
		return Live{}, errNoData
	}

	// car positions: last 10s for a live session, around the chequered flag otherwise
	at := time.Now().UTC().Truncate(10 * time.Second).Add(-10 * time.Second)
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
	if err := cached(fmt.Sprintf("location?session_key=%s&date>=%s&date<%s", key,
		at.Format(time.RFC3339), at.Add(10*time.Second).Format(time.RFC3339)), hot, &loc); err != nil {
		log.Printf("location: %v", err) // map dots are optional, timing is not
	}

	byNum := map[int]*Driver{}
	out := make([]Driver, 0, len(drivers))
	for _, d := range drivers {
		out = append(out, Driver{Number: d.Number, Acronym: d.Acronym, Name: d.Name, Team: d.Team, Colour: d.Colour})
	}
	for i := range out {
		byNum[out[i].Number] = &out[i]
	}
	// lap timeline: lap n starts when the first car starts it
	lapStart := map[int]time.Time{}
	maxLap := 0
	for _, l := range laps {
		t, err := time.Parse(time.RFC3339Nano, l.Start)
		if err != nil {
			continue
		}
		if cur, ok := lapStart[l.Lap]; !ok || t.Before(cur) {
			lapStart[l.Lap] = t
		}
		if l.Lap > maxLap {
			maxLap = l.Lap
		}
	}
	lapAt := func(t time.Time) int {
		n := 0
		for lap := 1; lap <= maxLap; lap++ {
			if st, ok := lapStart[lap]; ok && !st.After(t) {
				n = lap
			}
		}
		return n
	}
	for i := range out {
		out[i].ByLap = make([]int, maxLap)
	}
	// arrays are chronological: last write wins
	for _, p := range positions {
		d := byNum[p.Number]
		if d == nil {
			continue
		}
		d.Position = p.Position
		if t, err := time.Parse(time.RFC3339Nano, p.Date); err == nil {
			if lap := lapAt(t); lap > 0 {
				d.ByLap[lap-1] = p.Position
			}
		}
	}
	// carry positions forward through laps with no change
	for i := range out {
		for lap := 1; lap < len(out[i].ByLap); lap++ {
			if out[i].ByLap[lap] == 0 {
				out[i].ByLap[lap] = out[i].ByLap[lap-1]
			}
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
			to := s.LapEnd
			if to == 0 {
				to = d.Lap
			}
			d.Stints = append(d.Stints, Stint{Compound: s.Compound, From: s.LapStart, To: to})
		}
	}
	for i := range out {
		if out[i].Pits > 0 {
			out[i].Pits--
		}
	}
	for _, c := range champ {
		if d := byNum[c.Number]; d != nil {
			d.ChampPts = c.Pts
			if c.Pos != nil {
				d.ChampPos = *c.Pos
			}
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

	colour := map[string]string{}
	for _, d := range out {
		colour[d.Team] = d.Colour
	}
	teams := make([]Team, 0, len(champTeams))
	for _, t := range champTeams {
		team := Team{Name: t.Name, Colour: colour[t.Name], Points: t.Pts}
		if t.Pos != nil {
			team.Pos = *t.Pos
		}
		teams = append(teams, team)
	}

	live := Live{Session: sessions[0], Drivers: out, Teams: teams, Radio: []Radio{}, RaceControl: []RaceControl{}, UpdatedAt: time.Now().UTC().Format(time.RFC3339)}
	for i := len(radio) - 1; i >= 0; i-- {
		live.Radio = append(live.Radio, radio[i])
	}
	if len(weather) > 0 {
		live.Weather = &weather[len(weather)-1]
	}
	for i := len(rc) - 1; i >= 0; i-- {
		live.RaceControl = append(live.RaceControl, rc[i])
	}
	return live, nil
}

func upcomingLive(s Session) Live {
	return Live{
		Session: s, Upcoming: true, UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		Drivers: []Driver{}, Teams: []Team{}, Radio: []Radio{}, RaceControl: []RaceControl{}, // arrays, never null
	}
}

type query struct {
	path string
	ttl  time.Duration
	into any
}

func getAll(qs ...query) error {
	var wg sync.WaitGroup
	errs := make([]error, len(qs))
	for i, q := range qs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs[i] = cached(q.path, q.ttl, q.into)
		}()
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
