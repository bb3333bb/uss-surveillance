# Story 3.3: WebRTC Video Stream Transcoding

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an Operator,  
I want to watch the drone's live camera feed in ultra-low latency,  
so that I can inspect the perimeter in real time.

## Acceptance Criteria

1. **Given** the drone is streaming video.
2. **When** the feed (RTSP/RTMP) reaches the local SRS Media Server.
3. **Then** SRS transcodes the stream into WebRTC packets and plays it on the client UI video overlay widget with <200ms lag.

## Tasks / Subtasks

- [x] Task 1: SRS Media Server WebRTC SDK integration (AC: 2, 3)
  - [x] Subtask 1.1: Create `frontend/src/utils/srs.sdk.js` containing standard SDP peer exchange negotiation protocols (fetching WebRTC streams from `http://localhost:1985/rtc/v1/play/`).
  - [x] Subtask 1.2: Code fallback trigger blocks: if peer negotiations fail or timeout, fallback to play a simulated drone surveillance feed.
- [x] Task 2: React WebRTC Video Player Component (AC: 3)
  - [x] Subtask 2.1: Create `frontend/src/components/VideoPlayer.jsx` rendering the video tag and binding WebRTC stream tracks.
  - [x] Subtask 2.2: Implement HTML5 Canvas drawing loop for the simulated drone view when offline, featuring scrolling telemetry bars, target acquisition bounding boxes, scanning grids, and crosshairs.
- [x] Task 3: Sidebar Video Panel Playout (AC: 1, 3)
  - [x] Subtask 3.1: Replace the static placeholder card inside the Right Sidebar of `App.jsx` with the new `<VideoPlayer />` component.
  - [x] Subtask 3.2: Sync playout triggers with telemetry state: activate playout if `telemetry.is_flying` is true. If landed, show "WebRTC Playout Idle - Drone Docked".
- [x] Task 4: Unit Testing Media Integration (AC: 3)
  - [x] Subtask 4.1: Write unit tests in Go or frontend configurations verifying media routes are registered or handled gracefully.

## Dev Notes

- **Language & Frameworks:** JavaScript (React UI with MUI, Leaflet).
- **Source Paths to Create/Modify:**
  - Create: `frontend/src/utils/srs.sdk.js` (SRS WebRTC player SDK)
  - Create: `frontend/src/components/VideoPlayer.jsx` (Video player with canvas rendering fallbacks)
  - Modify: `frontend/src/App.jsx` (Embed video player inside Right Sidebar telemetry column)
- **Previous Learnings Integration:**
  - Playout must blend with Midnight Ocean dark mode rules.

### References

- [Source: _bmad-output/planning-artifacts/prds/prd-uss-surveillance-2026-07-08/prd.md#FR-6]
- [Source: _bmad-output/planning-artifacts/ux-designs/ux-uss-surveillance-2026-07-08/DESIGN.md#VideoOverlay]

## Dev Agent Record

### Agent Model Used

gemini-2.5-pro

### Debug Log References
- /home/bangnq/.gemini/antigravity-cli/brain/b29dd0a5-8fa1-4047-b14d-171d279485af/.system_generated/logs/transcript.jsonl

### Completion Notes List
- Created `frontend/src/utils/srs.sdk.js` implementing standard SDP exchange playout queries with SRS API.
- Created `frontend/src/components/VideoPlayer.jsx` rendering HTML5 Video element and fallback HTML5 Canvas elements.
- Coded canvas HUD telemetry trackers, tracking boxes, scanning grids, and overlay live badges.
- Embedded `<VideoPlayer />` in `App.jsx` right sidebar panel synced with telemetry flying flags.

### File List
- frontend/src/utils/srs.sdk.js
- frontend/src/components/VideoPlayer.jsx
- frontend/src/App.jsx
