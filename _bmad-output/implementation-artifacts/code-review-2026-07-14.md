# Code Review: All 16 Stories (Epics 1–4)

**Date:** 2026-07-14
**Method:** Adapted from `.claude/commands/BMad/tasks/review-story.md` (the full BMad QA-gate workflow needs `.bmad-core/` config that isn't vendored into this repo — see README.md — so this substitutes a direct acceptance-criteria-vs-code review, written up per story, without the gate YAML/quality-score machinery). Each epic was reviewed by an independent agent against the actual current code, not against story-file checkboxes or `sprint-status.yaml`, both of which were found to be stale in places (see notes below).

Nothing was fixed as part of this review pass — this is findings only. See the end of this doc for what actually happened next.

## Severity-ranked findings

### Critical — misrepresents safety-critical state to the operator

1. **"Pause Flight" is a no-op.** `commandHandler` (`backend/cmd/gateway/main.go`) publishes MQTT `hover`, but nothing subscribes to it and the handler's own state-mutation block only branches on `rth`. The drone keeps flying after an operator holds Pause for 1.5s and the UI reports success. (Story 3.4)
2. **`WEATHER_BREACH_WIND`/`WEATHER_BREACH_RAIN` alerts are fake.** They fire off `lat > 10.77`, completely disconnected from `backend/pkg/weather`. Presented to the operator as a weather-driven safety alert; it isn't. (Story 3.6)
3. **NFZ/geofence check only inspects the drawn boundary's vertices**, never the generated flight path's interior or edges. A polygon whose vertices sit outside the 800m radius but whose edge clips through it is accepted as safe. (Story 2.4)
4. **The suggestion engine (FR-5) is a fully static stub.** Ignores battery, lock state, and the actual request coordinates; always returns the same drone/dock. Not "state-based recommendation" in any real sense. (Story 2.3)
5. **Live weather safety check ignores precipitation.** Only wind speed feeds `Safe`; a live report of heavy rain/thunderstorm with acceptable wind is marked safe, contradicting the stated "wind OR heavy precipitation" requirement. (Story 2.2)
6. **Hardcoded default `JWT_SECRET` committed to source**, combined with an implicit, undocumented mock-auth fallback: any environment where `OIDC_ISSUER_URL` is unset or the IdP is briefly unreachable at boot silently serves an unauthenticated "operator" JWT to anyone hitting `/api/auth/login`, signed with a secret visible in the repo. No explicit `AUTH_MOCK_MODE` flag gates this. (Story 1.1)
7. **Real OIDC callback flow is broken.** `HandleCallback` writes a raw JSON body instead of redirecting back to the frontend with the token — only the mock-mode branch in `HandleLogin` actually works today. This means the SSO flow the epic is named after cannot complete end-to-end once a real IdP is configured, and nothing currently exercises that path to catch it. (Story 1.1)
8. **Hub-doors telemetry state is inaccurate for the entire duration of a flight.** Never set on launch (stays `"closed"` while the drone is airborne) and never resets out of `"recharging"` afterward. (Stories 3.1, 3.5)
9. **Heartbeat-timeout "pause" jumps the drone's displayed position back to the dock** rather than holding its last position — inconsistent with actual hover semantics and with the (also broken) manual pause command. (Story 3.2)

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
| 1.1 OIDC Authentication & JWT Validation | Needs Fixes |
| 1.2 Base 3-Column Dashboard Layout & Themes | Partially Done |
| 1.3 Role-Based Access Control Enforcer | Partially Done |
| 2.1 Polygon Drawing on 2D Map View | Done |
| 2.2 Automated Weather and Wind Assessment | Needs Fixes |
| 2.3 Fleet State Suggestion Engine | Needs Fixes |
| 2.4 Flight Path Grid Generator & Geofence Guard | Needs Fixes |
| 3.1 Hub Interlock & Takeoff Sequence | Partially Done |
| 3.2 1Hz Telemetry & Operator Mutex Lease | Needs Fixes |
| 3.3 WebRTC Video Stream Transcoding | Needs Fixes (placeholder) |
| 3.4 Manual Override Long-Press Controls | Needs Fixes |
| 3.5 Automatic Landing & Recharging | Partially Done |
| 3.6 Safety Alert HUD Indicator | Needs Fixes |
| 4.1 Synced Data Archiving | Partially Done |
| 4.2 Mission History Dashboard | Partially Done / Needs Fixes |
| 4.3 Interactive Timeline Scrubber & Map Sync | Partially Done / Needs Fixes |

None moved to "done" as part of this pass — see sprint-status.yaml, left unchanged pending a decision on what to fix first.
