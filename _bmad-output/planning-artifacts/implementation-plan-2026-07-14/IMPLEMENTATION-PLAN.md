---
title: USS Surveillance — Implementation & Deployment Plan
status: draft
created: 2026-07-14
based_on:
  - _bmad-output/planning-artifacts/prds/prd-uss-surveillance-2026-07-08/prd.md
  - _bmad-output/planning-artifacts/architecture/architecture-uss-surveillance-2026-07-09/HIGH-LEVEL-DESIGN.md
  - _bmad-output/planning-artifacts/architecture/architecture-uss-surveillance-2026-07-09/ARCHITECTURE-SPINE.md
  - _bmad-output/planning-artifacts/architecture/architecture-uss-surveillance-2026-07-09/DEPLOYMENT-BUDGETING-GUIDE.md
  - _bmad-output/planning-artifacts/epics.md
  - _bmad-output/implementation-artifacts/sprint-status.yaml
decisions:
  base: adopt-existing-design-and-code
  repo: continue-in-uss-surveillance
  hardware: simulator-only-for-now
  deployment_target: on-premises
---

# USS Surveillance — Implementation & Deployment Plan

## 0. Starting Point (why this plan looks the way it does)

This is **not a greenfield project**. A prior BMad-method planning pass already produced a PRD, HLD, UX design, epic/story breakdown, and a coded (simulator-based) implementation of all 16 stories across 4 epics, currently sitting at status `review` in [sprint-status.yaml](../../implementation-artifacts/sprint-status.yaml). Tech stack, architecture, and deployment target are already decided documents — this plan does not re-litigate them, it closes the gap between "coded" and "shipped."

Confirmed direction (per your answers):
- Reuse existing HLD/tech stack/code as the foundation.
- Continue in this repo (`git@github.com:bb3333bb/uss-surveillance.git`), not a new one.
- Hardware integration stays simulator-only until real DJI Dock credentials exist.
- Deployment target is on-premises, per the existing Deployment & Budgeting Guide.

## 1. Design & HLD — status: mostly done, two gaps to close

Already exists and is sound:
- Component boundaries, data flows (telemetry, command override, WebRTC video, mission replay) — [HIGH-LEVEL-DESIGN.md](../../architecture/architecture-uss-surveillance-2026-07-09/HIGH-LEVEL-DESIGN.md)
- ADR-style decisions — `ARCHITECTURE-SPINE.md`
- UX flows, mockups, midnight-ocean theme spec — `ux-designs/ux-uss-surveillance-2026-07-08/`

Gaps to close before implementation work continues:
1. **Arch-1 (Docker Stack) is unimplemented.** The epics document requires a Docker Compose stack (Go gateway, Python GIS/suggestion engine, SRS, Mosquitto, Postgres+PostGIS, Redis, frontend), but no `docker-compose.yml` or Dockerfiles exist anywhere in the repo. This blocks both local onboarding and deployment.
2. **PRD Open Questions still open** (§8 of the PRD) — need answers before the affected FRs can be considered "done," not just "coded":
   - Video protocol: HLD already assumes RTSP/RTMP ingest → SRS → WebRTC playout. Needs confirming against the actual target drone/dock model's supported protocols.
   - Offline DEM/tile loading: MVP scope explicitly excludes offline capability, so this is effectively answered ("online, self-hosted tile server for LAN security, not for offline use") — should be written back into the PRD/HLD explicitly so it stops appearing as open.
   - Max wind tolerance + weather API refresh cadence: currently **hardcoded/mocked** in `backend/pkg/weather` (a hardcoded lat threshold, not a live API call). Needs a real drone spec and a decision between OpenWeatherMap free tier (~10 min refresh) vs. a paid tier ($40–120/mo, per the budgeting guide) given the 15 m/s safety-block requirement (FR-4).

## 2. Tech Stack — already chosen, carried forward as-is

| Layer | Choice | Notes |
|---|---|---|
| Frontend | React + MUI, Leaflet/MapLibre | Midnight Ocean theme, 3-column dashboard |
| Backend gateway | Go | REST + WebSocket, JWT/OIDC, PostGIS geofence checks |
| GIS/suggestion service | Python, gRPC | Grid path generation, drone/hub suggestion logic |
| Video | SRS (Simple Realtime Server) | RTSP/RTMP ingest → WebRTC playout |
| Drone comms | Eclipse Mosquitto (MQTT over SSL) | Telemetry + command topics |
| Database | PostgreSQL + PostGIS | Mission archive, geofence data |
| Cache/locking | Redis | Operator mutex leases, live telemetry cache |
| OS | Ubuntu Server 22.04 LTS | On-prem target |
| Tiles | Self-hosted (Tegola/MapTiler Server) over OSM | LAN-hosted for network isolation, not offline use |

All open-source, $0 licensing (per Deployment & Budgeting Guide). No changes proposed — this is a good, coherent stack for the stated requirements.

## 3. Phased Plan

### Phase 0 — Repo hygiene (before anything else)
- `git status` currently shows an unstaged edit (`frontend/src/index.css`) and several untracked dirs (`.agent/`, `.claude/`, `.cursor/`, `.gemini/`, `.github/`, `.gitignore`) — leftovers from the BMad/Antigravity template scaffolding. Needs a decision pass: what's kept (e.g. `.github/` issue templates), what's gitignored (agent tool configs), before any new commits land on top.
- Confirm `master` is the intended long-lived branch and decide branch strategy (trunk-based + short-lived feature branches recommended, given a solo/small team).

### Phase 1 — Close the Docker/Arch-1 gap
- Write Dockerfiles for: Go gateway, Python suggestion engine, frontend (static build served via nginx or similar).
- Write `docker-compose.yml` wiring: gateway, suggestion engine, frontend, Postgres+PostGIS, Redis, Mosquitto, SRS, and (new) a self-hosted tile server.
- This unblocks both "run it locally with one command" and "deploy it on-prem" — currently RUN.md requires three manual terminals and a locally-installed toolchain per service.

### Phase 2 — GitHub / project requirements setup
- Add GitHub Actions CI: Go build+vet+test (`backend/pkg/*_test.go` already exist — good baseline), Python lint/test for the suggestion engine (currently no tests found there), frontend build+lint (currently no frontend tests found), and a proto-codegen-drift check for `proto/suggestion.proto`.
- Add branch protection on `master` requiring CI + 1 review (this is a GitHub repo-settings change — I'll need your go-ahead and admin access to the `bb3333bb/uss-surveillance` repo to apply it).
- Secrets/config inventory to provision: OIDC client ID/secret (SSO provider — org-specific, currently `[ASSUMPTION: SSO]` unverified), weather API key, MQTT/SRS credentials, DB credentials. Store as GitHub Actions secrets for CI and a `.env`/secrets file for the on-prem host (not committed).

### Phase 3 — Finish the implementation (16 stories: review → done)

**Done (2026-07-14):**
- ✅ Shared WGS84 distance helper (`backend/pkg/geo`) — also fixed a real bug it uncovered: the `RESTRICTED_AIRSPACE` telemetry alert was comparing raw lat/lng degree deltas against a magic-number threshold instead of a real meters-based radius.
- ✅ Mission list pagination (`GET /api/operator/missions?offset=&limit=`), backward compatible with the current no-params frontend call.
- ✅ WebSocket telemetry/command JSON schema documented (`docs/WEBSOCKET-API.md`).
- ✅ Historical telemetry/archive schema documented (`docs/DATA-SCHEMA.md`), including the Postgres migration shape.
- ✅ Migrated operator control leases from in-memory to Redis (`backend/pkg/lease.RedisManager`, `REDIS_URL`-gated with in-memory fallback), closing the multi-instance-safety gap. Verified against both a unit-test double (miniredis) and a standalone miniredis TCP server driving the real gateway binary end-to-end. `docker-compose.yml` now includes a `redis` service.
- ✅ Suggestion-engine (Python) test suite added in Phase 2 (was previously untested).

**Still open:**
- Run the BMad code-review workflow (or an equivalent fresh-context review) on each of the 16 `review`-status stories and resolve findings — not yet started, this is its own sizable pass.
- Wire the real weather API integration (replace the hardcoded threshold in `backend/pkg/weather`) — blocked on choosing a provider (OpenWeatherMap free vs. paid tier) and getting an API key; still an open PRD question (§8.3).
- Wire real Postgres+PostGIS and Mosquitto — larger lifts than originally scoped here (see the mock-infra note in §0); best tackled once Docker Compose itself is verified working (still pending user confirmation, see Phase 1).
- Frontend: no test suite exists yet (Vitest setup needed); the retro-flagged telemetry WebSocket hook/context refactor and timeline-scrubber coordinate caching are both real UI changes that need browser verification before landing, not done blind.
- "Mock test scripts simulating high-speed wind changes" retro item — low value while weather is still a hardcoded stub; revisit once real weather API is wired.

### Phase 4 — On-prem deployment
- Provision hardware per the existing sizing guide: 8–16 core CPU, 32–64GB ECC RAM, 1–2TB NVMe, GPU (T4/A2) recommended for WebRTC transcoding headroom, Ubuntu 22.04 LTS.
- Deploy the Phase 1 Docker Compose stack; put a reverse proxy (nginx/Caddy) in front for TLS termination on the operator-facing HTTP/WebSocket endpoints.
- Configure MQTT over TLS between the server and the Drone Hub subnet; confirm firewall rules isolate that subnet appropriately.
- Set up Postgres backups (mission archive is the audit trail for FR-17/UJ-1's post-mission auditing — losing it defeats a core PRD goal).
- Add basic operational monitoring/alerting (not currently specified anywhere in the architecture docs — worth a lightweight addition, e.g. Prometheus + Grafana or even just structured logs + uptime checks, given SM-C1's zero-incident bar).
- This phase stays simulator-driven per your answer — no real DJI Dock integration work until hardware/credentials exist. Flag this explicitly as the reason SM-1/SM-2/SM-3 can only be validated against the simulator, not real-world conditions, until then.

### Phase 5 — Validate against Success Metrics
- SM-1 (>95% autonomous mission completion), SM-2 (<500ms override latency), SM-3 (>90% suggestion accuracy): build a repeatable simulator-driven test harness that runs enough mission cycles to produce a real percentage, rather than eyeballing a manual RUN.md walkthrough.
- SM-C1 (zero safety incidents): a focused safety/security pass on the interlock logic (FR-12 door-open-before-launch, FR-11 override authorization) before this is ever pointed at real hardware.

## 4. What I need from you before I start executing

- **Phase 0/2**: confirm you want me working directly against `bb3333bb/uss-surveillance` on `master` (with feature branches), and whether you want PRs or direct commits for solo work.
- **Phase 2**: GitHub repo admin access if you want me to configure branch protection / CI secrets myself, or you can add secrets yourself from a list I provide.
- **Phase 1/3**: no blockers, can start immediately once Phase 0 hygiene is resolved.
- **Phase 4**: this is a real deployment to shared infrastructure — I will not run anything against a physical on-prem server without you confirming it's provisioned and explicitly approving each deployment step at the time.

## 5. Suggested order of execution

1. Phase 0 (repo hygiene) — quick, unblocks everything else.
2. Phase 1 (Docker Compose) — highest leverage, makes every subsequent phase easier to test.
3. Phase 2 (CI) in parallel with Phase 3 (finish stories) — CI should exist before more code lands.
3. Phase 3 (finish the 16 stories, close retro action items).
4. Phase 5 (metric validation harness) as Phase 3 stories close out.
5. Phase 4 (on-prem deploy) once hardware is actually provisioned — this is the one phase gated on something outside this repo.
