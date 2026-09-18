package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestBearerAttachedAndCached(t *testing.T) {
	tokens := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			tokens++
			if r.FormValue("username") != "u" || r.FormValue("password") != "p" {
				t.Fatalf("bad form: %v", r.Form)
			}
			w.Write([]byte(`{"access_token":"abc","expires_in":3600,"token_type":"bearer"}`))
			return
		}
		if r.Header.Get("Authorization") != "Bearer abc" {
			t.Fatalf("missing bearer: %q", r.Header.Get("Authorization"))
		}
		w.Write([]byte(`[]`))
	}))
	defer ts.Close()
	openf1User, openf1Pass, tokenURL, openf1 = "u", "p", ts.URL+"/token", ts.URL+"/"
	token, tokExp = "", time.Time{}
	defer func() { openf1User, openf1Pass = "", "" }()

	var v []any
	for i := 0; i < 3; i++ {
		if err := get("sessions", &v); err != nil {
			t.Fatal(err)
		}
	}
	if tokens != 1 {
		t.Fatalf("token requested %d times, want 1", tokens)
	}
}
