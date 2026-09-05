# Trackdate

A small, self-hosted app for planning and sharing track days with friends — built to replace the usual mess of shared spreadsheets and scattered WhatsApp messages.

Create a list of your track days, share it with a link, and let others see it read-only. No accounts, no passwords — just email magic-link sign-in.

> Status: early weekend project, not yet functional. This README describes the intended scope.

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

- Go
- Self-hosted on a homelab k3s cluster, deployed via Flux, exposed through a Cloudflare Tunnel

## License

[GPL-3.0](LICENSE)
