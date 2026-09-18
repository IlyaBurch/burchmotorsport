package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	rec := httptest.NewRecorder()
	handler().ServeHTTP(rec, httptest.NewRequest("GET", "/api/health", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != `{"status":"ok"}` {
		t.Fatalf("got %d %q", rec.Code, rec.Body.String())
	}
}

func TestFetchLiveMerges(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/sessions":
			w.Write([]byte(`[{"session_key":1,"session_name":"Race"}]`))
		case "/drivers":
			w.Write([]byte(`[{"driver_number":1,"name_acronym":"NOR"},{"driver_number":3,"name_acronym":"VER"}]`))
		case "/position":
			w.Write([]byte(`[{"driver_number":1,"position":1},{"driver_number":3,"position":2},{"driver_number":3,"position":1},{"driver_number":1,"position":2}]`))
		case "/intervals":
			w.Write([]byte(`[{"driver_number":1,"gap_to_leader":1.5,"interval":1.5}]`))
		case "/laps":
			w.Write([]byte(`[{"driver_number":1,"lap_number":1,"lap_duration":90.5},{"driver_number":1,"lap_number":2,"lap_duration":91.0}]`))
		}
	}))
	defer srv.Close()
	openf1 = srv.URL + "/"

	got, err := fetchLive()
	if err != nil {
		t.Fatal(err)
	}
	if got.Drivers[0].Acronym != "VER" || got.Drivers[0].Position != 1 {
		t.Fatalf("expected VER P1 after last position update, got %+v", got.Drivers[0])
	}
	nor := got.Drivers[1]
	if nor.Lap != 2 || *nor.LastLap != 91.0 || *nor.BestLap != 90.5 || nor.Gap != 1.5 {
		t.Fatalf("bad merge: %+v", nor)
	}
}
