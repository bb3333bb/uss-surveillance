# Story 4.2: Mission History Dashboard

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an Operator,  
I want to view a list of past missions on my left sidebar drawer,  
so that I can select a specific patrol log to audit.

## Acceptance Criteria

1. **Given** the operator clicks the "Mission Logs" tab.
2. **When** the sidebar tab is loaded.
3. **Then** the system queries database endpoints and lists completed missions with date, duration, and drone details.
4. **When** the operator selects a mission and clicks "View Replay", the dashboard transitions into Replay Mode.

## Tasks / Subtasks

- [x] Task 1: Go Gateway Mission Logs API Endpoint (AC: 3)
  - [x] Subtask 1.1: Register GET `/api/operator/missions` route in `backend/cmd/gateway/main.go` under operator role verification wrapper checks.
  - [x] Subtask 1.2: Read and parse records from the file database `backend/data/missions.json`, returning them sorted chronologically (newest first).
- [x] Task 2: React Sidebar History Tab Panels (AC: 1, 2, 3)
  - [x] Subtask 2.1: In `App.jsx`, fetch from `/api/operator/missions` whenever the "History" tab is active.
  - [x] Subtask 2.2: Render a scrollable list of completed missions in the sidebar, displaying Date/Time, Duration (formatted as MM:SS), and a "View Replay" action button.
- [x] Task 3: Replay Mode State Transitions (AC: 4)
  - [x] Subtask 3.1: Add `replayMode` and `activeReplay` state controls to `App.jsx`.
  - [x] Subtask 3.2: When Replay is engaged: clear active WebSocket map pins, display the complete static flight path of the selected mission on Leaflet, and render a Neon Cyan banner alert declaring: `⚠️ VIEWING HISTORICAL REPLAY: [MISSION_ID]`. Add an "Exit Replay" button to restore live operation widgets.
- [x] Task 4: Missions Endpoint Unit Testing (AC: 3)
  - [x] Subtask 4.1: Write Go tests verifying that the GET `/api/operator/missions` handler responds correctly with valid JSON arrays.

## Dev Notes

- **Language & Frameworks:** Go (Gateway), JavaScript (React UI with MUI).
- **Source Paths to Create/Modify:**
  - Modify: `backend/cmd/gateway/main.go` (Register `/api/operator/missions` routes and handlers)
  - Modify: `frontend/src/App.jsx` (Wire sidebar History tabs, Replay state banners, and Leaflet static trajectory lines)
- **Previous Learnings Integration:**
  - Ensure duration values format correctly as human-readable MM:SS (e.g. `05:12`) to keep dashboards polished.

### References

- [Source: _bmad-output/planning-artifacts/epics.md#### Story 4.2: Mission History Dashboard]

## Dev Agent Record

### Agent Model Used

gemini-2.5-pro

### Debug Log References
- /home/bangnq/.gemini/antigravity-cli/brain/b29dd0a5-8fa1-4047-b14d-171d279485af/.system_generated/logs/transcript.jsonl

### Completion Notes List
- Registered GET `/api/operator/missions` in `main.go` delivering reversed chronological arrays of completed mission logs.
- Coded HTTP retrieval logic unmarshaling missions lists from standard flat-file databases.
- Updated React `App.jsx` history drawer panel to fetch mission histories, displaying dates, durations, and replay triggers.
- Built interactive Replay Mode state overlays, rendering static Leaflet flight routes and floating neon cyan header alerts on selection.
- Appended unit tests in `archiver_test.go` checking mission list reversals.

### File List
- backend/cmd/gateway/main.go
- backend/pkg/archive/archiver_test.go
- frontend/src/App.jsx
