# Hosting & making Ladegast reachable for guests

Ladegast must be reachable by your guests' phones, i.e. from the internet. EVCC must **not** be. This page explains the options, from "no server needed" to "full control".

> Looking for the German operator notes (Kubernetes, env vars, backups)? See [DEPLOYMENT.md](DEPLOYMENT.md).

## The short version

- **You do not need your own public server.** Ladegast runs at home (next to EVCC) and a *tunnel service* gives it a public `https://` address. No router changes, works behind CGNAT / DS-Lite.
- **You do need a stable address.** Guest links contain the public address. If it changes, every link already sent breaks. Pick an option with a fixed hostname.
- **The admin area is never public** (blocked by Ladegast itself). See [Protect the admin area](#protect-the-admin-area).

## What has to talk to what

```
Guest phone ──(internet)──► public address ──► Ladegast (home) ──► EVCC (home, LAN only)
You (admin) ─────────(home network / VPN)────► Ladegast /admin
```

Ladegast runs on any always-on machine in your home network (Raspberry Pi, NAS, the machine that runs EVCC). It needs to reach EVCC (`LADEGAST_EVCC_BASE_URL`) and must be reachable by the tunnel/proxy on `LADEGAST_LISTEN_ADDR` (default `:8080`).

`LADEGAST_PUBLIC_BASE_URL` must be set to the **public address** (e.g. `https://ladegast.example.com`), because that is what ends up in the guest links.

## Choose an option

| | A. Cloudflare Tunnel | B. Tailscale Funnel | C. Your own VPS + reverse proxy | D. Port forwarding at home | E. Cloudflare quick tunnel |
|---|---|---|---|---|---|
| Own server needed | no | no | yes (a few €/month) | no | no |
| Own domain needed | yes (any registrar, DNS on Cloudflare) | no (`<name>.ts.net`) | recommended | recommended + DynDNS | no |
| Works behind CGNAT / DS-Lite | yes | yes | yes | **no** | yes |
| Router changes | none | none | none | yes | none |
| Stable address | yes | yes | yes | yes (with DynDNS) | **no, changes on restart** |
| `/admin` unreachable from internet | built in | built in | built in | built in | built in |
| Technical level | medium | easiest | highest | medium, risky | trivial, testing only |
| Cost | free tier | free (beta) | VPS | free | free |

**Recommendation for most users:** option **B** if you want the least effort and don't own a domain, option **A** if you own a domain (nicer address, supports blocking `/admin`). Use **C** if you already run a server. Avoid **D** unless you know how to secure it. Use **E** only to try things out. All options keep `/admin` local-only.

---

## A. Cloudflare Tunnel (own domain, free)

`cloudflared` runs next to Ladegast and opens an *outbound* connection to Cloudflare; Cloudflare serves your domain and forwards requests through it. Nothing is opened on your router.

Requirements: a free Cloudflare account and a domain whose DNS is managed by Cloudflare.

1. In the Cloudflare dashboard open **Zero Trust → Networks → Tunnels → Create a tunnel** (menu names change occasionally) (type *Cloudflared*), and copy the install/run command for your OS or Docker.
2. Run it on the machine that runs Ladegast (or any machine in the same network).
3. Add a **public hostname**: e.g. `ladegast.example.com` → service `http://<ladegast-host>:8080`.
4. Set `LADEGAST_PUBLIC_BASE_URL=https://ladegast.example.com` and restart Ladegast.
5. Nothing more to do for `/admin`: Ladegast blocks it for tunnel traffic, see [Protect the admin area](#protect-the-admin-area).

Docker Compose sketch (Ladegast + tunnel):

```yaml
services:
  ladegast:
    image: ghcr.io/mafeko/ladegast:latest
    env_file: .env
    volumes: [ "ladegast-data:/data" ]
    restart: unless-stopped
  cloudflared:
    image: cloudflare/cloudflared:latest
    command: tunnel --no-autoupdate run
    environment:
      TUNNEL_TOKEN: ${CLOUDFLARE_TUNNEL_TOKEN}
    restart: unless-stopped
volumes:
  ladegast-data:
```

In this setup the public hostname points to `http://ladegast:8080` (Compose service name).

Notes: Cloudflare terminates TLS and can technically see traffic. Cloudflare's terms are aimed at web content; Ladegast (HTML/JSON) is fine.

## B. Tailscale Funnel (no domain, no server)

[Tailscale](https://tailscale.com) is a VPN service. Its **Funnel** feature publishes a local port on a public `https://<machine>.<tailnet>.ts.net` address, with a valid certificate, without router changes.

Requirements: a free Tailscale account, Tailscale installed on the machine running Ladegast, MagicDNS and HTTPS enabled for your tailnet (Tailscale admin console → DNS), and Funnel allowed for your user (the CLI offers to enable it).

```bash
tailscale funnel --bg 8080
```

The command prints your public address. Put it into `LADEGAST_PUBLIC_BASE_URL` and restart Ladegast.

Good to know:

- Funnel is documented as **beta**, listens only on ports 443, 8443 and 10000, and has non-configurable bandwidth limits. That is plenty for Ladegast.
- The address is stable as long as the machine name and tailnet name stay the same.
- Funnel forwards all paths, but Ladegast refuses `/admin` for such requests (see [Protect the admin area](#protect-the-admin-area)). You can still reach the admin area privately through Tailscale itself.
- Guests need nothing installed; only you (the owner) need Tailscale.

Docs: [Tailscale Funnel](https://tailscale.com/kb/1223/funnel).

## C. Own server (VPS) as public entry point

A small rented server runs a reverse proxy (Caddy, nginx, Traefik) with a domain and forwards to Ladegast at home through a VPN such as Tailscale or WireGuard. This gives full control. `/admin` is already refused for proxied requests; the extra Caddy rule below is optional defense in depth:

```
ladegast.example.com {
    @admin path /admin /admin/*
    respond @admin 404
    reverse_proxy <home-tailscale-ip>:8080
}
```

(If you changed `LADEGAST_ADMIN_PATH_PREFIX`, use that prefix.) The detailed home-cluster scenario is described in [DEPLOYMENT.md](DEPLOYMENT.md) and the README.

## D. Port forwarding at home (not recommended)

Forward port 443 on your router to a reverse proxy at home, plus DynDNS. Needs a public IPv4 address (not possible behind CGNAT/DS-Lite), exposes your home IP, and makes you responsible for TLS certificates and updates. Prefer A or B.

## E. Cloudflare quick tunnel (testing only)

```bash
cloudflared tunnel --url http://localhost:8080
```

Creates a random `trycloudflare.com` address without account or domain. Cloudflare documents quick tunnels as [for testing and development, with no uptime guarantee](https://developers.cloudflare.com/cloudflare-one/networks/connectors/cloudflare-tunnel/do-more-with-tunnels/trycloudflare/); the address **changes every time you restart**, which would invalidate all guest links. Good for a demo with the mock EVCC, not for real guests.

---

## Protect the admin area

**The admin area is local-only by default.** Ladegast answers requests to `/admin` (and the admin API) with 404 unless they come directly from a trusted address: loopback, private LAN ranges, or Tailscale (`100.64.0.0/10`). Requests that arrive through a tunnel or reverse proxy carry headers such as `X-Forwarded-For` or `CF-Connecting-IP`; those are always refused, even if the proxy runs on the same machine. So with Cloudflare Tunnel, Tailscale Funnel or a VPS proxy, guests can reach their link and nobody can reach `/admin` from the internet. No extra proxy rule is needed.

You manage guests by opening `http://<machine>:8080/admin` from your home network, or from anywhere through your VPN (Tailscale).

Details and limits:

- The check uses the connection's source address plus proxy headers. It is not a replacement for keeping EVCC private: never expose EVCC itself.
- A reverse proxy *inside* your LAN (e.g. Traefik for `ladegast.lan`) adds headers and will be blocked too. If you need that, set `LADEGAST_ADMIN_ALLOW_REMOTE=true` and block the admin prefix at the public entry point yourself (Caddy example in option C; Cloudflare: public hostname with path `^/admin` and service `http_status:404`).
- `LADEGAST_ADMIN_ALLOW_REMOTE=true` also removes the guard completely; the admin login then has no rate limit, so use a long random password.

Always verify from outside your network (mobile data): guest link works, `/admin` returns 404.

## Checklist before sharing the first link

- [ ] `LADEGAST_PUBLIC_BASE_URL` is the final, stable, `https://` public address.
- [ ] Opening a guest link from mobile data shows the PIN screen.
- [ ] `https://<public address>/admin` returns 404 from mobile data.
- [ ] EVCC is not reachable from the internet.
- [ ] `LADEGAST_SESSION_SECRET` is random (32+ bytes) and the `ladegast.db` data volume is backed up.
- [ ] EVCC API key created (see [GETTING-STARTED.md](GETTING-STARTED.md)).
