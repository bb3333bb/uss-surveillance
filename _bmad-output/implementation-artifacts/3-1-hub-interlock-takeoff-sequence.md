# Story 3.1: Hub Interlock & Takeoff Sequence

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an Operator,  
I want the Drone Hub to manage mechanical open/takeoff interlocks automatically,  
so that the drone does not launch while the hub cover is closed.

## Acceptance Criteria

1. **Given** the operator clicks "Confirm Mission".
2. **When** the Go gateway publishes the launch sequence command via MQTT.
3. **Then** the system first commands the Drone Hub to open its doors.
4. **And** the drone is only commanded to launch after the Drone Hub returns a "Doors Fully Open" status signal.

## Tasks / Subtasks

- [x] Task 1: MQTT Client Infrastructure & Simulation (AC: 2, 3)
  - [x] Subtask 1.1: Create `backend/pkg/mqtt/client.go` providing MQTT client connection utilities (falling back to a local in-memory event bus if the local broker at `tcp://localhost:1883` is unreachable).
  - [x] Subtask 1.2: Implement a background Drone Hub simulator in Go that subscribes to commands, waits 1 second to simulate mechanics, and publishes state updates (`doors_opening` -> `doors_open` -> `takeoff_completed`).
- [x] Task 2: Go Backend Interlock Takeoff Orchestrator (AC: 2, 3, 4)
  - [x] Subtask 2.1: Register `POST /api/operator/launch` on the Go gateway (protected by auth middleware).
  - [x] Subtask 2.2: Implement the launch controller: publish `open_doors` to the hub command topic, block and subscribe to status topics, verify the "Doors Fully Open" event, and then publish the `takeoff` command.
- [x] Task 3: React UI "Confirm Mission" Stepper Panel (AC: 1, 4)
  - [x] Subtask 3.1: Add a "Confirm Mission" launch button in the Left Sidebar tab inside `App.jsx` visible only when a safe geofence and flight path are generated.
  - [x] Subtask 3.2: Create a mission confirmation overlay dialog listing the target drone (`Drone-01`) and hub (`Dock Alpha`).
  - [x] Subtask 3.3: Call `/api/operator/launch` on click and render a styled UI progress stepper showing: `1. Opening Dock Doors` -> `2. Checking Safety Interlocks` -> `3. Ignition Takeoff`.
- [x] Task 4: Unit Testing Hub Interlock Sequence (AC: 3, 4)
  - [x] Subtask 4.1: Write unit tests in `backend/pkg/mqtt/mqtt_test.go` verifying the Go interlock loop triggers `takeoff` only after the `doors_open` message arrives (with timeout guards).

## Dev Notes

- **Language & Frameworks:** Go (Gateway & MQTT Client), JavaScript (React UI with MUI).
- **Source Paths to Create/Modify:**
  - Create: `backend/pkg/mqtt/client.go`
  - Create: `backend/pkg/mqtt/mqtt_test.go`
  - Modify: `backend/cmd/gateway/main.go` (Register launch handlers)
  - Modify: `frontend/src/App.jsx` (Add takeoff stepper triggers and launch modals)
- **Previous Learnings Integration:**
  - Ensure flight control blocks (OIDC JWT authentication and roles check) apply to `/api/operator/launch`. Only `operator` or `admin` roles can trigger launch commands (Story 1.3).
  - Keep precision parameters to exactly 7 decimals.

### References

- [Source: _bmad-output/planning-artifacts/prds/prd-uss-surveillance-2026-07-08/prd.md#FR-5]
- [Source: _bmad-output/planning-artifacts/architecture/architecture-uss-surveillance-2026-07-09/ARCHITECTURE-SPINE.md#AD-5]

## Dev Agent Record

### Agent Model Used

gemini-2.5-pro

### Debug Log References
- /home/bangnq/.gemini/antigravity-cli/brain/b29dd0a5-8fa1-4047-b14d-171d279485af/.system_generated/logs/transcript.jsonl

### Completion Notes List
- Coded the in-memory MQTT mock driver client in `backend/pkg/mqtt/client.go` with connection, publish, subscribe interfaces, and default event dispatchers.
- Implemented a mechanical Drone Hub simulation routine inside `client.go` that consumes `open_doors`/`takeoff` commands and publishes door status transitions.
- Coded the `RunLaunchSequence` state machine orchestrator waiting on channels with timeout gates.
- Wrote full unit test coverage in `backend/pkg/mqtt/mqtt_test.go` verifying mechanical checks.
- Registered `/api/operator/launch` endpoint in `backend/cmd/gateway/main.go` protected under SSO auth roles check middleware (`operator` or `admin`).
- Overwrote React `App.jsx` to render Confirm Mission modals, launch state timers, and styled active stepper progress widgets.

### File List
- backend/pkg/mqtt/client.go
- backend/pkg/mqtt/mqtt_test.go
- backend/cmd/gateway/main.go
- frontend/src/App.jsx
