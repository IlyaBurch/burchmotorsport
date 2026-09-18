package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// Resolve maps a RuTube video to an openf1 session: publication date picks the
// weekend, the title picks the session type, a country in the title confirms.
type Resolve struct {
	Title      string   `json:"title"`
	Published  string   `json:"published"`
	Session    *Session `json:"session"`    // nil when nothing matched
	Confidence string   `json:"confidence"` // "date+title", "date", "title", ""
}

var sessionWords = []struct {
	re   *regexp.Regexp
	name string
}{
	{regexp.MustCompile(`(?i)спринт[-\s]*квалиф|sprint qual|shootout|шутаут`), "Sprint Qualifying"},
	{regexp.MustCompile(`(?i)спринт|sprint`), "Sprint"},
	{regexp.MustCompile(`(?i)квалиф|qualif`), "Qualifying"},
	{regexp.MustCompile(`(?i)(практик|тренир|свободн|practice|fp)\D{0,6}1`), "Practice 1"},
	{regexp.MustCompile(`(?i)(практик|тренир|свободн|practice|fp)\D{0,6}2`), "Practice 2"},
	{regexp.MustCompile(`(?i)(практик|тренир|свободн|practice|fp)\D{0,6}3`), "Practice 3"},
	{regexp.MustCompile(`(?i)гонк|race`), "Race"},
}

// ru title fragment → openf1 country_name. Lowercase, stem-ish.
var countries = map[string]string{
	"австрал": "Australia", "кита": "China", "япон": "Japan", "бахрейн": "Bahrain", "саудов": "Saudi Arabia",
	"майами": "United States", "эмилия": "Italy", "имол": "Italy", "монако": "Monaco", "испан": "Spain", "барселон": "Spain",
	"канад": "Canada", "австри": "Austria", "великобритан": "United Kingdom", "британ": "United Kingdom", "сильверстоун": "United Kingdom",
	"бельги": "Belgium", "венгр": "Hungary", "нидерланд": "Netherlands", "голланд": "Netherlands", "итали": "Italy", "монц": "Italy",
	"азербайджан": "Azerbaijan", "баку": "Azerbaijan", "сингапур": "Singapore", "сша": "United States", "остин": "United States",
	"мексик": "Mexico", "бразил": "Brazil", "лас-вегас": "United States", "вегас": "United States", "катар": "Qatar", "абу-даби": "United Arab Emirates", "абу даби": "United Arab Emirates",
	"мадрид": "Spain", "малайз": "Malaysia",
}

func sessionFromTitle(title string) string {
	for _, w := range sessionWords {
		if w.re.MatchString(title) {
			return w.name
		}
	}
	return "Race"
}

func countryFromTitle(title string) string {
	t := strings.ToLower(title)
	for frag, c := range countries {
		if strings.Contains(t, frag) {
			return c
		}
	}
	return ""
}

var yearRe = regexp.MustCompile(`\b(20\d\d)\b`)

// GET /api/resolve?rutube=<id>
func resolveHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("rutube")
	if !regexp.MustCompile(`^[0-9a-f]{32}$`).MatchString(id) {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	v, err := store.get("resolve:"+id, func() (any, time.Duration, error) {
		res, err := resolveRutube(id)
		return res, forever, err
	})
	writeJSON(w, v, err)
}

var rutubeAPI = "https://rutube.ru/api/video/"

func resolveRutube(id string) (Resolve, error) {
	req, _ := http.NewRequest("GET", rutubeAPI+id+"/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 burchmotorsport") // rutube 403s the default Go UA
	resp, err := client.Do(req)
	if err != nil {
		return Resolve{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return Resolve{}, fmt.Errorf("rutube %s: %s", id, resp.Status)
	}
	var meta struct {
		Title       string `json:"title"`
		Publication string `json:"publication_ts"`
		Created     string `json:"created_ts"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return Resolve{}, err
	}
	pub := meta.Publication
	if pub == "" {
		pub = meta.Created
	}
	res := Resolve{Title: meta.Title, Published: pub}

	want := sessionFromTitle(meta.Title)
	country := countryFromTitle(meta.Title)

	// 1. by date: sessions that started in the 4 days before publication
	if t, e := time.Parse("2006-01-02T15:04:05", pub); e == nil {
		var s []Session
		q := fmt.Sprintf("sessions?date_start>=%s&date_start<=%s", t.Add(-4*24*time.Hour).Format("2006-01-02"), t.Add(6*time.Hour).Format("2006-01-02T15:04:05"))
		if err := cached(q, forever, &s); err == nil {
			if best := pick(s, want, country); best != nil {
				res.Session, res.Confidence = best, "date"
				if country != "" && best.Country == country {
					res.Confidence = "date+title"
				}
				return res, nil
			}
		}
	}
	// 2. by title: country + year (re-uploads published long after the weekend)
	if year := yearRe.FindString(meta.Title); year != "" && country != "" {
		var s []Session
		q := fmt.Sprintf("sessions?year=%s&country_name=%s&session_name=%s", year, strings.ReplaceAll(country, " ", "%20"), strings.ReplaceAll(want, " ", "%20"))
		if err := cached(q, forever, &s); err == nil && len(s) > 0 {
			res.Session, res.Confidence = &s[len(s)-1], "title"
		}
	}
	return res, nil
}

// pick prefers the wanted session type in the wanted country, then the wanted
// type anywhere, then the latest session. Latest wins on ties.
func pick(s []Session, want, country string) *Session {
	var byType, any *Session
	for i := range s {
		x := &s[i]
		if any == nil || x.Start > any.Start {
			any = x
		}
		if x.Name != want {
			continue
		}
		if country != "" && x.Country == country {
			return x
		}
		if byType == nil || x.Start > byType.Start {
			byType = x
		}
	}
	if byType != nil {
		return byType
	}
	return any
}
