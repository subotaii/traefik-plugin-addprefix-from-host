# Traefik Plugin — Add Prefix From Host

**Goal:** behave exactly like Traefik's native `AddPrefix`, but compute the prefix from the request subdomain — e.g.

https://site1.domain.com/news -> upstream path: /clubs/site1/news  
https://site2.domain.com -> upstream path: /clubs/site2

- **Request-only** rewrite (like `AddPrefix`).
- **No response rewriting** (no Location/Set-Cookie changes) — keeps CSRF/session behavior identical to `AddPrefix`.
- Built-in **double-prefix guard** and **bypass paths**.

## Configuration

### Static
```yaml
experimental:
  plugins:
    addPrefixFromHost:
      moduleName: github.com/subotaii/traefik-plugin-addprefix-from-host
      version: v0.1.0
```

### Dynamic

```yaml
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: addprefix-from-host
  namespace: your-namespace
spec:
  plugin:
    addPrefixFromHost:
      basePrefix: "/clubs"
      domainSuffix: "domain.com"
      # alternatively:
      # hostPattern: "^([a-z0-9-]+)\\.domain\\.com$"
      bypassPathPrefixes:
        - "/healthz"
        - "/metrics"
      allowRootHost: false
```

### Options

- basePrefix (string, default "/clubs"): base path to prepend.
- domainSuffix (string): limit to hosts *.domainSuffix (e.g., domain.com).
- hostPattern (regex): alternative to domainSuffix; must capture the site in group 1.
- bypassPathPrefixes ([]string): paths left untouched (e.g., /healthz, /metrics).
- allowRootHost (bool): if true, also rewrite bare host (domainSuffix with no subdomain).