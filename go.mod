module github.com/subotaii/traefik-plugin-addprefix-from-host

go 1.22

require (
    // Pas de SDK spécial nécessaire : un plugin Traefik est un http.Handler.
    // Garder une dépendance clean aide au chargement dynamique.
)
