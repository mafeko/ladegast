# Ladegast

Let friends charge their car at your home wallbox without giving them access to your [EVCC](https://evcc.io) instance.

## Goal

A home-wallbox owner runs EVCC. Friends occasionally charge there. Ladegast sits in front of EVCC and gives each friend a minimal, private view:

- **Charge control:** switch the charging mode between `smart` (solar surplus) and `now` (fast charge).
- **Usage insight:** sessions, energy (solar vs. grid), and cost over a selectable time range.
- **Own tariffs:** cost is computed with per-friend solar/grid rates, not EVCC's tariff.

The owner (admin) manages friends and tariffs. Friends never see EVCC itself.

## How it works

1. Admin creates a friend: name, EVCC vehicle, rates, PIN.
2. Friend receives a personal one-time link plus PIN.
3. Friend opens the link, enters the PIN, and gets a dashboard for their own vehicle only.
4. Ladegast reads sessions live from EVCC, aggregates them, and applies the friend's rates.

Admin can regenerate links, reset PINs, delete friends, and preview any friend's view.

## Architecture

```
Browser (Vue SPA) ──► Ladegast (Go binary, embeds SPA) ──► EVCC (REST API, API key)
                              │
                              └─► bbolt file (friends, admins, login rate limits)
```

| Part | Location | Role |
|---|---|---|
| HTTP server | `internal/httpserver` | Friend and admin APIs, sessions, SPA serving |
| EVCC client | `internal/evcc` | Fetch sessions, switch modes, aggregate |
| Store | `internal/store` | Persistent friends/admins, PIN rate limiting |
| Frontend | `web/` | Vue + Tailwind: friend dashboard, usage explorer, admin UI |
| Mock EVCC | `cmd/evccmock`, `dev/` | Fake EVCC with demo data for local development |

No charging data is stored; only Ladegast's own entities are persisted.

## Security

**Authentication**

- **Friend:** opens a personal link (`/public/l/<uuid>`) and enters a PIN. The PIN is stored as a bcrypt hash. Failed attempts are rate-limited per IP. On success a signed, `HttpOnly`, `SameSite=Strict` session cookie is issued; every further request is checked against that friend's ID. A friend can only see their own vehicle.
- **Admin:** username and password (bcrypt) via `/admin/login`, also cookie-based. Initial credentials come from environment variables on first start.
- **Revocation:** the admin can regenerate a link, reset a PIN, or delete a friend; the old link/PIN stops working immediately.

**Exposing the service**

Ladegast does not restrict the admin area itself. The infrastructure must. Only `/public/*` and the static frontend may be reachable from the internet; `/admin/*` stays internal.

```
Internet ──► reverse proxy ──► Ladegast
              /admin/* → 404      (admin only via LAN / VPN)
```

The proxy (Caddy, nginx, Traefik, ...) rejects the admin prefix and forwards everything else.

**Option: local service + cloud proxy**

Run Ladegast (and EVCC) at home, and a small cloud server as the public entry point, connected through a VPN such as Tailscale:

```
Friend ──► cloud server (reverse proxy, blocks /admin) ══ Tailscale ══► home: Ladegast ──► EVCC
Admin  ──► home network / Tailscale ──────────────────────────────────► Ladegast /admin
```

- No inbound ports opened at home; EVCC is never exposed.
- Only the cloud proxy can reach Ladegast over the VPN, and it only forwards the public part.
- The admin uses Ladegast directly from the home network or the VPN.

Details and a worked example: [DEPLOYMENT.md](DEPLOYMENT.md).

## Quick start

```bash
cp .env.example .env     # fill in values, see schema/config.md
task web:install && task web:build
task run
```

For a demo without real EVCC: `task evcc:mock` and point `LADEGAST_EVCC_BASE_URL` at `http://localhost:7070` (see [dev/README.md](dev/README.md)).

Common tasks: `task test`, `task web:test`, `task check`, `task docker:build`.

## Further docs

- [DESIGN.md](DESIGN.md): architecture, data model, API routes, auth flow (German)
- [DEPLOYMENT.md](DEPLOYMENT.md): running the binary or Docker image (German)
- [schema/config.md](schema/config.md): environment variable reference

## Versioning & releases

[Semantic Versioning](https://semver.org) driven by [Conventional Commits](https://www.conventionalcommits.org). Run `npm ci` once at the repo root to enable the `commit-msg` hook (husky + commitlint).

| Commit | Release |
| --- | --- |
| `fix: ...` | patch |
| `feat: ...` | minor |
| `feat!: ...` / `BREAKING CHANGE:` footer | major |
| `chore:` `docs:` `ci:` `test:` `refactor:` ... | none |

On every push to `main` CI runs [semantic-release](https://semantic-release.gitbook.io): it computes the next version from the commits since the last `v*` tag, updates `package.json`, `web/package.json` and `CHANGELOG.md` (committed back as `chore(release): x.y.z [skip ci]`), tags `vX.Y.Z`, creates the GitHub release and then builds the Docker image tagged `X.Y.Z`, `X.Y` and `latest`. No releasable commit, no new tag, no image. Use squash merges with a conventional PR title (PR titles are linted).

Preview locally: `npm run release:dry` (only meaningful on `main`).
