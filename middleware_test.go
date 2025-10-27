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

func Test_NoRewrite_WhenBypassPath(t *testing.T) {
	cfg := &Config{
		BasePrefix:         "/clubs",
		DomainSuffix:       "domain.com",
		BypassPathPrefixes: []string{"/healthz"},
	}
	rec := httptest.NewRecorder()

	var gotPath string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})

	mw := newMiddleware(t, cfg, next)
	req := httptest.NewRequest(http.MethodGet, "https://site1.domain.com/healthz", nil)
	mw.ServeHTTP(rec, req)

	if gotPath != "/healthz" {
		t.Fatalf("expected no rewrite for bypass path, got %q", gotPath)
	}
}

func Test_Rewrite_FromSubdomain(t *testing.T) {
	cfg := &Config{
		BasePrefix:   "/clubs",
		DomainSuffix: "domain.com",
	}
	rec := httptest.NewRecorder()

	var gotPath string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})

	mw := newMiddleware(t, cfg, next)
	req := httptest.NewRequest(http.MethodGet, "https://site2.domain.com/news", nil)
	mw.ServeHTTP(rec, req)

	want := "/clubs/site2/news"
	if gotPath != want {
		t.Fatalf("rewrite mismatch: got %q want %q", gotPath, want)
	}
}

func Test_Rewrite_FromSubdomainDash(t *testing.T) {
	cfg := &Config{
		BasePrefix:   "/clubs",
		DomainSuffix: "domain.com",
	}
	rec := httptest.NewRecorder()

	var gotPath string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})

	mw := newMiddleware(t, cfg, next)
	req := httptest.NewRequest(http.MethodGet, "https://site-dash.domain.com/news", nil)
	mw.ServeHTTP(rec, req)

	want := "/clubs/site-dash/news"
	if gotPath != want {
		t.Fatalf("rewrite mismatch: got %q want %q", gotPath, want)
	}
}

func Test_NoDoublePrefix(t *testing.T) {
	cfg := &Config{
		BasePrefix:   "/clubs",
		DomainSuffix: "domain.com",
	}
	rec := httptest.NewRecorder()

	var gotPath string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})

	mw := newMiddleware(t, cfg, next)
	req := httptest.NewRequest(http.MethodGet, "https://site3.domain.com/clubs/site3/ping", nil)
	mw.ServeHTTP(rec, req)

	want := "/clubs/site3/ping"
	if gotPath != want {
		t.Fatalf("should not double-prefix: got %q want %q", gotPath, want)
	}
}

func Test_HostPattern_Regex(t *testing.T) {
	cfg := &Config{
		BasePrefix:  "/clubs",
		HostPattern: "^([a-z0-9-]+)\\.domain\\.com$",
		// DomainSuffix intentionally empty so HostPattern is used
	}
	rec := httptest.NewRecorder()

	var gotPath string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})

	mw := newMiddleware(t, cfg, next)
	req := httptest.NewRequest(http.MethodGet, "https://alpha.domain.com/checkout", nil)
	mw.ServeHTTP(rec, req)

	want := "/clubs/alpha/checkout"
	if gotPath != want {
		t.Fatalf("regex capture mismatch: got %q want %q", gotPath, want)
	}
}

func Test_HostPattern_Regex_Dash(t *testing.T) {
	cfg := &Config{
		BasePrefix:  "/clubs",
		HostPattern: "^([a-z0-9-]+)\\.domain\\.com$",
		// DomainSuffix intentionally empty so HostPattern is used
	}
	rec := httptest.NewRecorder()

	var gotPath string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})

	mw := newMiddleware(t, cfg, next)
	req := httptest.NewRequest(http.MethodGet, "https://alpha-dash.domain.com/checkout", nil)
	mw.ServeHTTP(rec, req)

	want := "/clubs/alpha-dash/checkout"
	if gotPath != want {
		t.Fatalf("regex capture mismatch: got %q want %q", gotPath, want)
	}
}

func Test_RootHost_NoRewrite_ByDefault(t *testing.T) {
	cfg := &Config{
		BasePrefix:   "/clubs",
		DomainSuffix: "domain.com",
		// allowRootHost defaults to false
	}
	rec := httptest.NewRecorder()

	var gotPath string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})

	mw := newMiddleware(t, cfg, next)
	req := httptest.NewRequest(http.MethodGet, "https://domain.com/", nil)
	mw.ServeHTTP(rec, req)

	if gotPath != "/" {
		t.Fatalf("root host should not be rewritten by default, got %q", gotPath)
	}
}
