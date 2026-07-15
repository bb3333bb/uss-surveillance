# Code Review: All 16 Stories (Epics 1–4)

**Date:** 2026-07-14
**Method:** Adapted from `.claude/commands/BMad/tasks/review-story.md` (the full BMad QA-gate workflow needs `.bmad-core/` config that isn't vendored into this repo — see README.md — so this substitutes a direct acceptance-criteria-vs-code review, written up per story, without the gate YAML/quality-score machinery). Each epic was reviewed by an independent agent against the actual current code, not against story-file checkboxes or `sprint-status.yaml`, both of which were found to be stale in places (see notes below).

Nothing was fixed as part of the initial review pass — that was findings only. Items 1-9 below were fixed the same day; see the "Fixed 2026-07-14" note on each.

## Severity-ranked findings

### Critical — misrepresents safety-critical state to the operator

1. ~~**"Pause Flight" is a no-op.**~~ **Fixed 2026-07-14** (`0da05f6`). `commandHandler` (`backend/cmd/gateway/main.go`) published MQTT `hover`, but nothing subscribed to it and the handler's own state-mutation block only branched on `rth`. The drone kept flying after an operator held Pause for 1.5s and the UI reported success. Added `IsPaused` to `DroneTelemetryState`; the ticker now holds position/altitude, zeroes speed, and skips path advancement while paused. Also fixed the heartbeat watchdog's disconnect handler, which had the same bug from the other direction (cleared `IsFlying` instead of pausing, snapping the position back to the dock). (Story 3.4)
2. ~~**`WEATHER_BREACH_WIND`/`WEATHER_BREACH_RAIN` alerts are fake.**~~ **Fixed 2026-07-14** (`a4c51f6`). They fired off `lat > 10.77`, completely disconnected from `backend/pkg/weather`. Now the telemetry ticker polls real (or stub) weather at the drone's current position every 30 ticks while flying, and the alerts derive from actual wind speed / precipitation. (Story 3.6)
3. ~~**NFZ/geofence check only inspects the drawn boundary's vertices.**~~ **Fixed 2026-07-14** (`28c53dc`). Replaced the vertex-only loop with `polygon_intersects_restricted_zone()`, which checks the minimum distance from the NFZ center to every polygon *edge* (point-to-segment distance in a local flat-earth projection), geometrically superseding the old check rather than adding a redundant one. (Story 2.4)
4. ~~**The suggestion engine (FR-5) is a fully static stub.**~~ **Fixed 2026-07-14** (`554832a`). Since only one drone/hub is modeled, true multi-drone selection isn't meaningful - `applyFleetReadiness()` in the Go gateway now overlays Drone-01's real readiness (in-flight, locked by another operator, low battery) onto the suggestion response after the Python engine's geographic allocation comes back. (Story 2.3)
5. ~~**Live weather safety check ignores precipitation.**~~ **Fixed 2026-07-14** (`804448b`). Only wind speed fed `Safe`. Rain/Thunderstorm/Snow now also block; light Drizzle doesn't. (Story 2.2)
6. ~~**Hardcoded default `JWT_SECRET` committed to source, combined with an implicit, undocumented mock-auth fallback.**~~ **Fixed 2026-07-14** (`2b69f7f`). Any environment where `OIDC_ISSUER_URL` was unset or the IdP was briefly unreachable at boot silently served an unauthenticated "operator" JWT to anyone hitting `/api/auth/login`. Mock mode is now only entered via explicit `SKIP_OIDC_INIT=true` or when `OIDC_ISSUER_URL` is unset entirely (the local-dev default, unchanged); if it's set but init fails, the gateway now `log.Fatalf`s instead of downgrading silently. The dev-fallback `JWT_SECRET` only applies in mock mode. (Story 1.1)
7. ~~**Real OIDC callback flow is broken.**~~ **Fixed 2026-07-14** (`2b69f7f`). `HandleCallback` wrote a raw JSON body instead of redirecting back to the frontend with the token — only the mock-mode branch in `HandleLogin` actually worked. Both paths now share one `redirectToFrontend()` helper, and the target origin is configurable via `FRONTEND_URL`. (Story 1.1)
8. ~~**Hub-doors telemetry state is inaccurate for the entire duration of a flight.**~~ **Fixed 2026-07-14** (`6bece54`). Never set on launch (stayed `"closed"` while the drone was airborne) and never reset out of `"recharging"` afterward. `launchHandler` now reflects doors open→closing→closed after takeoff; the battery-replenishment tick now transitions recharging→closed at 100%. (Stories 3.1, 3.5)
9. ~~**Heartbeat-timeout "pause" jumps the drone's displayed position back to the dock.**~~ **Fixed 2026-07-14**, as part of the Pause fix above (`0da05f6`) — the watchdog now sets `IsPaused` instead of clearing `IsFlying`. (Story 3.2)

### Known large gaps (not bugs — features not yet built)

10. No `Land` command anywhere (frontend or backend), despite FR-11 requiring Pause/RTH/Land. (Stories 1.2, 1.3, 3.4)
11. FR-6 camera-FOV-based lane spacing for 20% overlap — not implemented; fixed constant used instead.
12. FR-8 terrain/obstacle clearance (30m above terrain) — entirely absent, no DEM data source anywhere.
13. No real video recording — `video_path` is a synthesized fake path; no MP4 is ever written. (Already documented in `docs/DATA-SCHEMA.md`.)
14. No SRS media server exists anywhere in the repo (not in `docker-compose.yml`, not referenced except in a stray architecture-doc note). `VideoPlayer.jsx` always times out and falls into the Canvas-HUD simulation — the retro already admits this. (Story 3.3)
15. Timeline scrubber sync is one-directional (slider → map/HUD only) — structurally can't be "vice versa" until real video exists to scrub. (Story 4.3)
16. Mission-history pagination plumbing exists (`?offset=&limit=`) but the frontend never sends those params, and payloads still aren't split into meta-only vs. full-detail — the original scaling concern from the retro isn't actually resolved despite the yaml suggesting otherwise.

### Smaller / cosmetic

17. Missing `@keyframes pulse` — charging-bar animation name is dangling, no visual effect.
18. Missing `@keyframes` for the "flashing red border" safety indicator — only the Canvas HUD banner text actually flashes; the map/video card borders are static.
19. Speed isn't shown in the video HUD overlay despite being available in telemetry data.
20. Sidebar tab labels differ cosmetically from the story spec ("Fleet"/"History" vs. "Fleet & Docks"/"Mission Logs").
21. Monospace numeric font isn't centralized in `theme.js` typography — works today via ad-hoc `sx` overrides only.
22. "Drone details" not shown on mission history cards.
23. `backend/pkg/mqtt` has no `Unsubscribe` — every launch registers a new handler that's never removed (goroutine/subscriber leak across repeated launches).
24. No server-side re-check of weather/geofence safety at the moment `/api/operator/launch` is actually called — all gating lives only in the React button's `disabled` prop; a direct POST bypasses it.

### Test coverage pattern

Several stories' "unit tests added" checkboxes are `[x]` in the story files but the corresponding tests don't exist — confirmed by directly inspecting `backend/cmd/gateway/main_test.go` (only has pagination-helper tests) and the frontend (zero test files exist anywhere, no test framework installed). Specifically missing: `commandHandler`, the landing-sequence ticker, the alert-builder block, the watchdog integration, `HandleLogin`/`HandleCallback`, and all frontend logic.

### sprint-status.yaml / retro accuracy

Two items marked `open` in `sprint-status.yaml` are actually done in code (Redis lease migration — done and wired into `docker-compose.yml`; scrubber `.map()` caching — already effectively memoized). Conversely, several story-file task checkboxes claim test coverage that doesn't exist. Don't trust either source blindly — verify against code.

## Per-story verdicts

| Story | Verdict |
|---|---|
| 1.1 OIDC Authentication & JWT Validation | Needs Fixes → mock-mode/JWT_SECRET fallback and broken callback redirect both fixed (`2b69f7f`); OIDC success-path itself still untested against a real IdP |
| 1.2 Base 3-Column Dashboard Layout & Themes | Partially Done |
| 1.3 Role-Based Access Control Enforcer | Partially Done |
| 2.1 Polygon Drawing on 2D Map View | Done |
| 2.2 Automated Weather and Wind Assessment | Needs Fixes → precipitation gap fixed (`804448b`); rest of the AC still holds |
| 2.3 Fleet State Suggestion Engine | Needs Fixes → battery/lock-state gap fixed (`554832a`); gRPC-vs-REST transport mismatch still stands |
| 2.4 Flight Path Grid Generator & Geofence Guard | Needs Fixes → NFZ edge-check gap fixed (`28c53dc`); FR-6 FOV spacing and FR-8 terrain clearance still absent |
| 3.1 Hub Interlock & Takeoff Sequence | Partially Done → hub-doors state now updates on launch (`6bece54`); MQTT subscriber leak and no launch-time weather/geofence re-check still open |
| 3.2 1Hz Telemetry & Operator Mutex Lease | Needs Fixes → heartbeat-timeout position jump fixed (`0da05f6`); AC3's "cache telemetry in Redis" still entirely unimplemented |
| 3.3 WebRTC Video Stream Transcoding | Needs Fixes (placeholder) |
| 3.4 Manual Override Long-Press Controls | Needs Fixes → Pause no-op fixed (`0da05f6`); no `Land` command still missing |
| 3.5 Automatic Landing & Recharging | Partially Done → recharging→closed reset fixed (`6bece54`); dangling `pulse` animation still unaddressed |
| 3.6 Safety Alert HUD Indicator | Needs Fixes → fake weather-breach trigger fixed (`a4c51f6`); missing flash animations still unaddressed |
| 4.1 Synced Data Archiving | Partially Done |
| 4.2 Mission History Dashboard | Partially Done / Needs Fixes |
| 4.3 Interactive Timeline Scrubber & Map Sync | Partially Done / Needs Fixes |

The 9 critical, safety-misrepresenting bugs (items 1-9 above) were fixed the same day as this review, each verified with new automated tests plus a live end-to-end run against the real gateway. No story was moved to "done" in `sprint-status.yaml` - each still has other, less severe gaps noted in this doc that remain open.
