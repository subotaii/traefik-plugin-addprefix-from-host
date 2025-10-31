package traefik_plugin_addprefix_from_host

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newMiddleware(t *testing.T, cfg *Config, next http.Handler) http.Handler {
	t.Helper()
	h, err := New(nil, next, cfg, "test")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return h
}

func Test_Rewrite_Root(t *testing.T) {
	cfg := &Config{
		BasePrefix: "/clubs",
	}
	rec := httptest.NewRecorder()

	var gotPath string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})

	mw := newMiddleware(t, cfg, next)
	req := httptest.NewRequest(http.MethodGet, "https://site1.domain.com", nil)
	mw.ServeHTTP(rec, req)

	want := "/clubs/site1"
	if gotPath != want {
		t.Fatalf("rewrite mismatch: got %q want %q", gotPath, want)
	}
}

func Test_Rewrite_SlashNews(t *testing.T) {
	cfg := &Config{
		BasePrefix: "/clubs",
	}
	rec := httptest.NewRecorder()

	var gotPath string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})

	mw := newMiddleware(t, cfg, next)
	req := httptest.NewRequest(http.MethodGet, "https://site1.domain.com/news", nil)
	mw.ServeHTTP(rec, req)

	want := "/clubs/site1/news"
	if gotPath != want {
		t.Fatalf("rewrite mismatch: got %q want %q", gotPath, want)
	}
}

func Test_Rewrite_SubdomainAndDomainWithDash_Root(t *testing.T) {
	cfg := &Config{
		BasePrefix: "/clubs",
	}
	rec := httptest.NewRecorder()

	var gotPath string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})

	mw := newMiddleware(t, cfg, next)
	req := httptest.NewRequest(http.MethodGet, "https://site-1.domain-a.com", nil)
	mw.ServeHTTP(rec, req)

	want := "/clubs/site-1"
	if gotPath != want {
		t.Fatalf("rewrite mismatch: got %q want %q", gotPath, want)
	}
}

func Test_Rewrite_SubdomainAndDomainWithDash_SlashNews(t *testing.T) {
	cfg := &Config{
		BasePrefix: "/clubs",
	}
	rec := httptest.NewRecorder()

	var gotPath string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})

	mw := newMiddleware(t, cfg, next)
	req := httptest.NewRequest(http.MethodGet, "https://site-1.domain-a.com/news", nil)
	mw.ServeHTTP(rec, req)

	want := "/clubs/site-1/news"
	if gotPath != want {
		t.Fatalf("rewrite mismatch: got %q want %q", gotPath, want)
	}
}

func Test_NoDoublePrefix(t *testing.T) {
	cfg := &Config{
		BasePrefix: "/clubs",
	}
	rec := httptest.NewRecorder()

	var gotPath string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})

	mw := newMiddleware(t, cfg, next)
	req := httptest.NewRequest(http.MethodGet, "https://site1.domain.com/clubs/site1/ping", nil)
	mw.ServeHTTP(rec, req)

	want := "/clubs/site1/ping"
	if gotPath != want {
		t.Fatalf("should not double-prefix: got %q want %q", gotPath, want)
	}
}
