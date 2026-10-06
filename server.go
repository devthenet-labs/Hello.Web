package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"time"
)

// page is the service's one page.
const page = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>%[1]s</title>
<style>%[5]s</style>
</head>
<body>
<div class="card %[6]s">
<h1>%[1]s</h1>
<p>%[4]s</p>
<p><small>%[3]s</small></p>
<p>Revision <code>%[2]s</code>.</p>
</div>
</body>
</html>
`

// styleCSS is the page's one stylesheet, inlined in a <style> element so
// the CSP's style-src can allow it by its sha256 hash alone, with no
// 'unsafe-inline' and no external file.
const styleCSS = `:root { color-scheme: light dark; }
body {
  margin: 0;
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: system-ui, -apple-system, "Segoe UI", Roboto, sans-serif;
  background: #f5f5f7;
  color: #1a1a1a;
}
.card {
  border: 1px solid #d0d0d5;
  border-radius: 12px;
  padding: 2rem 2.5rem;
  text-align: center;
  max-width: 24rem;
  margin: 1rem;
}
.card.morning {
  background: #fff6e0;
  color: #6b4b00;
  border-color: #f0d998;
}
.card.afternoon {
  background: #e7f3ff;
  color: #0b4a6f;
  border-color: #b8ddfb;
}
.card.evening {
  background: #2a2d42;
  color: #eef1fb;
  border-color: #454a6b;
}
h1 {
  margin: 0 0 0.5rem 0;
  font-size: 1.5rem;
}
p {
  margin: 0.25rem 0;
}
`

// styleHash is the CSP hash-source for styleCSS: the base64 encoding of
// its SHA-256 digest, computed once at package init so it always matches
// the stylesheet the page serves.
var styleHash = func() string {
	sum := sha256.Sum256([]byte(styleCSS))
	return base64.StdEncoding.EncodeToString(sum[:])
}()

// cspHeader is the Content-Security-Policy every response carries:
// default deny, no framing, and the page's own stylesheet allowed by
// hash alone.
var cspHeader = fmt.Sprintf("default-src 'none'; frame-ancestors 'none'; style-src 'sha256-%s'", styleHash)

// period returns the time-of-day bucket for a UTC hour (0-23): the same
// clock the date line already uses. Morning is 05:00-11:59, afternoon is
// 12:00-16:59, evening is the rest (17:00-04:59). It names both the
// greeting and the card's colour scheme, so the two never fall out of
// sync with each other.
func period(hour int) string {
	switch {
	case hour >= 5 && hour < 12:
		return "morning"
	case hour >= 12 && hour < 17:
		return "afternoon"
	default:
		return "evening"
	}
}

// greeting returns the time-of-day greeting for a UTC hour (0-23).
func greeting(hour int) string {
	switch period(hour) {
	case "morning":
		return "Good morning"
	case "afternoon":
		return "Good afternoon"
	default:
		return "Good evening"
	}
}

// newHandler serves the page at / and the readiness check at /healthz.
func newHandler(revision string, now func() time.Time) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "revision": revision})
	})
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		t := now().UTC()
		today := t.Format("2 January 2006")
		_, _ = fmt.Fprintf(w, page,
			html.EscapeString("Hello.Web"),
			html.EscapeString(revision),
			html.EscapeString(today),
			html.EscapeString(greeting(t.Hour())),
			styleCSS,
			html.EscapeString(period(t.Hour())),
		)
	})
	return withSecurityHeaders(mux)
}

// withSecurityHeaders sets the headers every response carries.
func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", cspHeader)
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}
