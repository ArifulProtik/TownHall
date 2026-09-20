# TownHall UI

React 19 + Vite 8 + TypeScript + Tailwind v4 + Redux Toolkit (RTK Query) + React Router 7, managed with bun.

## Prerequisites

- bun 1.x (`bun --version`)
- Backend running locally (`make dev` in repo root → `http://localhost:8080`)

## Commands (run in `ui/`)

- `bun install` — install dependencies
- `bun run dev` — Vite dev server on `:5173`; `/api/*` is proxied to `http://localhost:8080`
- `bun run build` — `tsc -b && vite build` → `ui/dist/`
- `bun run lint` / `bun run format` — ESLint / Prettier
- `bun run test` / `bun run test:watch` — Vitest (single run / watch)
- `bun run typecheck` — `tsc --noEmit`

## Auth model

- Access token lives in Redux memory only (never persisted, never logged).
- Refresh travels by HttpOnly cookie (`credentials: "include"`); 401s funnel through `baseQueryWithReauth` (refresh → retry once → else logout).
- TS types in `src/features/auth/authApi.ts` mirror `internal/auth/auth_model.go` exactly — update both sides together when the contract changes.
