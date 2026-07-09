---
title: USS Surveillance - Experience Specification
status: final
created: 2026-07-08
updated: 2026-07-08
---

# Interactive Experience Specification: USS Surveillance

Interactive Mockup Reference: [screen-dashboard.html](file:///home/bangnq/wip/uss-surveillance/_bmad-output/planning-artifacts/ux-designs/ux-uss-surveillance-2026-07-08/mockups/screen-dashboard.html)

> [!NOTE]
> In case of any conflict between the mockup interactive behavior and the specifications written in this document (EXPERIENCE.md), this document wins.

## 1. Foundation
- **Form-Factor:** Desktop web application optimized for 1920x1080 resolution. Avoid touch-based scaling; layout is optimized for cursor precision.
- **UI System:** Material UI (MUI). All visual tokens (colors, margins, borders) reference definitions in `DESIGN.md` using the `{tokens.colors.<token>}` syntax.

## 2. Information Architecture (IA)
The dashboard is divided into four main functional regions:
1.  **Header:** Displays global system status, active local wind/weather telemetry, current operator identity, and platform branding.
2.  **Left Drawer (Collapsible):** Displays tabbed navigation:
    - **Fleet & Docks Tab:** Displays physical Drone Hub positions, their operational status, and details of housed drones.
    - **Mission Logs Tab:** Lists archived and completed missions for search and analysis.
3.  **Center Workspace (Map):** A full-size OpenStreetMap display. Houses polygon drawing tools, flight trajectory overlays, and the 2D/3D toggle button.
4.  **Right Panel (Contextual Controls):** 
    - In **Active Flight Mode**, displays live numerical telemetry (altitude, battery, speed), low-latency live video, and emergency override buttons.
    - In **Replay Mode**, displays historical telemetry indicators, archived video playback, and scrubbing details.
5.  **Bottom Panel (Mission Replay):** Collapsible timeline scrubber, only visible in Replay Mode.

## 3. Voice and Tone
- **Tone:** Objective, technical, and alert-driven. Avoid descriptive fluff; present exact numbers and statuses.
- **Alert Microcopy:** 
  - Normal state: `🌬️ Wind: 4.2 m/s (Safe)` (green status).
  - Approaching limit: `⚠️ Wind: 12.8 m/s (Caution)` (yellow status).
  - Limit exceeded: `🚨 CRITICAL WIND SPEED: 16.5 m/s` (flashing red status, trigger prompt).

## 4. Component Patterns
- **Safety override actions (FR-11):** To prevent accidental trigger, RTH and Emergency Land buttons require a 1-second long-press or a confirmation pop-up to execute.
- **Map View Mode Switch (FR-2):** Toggling to "3D Elevation View" tilts the map angle by 45 degrees, rendering digital elevation model (DEM) grids to highlight drone altitude relative to local buildings.
- **Telemetry update refresh (FR-9):** Numbers must transition using a smooth cross-fade to prevent visual jitter at 1 Hz updates.

## 5. State Patterns
The application transitions between three global screen states:

### 5.1 Active Flight Mode (Default)
- **Visible elements:** Left Sidebar (Fleet tab), Center Map (2D/3D), Right Sidebar (Telemetry, Live Video, Override Commands).
- **Hidden elements:** Bottom Replay Panel.

### 5.2 Replay Mode (Triggered)
- **Activation:** Initiated when Sarah clicks a completed patrol in the "Mission Logs" tab and selects "Replay".
- **Visible elements:** Center Map (drawing historical path), Right Sidebar (archive video, playback rate toggles), Bottom Replay Panel (scrubber, play/pause timeline).
- **Hidden elements:** Real-time wind alerts in header, active flight override buttons (replaced by play controls).

### 5.3 Emergency Alert Mode (Automatic)
- **Activation:** Triggered when telemetry exceeds drone thresholds (e.g. wind > 15 m/s or battery < safe return threshold).
- **UI behavior:** Plays a system chime, flashes a red warning border around the Map View, and overlays a high-priority prompt allowing Sarah to confirm RTH or let the system execute the auto-RTH countdown (10 seconds).

## 6. Interaction Primitives
- **Polygon drawing:** Operator clicks to define vertices, system draws dashed lines, and double-clicking the starting vertex closes the polygon.
- **Timeline scrubbing:** Dragging the slider dynamically moves the drone marker along the map path while seeking the video player frame to the corresponding timestamp.

## 7. Accessibility Floor
- **Contrast:** Telemetry values must maintain a contrast ratio of at least 4.5:1 against the `{tokens.colors.surface}` background.
- **Keyboard Overrides:** 
  - `Spacebar` acts as play/pause in Replay Mode, and hover-hold in Active Flight Mode.
  - `Escape` closes Replay Mode and returns the operator to the active fleet view.

## 8. Key Flows

### Flow 1: Sarah schedules and launches a mission (UJ-1)
1. Sarah logs into the dashboard via SSO.
2. She selects "Draw Area" and clicks 4 points on the map to define the perimeter.
3. The Suggestion Engine selects "Drone-3 (Dock South)" based on weather and battery.
4. Sarah clicks "Confirm Mission" and confirms takeoff. The left sidebar shows Dock South opening doors.

### Flow 2: Sarah audits a completed flight
1. Sarah clicks the "Mission Logs" tab in the left sidebar.
2. She selects "Mission #142 (Grid Search)" and clicks "View Replay".
3. The dashboard transitions to **Replay Mode**: the map zooms to the flight path, the bottom timeline panel slides up, and the recorded video loads.
4. Sarah drags the slider to 05:40 to inspect a specific section of the perimeter; the map icon jumps to the exact coordinates where the video was captured.
5. Sarah clicks "Exit Replay" to return to live monitoring.
