# Configuration & Secrets Inventory

Environment variables the running system reads today, plus what will be needed once Phase 3/4 wire in real infrastructure. Nothing here is a secret value — this is the inventory, not the store. Actual values belong in GitHub Actions repo/environment secrets (for CI) and an untracked `.env` / secrets file on the on-prem host (for deployment), never committed.

## Consumed today (`backend/cmd/gateway/main.go`)

| Variable | Purpose | Dev default if unset |
|---|---|---|
| `PORT` | Gateway HTTP listen port | `8080` |
| `SUGGESTION_ENGINE_URL` | Base URL of the Python suggestion/planner service | `http://localhost:50051` |
| `JWT_SECRET` | HMAC signing key for session JWTs | insecure dev fallback string — **must** be overridden outside local dev |
| `SKIP_OIDC_INIT` | When `"true"`, skips real OIDC discovery at startup (dev/CI bypass) | unset (OIDC init attempted) |
| `OIDC_ISSUER_URL` | SSO provider issuer URL | `http://localhost:8080/realms/uss-surveillance` (placeholder, not a real IdP) |
| `OIDC_CLIENT_ID` | SSO client ID | `uss-surveillance-client` (placeholder) |
| `OIDC_CLIENT_SECRET` | SSO client secret | none — required once a real IdP is wired in |
| `OIDC_REDIRECT_URI` | OAuth2 callback URL | none |

**Open item:** `[ASSUMPTION: SSO]` in the PRD is still unverified — we don't yet know which real OIDC/OAuth2 provider the org uses. `OIDC_CLIENT_SECRET`/`OIDC_ISSUER_URL`/`OIDC_CLIENT_ID` can't be set to real values until that's resolved.

## Not yet consumed by any code (planned for Phase 3/4)

These appear in the HLD but nothing in the codebase reads them yet — listed here so they aren't forgotten when Phase 3 wires the real integrations in:

| Planned variable | Purpose | Blocked on |
|---|---|---|
| `WEATHER_API_KEY` | Real wind/weather lookup for FR-4 (currently hardcoded lat threshold in `backend/pkg/weather`) | Choosing a provider — OpenWeatherMap free tier vs. paid, per `DEPLOYMENT-BUDGETING-GUIDE.md` |
| `DATABASE_URL` (Postgres+PostGIS) | Mission archive, geofence data (currently a local JSON file in `backend/pkg/archive`) | Phase 3 real-infra wiring |
| `REDIS_URL` | Operator mutex leases, telemetry cache (currently in-process) | Phase 3 real-infra wiring |
| `MQTT_BROKER_URL` + TLS cert paths | Real Mosquitto broker connection (currently an in-memory mock in `backend/pkg/mqtt`) | Phase 3 real-infra wiring, and eventually real Drone Hub hardware |
| `SRS_*` (media server ingest/playout URLs) | WebRTC video streaming (not yet implemented in the gateway at all) | Phase 3/4 |

## CI (GitHub Actions)

`.github/workflows/ci.yml` currently needs no secrets — `SKIP_OIDC_INIT=true` isn't even required because CI only runs `go vet`/`go build`/`go test`, not the actual server binary. If a future workflow job starts the gateway (e.g. for an integration test), set `SKIP_OIDC_INIT: "true"` as a job env var rather than a secret, since it's not sensitive.
