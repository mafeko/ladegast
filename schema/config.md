# Konfigurations-Referenz

Ladegast wird ausschliesslich ueber Umgebungsvariablen konfiguriert. Keine
Konfigurationsdatei mit echten Werten gehoert ins Repository (siehe
`.env.example` als Template).

| Variable                          | Typ      | Pflicht | Default   | Beschreibung |
|------------------------------------|----------|---------|-----------|--------------|
| `LADEGAST_LISTEN_ADDR`              | string   | nein    | `:8080`   | Bind-Adresse des HTTP-Servers |
| `LADEGAST_PUBLIC_BASE_URL`           | URL      | ja      | –         | Basis-URL, unter der Ladegast von aussen erreichbar ist (fuer Link-Generierung) |
| `LADEGAST_PUBLIC_PATH_PREFIX`        | string   | nein    | `/public` | Pfadpraefix fuer Freund-Routen |
| `LADEGAST_ADMIN_PATH_PREFIX`         | string   | nein    | `/admin`  | Pfadpraefix fuer Admin-Routen |
| `LADEGAST_DATA_DIR`                  | path     | nein    | `./data`  | Verzeichnis fuer die bbolt-Datenbankdatei |
| `LADEGAST_SESSION_SECRET`            | string   | ja      | –         | Signierschluessel fuer Session-Cookies (min. 32 Byte zufaellig) |
| `LADEGAST_EVCC_BASE_URL`             | URL      | ja      | –         | Basis-URL der EVCC-Instanz |
| `LADEGAST_EVCC_API_KEY`              | string   | ja      | –         | Bearer-Token fuer die EVCC-REST-API |
| `LADEGAST_ADMIN_USERNAME`            | string   | nein¹   | –         | Admin-Benutzername, nur beim allerersten Start verwendet |
| `LADEGAST_ADMIN_PASSWORD`            | string   | nein¹   | –         | Admin-Passwort im Klartext, wird beim ersten Start gehasht in die DB uebernommen |
| `LADEGAST_RATE_LIMIT_MAX_ATTEMPTS`   | int      | nein    | `5`       | Fehlversuche PIN-Login bis Lockout |
| `LADEGAST_RATE_LIMIT_WINDOW`         | duration | nein    | `15m`     | Zeitfenster fuer das Rate-Limit |

¹ Pflicht nur, wenn noch kein Admin-Account in der Datenbank existiert.

## Deployment-Unabhaengigkeit

Diese Liste ist die einzige Schnittstelle zwischen Ladegast und seiner
Umgebung. Jede Deployment-Variante (Kubernetes, systemd, Docker, lokal)
muss nur diese Variablen setzen — der Rest ist Betreiber-Sache. Beispiele
dazu liegen in `deploy/reference/`.
