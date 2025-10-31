# Traefik Plugin — Add Prefix From Host

**Goal:** behave exactly like Traefik's native `AddPrefix`, but compute the prefix from the request subdomain — e.g.  
With basePrefix = "/clubs" :  
https://site1.domain.com -> upstream path: /clubs/site1  
https://site1.domain.com/news -> upstream path: /clubs/site1/news    
https://site2.domain.com -> upstream path: /clubs/site2  
https://site2.domain.com/pages/1 -> upstream path: /clubs/site2/pages/1  

- **Request-only** rewrite (like `AddPrefix`).
- **No response rewriting** (no Location/Set-Cookie changes) — keeps CSRF/session behavior identical to `AddPrefix`.
- Built-in **double-prefix guard**

## Configuration

### Static
```yaml
experimental:
  plugins:
    addPrefixFromHost:
      moduleName: github.com/subotaii/traefik-plugin-addprefix-from-host
      version: v0.3.0
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
```

### Options

- basePrefix (string, default "/clubs"): base path to prepend.