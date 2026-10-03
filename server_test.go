package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestHandler(t *testing.T) {
	handler := newHandler("abc123", fixedClock(time.Date(2026, time.October, 3, 15, 4, 5, 0, time.UTC)))
	cases := []struct {
		name, method, path string
		status             int
		contentType, body  string
	}{
		{"page", http.MethodGet, "/", http.StatusOK, "text/html; charset=utf-8", "<h1>Hello.Web</h1>"},
		{"readiness", http.MethodGet, "/healthz", http.StatusOK, "application/json", `"status":"ok"`},
		{"unknown path", http.MethodGet, "/missing", http.StatusNotFound, "", ""},
		{"wrong method", http.MethodPost, "/", http.StatusMethodNotAllowed, "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, nil))
			if rec.Code != tc.status {
				t.Fatalf("%s %s = %d, want %d", tc.method, tc.path, rec.Code, tc.status)
			}
			if got := rec.Header().Get("Content-Type"); tc.contentType != "" && got != tc.contentType {
				t.Errorf("Content-Type = %q, want %q", got, tc.contentType)
			}
			if !strings.Contains(rec.Body.String(), tc.body) {
				t.Errorf("body lacks %q:\n%s", tc.body, rec.Body.String())
			}
			if tc.name == "page" && !strings.Contains(rec.Body.String(), "3 October 2026") {
				t.Errorf("body lacks rendered date:\n%s", rec.Body.String())
			}
			if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
				t.Errorf("X-Content-Type-Options = %q, want nosniff", got)
			}
		})
	}
}

func TestReadinessReportsRevision(t *testing.T) {
	rec := httptest.NewRecorder()
	newHandler("abc123", fixedClock(time.Date(2026, time.October, 3, 15, 4, 5, 0, time.UTC))).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	var got map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("readiness body is not JSON: %v", err)
	}
	if got["revision"] != "abc123" {
		t.Errorf("revision = %q, want abc123", got["revision"])
	}
}

func TestHomePageShowsDate(t *testing.T) {
	now := fixedClock(time.Date(2027, time.January, 9, 23, 59, 0, 0, time.UTC))
	rec := httptest.NewRecorder()
	newHandler("abc123", now).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	body := rec.Body.String()
	want := "<h1>Hello.Web</h1>\n<p><small>9 January 2027</small></p>"
	if !strings.Contains(body, want) {
		t.Errorf("body = %s\nwant substring %q", body, want)
	}
}

func TestHomePageShowsDateConvertsToUTC(t *testing.T) {
	now := fixedClock(time.Date(2026, time.October, 3, 23, 30, 0, 0, time.FixedZone("UTC-5", -5*3600)))
	rec := httptest.NewRecorder()
	newHandler("abc123", now).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	body := rec.Body.String()
	if !strings.Contains(body, "4 October 2026") {
		t.Errorf("body = %s\nwant substring %q", body, "4 October 2026")
	}
}
