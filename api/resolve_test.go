package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResolveTitle(t *testing.T) {
	resetCache()
	of := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.RawQuery
		if !strings.Contains(q, "year=2025") || !strings.Contains(q, "country_name=United%20Kingdom") || !strings.Contains(q, "session_name=Race") {
			t.Fatalf("unexpected query %s", q)
		}
		w.Write([]byte(`[{"session_key":9947,"session_name":"Race","country_name":"United Kingdom","date_start":"2025-07-06T14:00:00+00:00"}]`))
	}))
	defer of.Close()
	openf1 = of.URL + "/"

	got, err := resolveTitle("Формула 1 - Гран-При Великобритании 2025 - Гонка | Сильверстоун")
	if err != nil || got.Session == nil || got.Session.Key != 9947 || !got.F1 {
		t.Fatalf("got %+v err %v", got, err)
	}

	got, _ = resolveTitle("Котики играют")
	if got.F1 || got.Session != nil {
		t.Fatalf("cats are not F1: %+v", got)
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
	for title, want := range map[string]bool{
		"Формула 1 - Гран-При Италии": true, "F1 Monza race": true, "Ф1 Монца": true, "Котики играют": false,
	} {
		if f1Re.MatchString(title) != want {
			t.Errorf("f1 %q want %v", title, want)
		}
	}
}
