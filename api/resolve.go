package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// Resolve maps a video title to an openf1 session. The title is all we get:
// the RuTube player hands it to the browser via getPlayOptions, and RuTube's
// own API is unreachable from outside Russia, so the server never talks to it.
type Resolve struct {
	Title   string   `json:"title"`
	Session *Session `json:"session"` // nil when nothing matched
	F1      bool     `json:"f1"`      // looks like Formula 1 at all
}

// \b is ASCII-only in Go, so word edges are spelled out for the Cyrillic forms
var f1Re = regexp.MustCompile(`(?i)формул|formula|grand prix|гран[\s-]*при|(?:^|[^\pL\pN])(?:f1|ф1|ф-1)(?:[^\pL\pN]|$)`)

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

// ru/en title fragment → openf1 country_name. Lowercase, stem-ish.
var countries = map[string]string{
	"австрал": "Australia", "australia": "Australia", "кита": "China", "china": "China", "япон": "Japan", "japan": "Japan",
	"бахрейн": "Bahrain", "bahrain": "Bahrain", "саудов": "Saudi Arabia", "saudi": "Saudi Arabia",
	"майами": "United States", "miami": "United States", "эмилия": "Italy", "имол": "Italy", "imola": "Italy",
	"монако": "Monaco", "monaco": "Monaco", "испан": "Spain", "spain": "Spain", "барселон": "Spain", "мадрид": "Spain", "madrid": "Spain",
	"канад": "Canada", "canada": "Canada", "австри": "Austria", "austria": "Austria",
	"великобритан": "United Kingdom", "британ": "United Kingdom", "british": "United Kingdom", "сильверстоун": "United Kingdom", "silverstone": "United Kingdom",
	"бельги": "Belgium", "belgi": "Belgium", "венгр": "Hungary", "hungar": "Hungary", "нидерланд": "Netherlands", "голланд": "Netherlands", "dutch": "Netherlands",
	"итали": "Italy", "монц": "Italy", "monza": "Italy", "italian": "Italy",
	"азербайджан": "Azerbaijan", "баку": "Azerbaijan", "baku": "Azerbaijan", "сингапур": "Singapore", "singapore": "Singapore",
	"сша": "United States", "остин": "United States", "austin": "United States", "лас-вегас": "United States", "вегас": "United States", "vegas": "United States",
	"мексик": "Mexico", "mexic": "Mexico", "бразил": "Brazil", "brazil": "Brazil", "катар": "Qatar", "qatar": "Qatar",
	"абу-даби": "United Arab Emirates", "абу даби": "United Arab Emirates", "abu dhabi": "United Arab Emirates", "малайз": "Malaysia", "malaysia": "Malaysia",
}

var yearRe = regexp.MustCompile(`\b(20\d\d)\b`)

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

// GET /api/resolve?title=<video title>
func resolveHandler(w http.ResponseWriter, r *http.Request) {
	title := strings.TrimSpace(r.URL.Query().Get("title"))
	if title == "" || len(title) > 300 {
		http.Error(w, "bad title", http.StatusBadRequest)
		return
	}
	v, err := store.get("resolve:"+title, func() (any, time.Duration, error) {
		res, err := resolveTitle(title)
		return res, 24 * time.Hour, err
	})
	writeJSON(w, v, err)
}

func resolveTitle(title string) (Resolve, error) {
	country := countryFromTitle(title)
	res := Resolve{Title: title, F1: f1Re.MatchString(title) || country != ""}
	if country == "" {
		return res, nil
	}
	year := yearRe.FindString(title)
	if year == "" {
		year = fmt.Sprint(time.Now().Year())
	}
	var s []Session
	q := fmt.Sprintf("sessions?year=%s&country_name=%s&session_name=%s",
		year, strings.ReplaceAll(country, " ", "%20"), strings.ReplaceAll(sessionFromTitle(title), " ", "%20"))
	if err := cached(q, forever, &s); err != nil {
		return res, err
	}
	if len(s) > 0 {
		res.Session, res.F1 = &s[len(s)-1], true
	}
	return res, nil
}
