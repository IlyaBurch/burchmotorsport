package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResolveRutube(t *testing.T) {
	resetCache()
	rt := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"title":"Формула 1 - Гран-При Великобритании 2025 - Гонка | Сильверстоун","publication_ts":"2025-07-06T16:27:21"}`))
	}))
	defer rt.Close()
	of := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "date_start") {
			t.Fatalf("expected a date query first, got %s", r.URL.RawQuery)
		}
		w.Write([]byte(`[
			{"session_key":1,"session_name":"Qualifying","country_name":"United Kingdom","date_start":"2025-07-05T14:00:00+00:00"},
			{"session_key":2,"session_name":"Race","country_name":"United Kingdom","date_start":"2025-07-06T14:00:00+00:00"}]`))
	}))
	defer of.Close()
	rutubeAPI, openf1 = rt.URL+"/", of.URL+"/"

	got, err := resolveRutube("ee1a3832d90cc60b507ff90a4d978a43")
	if err != nil || got.Session == nil || got.Session.Key != 2 || got.Confidence != "date+title" {
		t.Fatalf("got %+v err %v", got, err)
	}
}

func TestSessionFromTitle(t *testing.T) {
	cases := map[string]string{
		"Квалификация Гран-при Монако":   "Qualifying",
		"Спринт-квалификация Китай 2026": "Sprint Qualifying",
		"Спринт Майами":                  "Sprint",
		"Свободная практика 2 Бахрейн":   "Practice 2",
		"FP3 Silverstone": "Practice 3",
		"ГОНКА ГООННККААААА Великобритания 2026": "Race",
		"Формула 1 Монца":                        "Race",
	}
	for title, want := range cases {
		if got := sessionFromTitle(title); got != want {
			t.Errorf("%q → %q, want %q", title, got, want)
		}
	}
	if countryFromTitle("ГОНКА ГООННККААААА Великобритания 9 этап 2026") != "United Kingdom" {
		t.Error("country from title")
	}
}
