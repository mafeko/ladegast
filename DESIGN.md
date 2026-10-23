# Ladegast — Architektur-Entwurf

Service, über den Freunde per Einmal-Link + PIN ihr Fahrzeug an der heimischen
Wallbox laden, ohne vollen EVCC-Zugriff zu bekommen. Verifiziert gegen eine
reale EVCC-Instanz (siehe Abschnitt "EVCC-Integration").

**Grundprinzip:** Ladegast ist eine eigenständige, deployment-agnostische
Anwendung (ein Go-Binary + eigenständiges Vue-Frontend, siehe `web/`). Sie
weiß nichts von Kubernetes, Tailscale oder Caddy — alles Umgebungsspezifische
kommt über Umgebungsvariablen rein. Wie eine konkrete Umgebung das Binary
oder Docker-Image betreibt, ist Betreiber-Sache und bewusst nur
dokumentiert, nicht als Beispiel-Manifeste im Repo (siehe
[DEPLOYMENT.md](DEPLOYMENT.md)).

## 0. Keine privaten Daten im Repository

- Reale Domains, Freundesnamen, Fahrzeugnamen, API-Keys, PINs, Tarife etc.
  landen **nie** im Git-Verlauf.
- Jede Konfigurationsdatei mit echten Werten (`.env`, `config.local.*`,
  Secrets) ist in `.gitignore` gelistet.
- Für jede solche Datei existiert ein versioniertes Template mit
  Platzhaltern, z. B. `.env.example`.
- Dokumentation/Beispiele nutzen ausschließlich generische Platzhalter
  (`example.com`, `vehicle_1`, `friend-name`) — nie echte Daten aus diesem
  Projekt.

```gitignore
.env
.env.*
!.env.example
config.local.*
*.secrets.*
```

## 1. Komponentenübersicht (generisch)

```
          beliebiger Reverse Proxy / Ingress
          (siehe DEPLOYMENT.md für ein dokumentiertes Beispiel)
                       │
                       │  leitet nur PUBLIC_PATH_PREFIX weiter
                       ▼
          ┌─────────────────────────────────┐
          │  ladegast (1 Binary/Prozess)     │
          │  Go-Backend + Vue-Frontend       │
          │  konfiguriert nur über ENV-Vars  │
          │  Storage: Datei unter DATA_DIR   │
          │  Admin-Bereich: ADMIN_PATH_PREFIX│
          └───────────────┬──────────────────┘
                           │ HTTP + Bearer-API-Key
                           ▼
                    EVCC-Instanz
             (Adresse/Key kommen aus ENV)
```

Wie der Reverse Proxy, wie Ladegast erreicht wird (K8s-Service, systemd auf
einer VM, Docker Compose, …) und wie es zur EVCC-Instanz kommt (gleiches
Netz, VPN, …) ist austauschbar und für den Kern der Anwendung irrelevant.
Die einzige Kopplung: Ladegast muss den EVCC-Endpoint per HTTP erreichen und
einen gültigen API-Key haben.

**Netzwerk-Trennung Admin vs. Public ist Betreiber-Verantwortung:** Ladegast
exponiert intern zwei Pfadpräfixe (`PUBLIC_PATH_PREFIX`, `ADMIN_PATH_PREFIX`,
per ENV konfigurierbar, Default z. B. `/public` und `/admin`). Ob und wie der
Admin-Pfad von außen abgeschnitten wird (kein Ingress-Rule, separates
internes Service-Objekt, Firewall-Regel, VPN-only-Netz, Caddy-Proxy, der nur
`/public/*` weiterleitet …) entscheidet die jeweilige Deployment-Umgebung.
[DEPLOYMENT.md](DEPLOYMENT.md) beschreibt eine Variante davon als Beispiel.

## 2. Konfiguration (Umgebungsvariablen)

Alle Einstellungen kommen über ENV-Vars, mit sinnvollen Defaults für die
lokale Entwicklung (keine Pflicht-Secrets außer EVCC-Zugang). Vollständige
Liste + Typen in `schema/config.md` (bzw. maschinenlesbar in
`schema/config.schema.json`, falls gewünscht) — hier die Kern-Variablen:

| Variable                         | Zweck                                             | Beispiel/Default        |
|-----------------------------------|----------------------------------------------------|--------------------------|
| `LADEGAST_LISTEN_ADDR`            | Bind-Adresse HTTP-Server                           | `:8080`                  |
| `LADEGAST_PUBLIC_BASE_URL`        | Basis-URL, unter der Ladegast von außen erreichbar ist (für Link-Generierung) | *(muss gesetzt werden, keine Annahme)* |
| `LADEGAST_PUBLIC_PATH_PREFIX`     | Pfadpräfix für Freund-Routen                       | `/public`                |
| `LADEGAST_ADMIN_PATH_PREFIX`      | Pfadpräfix für Admin-Routen                        | `/admin`                 |
| `LADEGAST_DATA_DIR`               | Verzeichnis für die bbolt-Datei                     | `./data`                 |
| `LADEGAST_SESSION_SECRET`         | Signierschlüssel für Session-Cookies                | *(Pflicht, kein Default)*|
| `LADEGAST_EVCC_BASE_URL`          | Basis-URL der EVCC-Instanz                          | *(Pflicht)*              |
| `LADEGAST_EVCC_API_KEY`           | Bearer-Token für EVCC                               | *(Pflicht, Secret)*      |
| `LADEGAST_ADMIN_USERNAME`         | Initialer Admin-Benutzername                        | *(Pflicht beim ersten Start)* |
| `LADEGAST_ADMIN_PASSWORD`         | Initiales Admin-Passwort (wird beim ersten Start gehasht und verworfen) | *(Pflicht beim ersten Start)* |
| `LADEGAST_RATE_LIMIT_MAX_ATTEMPTS`| Fehlversuche PIN-Login bis Lockout                  | `5`                      |
| `LADEGAST_RATE_LIMIT_WINDOW`      | Zeitfenster für Rate-Limit                          | `15m`                    |

- `.env.example` im Repo-Root dokumentiert alle Variablen mit Platzhaltern;
  echte Werte liegen in `.env` (gitignored) oder werden vom jeweiligen
  Deployment injiziert (K8s Secret/ConfigMap, systemd `EnvironmentFile`,
  Docker `--env-file`, …).
- Kein Code verweist hart auf einen bestimmten Deployment-Mechanismus —
  `os.Getenv`/eine kleine Config-Struct im Go-Code genügt.

## 3. EVCC-Integration (verifiziert)

- `GET /api/sessions?year=&month=` liefert pro Ladevorgang u. a.:
  `created, finished, loadpoint, vehicle, chargedEnergy, solarPercentage,
  price, pricePerKWh`. `solarPercentage` ist real befüllt (Live-Check: 0 % bis
  98 % je Session).
- Kein direktes Grid-Feld — berechnet sich aus:
  `solar_kwh = chargedEnergy * solarPercentage / 100`
  `grid_kwh  = chargedEnergy - solar_kwh`
- Preis-Felder von EVCC (`price`, `pricePerKWh`) basieren auf EVCC's eigenem
  Tarif und werden **ignoriert** — Ladegast rechnet mit eigenen,
  pro-Freund-konfigurierbaren Sätzen (`rate_solar_per_kwh`,
  `rate_grid_per_kwh`, in der Friend-Entität gespeichert, nicht per ENV).
- Modi (diese Version): `off`, `smart` (PV-Überschuss), `now` (Schnellladen).
  Für MVP reichen `smart` ↔ `now`.
- Mode-Switch: `POST /api/vehicles/{name}/mode/{mode}`, wobei `{name}` der
  **Config-Name** aus der EVCC-Konfiguration ist (nicht `db:XX`, nicht der
  Titel) — vor Implementierung einmal gegen die reale Config prüfen.
- Auth: Long-lived API-Key (`evcc_...`, Bearer), verwaltet über
  `/auth/apikey`. Ladegast hält genau einen solchen Key (`LADEGAST_EVCC_API_KEY`).
- Kein eigener Datenspeicher für Ladedaten nötig — Sessions werden bei jedem
  Dashboard-Request live von EVCC gezogen und aggregiert (Tag/Woche/
  Monat/Jahr). Nur Ladegast-eigene Entitäten werden persistiert (s. u.).

## 4. Datenmodell (bbolt, Go-Structs)

```go
type Friend struct {
    ID                string    // uuid, auch Link-Slug
    Name              string
    VehicleName       string    // EVCC-Config-Name, z.B. "vehicle_1"
    PINHash           string    // bcrypt
    RateSolarPerKWh   float64
    RateGridPerKWh    float64
    CreatedAt         time.Time
    LinkRegeneratedAt time.Time
}

type Admin struct {
    Username     string
    PasswordHash string
}

type FailedLoginAttempt struct {
    IP        string
    FriendID  string
    Timestamp time.Time
}
```

Buckets: `friends`, `admins`, `rate_limit`. Ladedaten selbst: kein Bucket,
nur Durchleitung/Aggregation aus EVCC. Speicherort: `LADEGAST_DATA_DIR`
(austauschbar: lokales Verzeichnis, gemountetes Volume, Netzwerk-Share — für
den Code irrelevant, solange der Pfad beschreibbar ist).

## 5. Auth-Flow (Freund)

1. Freund öffnet `{PUBLIC_BASE_URL}{PUBLIC_PATH_PREFIX}/l/{uuid}`.
2. Seite fragt PIN ab. POST an `.../l/{uuid}/login` mit PIN.
3. Backend: IP-Rate-Limit prüfen (`LADEGAST_RATE_LIMIT_*`), PIN-Hash
   vergleichen.
4. Bei Erfolg: httpOnly-Session-Cookie (kurze TTL, z. B. 7 Tage, kein
   Passwort-Reset-Flow nötig da Link dauerhaft gültig bleibt), signiert mit
   `LADEGAST_SESSION_SECRET`.
5. Alle weiteren `.../l/{uuid}/*`-Requests prüfen Cookie gegen Friend-ID.

**Admin-Login:** klassisch Username/Passwort (bcrypt) + Session-Cookie,
ausschließlich über `{ADMIN_PATH_PREFIX}/login`.

## 6. API-Routen

Präfixe sind konfigurierbar (`LADEGAST_PUBLIC_PATH_PREFIX` /
`LADEGAST_ADMIN_PATH_PREFIX`); Defaults `/public` und `/admin` unten
verwendet.

### Public (`/public/*`)
```
GET  /public/l/{uuid}              Login-Seite
POST /public/l/{uuid}/login         { pin } → Cookie
GET  /public/l/{uuid}/dashboard     Übersicht (Tag/Woche/Monat/Jahr)
GET  /public/l/{uuid}/sessions      Session-Liste + Kosten
POST /public/l/{uuid}/mode          { mode: "smart"|"now" }
```

### Admin (`/admin/*`)
```
GET    /admin/login
POST   /admin/login
GET    /admin/friends
POST   /admin/friends                 neuen Freund anlegen (generiert UUID+PIN)
PUT    /admin/friends/{id}            Tarife/Name/Fahrzeug ändern
POST   /admin/friends/{id}/regenerate-link   neue UUID, alte ungültig
POST   /admin/friends/{id}/reset-pin         neue PIN setzen/anzeigen
GET    /admin/friends/{id}/view-as           identische Ansicht wie /public/l/{uuid}/dashboard,
                                              ohne erneuten PIN (Admin-Session reicht)
```

## 7. Dashboard-Aggregation

Pro Request: `GET /api/sessions` von EVCC holen, nach `vehicle ==
friend.VehicleName` filtern (serverseitig, Freund sieht nie fremde
Fahrzeuge), dann gruppieren:

- Tag: `created` auf Kalendertag runden
- Woche/Monat/Jahr: analog, Summen über `chargedEnergy`, `solar_kwh`,
  `grid_kwh`, Kosten = `solar_kwh * RateSolarPerKWh + grid_kwh *
  RateGridPerKWh`

Kein Caching nötig für MVP (Session-Liste ist klein); später ggf.
In-Memory-Cache mit kurzer TTL, falls EVCC-Calls spürbar werden.

## 8. Repository-Layout (Vorschlag)

```
/                       Go-Module-Root, main.go, internal/...
  internal/config/      ENV-Parsing → Config-Struct, mit Defaults & Validierung
  internal/evcc/        EVCC-Client (sessions, vehicle-mode)
  internal/store/       bbolt-Zugriff (Friend, Admin, RateLimit)
  internal/httpserver/  Router, Public- und Admin-Handler
  web/                  Eigenständiges Vue-Frontend (Vite, eigenes package.json)
schema/
  config.md             Vollständige ENV-Var-Referenz (generisch, Typen, Defaults)
.env.example             Platzhalter für alle ENV-Vars
.gitignore               .env, config.local.*, *.secrets.*
Dockerfile               Multi-Stage-Build des Go-Binaries
DEPLOYMENT.md            Betriebsdokumentation (kein Code/Manifeste)
```

`schema/` dokumentiert, was zu bauen ist (Konfigurationsvertrag,
API-Shape) unabhängig vom Deployment.

## 9. Deployment

Bewusst **nur dokumentiert**, nicht als Beispiel-Manifeste im Repo — siehe
[DEPLOYMENT.md](DEPLOYMENT.md) für die Betriebsanleitung (Binary/Docker
ausführen, benötigte ENV-Vars, Anforderungen an den Reverse Proxy, und ein
beschriebenes, aber nicht eingechecktes Beispiel mit Kubernetes/Caddy/
Tailscale).

## 10. Offene Punkte / Risiken

- `{name}`-Format für `/api/vehicles/{name}/mode/{mode}` vor dem ersten
  Schreibtest gegen die echte EVCC-Konfiguration verifizieren (bisher nur
  Lesezugriffe erlaubt/getestet).
- Was passiert, wenn ein Freund `now` setzt und "vergisst", zurück auf
  `smart` zu schalten? (Timeout/Auto-Reset nach X Stunden? Für MVP evtl.
  bewusst weglassen und später nachrüsten.)
- Mehrere Freunde mit gleichem Fahrzeug-Namen in EVCC (Edge Case, aktuell
  nicht vorgesehen — 1 Freund : 1 Fahrzeug fix).
- Backup/Recovery der Datenbank-Datei testen, bevor produktiv genutzt
  (unabhängig vom konkreten Storage-Backend).
