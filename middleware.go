package traefik_plugin_addprefix_from_host

import (
	"context"
	"net/http"
	"strings"
)

// Config is the configuration for the middleware.
type Config struct {
	// BasePrefix is the static prefix to add before the subdomain.
	// Example: "/clubs"
	BasePrefix string `json:"basePrefix,omitempty"`
}

// CreateConfig creates the default configuration.
func CreateConfig() *Config {
	return &Config{
		BasePrefix: "/clubs",
	}
}

// New creates a new middleware handler.
func New(_ context.Context, next http.Handler, cfg *Config, _ string) (http.Handler, error) {
	prefix := cfg.BasePrefix
	if prefix == "" {
		prefix = "/"
	}
	// normalize once
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}

	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// 1) extraire le host (sans port)
		host := req.Host
		if i := strings.IndexByte(host, ':'); i >= 0 {
			host = host[:i]
		}

		// 2) extraire le sous-domaine: on prend tout avant le premier "."
		//    ex: "site1.domain.com" -> "site1"
		subdomain := ""
		if p := strings.IndexByte(host, '.'); p > 0 {
			subdomain = host[:p]
		}

		// si pas de sous-domaine (host sans point) -> on ne fait rien
		if subdomain == "" {
			next.ServeHTTP(rw, req)
			return
		}

		// préfixe complet = /clubs/<subdomain>
		fullPrefix := prefix
		if !strings.HasSuffix(fullPrefix, "/") {
			fullPrefix += "/"
		}
		fullPrefix += subdomain

		// ne pas préfixer si c'est déjà le cas
		if strings.HasPrefix(req.URL.Path, fullPrefix) {
			next.ServeHTTP(rw, req)
			return
		}

		// comportement addPrefix: on préfixe le path (et le RawPath si présent)
		req.URL.Path = joinPath(fullPrefix, req.URL.Path)
		if req.URL.RawPath != "" {
			req.URL.RawPath = joinPath(fullPrefix, req.URL.RawPath)
		}

		next.ServeHTTP(rw, req)
	}), nil
}

// joinPath est la même logique que dans addPrefix natif (éviter les //)
func joinPath(a, b string) string {
	switch {
	case a == "" && b == "":
		return "/"
	case a == "":
		if strings.HasPrefix(b, "/") {
			return b
		}
		return "/" + b
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
