# Story 2.4: Flight Path Grid Generator & Geofence Guard

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an Operator,  
I want the system to automatically generate a lawnmower flight path grid inside the geofence and check for restricted zone intersections,  
so that the drone can perform autonomous surveillance without entering forbidden airspace.

## Acceptance Criteria

1. **Given** the operator has drawn a geofence patrol boundary.
2. **When** the geofence allocation suggestion is accepted.
3. **Then** the Go gateway queries the Python Planner service to generate a flight path grid.
4. **And** the Python service returns a set of flight coordinates representing a lawnmower patrol grid inside the polygon.
5. **And** the Python service verifies the geofence vertices do not intersect a mock restricted air zone (No-Fly Zone). If it intersects, it returns a geofence warning.
6. **And** the UI renders the generated flight path grid lines inside the polygon on the Leaflet map.

## Tasks / Subtasks

- [x] Task 1: Go Backend Path Planning Rest Endpoint (AC: 3, 5)
  - [x] Subtask 1.1: Register `POST /api/operator/plan` on the Go gateway expecting JSON payload `{"vertices": [{"lat": float64, "lng": float64}, ...]}`.
  - [x] Subtask 1.2: Forward request payload to Python Suggestion & Planner service and return the calculated flight coordinates.
- [x] Task 2: Python Lawnmower Grid & No-Fly Zone Inspector (AC: 4, 5)
  - [x] Subtask 2.1: Add `POST /api/plan` handler in the Python Suggestion & Planner service.
  - [x] Subtask 2.2: Implement geofence safety checks: if any vertex lies within 800 meters of a simulated No-Fly Zone centered at `[10.7725, 106.69]`, reject with safety warning: `{"safe": false, "message": "Geofence intersects restricted airport airspace"}`.
  - [x] Subtask 2.3: Implement lawnmower grid generator: find bounding box of the geofence vertices, generate horizontal scanlines spaced by 0.0003 degrees, check containment of points inside the geofence polygon using ray-casting, and return the filtered path vertices.
- [x] Task 3: Leaflet Flight Path Polyline Rendering (AC: 6)
  - [x] Subtask 3.1: In `App.jsx`, query `/api/operator/plan` after geofence allocation suggestions complete successfully.
  - [x] Subtask 3.2: If geofence air safety check fails, render a red banner: `RESTRICTED AIRSPACE INTERSECTION` and lock/disable launch overrides.
  - [x] Subtask 3.3: If safe, render the path array as a solid neon cyan Leaflet polyline inside the geofence polygon.
- [x] Task 4: Unit Testing Path Planning & Geofence Checks (AC: 5)
  - [x] Subtask 4.1: Write unit tests in Go or Python verifying coordinate calculations and No-Fly Zone intersection alerts.

## Dev Notes

- **Language & Frameworks:** Go (Gateway), Python (Planner service), and JavaScript (React UI with MUI).
- **Source Paths to Modify:**
  - Modify: `suggestion-engine/main.py` (Add plan calculations and NFZ checks)
  - Modify: `backend/cmd/gateway/main.go` (Add REST planning handler proxying requests)
  - Modify: `frontend/src/App.jsx` (Dispatch path planning queries and render map polylines)
- **Previous Learnings Integration:**
  - Coordinate variables must utilize WGS84 projection (EPSG:4326) serialized as floats with exactly 7 decimal places (AD-6).
  - Geofence air safety alerts must disable flight overrides.

### References

- [Source: _bmad-output/planning-artifacts/prds/prd-uss-surveillance-2026-07-08/prd.md#FR-4]
- [Source: _bmad-output/planning-artifacts/architecture/architecture-uss-surveillance-2026-07-09/ARCHITECTURE-SPINE.md#AD-6]

## Dev Agent Record

### Agent Model Used

gemini-2.5-pro

### Debug Log References
- /home/bangnq/.gemini/antigravity-cli/brain/b29dd0a5-8fa1-4047-b14d-171d279485af/.system_generated/logs/transcript.jsonl

### Completion Notes List
- Implemented `POST /api/plan` handler in the Python Suggestion & Planner service.
- Programmed geofence safety inspections comparing vertices against a mock Airport No-Fly Zone centered at `[10.7725, 106.69]` using the Haversine formula.
- Programmed lawnmower sweep path calculations using polygon bounding boxes and ray-casting point-in-polygon containment checks. Alternated sweep directions on consecutive lines.
- Coded Go client `GetPlan` lookup calls in `backend/pkg/suggestion/client.go` with full mock test suites.
- Registered `/api/operator/plan` proxy endpoint in `backend/cmd/gateway/main.go`.
- Overwrote React `App.jsx` to fetch flight grids, toggle restricted airspace banners, and render solid cyan paths on Leaflet maps with Takeoff/Landing overlays.
- Verified compilation builds.

### File List
- suggestion-engine/main.py
- backend/pkg/suggestion/client.go
- backend/pkg/suggestion/client_test.go
- backend/cmd/gateway/main.go
- frontend/src/App.jsx
