# Story 3.6: Safety Alert HUD Indicator

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an Operator,  
I want the HUD and Map borders to flash neon red when weather limits, geofences, or leases are breached during a flight,  
so that I am instantly notified of safety hazards and can take manual command.

## Acceptance Criteria

1. **Given** a telemetry payload streams to the browser.
2. **When** the payload contains safety alerts (weather threshold breach, geofence boundary warning, or OIDC role lease violations).
3. **Then** the UI flashes a red warning bar on the video player HUD and map border (Midnight Ocean Theme Neon Red `#ff3d00`).

## Tasks / Subtasks

- [x] Task 1: Go Gateway Telemetry Alert Annotator (AC: 2)
  - [x] Subtask 1.1: Add `alerts []string` variable arrays inside the WebSockets JSON payload structure.
  - [x] Subtask 1.2: Check safety parameters dynamically on every broadcast tick: append warning codes (`WEATHER_BREACH_WIND`, `WEATHER_BREACH_RAIN`, `RESTRICTED_AIRSPACE`, `LEASE_CONFLICT`) based on weather client state, planning client safety status, and active leaseholder mismatch checks.
- [x] Task 2: React HUD Flashing Border Overlay (AC: 3)
  - [x] Subtask 2.1: Add CSS animation rules to `App.jsx` styled blocks for neon red flashing keyframe pulses (`#ff3d00`).
  - [x] Subtask 2.2: Bind conditional styled borders around the Map panel container and Video player card container when active alerts are received.
- [x] Task 3: Video Canvas Simulation Overlay Alerts (AC: 1, 3)
  - [x] Subtask 3.1: Pass the alerts array down to the `<VideoPlayer />` component.
  - [x] Subtask 3.2: Update the HTML5 Canvas rendering loop to overlay a neon red flashing banner at the top of the HUD stating: `⚠️ WARNING: ALERT STATUS ACTIVE` with sub-details outlining the active codes.
- [x] Task 4: Go Alerts Generator Unit Testing (AC: 2)
  - [x] Subtask 4.1: Write unit tests verifying Go telemetry builders append the correct warning string arrays under threshold violations.

## Dev Notes

- **Language & Frameworks:** Go (Gateway), JavaScript (React UI with MUI).
- **Source Paths to Create/Modify:**
  - Modify: `backend/cmd/gateway/main.go` (Inspect weather/geofence alerts and package in WebSocket payloads)
  - Modify: `frontend/src/App.jsx` (Add flashing alert CSS styles and apply borders conditionally)
  - Modify: `frontend/src/components/VideoPlayer.jsx` (Render red warning banners and descriptions on Canvas)
- **Previous Learnings Integration:**
  - Alerts color must use Midnight Ocean theme Neon Red (`#ff3d00`).

### References

- [Source: _bmad-output/planning-artifacts/prds/prd-uss-surveillance-2026-07-08/prd.md#FR-8]
- [Source: _bmad-output/planning-artifacts/architecture/architecture-uss-surveillance-2026-07-09/ARCHITECTURE-SPINE.md#AD-5]

## Dev Agent Record

### Agent Model Used

gemini-2.5-pro

### Debug Log References
- /home/bangnq/.gemini/antigravity-cli/brain/b29dd0a5-8fa1-4047-b14d-171d279485af/.system_generated/logs/transcript.jsonl

### Completion Notes List
- Coded active warning monitors appending `WEATHER_BREACH_WIND`, `WEATHER_BREACH_RAIN`, `RESTRICTED_AIRSPACE`, and `LEASE_CONFLICT` alerts inside the Go gateway.
- Added dynamic safety warnings and flashing red CSS borders (`#ff3d00`) around the 2D Map canvas and Video Player container card.
- Updated HTML5 Canvas camera feeds to render neon red flashing warning banners outlining active warning codes.

### File List
- backend/cmd/gateway/main.go
- frontend/src/App.jsx
- frontend/src/components/VideoPlayer.jsx
