# Story 1.2: Base 3-Column Dashboard Layout & Themes

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an Operator,  
I want to see the workspace organized in a high-density 3-column dark layout,  
so that I can observe telemetry, maps, and video with low eye fatigue.

## Acceptance Criteria

1. **Given** the operator is authenticated.
2. **When** the dashboard page renders.
3. **Then** the visual identity applies the Midnight Ocean theme (Deep Navy `#0a192f` background, Steel Blue `#172a45` panels, Ocean Blue `#00b0ff` borders).
4. **And** the layout displays the three defined drawers (Left Sidebar, Map Center Panel, Right Control Sidebar).

## Tasks / Subtasks

- [x] Task 1: MUI Theme Customization (AC: 3)
  - [x] Subtask 1.1: Create a shared theme configuration file `frontend/src/theme.js` containing custom color definitions for the Midnight Ocean theme (Deep Navy `#0a192f`, Steel Blue `#172a45`, Primary Blue `#00b0ff`, Secondary Cyan `#00e5ff`, Text primary `#e6f1ff`).
  - [x] Subtask 1.2: Set up custom font rules: Roboto for general layout and a strict Monospace font family for all numeric values (coordinates, telemetry metrics) to prevent layout shifting.
- [x] Task 2: 3-Column Grid Scaffolding (AC: 4)
  - [x] Subtask 2.1: Replace the basic landing container in `App.jsx` with a responsive 3-column Flex/Grid layout container.
  - [x] Subtask 2.2: Scaffolding dimensions: Left Sidebar (width 320px), Center Map Panel (flex-grow: 1), Right Sidebar (width 380px) with collapsible states.
- [x] Task 3: Left Sidebar Content Tabs (AC: 4)
  - [x] Subtask 3.1: Implement Tab handlers on the Left sidebar: Tab A ("Fleet & Docks") and Tab B ("Mission Logs").
  - [x] Subtask 3.2: Render a mock list of active Docks and drones within Tab A to prepare for telemetry ingestion.
- [x] Task 4: Right Sidebar Control Panel (AC: 3, 4)
  - [x] Subtask 4.1: Render placeholder cards in the Right Sidebar: Live Video overlay slot, active Telemetry grid, and manual override command slots (Pause, RTH, Land).
  - [x] Subtask 4.2: Ensure all manual command override buttons are styled in bright action colors (Warning `#ff9100`, Danger `#ff3d00`).

## Dev Notes

- **Language & Frameworks:** JavaScript (React UI with MUI).
- **Libraries Recommended:**
  - `@mui/material` and `@mui/icons-material` (already installed).
- **Database/Entity Impact:** None. All changes are frontend visual layout scaffolding.
- **Source Paths to Create/Modify:**
  - Create: `frontend/src/theme.js` (Theme configuration file)
  - Modify: `frontend/src/App.jsx` (Replace landing wrapper with 3-column Grid dashboard)
- **Previous Learnings Integration:**
  - Story 1.1 successfully integrated the OIDC AuthContext. The new dashboard components must check `isAuthenticated` from `useAuth()` and only render this dashboard layout once logged in.

### Project Structure Notes

- Use the shared MUI `ThemeProvider` to inject the custom theme at the root of `App.jsx` or `main.jsx`.

### References

- [Source: _bmad-output/planning-artifacts/ux-designs/ux-uss-surveillance-2026-07-08/DESIGN.md#ColorPalettes]
- [Source: _bmad-output/planning-artifacts/ux-designs/ux-uss-surveillance-2026-07-08/EXPERIENCE.md#LayoutGrid]

## Dev Agent Record

### Agent Model Used

gemini-2.5-pro

### Debug Log References
- /home/bangnq/.gemini/antigravity-cli/brain/b29dd0a5-8fa1-4047-b14d-171d279485af/.system_generated/logs/transcript.jsonl

### Completion Notes List
- Created `frontend/src/theme.js` implementing Midnight Ocean styling rules, overrides for buttons/papers, and custom typography headers.
- Overwrote `App.jsx` to render the authenticated 3-column docking card dashboard layout.
- Integrated left panel tab logic showing mock Drone Hubs and active status listings.
- Rendered right-hand panel widgets, incorporating monospace font families for all telemetry variables (Battery, Altitude, Speed, Wind).
- Added bright warning/danger styles for override buttons (Pause/RTH).
- Executed compilation test verifying error-free Vite builds.

### File List
- frontend/src/theme.js
- frontend/src/App.jsx
