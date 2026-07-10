# Story 4.3: Interactive Timeline Scrubber & Map Sync

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an Operator,  
I want my video playback to be synchronized with the map coordinates during replay,  
so that scrubbing the timeline shows me the exact spot where that video segment was recorded.

## Acceptance Criteria

1. **Given** the dashboard is in Replay Mode.
2. **When** the operator drags the bottom timeline slider.
3. **Then** the video player jumps to the matching frame, and the drone map marker moves along the drawn trajectory path to show the coordinates at that millisecond index (and vice versa).

## Tasks / Subtasks

- [x] Task 1: React Replay Timeline Slider Control (AC: 2)
  - [x] Subtask 1.1: Render a styled Slider bar panel at the bottom center of the viewport (Midnight Ocean theme steel blue background) when `replayMode` is active.
  - [x] Subtask 1.2: Set the slider min/max bounds to map index counts (`0` to `telemetry.length - 1`). Wire state controls `replayIndex` to track scrubbing changes.
- [x] Task 2: Playback Map Marker Sync (AC: 3)
  - [x] Subtask 2.1: Render a historical replay drone marker (using color Neon Cyan `#00e5ff`) on Leaflet at the coordinates corresponding to the active `replayIndex`.
  - [x] Subtask 2.2: Update the marker position dynamically in real-time as the slider values change.
- [x] Task 3: Video HUD Playout Sync (AC: 3)
  - [x] Subtask 3.1: Pass the selected `replayIndex` and historical telemetry frames into `<VideoPlayer />`.
  - [x] Subtask 3.2: Update the Video HUD canvas rendering loop: if `replayMode` is active, render the static HUD telemetry frame corresponding to the index (altitude, battery, speed, coordinates) alongside a red warning label: `⚠️ REPLAY PLAYBACK [MM:SS]`.
- [x] Task 4: Play/Pause Auto-Scrubber Actions (AC: 2)
  - [x] Subtask 4.1: Render Play/Pause icon buttons next to the slider. When active, auto-advance `replayIndex` at 1 Hz intervals to simulate active flight replay.

## Dev Notes

- **Language & Frameworks:** JavaScript (React UI with MUI/Leaflet).
- **Source Paths to Create/Modify:**
  - Modify: `frontend/src/App.jsx` (Add timeline slider controls, play/pause ticks, and map marker syncs)
  - Modify: `frontend/src/components/VideoPlayer.jsx` (Render historical playback frame overlays on canvas)
- **Previous Learnings Integration:**
  - Playout slider and player alerts must align with the Steel Blue (`#172a45`) and Neon Cyan (`#00e5ff`) visual palettes.

### References

- [Source: _bmad-output/planning-artifacts/epics.md#### Story 4.3: Interactive Timeline Scrubber & Map Sync]

## Dev Agent Record

### Agent Model Used

gemini-2.5-pro

### Debug Log References
- /home/bangnq/.gemini/antigravity-cli/brain/b29dd0a5-8fa1-4047-b14d-171d279485af/.system_generated/logs/transcript.jsonl

### Completion Notes List
- Developed Bottom Timeline Scrubber Slider panel floated over the Leaflet Map container using Midnight Ocean theme components.
- Coded active play/pause scrubber triggers auto-advancing playback indices at 1 Hz intervals.
- Integrated Leaflet Map marker syncs mapping Neon Cyan (`#00e5ff`) circle pins dynamically along historical paths on timeline change.
- Wired HTML5 canvas video HUD rendering to display static parameter logs (altitude, battery, coordinates) alongside flashing Replay Playback banners.

### File List
- frontend/src/App.jsx
- frontend/src/components/VideoPlayer.jsx
