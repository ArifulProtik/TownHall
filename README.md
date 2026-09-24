# TownHall

TownHall is an interactive social media platform that blends community spaces with a public timeline — Discord-style servers and channels meets a Twitter-style feed: realtime conversation, posts, and the communities around them.

## Tech stack

- Golang
- PostgreSQL
- Echo
- EntGo
- React

## Prerequisites

- Go 1.26+
- PostgreSQL
- Node.js (for the React frontend)
- [air](https://github.com/air-verse/air) for live reload (optional)

## How to run

```bash
cp .env.example .env   # point DATABASE_URL at your Postgres
make dev               # live reload on :8080
```

```bash
curl localhost:8080/api/v1/auth/health
```

## Observability (local)

`make up` starts the prod-shape stack (app + pg + redis + Grafana + Loki +
Prometheus + Alloy; needs Docker; stop `make dev` first — both want
`:8080`). Grafana at
`http://localhost:3000` (admin/admin) opens the TownHall dashboard:
request rate, 5xx share, p95 latency, and app logs correlated by
`request_id`. `make logs` follows the app container; `make down` stops
everything. Compose Postgres is separate from your native one on
`:5432`, which stays untouched.

## Docker dev (hot reload)

Prefer containers over native Postgres? `make dev-docker` starts a dev
stack (needs Docker; stop `make dev-all` first): `api` hot-reloads via
air on `:8080`, plus throwaway pg + redis on dev-only volumes. Run the
frontend natively alongside it (`bun --cwd=ui run dev`, lighter than a
container) — its `/api` proxy already points at `localhost:8080`.
`make dev-docker-full` additionally dockerizes the web frontend
(`--profile web`, heavier). `make dev-docker-logs` follows the api;
`make dev-docker-down` stops everything.

## On the roadmap

- Community spaces (servers, channels, roles) and realtime chat
- Public timeline: posts, replies, likes, follows
- Direct messages, notifications, moderation
