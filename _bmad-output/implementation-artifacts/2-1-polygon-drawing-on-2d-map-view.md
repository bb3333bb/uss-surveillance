# Story 2.1: Polygon Drawing on 2D Map View

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an Operator,  
I want to draw a patrol boundary directly on a 2D map,  
so that I can define the geographic area to be monitored.

## Acceptance Criteria

1. **Given** the operator is viewing the active 2D Map View.
2. **When** they select "Draw Area" and click vertices on the map.
3. **Then** the UI draws a dashed polygon boundary following their cursor points.
4. **And** closing the polygon outputs the array of WGS84 coordinates to the client state.

## Tasks / Subtasks

- [x] Task 1: Leaflet Map Integration (AC: 1)
  - [x] Subtask 1.1: Install `leaflet` and `react-leaflet` package in the `frontend` project.
  - [x] Subtask 1.2: Import Leaflet's stylesheet (`leaflet/dist/leaflet.css`) into `frontend/src/main.jsx` to resolve tile layout errors.
  - [x] Subtask 1.3: Render a basic 2D Leaflet Map inside the Center Map Panel of `App.jsx`, loading free OpenStreetMap base tiles.
- [x] Task 2: Drawing Control Handlers (AC: 2)
  - [x] Subtask 2.1: Render a "Draw Area" button overlay in the top-left corner of the map viewport.
  - [x] Subtask 2.2: Bind click actions to toggle "Drawing Mode" on the Leaflet map instance.
- [x] Task 3: Interactive Polygon Drawing & Output (AC: 2, 3, 4)
  - [x] Subtask 3.1: In drawing mode, clicking the map adds coordinate vertices, rendering temporary markers and dashed connecting lines.
  - [x] Subtask 3.2: Render a dynamic dashed line connecting the last vertex to the operator's current mouse position.
  - [x] Subtask 3.3: Clicking the first vertex closes the polygon, renders a solid border line polygon, exits drawing mode, and logs the list of WGS84 lat/lng coordinates (with exactly 7 decimal places) to the React console state.
  - [x] Subtask 3.4: Render a "Clear Area" button on the map overlay if a closed polygon exists, which clears the map layer and resets coordinates.

## Dev Notes

- **Language & Frameworks:** JavaScript (React UI with MUI).
- **Libraries Recommended:**
  - Leaflet: `leaflet` and `react-leaflet` or use Vanilla Leaflet mapping objects directly on a React `ref` div container.
- **Database/Entity Impact:** None. All interactions are contained in client state variables.
- **Source Paths to Modify:**
  - Modify: `frontend/package.json` (Inject leaflet packages)
  - Modify: `frontend/src/App.jsx` (Integrate Leaflet map canvas inside the center panel)
- **Previous Learnings Integration:**
  - Ensure the "Draw Area" map overlay is hidden if `role === 'viewer'` (checked in Story 1.3).
  - Telemetry coordinates are WGS84 EPSG:4326. Ensure precision uses exactly 7 decimal places (as specified in AD-6 and WGS84 precision rule).

### References

- [Source: _bmad-output/planning-artifacts/prds/prd-uss-surveillance-2026-07-08/prd.md#FR-1]
- [Source: _bmad-output/planning-artifacts/ux-designs/ux-uss-surveillance-2026-07-08/EXPERIENCE.md#MapControls]
- [Source: _bmad-output/planning-artifacts/architecture/architecture-uss-surveillance-2026-07-09/ARCHITECTURE-SPINE.md#AD-6]

## Dev Agent Record

### Agent Model Used

gemini-2.5-pro

### Debug Log References
- /home/bangnq/.gemini/antigravity-cli/brain/b29dd0a5-8fa1-4047-b14d-171d279485af/.system_generated/logs/transcript.jsonl

### Completion Notes List
- Installed Leaflet packages in the React app.
- Imported Leaflet CSS in main.jsx to resolve broken tile layouts.
- Integrated Vanilla Leaflet instance on a React DOM ref inside the center panel in App.jsx.
- Coded custom drawing controls: "Draw Area" toggles drawing states; "Clear Area" resets polygon vertices.
- Built interactive point marker drawing with circular styled markers, dashed polylines, and dynamic mouse tracking lines.
- Implemented click handlers on the origin marker to close the shape, rendering a filled solid border polygon.
- Captured polygon coordinate outputs as WGS84 floats with exactly 7 decimals precision, printing them to the developer console and rendering them in the left sidebar geofence details list.
- Verified error-free client compilation in production build.

### File List
- frontend/package.json
- frontend/src/main.jsx
- frontend/src/App.jsx
