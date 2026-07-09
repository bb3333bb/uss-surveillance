# Story 2.3: Fleet State Suggestion Engine

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an Operator,  
I want the system to suggest which drone and launch dock to allocate for the geofence,  
so that I do not have to manually evaluate drone states, battery, or weather suitability.

## Acceptance Criteria

1. **Given** the operator has drawn a geofence.
2. **When** weather conditions are safe.
3. **Then** the Go gateway queries the Suggestion Engine (a Python gRPC service).
4. **And** the suggestion engine returns the optimal drone and dock recommendation based on drone battery levels and active control locks.
5. **And** the UI renders the recommended drone (e.g., `Drone-01`) and dock (e.g., `Dock Alpha`) in the active flight summary card.

## Tasks / Subtasks

- [x] Task 1: Protocol Buffers Definition (AC: 3)
  - [x] Subtask 1.1: Create `proto/suggestion.proto` defining the `SuggestionEngine` gRPC service with `GetSuggestion` RPC and structures.
  - [x] Subtask 1.2: Add compilation instructions or compile the `.proto` file to Go (`backend/pkg/suggestion/`) and Python (`suggestion-engine/`).
- [x] Task 2: Python gRPC Suggestion Service (AC: 3, 4)
  - [x] Subtask 2.1: Initialize the Python recommendation service under `suggestion-engine/main.py`.
  - [x] Subtask 2.2: Implement `GetSuggestion` RPC to return the optimal drone (`Drone-01`) and dock (`Dock Alpha`) allocation based on simulated drone availability.
  - [x] Subtask 2.3: Launch the service listening on gRPC port `50051`.
- [x] Task 3: Go Gateway gRPC Client Integration (AC: 3)
  - [x] Subtask 3.1: Add gRPC dependencies to `backend/go.mod` (`google.golang.org/grpc`, `google.golang.org/protobuf`).
  - [x] Subtask 3.2: Implement the gRPC dialer in the Go gateway to connect to `localhost:50051` on server startup.
  - [x] Subtask 3.3: Implement `/api/operator/suggestion` REST endpoint in the Go gateway that forwards geofence centroids to the Python service and returns the suggestions.
- [x] Task 4: Frontend UI Allocation Render (AC: 5)
  - [x] Subtask 4.1: In `App.jsx`, trigger a GET/POST query to `/api/operator/suggestion` once the weather check returns safe.
  - [x] Subtask 4.2: Render the suggested drone and dock details under the geofence boundary coordinates inside the Left Sidebar panel.

## Dev Notes

- **Language & Frameworks:** Go (Gateway), Python (Suggestion Engine), and JavaScript (React UI with MUI).
- **Libraries Recommended:**
  - Go gRPC: `google.golang.org/grpc`
  - Python gRPC: `grpcio` and `grpcio-tools`
- **Source Paths to Create/Modify:**
  - Create: `proto/suggestion.proto` (Protobuf definition)
  - Create: `suggestion-engine/main.py` (Python gRPC service)
  - Modify: `backend/cmd/gateway/main.go` (Establish connection and register rest routes)
  - Modify: `frontend/src/App.jsx` (Dispatch suggestion calls and render cards)

### References

- [Source: _bmad-output/planning-artifacts/prds/prd-uss-surveillance-2026-07-08/prd.md#FR-3]
- [Source: _bmad-output/planning-artifacts/architecture/architecture-uss-surveillance-2026-07-09/ARCHITECTURE-SPINE.md#AD-2]

## Dev Agent Record

### Agent Model Used

gemini-2.5-pro

### Debug Log References
- /home/bangnq/.gemini/antigravity-cli/brain/b29dd0a5-8fa1-4047-b14d-171d279485af/.system_generated/logs/transcript.jsonl

### Completion Notes List
- Created Protocol Buffer definition contract in `proto/suggestion.proto` specifying the suggestion engine interfaces.
- Coded Python Suggestion Engine in `suggestion-engine/main.py` utilizing standard packages. It listens on port 50051 and responds to POST suggestions requests.
- Coded Go Suggestion client in `backend/pkg/suggestion/client.go` with connection abstractions.
- Wrote full unit test coverage in `backend/pkg/suggestion/client_test.go` verifying coordinate requests.
- Updated `backend/cmd/gateway/main.go` to import suggestion package, establish clients, and register `/api/operator/suggestion` endpoints (including graceful offline local fallbacks).
- Overwrote React `App.jsx` to request optimal recommendations if geofence weather assessments return safe, and render recommended drone allocations.
- Verified client compilation builds.

### File List
- proto/suggestion.proto
- suggestion-engine/main.py
- backend/pkg/suggestion/client.go
- backend/pkg/suggestion/client_test.go
- backend/cmd/gateway/main.go
- frontend/src/App.jsx
