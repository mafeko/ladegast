# Getting started (for EVCC owners)

Goal: your guests open a personal link on their phone, enter a PIN, and can see their own charging sessions, solar/grid share and cost (charge mode switching is on the [roadmap](README.md#roadmap)). They never see your EVCC.

Time: ~30 minutes. You need an EVCC instance with charging sessions enabled and a machine at home that is always on (Raspberry Pi, NAS, mini PC, or the EVCC host itself) running Docker.

## 1. What you need

| Requirement | Why |
|---|---|
| EVCC with a working wallbox | Ladegast reads sessions from it and switches charge mode |
| EVCC API key | Lets Ladegast talk to EVCC without your admin password |
| Vehicles configured in EVCC | Each guest is linked to one EVCC vehicle |
| Docker on an always-on machine | Easiest way to run Ladegast |
| A way to make Ladegast reachable from the internet | See [HOSTING.md](HOSTING.md). **No own server needed.** |

## 2. Create an EVCC API key

Ladegast authenticates to EVCC with a Bearer token (`evcc_...`). EVCC generates it via its REST API (`POST /auth/apikey`, requires your EVCC admin password; the key is shown only once, so save it). See the [EVCC REST API docs](https://docs.evcc.io/integrations/rest-api) for the exact call and the current way to do it in your EVCC version.

Check it works (replace address and key):

```bash
curl -H "Authorization: Bearer evcc_xxx" "http://<evcc-host>:7070/api/sessions?year=2026&month=11"
```

You should get a JSON list of charging sessions.

## 3. Run Ladegast

Create a folder with `.env` and `compose.yaml`.

`.env` (full reference: [schema/config.md](schema/config.md)):

```bash
LADEGAST_PUBLIC_BASE_URL=https://ladegast.example.com   # the public address from step 4
LADEGAST_EVCC_BASE_URL=http://<evcc-host>:7070          # EVCC address as seen from Ladegast
LADEGAST_EVCC_API_KEY=evcc_xxx
LADEGAST_SESSION_SECRET=<output of: openssl rand -hex 32>
LADEGAST_ADMIN_USERNAME=admin                           # first start only
LADEGAST_ADMIN_PASSWORD=<long random password>          # first start only
```

`compose.yaml`:

```yaml
services:
  ladegast:
    image: ghcr.io/mafeko/ladegast:latest
    env_file: .env
    ports: [ "8080:8080" ]
    volumes: [ "ladegast-data:/data" ]
    restart: unless-stopped
volumes:
  ladegast-data:
```

```bash
docker compose up -d
```

Open `http://<machine>:8080/admin` in your home network and log in. After the first start you may remove `LADEGAST_ADMIN_USERNAME/PASSWORD` from `.env`.

> Tip: if EVCC runs in Docker on the same machine, put both in one Compose network and use `http://evcc:7070`.

## 4. Make it reachable for guests

Follow [HOSTING.md](HOSTING.md). Short version: use Cloudflare Tunnel (own domain) or Tailscale Funnel (no domain), then set `LADEGAST_PUBLIC_BASE_URL` to the resulting `https://` address and restart Ladegast (`docker compose up -d`). `/admin` stays local-only automatically: open it from your home network or VPN.

## 5. Add your first guest

1. In the admin area click to create a guest: name, EVCC vehicle, solar and grid price per kWh, PIN (or let Ladegast generate one).
2. The vehicle field suggests EVCC vehicles that have no guest yet. A vehicle appears there only after its first charging session; otherwise type its EVCC name exactly as in `evcc.yaml`.
3. Send the guest the personal link and the PIN, ideally in two separate messages.
4. Test with your own phone on mobile data (not your Wi-Fi).

If a link leaks or a guest should lose access: regenerate the link, reset the PIN or delete the guest. The old link stops working immediately.

## Troubleshooting

| Symptom | Check |
|---|---|
| Guest link opens a blank page or wrong host | `LADEGAST_PUBLIC_BASE_URL` wrong or has trailing path; restart after changing, regenerate links |
| Admin shows no vehicles / sessions | EVCC address/API key wrong (see the curl test above); sessions only exist after the first charge |
| Container exits with "permission denied" | `/data` must be a writable volume |
| PIN screen says too many attempts | Rate limit (default 5 per 15 min per IP); wait or adjust `LADEGAST_RATE_LIMIT_*` |
| Works at home, not on mobile data | Tunnel not running or points to the wrong host/port, see [HOSTING.md](HOSTING.md) |
| Guest cost differs from EVCC | Intended: cost uses the guest's own solar/grid rates |

## Updating and backups

```bash
docker compose pull && docker compose up -d
```

Back up the `ladegast-data` volume (`ladegast.db`: guests, admin, settings). Charging data stays in EVCC.
