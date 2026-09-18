package main

import (
	"fmt"
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
			w.Write([]byte(`[{"driver_number":1,"name_acronym":"NOR","team_name":"McLaren","team_colour":"FF8000"},{"driver_number":3,"name_acronym":"VER"}]`))
		case "/position":
			w.Write([]byte(`[{"driver_number":1,"position":1,"date":"2025-01-01T00:00:00Z"},{"driver_number":3,"position":2,"date":"2025-01-01T00:00:00Z"},{"driver_number":3,"position":1,"date":"2025-01-01T00:01:30Z"},{"driver_number":1,"position":2,"date":"2025-01-01T00:01:30Z"}]`))
		case "/intervals":
			w.Write([]byte(`[{"driver_number":1,"gap_to_leader":1.5,"interval":1.5}]`))
		case "/laps":
			w.Write([]byte(`[{"driver_number":1,"lap_number":1,"date_start":"2025-01-01T00:00:00Z","lap_duration":90.5,"duration_sector_1":30.0},{"driver_number":1,"lap_number":2,"date_start":"2025-01-01T00:01:30Z","lap_duration":91.0,"duration_sector_1":30.5}]`))
		case "/stints":
			w.Write([]byte(`[{"driver_number":1,"lap_start":1,"lap_end":1,"compound":"SOFT","tyre_age_at_start":2},{"driver_number":1,"lap_start":2,"compound":"HARD","tyre_age_at_start":0}]`))
		case "/weather":
			w.Write([]byte(`[{"air_temperature":20},{"air_temperature":25}]`))
		case "/race_control":
			w.Write([]byte(`[{"message":"old"},{"message":"new"}]`))
		case "/championship_drivers":
			w.Write([]byte(`[{"driver_number":1,"position_start":2,"points_start":100.5}]`))
		case "/championship_teams":
			w.Write([]byte(`[{"team_name":"McLaren","position_start":1,"points_start":200}]`))
		case "/team_radio":
			w.Write([]byte(`[{"driver_number":1,"date":"a","recording_url":"u1"},{"driver_number":1,"date":"b","recording_url":"u2"}]`))
		case "/pit":
			w.Write([]byte(`[{"driver_number":1,"lap_number":14,"pit_duration":21.5}]`))
		case "/location":
			w.Write([]byte(`[{"driver_number":1,"x":10,"y":20},{"driver_number":1,"x":11,"y":21}]`))
		default:
			w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()
	openf1 = srv.URL + "/"

	got, err := fetchLive("latest")
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
	if *nor.Sectors[0] != 30.5 || *nor.BestSect[0] != 30.0 {
		t.Fatalf("bad sectors: %+v", nor)
	}
	if nor.Compound != "HARD" || nor.TyreAge != 1 || nor.Pits != 1 {
		t.Fatalf("bad stint: %+v", nor)
	}
	if len(nor.Stints) != 2 || nor.Stints[1] != (Stint{"HARD", 2, 2}) {
		t.Fatalf("bad stints: %+v", nor.Stints)
	}
	if fmt.Sprint(nor.ByLap) != "[1 2]" || fmt.Sprint(got.Drivers[0].ByLap) != "[2 1]" {
		t.Fatalf("bad positions by lap: NOR %v VER %v", nor.ByLap, got.Drivers[0].ByLap)
	}
	if nor.ChampPos != 2 || nor.ChampPts != 100.5 {
		t.Fatalf("bad championship: %+v", nor)
	}
	if len(got.Teams) != 1 || got.Teams[0].Colour != "FF8000" || got.Teams[0].Pos != 1 {
		t.Fatalf("bad teams: %+v", got.Teams)
	}
	if len(got.Pits) != 1 || *got.Pits[0].Duration != 21.5 {
		t.Fatalf("bad pits: %+v", got.Pits)
	}
	if got.Radio[0].URL != "u2" {
		t.Fatalf("radio not newest first: %+v", got.Radio)
	}
	if nor.X != 11 || nor.Y != 21 {
		t.Fatalf("bad location: %+v", nor)
	}
	if got.Weather.Air != 25 || got.RaceControl[0].Message != "new" {
		t.Fatalf("bad weather/rc: %+v %+v", got.Weather, got.RaceControl)
	}
}
