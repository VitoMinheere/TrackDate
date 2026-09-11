# Trackdate

A small, self-hosted app for planning and sharing track days with friends — built to replace the usual mess of shared spreadsheets and scattered WhatsApp messages.

Create a list of your track days, share it with a link, and let others see it read-only. No accounts, no passwords — just email magic-link sign-in.

> Status: MVP implemented — passwordless login, adding track days, and the read-only share link all work.

## Why

Track day plans usually end up split across a shared spreadsheet with a training partner and a pile of WhatsApp/Discord messages with a wider group. Trackdate is a minimal alternative: one shareable list per person, kept up to date in one place.

## Planned features

**Minimum version**
- Passwordless login via email magic link
- Add track days with date, track, organization, and cost
- Shareable read-only link to your own list

**Planned next**
- Provisional vs. booked/confirmed status per track day (most organizations don't publish next year's calendar until November/December)
- Notes/link field per track day for registration deadlines or requirements
- Public share link viewable with no login, for sharing outside the group (e.g. on Instagram)
- A lightweight way for others to mark themselves as attending
- Pit box coordination: claim a spot in a pit box, let others join it, and see who and how many are in it
- Per-viewer visibility, so specific track days can be hidden from specific people

**Explicitly out of scope**
- File/document uploads — avoided to sidestep storing other people's data

## Tech stack

- Go, standard library only aside from a pure-Go SQLite driver (`modernc.org/sqlite`) — no CGO, no web framework, no ORM
- SQLite file on a mounted volume — plenty for a single-container, low-traffic self-hosted app
- Magic-link email via SMTP (`net/smtp`), configured with env vars; falls back to logging the link to stdout when unconfigured, for local development
- Self-hosted on a homelab k3s cluster, deployed via Flux, exposed through a Cloudflare Tunnel

## Running locally

```sh
go run ./cmd/trackdate
```

The server listens on `:8080` by default. Without `SMTP_HOST` set, sign-in
links are logged to stdout instead of emailed — copy the URL from the logs
to sign in during local development.

## Configuration

All configuration is via environment variables:

| Variable | Default | Notes |
|---|---|---|
| `TRACKDATE_ADDR` | `:8080` | Listen address |
| `TRACKDATE_DB_PATH` | `trackdate.db` | Path to the SQLite file; point this at a mounted volume in production |
| `TRACKDATE_BASE_URL` | `http://localhost:8080` | Used to build magic-link and share URLs; set this to the public HTTPS URL in production (the session cookie's `Secure` flag is derived from this) |
| `TRACKDATE_CURRENCY` | `EUR` | Display-only currency label shown next to costs |
| `SMTP_HOST` | *(unset)* | SMTP relay host; if unset, magic links are logged instead of emailed |
| `SMTP_PORT` | `587` | |
| `SMTP_USERNAME` / `SMTP_PASSWORD` | *(unset)* | Omit both for an unauthenticated relay |
| `SMTP_FROM` | `Trackdate <trackdate@localhost>` | `From:` header on the sign-in email |

## Deployment

The included `Dockerfile` builds a static binary (CGO disabled) onto a
distroless base and runs as non-root. Mount a volume at `/data` for the
SQLite file (`TRACKDATE_DB_PATH` defaults to `/data/trackdate.db` in the
image) and set `TRACKDATE_BASE_URL` to the public URL the app is reached at
through the Cloudflare Tunnel. The app itself speaks plain HTTP — TLS
termination is expected to happen at the tunnel.

## License

[GPL-3.0](LICENSE)
