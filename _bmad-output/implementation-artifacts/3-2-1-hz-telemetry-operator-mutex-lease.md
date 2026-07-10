# Story 3.2: 1 Hz Telemetry & Operator Mutex Lease

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an Operator,  
I want to see live drone parameters refreshed at 1 Hz and maintain an exclusive control lease,  
so that I have full situational awareness and prevent conflicting commands from other users.

## Acceptance Criteria

1. **Given** a drone is flying.
2. **When** telemetry streams from the drone over MQTT.
3. **Then** the Go gateway caches the state in Redis and streams it at 1 Hz to the browser via WebSockets.
4. **And** the Go gateway secures an exclusive command mutex lock for the initiating operator in Redis.
5. **And** automatically pauses the drone if the operator's browser heartbeat ceases for >10 seconds.

## Tasks / Subtasks

- [x] Task 1: Go Backend Redis Lease Manager (AC: 4)
  - [x] Subtask 1.1: Create `backend/pkg/lease/manager.go` implementing Redis lock leases (with in-memory fallback mappings if Redis is unreachable).
  - [x] Subtask 1.2: Coder lease acquisition handlers: `AcquireLease(operator string, droneID string, duration time.Duration) bool` and `ReleaseLease(droneID string)`.
- [x] Task 2: 1 Hz WebSocket Telemetry Broadcaster (AC: 2, 3)
  - [x] Subtask 2.1: Register `/api/operator/telemetry` WebSocket handler on the Go gateway.
  - [x] Subtask 2.2: Implement 1 Hz background subscription routine reading telemetry coordinates from the MQTT broker simulator and broadcasting to WebSocket clients.
- [x] Task 3: Connection Heartbeat Watchdog - AD-5 (AC: 5)
  - [x] Subtask 3.1: Code connection watchdogs inside the WebSocket reader loop: track incoming pings.
  - [x] Subtask 3.2: If no heartbeats are received from the active leaseholder for >10 seconds, release their control lease and publish a Pause/Hover command (`drone/hub/command` -> `hover`).
- [x] Task 4: React UI Telemetry & Heartbeat Feeds (AC: 1, 3)
  - [x] Subtask 4.1: Connect WebSocket clients in `App.jsx` on SSO login.
  - [x] Subtask 4.2: Feed live coordinates, battery percentage, altitude, and wind speeds to the telemetry sidebar panel and update drone leaflet marker paths.
  - [x] Subtask 4.3: Dispatch recurring heartbeat pings from the browser client to the gateway.
- [x] Task 5: Lease & Telemetry Unit Testing (AC: 4, 5)
  - [x] Subtask 5.1: Write unit tests in `backend/pkg/lease/lease_test.go` verifying lease exclusivity and automatic expirations.

## Dev Notes

- **Language & Frameworks:** Go (Gateway & WebSockets), JavaScript (React UI).
- **Source Paths to Create/Modify:**
  - Create: `backend/pkg/lease/manager.go`
  - Create: `backend/pkg/lease/lease_test.go`
  - Modify: `backend/cmd/gateway/main.go` (WebSocket registration and handlers)
  - Modify: `frontend/src/App.jsx` (WS clients, sidebars, and markers)
- **Previous Learnings Integration:**
  - Ensure WebSocket coordinate frames match WGS84 7 decimal precision rules (AD-6).
  - Ensure websocket connections support initial token check headers.

### References

- [Source: _bmad-output/planning-artifacts/prds/prd-uss-surveillance-2026-07-08/prd.md#FR-5]
- [Source: _bmad-output/planning-artifacts/architecture/architecture-uss-surveillance-2026-07-09/ARCHITECTURE-SPINE.md#AD-5]
- [Source: _bmad-output/planning-artifacts/architecture/architecture-uss-surveillance-2026-07-09/ARCHITECTURE-SPINE.md#AD-6]

## Dev Agent Record

### Agent Model Used

gemini-2.5-pro

### Debug Log References
- /home/bangnq/.gemini/antigravity-cli/brain/b29dd0a5-8fa1-4047-b14d-171d279485af/.system_generated/logs/transcript.jsonl

### Completion Notes List
- Programmed exclusive operator command leases in `backend/pkg/lease/manager.go` including concurrency mutex locks.
- Wrote full unit test coverage verifying lease locks in `backend/pkg/lease/lease_test.go`.
- Installed gorilla/websocket and registered `/api/operator/telemetry` WebSocket endpoint on the Go gateway.
- Coded 1 Hz background updates loop reading path positions and broadcasting JSON telemetry packets.
- Implemented AD-5 safety watchdog: if a connection heartbeat ping is missing for >10 seconds, eviction releases operator leases and issues an MQTT hover command.
- Overwrote React `App.jsx` to establish WS connections, send heartbeats, and render moving drone markers and lease locks status badges.

### File List
- backend/pkg/lease/manager.go
- backend/pkg/lease/lease_test.go
- backend/cmd/gateway/main.go
- frontend/src/App.jsx
