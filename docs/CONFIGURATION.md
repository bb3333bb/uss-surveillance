# Configuration & Secrets Inventory

Environment variables the running system reads today, plus what will be needed once Phase 3/4 wire in real infrastructure. Nothing here is a secret value — this is the inventory, not the store. Actual values belong in GitHub Actions repo/environment secrets (for CI) and an untracked `.env` / secrets file on the on-prem host (for deployment), never committed.

## Consumed today (`backend/cmd/gateway/main.go`)

| Variable | Purpose | Dev default if unset |
|---|---|---|
| `PORT` | Gateway HTTP listen port | `8080` |
| `SUGGESTION_ENGINE_URL` | Base URL of the Python suggestion/planner service | `http://localhost:50051` |
| `FRONTEND_URL` | Where the browser is redirected after login (mock or real OIDC), carrying the issued token/user/role as query params | `http://localhost:5173` (Vite dev server) |
| `JWT_SECRET` | HMAC signing key for session JWTs | Only defaults to an insecure dev key in mock mode (see `SKIP_OIDC_INIT`/`OIDC_ISSUER_URL` below) — **required** (gateway refuses to start without it) whenever `OIDC_ISSUER_URL` is set |
| `SKIP_OIDC_INIT` | When `"true"`, explicitly forces mock-mode auth (unauthenticated "operator" login) regardless of `OIDC_ISSUER_URL` | unset |
| `OIDC_ISSUER_URL` | SSO provider issuer URL | unset — leaving this unset is what puts the gateway into mock mode for local dev/CI. If it **is** set and OIDC initialization fails, the gateway refuses to start rather than silently falling back to mock mode (set `SKIP_OIDC_INIT=true` if that's actually what you want) |
| `OIDC_CLIENT_ID` | SSO client ID | required once `OIDC_ISSUER_URL` is set |
| `OIDC_CLIENT_SECRET` | SSO client secret | none — required once a real IdP is wired in |
| `OIDC_REDIRECT_URI` | OAuth2 callback URL | none |
| `REDIS_URL` | Redis connection string (e.g. `redis://redis:6379`) for the operator control lease mutex (`backend/pkg/lease.RedisManager`) | unset — falls back to an in-process, single-instance lease manager. Set this for any multi-instance gateway deployment. |
| `CORS_ALLOWED_ORIGIN` | Origin allowed to call the API cross-origin (`backend/pkg/cors`) | `http://localhost:5173` (Vite dev server). Set to the real frontend origin in any non-local deployment. |
| `WEATHER_API_KEY` | OpenWeatherMap API key for the FR-4 wind-safety check (`backend/pkg/weather.Client`) | unset — falls back to the deterministic dev/CI stub (also used if a live fetch errors). Get a free-tier key at openweathermap.org and set this to go live. |

**Open item:** `[ASSUMPTION: SSO]` in the PRD is still unverified — we don't yet know which real OIDC/OAuth2 provider the org uses. `OIDC_CLIENT_SECRET`/`OIDC_ISSUER_URL`/`OIDC_CLIENT_ID` can't be set to real values until that's resolved.

## Not yet consumed by any code (planned for Phase 3/4)

These appear in the HLD but nothing in the codebase reads them yet — listed here so they aren't forgotten when Phase 3 wires the real integrations in:

| Planned variable | Purpose | Blocked on |
|---|---|---|
| `DATABASE_URL` (Postgres+PostGIS) | Mission archive, geofence data (currently a local JSON file in `backend/pkg/archive`) | Phase 3 real-infra wiring |
| `MQTT_BROKER_URL` + TLS cert paths | Real Mosquitto broker connection (currently an in-memory mock in `backend/pkg/mqtt`) | Phase 3 real-infra wiring, and eventually real Drone Hub hardware |
| `SRS_*` (media server ingest/playout URLs) | WebRTC video streaming (not yet implemented in the gateway at all) | Phase 3/4 |

## CI (GitHub Actions)

`.github/workflows/ci.yml` currently needs no secrets — `SKIP_OIDC_INIT=true` isn't even required because CI only runs `go vet`/`go build`/`go test`, not the actual server binary. If a future workflow job starts the gateway (e.g. for an integration test), set `SKIP_OIDC_INIT: "true"` as a job env var rather than a secret, since it's not sensitive.
