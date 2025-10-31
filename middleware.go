package traefik_plugin_addprefix_from_host

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

const (
	typeName = "AddPrefixFromHost"
)

// Config defines the plugin configuration.
type Config struct {
	// BasePrefix is the static part to prepend, e.g. "/clubs".
	BasePrefix string `json:"basePrefix,omitempty"`
}

// CreateConfig returns the default config.
func CreateConfig() *Config {
	return &Config{
		BasePrefix: "/clubs",
	}
}

// addPrefixFromHost is the middleware used to add "/<basePrefix>/<subdomain>" to the request path.
type addPrefixFromHost struct {
	next       http.Handler
	basePrefix string
	name       string
}

// New creates a new middleware handler.
func New(_ context.Context, next http.Handler, cfg *Config, name string) (http.Handler, error) {
	if cfg.BasePrefix == "" {
		return nil, errors.New("basePrefix cannot be empty")
	}

	return &addPrefixFromHost{
		next:       next,
		basePrefix: cfg.BasePrefix,
		name:       name,
	}, nil
}

// ServeHTTP adds the prefix + subdomain to the request URL path.
func (a *addPrefixFromHost) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// 1) extract host without port
	host := req.Host
	if i := strings.IndexByte(host, ':'); i >= 0 {
		host = host[:i]
	}

	// 2) extract subdomain (before first dot)
	var subdomain string
	if p := strings.IndexByte(host, '.'); p > 0 {
		subdomain = host[:p]
	}

	// no subdomain -> do nothing
	if subdomain == "" {
		a.next.ServeHTTP(rw, req)
		return
	}

	// 3) build full prefix: /clubs/<subdomain>
	fullPrefix := ensureLeadingSlash(a.basePrefix)
	if !strings.HasSuffix(fullPrefix, "/") {
		fullPrefix += "/"
	}
	fullPrefix += subdomain

	// avoid double prefix
	if strings.HasPrefix(req.URL.Path, fullPrefix) {
		a.next.ServeHTTP(rw, req)
		return
	}

	// 4) rewrite path (like native AddPrefix)
	req.URL.Path = ensureLeadingSlash(fullPrefix + req.URL.Path)

	// rewrite RawPath too if present
	if req.URL.RawPath != "" {
		req.URL.RawPath = ensureLeadingSlash(fullPrefix + req.URL.RawPath)
	}

	// keep RequestURI consistent
	req.RequestURI = req.URL.RequestURI()

	// 5) forward
	a.next.ServeHTTP(rw, req)
}

// ensureLeadingSlash ensures the string starts with "/".
func ensureLeadingSlash(str string) string {
	if str == "" {
		return str
	}
	if str[0] == '/' {
		return str
	}
	return "/" + str
}
