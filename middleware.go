package traefik_plugin_addprefix_from_host

import (
	"context"
	"net/http"
	"regexp"
	"strings"
)

// Config holds the plugin configuration.
type Config struct {
	// BasePrefix is the static base to prepend (e.g., "/clubs").
	BasePrefix string `json:"basePrefix,omitempty"`

	// DomainSuffix narrows host matching to "*.domain.tld".
	// Example: "domain.com". If empty, HostPattern is used instead.
	DomainSuffix string `json:"domainSuffix,omitempty"`

	// HostPattern is an optional explicit regex to capture the "site" in group 1.
	// Example: "^([a-z0-9-]+)\\.domain\\.com$"
	HostPattern string `json:"hostPattern,omitempty"`

	// BypassPathPrefixes will not be rewritten (e.g., "/healthz", "/metrics").
	BypassPathPrefixes []string `json:"bypassPathPrefixes,omitempty"`

	// AllowRootHost allows rewriting when the host equals DomainSuffix (no subdomain).
	// If false, requests to the bare domain are not rewritten.
	AllowRootHost bool `json:"allowRootHost,omitempty"`
}

// CreateConfig initializes the default plugin configuration.
// Yaegi expects this symbol to be exported from the package whose name
// is the repo name with dashes replaced by underscores.
func CreateConfig() *Config {
	return &Config{
		BasePrefix:         "/clubs",
		BypassPathPrefixes: []string{"/healthz", "/metrics"},
	}
}

// addPrefixFromHost is the middleware instance.
type addPrefixFromHost struct {
	next           http.Handler
	basePrefix     string
	bypassPrefixes []string
	hostRe         *regexp.Regexp
	domainSuffix   string
	allowRootHost  bool
}

// New creates a new middleware.
// Signature must match Traefik plugin expectations.
func New(ctx context.Context, next http.Handler, cfg *Config, name string) (http.Handler, error) {
	// Avoid unused warnings in some build contexts.
	_ = ctx
	_ = name

	var hostRe *regexp.Regexp
	if cfg.HostPattern != "" {
		hostRe = regexp.MustCompile(cfg.HostPattern)
	}

	return &addPrefixFromHost{
		next:           next,
		basePrefix:     ensureLeadingSlash(cfg.BasePrefix),
		bypassPrefixes: append([]string{}, cfg.BypassPathPrefixes...),
		hostRe:         hostRe,
		domainSuffix:   strings.ToLower(cfg.DomainSuffix),
		allowRootHost:  cfg.AllowRootHost,
	}, nil
}

func (m *addPrefixFromHost) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Bypass explicit paths (cheap and deterministic).
	for _, p := range m.bypassPrefixes {
		if p != "" && strings.HasPrefix(req.URL.Path, p) {
			m.next.ServeHTTP(rw, req)
			return
		}
	}

	host := req.Host
	if i := strings.IndexByte(host, ':'); i >= 0 {
		host = host[:i]
	}
	host = strings.ToLower(host)

	sub, ok := m.extractSubdomain(host)
	if !ok || sub == "" {
		// No subdomain or not matching -> no rewrite.
		m.next.ServeHTTP(rw, req)
		return
	}

	// Guard: already prefixed? (avoid double-prefix)
	expected := m.basePrefix + "/" + sub
	if strings.HasPrefix(req.URL.Path, expected) {
		m.next.ServeHTTP(rw, req)
		return
	}

	// AddPrefix behavior: prepend to upstream path (request only).
	req.URL.Path = singleSlashJoin(expected, req.URL.Path)
	if req.URL.RawPath != "" {
		req.URL.RawPath = singleSlashJoin(expected, req.URL.RawPath)
	}

	m.next.ServeHTTP(rw, req)
}

func (m *addPrefixFromHost) extractSubdomain(host string) (string, bool) {
	// 1) Prefer explicit regex.
	if m.hostRe != nil {
		if matches := m.hostRe.FindStringSubmatch(host); len(matches) >= 2 {
			return matches[1], true
		}
		return "", false
	}

	// 2) Otherwise rely on domain suffix.
	if m.domainSuffix == "" {
		return "", false
	}
	if host == m.domainSuffix {
		return "", m.allowRootHost
	}
	if !strings.HasSuffix(host, "."+m.domainSuffix) {
		return "", false
	}
	// "site1.domain.com" -> "site1" (leftmost label of the remainder)
	rest := strings.TrimSuffix(host, "."+m.domainSuffix)
	if rest == "" {
		return "", false
	}
	parts := strings.Split(rest, ".")
	return parts[0], true
}

func ensureLeadingSlash(s string) string {
	if s == "" {
		return "/"
	}
	if !strings.HasPrefix(s, "/") {
		return "/" + s
	}
	return s
}

func singleSlashJoin(a, b string) string {
	switch {
	case a == "" && b == "":
		return "/"
	case a == "":
		return ensureLeadingSlash(b)
	case b == "":
		return a
	default:
		aslash := strings.HasSuffix(a, "/")
		bslash := strings.HasPrefix(b, "/")
		switch {
		case aslash && bslash:
			return a + b[1:]
		case !aslash && !bslash:
			return a + "/" + b
		default:
			return a + b
		}
	}
}
