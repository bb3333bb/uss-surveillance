# Story 4.1: Synced Data Archiving

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an Operator,  
I want the system to save flight paths and video recordings automatically,  
so that past missions can be analyzed later.

## Acceptance Criteria

1. **Given** a mission has completed.
2. **When** the drone docks.
3. **Then** the Go gateway compiles the 1 Hz telemetry logs into a JSON log file (indexed from `t=0` relative to takeoff) and stores it in a persistent database alongside the MP4 video recording reference.

## Tasks / Subtasks

- [x] Task 1: Go Backend Telemetry Archiver Package (AC: 3)
  - [x] Subtask 1.1: Create `backend/pkg/archive/archiver.go` implementing thread-safe telemetry coordinate buffers.
  - [x] Subtask 1.2: Capture telemetry points at 1 Hz intervals when `IsFlying == true`, indexing points relative to takeoff time (`t=0`).
- [x] Task 2: Go Backend Mission Persistent Database (AC: 3)
  - [x] Subtask 2.1: Setup a JSON/flat-file database in `backend/data/missions.json` to store mission records (representing PostgreSQL rows).
  - [x] Subtask 2.2: Define the schema: mission ID, start timestamp, end timestamp, duration, raw telemetry log array, and simulated MP4 video path link.
- [x] Task 3: Mission Completion Landing Trigger (AC: 1, 2, 3)
  - [x] Subtask 3.1: In the background simulation ticker (Story 3.5), trigger the archival compilation when `IsFlying` transitions to false on precision docking alignment.
  - [x] Subtask 3.2: Clear active flight buffers post-write to reset the tracker for the next launch.
- [x] Task 4: Unit Testing Telemetry Archiving (AC: 3)
  - [x] Subtask 4.1: Write unit tests verifying that telemetry points index from `t=0` and persist correct coordinate payloads.

## Dev Notes

- **Language & Frameworks:** Go (Gateway).
- **Source Paths to Create/Modify:**
  - Create: `backend/pkg/archive/archiver.go` (Telemetry collector and file writer)
  - Create: `backend/pkg/archive/archiver_test.go` (Time-index unit validation tests)
  - Modify: `backend/cmd/gateway/main.go` (Trigger archiver on landing sequence transitions)
- **Previous Learnings Integration:**
  - Ensure all database file writes use mutex locks to prevent race conditions during parallel client connections.

### References

- [Source: _bmad-output/planning-artifacts/epics.md#### Story 4.1: Synced Data Archiving]

## Dev Agent Record

### Agent Model Used

gemini-2.5-pro

### Debug Log References
- /home/bangnq/.gemini/antigravity-cli/brain/b29dd0a5-8fa1-4047-b14d-171d279485af/.system_generated/logs/transcript.jsonl

### Completion Notes List
- Programmed Go thread-safe telemetry archiver mapping point buffers relative to takeoff time (`t=0`).
- Implemented file database serialization structure in `backend/data/missions.json` tracking start/end times, durations, and logs.
- Wired precision landing alignment and manual return-to-home overrides triggers to persist mission recordings.
- Created unit tests verifying telemetry parsing, time-indexing, and directory/JSON writes.

### File List
- backend/pkg/archive/archiver.go
- backend/pkg/archive/archiver_test.go
- backend/cmd/gateway/main.go
