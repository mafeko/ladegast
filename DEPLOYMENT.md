# Betrieb von Ladegast

Diese Anleitung ist bewusst generisch: Ladegast selbst kennt keine
bestimmte Plattform (siehe [DESIGN.md](DESIGN.md)). Es gibt im Repository
keine Kubernetes-Manifeste, keine Caddy-Config und keine
Docker-Compose-Datei zum Abschreiben — nur diese Dokumentation. Jede
Umgebung (Kubernetes, ein einzelner Server mit systemd, Docker Compose, …)
setzt lediglich die unten beschriebenen Umgebungsvariablen.

## 1. Binary direkt ausführen

```bash
task build        # baut ./bin/ladegast
export $(grep -v '^#' .env | xargs)   # oder jeden anderen Weg, ENV-Vars zu setzen
./bin/ladegast
```

Pflichtvariablen ohne Default: `LADEGAST_PUBLIC_BASE_URL`,
`LADEGAST_SESSION_SECRET`, `LADEGAST_EVCC_BASE_URL`,
`LADEGAST_EVCC_API_KEY`. Vollständige Referenz: [schema/config.md](schema/config.md).
Beim allerersten Start zusätzlich `LADEGAST_ADMIN_USERNAME` /
`LADEGAST_ADMIN_PASSWORD`, danach kann beides wieder entfernt werden.

## 2. Docker-Image

```bash
task docker:build
docker run --rm -p 8080:8080 \
  -v ladegast-data:/data \
  --env-file .env \
  ladegast:dev
```

- Das Image läuft als non-root (distroless `nonroot`, UID/GID 65532) und
  schreibt nach `/data` (siehe `LADEGAST_DATA_DIR=/data`, im Image
  vorgegeben). Dieses Verzeichnis muss ein beschreibbares Volume sein,
  sonst schlägt der Start fehl ("permission denied" beim Öffnen der
  Datenbankdatei).
- Es gibt keinen eingebetteten Healthcheck; `GET /` antwortet mit 200
  (SPA-Shell) sobald der Prozess läuft und eignet sich als einfacher
  Liveness-Check. `GET {PUBLIC_PATH_PREFIX}/l/<irgendeine-uuid>/status`
  antwortet mit 404 für eine unbekannte UUID, falls ein spezifischerer
  Check gewünscht ist.

## 3. Reverse Proxy / Ingress — die einzige harte Anforderung

Ladegast exponiert intern zwei API-Pfadpräfixe (Default `/public` und
`/admin`, konfigurierbar über `LADEGAST_PUBLIC_PATH_PREFIX` /
`LADEGAST_ADMIN_PATH_PREFIX`) sowie das eingebettete Frontend an der Root
(`/`, `/assets/*`, die Seiten-Routen — siehe DESIGN.md Abschnitt 6).
**Von außen (Internet) darf alles außer `ADMIN_PATH_PREFIX` erreichbar
sein.** Die Admin-Oberfläche selbst liegt unter `ADMIN_PATH_PREFIX` und ist
damit automatisch mitgeschützt, sobald dieser Pfad abgeschnitten wird.
`ADMIN_PATH_PREFIX` ist für den Betreiber selbst gedacht und muss durch
die Infrastruktur abgeschnitten werden — Ladegast selbst erzwingt das
nicht (kein IP-Allowlisting o. Ä. im Code, siehe DESIGN.md Abschnitt 1).

Wie das konkret aussieht, ist Betreiber-Sache. Optionen, ohne Anspruch auf
Vollständigkeit:

- **Separate Services/Ingress-Objekte:** ein öffentlicher Ingress routet
  auf alles außer `ADMIN_PATH_PREFIX`; der Admin-Pfad wird über ein
  zweites, rein clusterinternes Service-Objekt erreicht (z. B. `kubectl
  port-forward` oder ein internes Ingress nur im VPN).
- **Reverse-Proxy-Routing-Regel:** der öffentlich erreichbare Proxy
  (Caddy, nginx, Traefik, …) bekommt eine einzelne Regel, die Pfade unter
  `ADMIN_PATH_PREFIX` mit 404 beantwortet (oder gar nicht erst weiterleitet)
  und alles andere unverändert durchlässt, bevor es überhaupt bei Ladegast
  ankommt.
- **Netzwerk-Ebene:** Ladegast läuft nur in einem Netz, das sowieso nicht
  von außen erreichbar ist (z. B. ein Tailscale-only-Netz), und nur ein
  einzelner externer Reverse Proxy mit der Regel aus dem vorigen Punkt
  bekommt über das VPN Zugriff.

## 4. Beispiel-Szenario (beschrieben, nicht eingecheckt)

Ein mögliches konkretes Setup — Heim-Kubernetes-Cluster, verbunden über
Tailscale mit einem Hetzner-Server, auf dem bereits Caddy läuft:

1. Ladegast läuft als Deployment im Heim-Cluster, ENV-Vars kommen aus
   einem Secret (`LADEGAST_EVCC_API_KEY`, `LADEGAST_SESSION_SECRET`,
   `LADEGAST_ADMIN_PASSWORD`) + einer ConfigMap (alles Nicht-Geheime).
   Persistenz über ein PVC für `LADEGAST_DATA_DIR` (z. B. Longhorn, mit
   Backup z. B. über restic — reine Betriebsentscheidung, von Ladegast
   völlig unabhängig).
2. Ein ClusterIP-Service macht Ladegast im Heimnetz erreichbar. Es gibt
   **kein** Ingress-Objekt für `ADMIN_PATH_PREFIX` an der externen Kante —
   der Admin-Bereich wird nur über das Heimnetz/VPN direkt angesprochen.
3. Auf dem Hetzner-Server bekommt die bestehende Caddy-Instanz einen
   zusätzlichen `reverse_proxy`-Block für die gewünschte Subdomain, der
   über den Tailscale-Tunnel an den ClusterIP-Service weiterleitet - mit
   einer vorgeschalteten Regel, die Pfade unter `ADMIN_PATH_PREFIX`
   ablehnt.
4. EVCC selbst bleibt wie bisher rein clusterintern (kein eigenes
   Ingress); Ladegast erreicht es direkt über `LADEGAST_EVCC_BASE_URL`
   im selben Netz.

Dieses Szenario ist ein Beispiel, keine Vorgabe — jede andere Kombination
aus Reverse Proxy, Orchestrierung und Netzwerk funktioniert genauso, solange
Punkt 3 (`ADMIN_PATH_PREFIX` nicht von außen erreichbar) eingehalten wird.

## 5. Backups

Ladegast selbst hat keine Backup-Logik. Zu sichern ist ausschließlich die
Datei unter `LADEGAST_DATA_DIR` (per Default `ladegast.db`) — sie enthält
alle Freund-/Admin-/Rate-Limit-Daten. Ladedaten selbst liegen bei EVCC und
müssen dort separat gesichert werden.
