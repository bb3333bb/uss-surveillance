---
stepsCompleted:
  - step-01
  - step-02
  - step-03
  - step-04
inputDocuments:
  - _bmad-output/planning-artifacts/prds/prd-uss-surveillance-2026-07-08/prd.md
  - _bmad-output/planning-artifacts/architecture/architecture-uss-surveillance-2026-07-09/ARCHITECTURE-SPINE.md
  - _bmad-output/planning-artifacts/ux-designs/ux-uss-surveillance-2026-07-08/DESIGN.md
  - _bmad-output/planning-artifacts/ux-designs/ux-uss-surveillance-2026-07-08/EXPERIENCE.md
---

# USS Surveillance - Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for USS Surveillance, decomposing the requirements from the PRD, UX Design, and Architecture requirements into implementable stories.

## Requirements Inventory

### Functional Requirements

- **FR-1:** Polygon Area Drawing - Sarah draws polygon coordinates on the 2D Map View.
- **FR-2:** 2D/3D View Toggle - Toggle between 2D Map View and 3D elevation map rendering.
- **FR-3:** Drone & Hub Suggestion Display - Overlay suggested Drone Hub, selected drone, and calculated Flight Plan on the map.
- **FR-4:** Environmental Check - Fetch wind/weather data; block launch and warn operator if local wind speed exceeds 15 m/s.
- **FR-5:** State-Based Recommendation - Suggest optimal drone based on battery, weather, payload, and readiness state.
- **FR-6:** Grid Route Optimization - Automatically calculate search grid path with 20% video overlap bounds.
- **FR-7:** Hub Handoff Pathing - Prepend/append vertical ascent/descent paths starting and ending at the Drone Hub.
- **FR-8:** Terrain & Obstacle Clearance - Validate waypoints remain at least 30 meters above local digital elevation (DEM) markers.
- **FR-9:** Telemetry Visualization - Stream real-time telemetry (Altitude, Battery, Speed, Wind) to UI dashboard at 1 Hz.
- **FR-10:** Live Video Overlay - Render live WebRTC drone camera stream in picture-in-picture overlay.
- **FR-11:** Manual Override Controls - Issue Pause, Return-to-Home (RTH), and Land override commands.
- **FR-12:** Mechanical Handoff & Safety Interlock - Query Drone Hub status to ensure doors are 100% open before launch.
- **FR-13:** Automatic Charging Management - Trigger charging contact cycles upon drone landing confirmation and door closure.
- **FR-14:** Hub Telemetry - Stream Drone Hub temperature, door state, and backup battery metrics to server database.
- **FR-15:** SSO Integration - Authenticate operators via OIDC Single Sign-On.
- **FR-16:** Role-Based Access Control - Enforce permissions per role (Viewer, Operator, Admin).
- **FR-17:** Synced Data Archival - Save completed missions as paired MP4 video and JSON telemetry timelines.
- **FR-18:** Interactive Replay Map - Render historical flight paths and load synchronized replay workspace.
- **FR-19:** Time-Sync Scrubbing - Link video player scrubber to map drone marker location playback.

### NonFunctional Requirements

- **NFR-1:** Telemetry Latency - Server-to-client telemetry and video stream delay must remain below 200ms.
- **NFR-2:** Command Propagation - Manual override commands must reach the physical drone in less than 500ms.
- **NFR-3:** Contrast Ratio - Ensure text contrast is at least 4.5:1 on dark surfaces for control room safety.

### Additional Requirements

- **Arch-1:** Docker Stack & Services - Multi-service build containerized locally using Docker Compose (Go Gateway, Python GIS Planner, SRS WebRTC Server, Mosquitto MQTT, Postgres, Redis).
- **Arch-2:** Go-Python gRPC API - Communication between Go backend and Python route generator uses typed gRPC calls.
- **Arch-3:** Redis Mutex Locking - Exclusive write-lock session in Redis per active drone. Heart-beat timeouts (>10s) trigger automatic Pause commands.
- **Arch-4:** GPS Precision Mapping - Coordinates exchanged between backend services must use WGS84 EPSG:4326 format with exactly 7 decimal places.

### UX Design Requirements

- **UX-DR-1:** Midnight Ocean Theme - Implement custom MUI theme with Deep Navy (`#0a192f`) and Steel Blue (`#172a45`) base dark mode.
- **UX-DR-2:** 3-Column Docking Layout - Desktop layout utilizing collapsible sidebar cards (Left: Fleet vs History tabs; Center: Map; Right: Telemetry & Controls).
- **UX-DR-3:** Double-Action Override - RTH and Land buttons require a 1-second long-press or confirmation modal to trigger.
- **UX-DR-4:** Emergency Flash Border - Flash screen borders red, play chime sound, and initiate a 10s auto-RTH countdown when emergency wind or battery events fire.
- **UX-DR-5:** Telemetry Mono Typography - Numerical values (battery, speed, altitude) must render in monospace font families to prevent UI layout shift.

### FR Coverage Map

- **FR-1:** Epic 2 - Drawing polygon coordinates on 2D map.
- **FR-2:** Epic 3 - Toggle between 2D and 3D map views.
- **FR-3:** Epic 2 - Render suggested drone, hub and flight path.
- **FR-4:** Epic 2 - Block launch if wind speed exceeds 15 m/s.
- **FR-5:** Epic 2 - Recommend optimal drone based on battery, weather and readiness.
- **FR-6:** Epic 2 - Automated grid route optimization with 20% video overlap.
- **FR-7:** Epic 2 - Prepend/append vertical corridors to Drone Hub.
- **FR-8:** Epic 2 - Validate path waypoints are >30m above terrain DEM.
- **FR-9:** Epic 3 - Render 1 Hz real-time flight telemetry on UI dashboard.
- **FR-10:** Epic 3 - Low-latency WebRTC video stream player widget.
- **FR-11:** Epic 3 - Dispatch Pause, RTH, and Land command overrides.
- **FR-12:** Epic 3 - Confirm Drone Hub doors are 100% open before launch.
- **FR-13:** Epic 3 - Command Drone Hub door close and charge contacts on landing.
- **FR-14:** Epic 3 - Ingest and log Drone Hub state and temperature.
- **FR-15:** Epic 1 - Single Sign-On OIDC authentication.
- **FR-16:** Epic 1 - Role-Based Access Control (Viewer, Operator, Admin).
- **FR-17:** Epic 4 - Archive flight video (MP4) and telemetry (JSON) relative to t=0.
- **FR-18:** Epic 4 - Load completed mission trajectory paths and replay workspace.
- **FR-19:** Epic 4 - Sync video scrubber to map marker coordinates.

## Epic List

### Epic 1: Operator Portal & SSO Authentication
Operator (Sarah) can securely authenticate via OIDC Single Sign-On and access the base 3-column dashboard workspace with active Role-Based Access Control permissions.
**FRs covered:** FR-15, FR-16.

### Epic 2: Intelligent Mission Planner & Suggestion Engine
Sarah can draw Monitoring Areas on the 2D map. The system automatically fetches weather, checks drone states, and suggests an optimized search-grid flight path for Sarah to confirm.
**FRs covered:** FR-1, FR-3, FR-4, FR-5, FR-6, FR-7, FR-8.

### Epic 3: Automated Launch & Real-Time Flight Monitoring
Sarah can launch flights. The Drone Hub opens doors, launches the drone, and triggers charging upon landing, while Sarah supervises live WebRTC video, 1 Hz telemetry, and can trigger safety overrides.
**FRs covered:** FR-2, FR-9, FR-10, FR-11, FR-12, FR-13, FR-14.

### Epic 4: Post-Mission Synced Replay Engine
Sarah can access completed mission logs to play back and review past flights, scrubbing a timeline that links recorded video directly to map drone coordinates.
**FRs covered:** FR-17, FR-18, FR-19.

---

## Epic 1: Operator Portal & SSO Authentication

The goal of this epic is to deploy the core Docker local environment container structure, authenticate dashboard users via SSO, and render the base 3-column desktop layout using our customized Midnight Ocean dark theme.

### Story 1.1: OIDC Authentication & JWT Validation
**As an** Operator,  
**I want** to authenticate using my organization's Single Sign-On,  
**So that** I can securely access the dashboard console.

**Acceptance Criteria:**
*   **Given** the user navigates to the application root URL.
*   **When** they click the login trigger button.
*   **Then** the application redirects them to the OpenID Connect (OIDC) identity provider login flow.
*   **When** authentication completes successfully and they return to the app.
*   **Then** the Go gateway signs a JWT session token containing user details and role scope, saving it securely in browser session storage.

### Story 1.2: Base 3-Column Dashboard Layout & Themes
**As an** Operator,  
**I want** to see the workspace organized in a high-density 3-column dark layout,  
**So that** I can observe telemetry, maps, and video with low eye fatigue.

**Acceptance Criteria:**
*   **Given** the operator is authenticated.
*   **When** the dashboard page renders.
*   **Then** the visual identity applies the Midnight Ocean theme (Deep Navy `#0a192f` background, Steel Blue `#172a45` panels, Ocean Blue `#00b0ff` borders).
*   **And** the layout displays the three defined drawers (Left Sidebar, Map Center Panel, Right Control Sidebar).

### Story 1.3: Role-Based Access Control (RBAC) Enforcer
**As an** Administrator or Operator,  
**I want** the system to enforce role scopes,  
**So that** unprivileged users cannot issue commands or alter flight plans.

**Acceptance Criteria:**
*   **Given** an operator has logged in.
*   **When** their JWT claims contain the role `viewer`.
*   **Then** the right sidebar command buttons (Pause, RTH, Land) are disabled on the UI and HTTP request validation on the Go gateway blocks command execution.
*   **When** their JWT contains `operator`, all flight planning and launch override actions are enabled.

---

## Epic 2: Intelligent Mission Planner & Suggestion Engine

The goal of this epic is to integrate the Go Gateway with the Python GIS Planner over gRPC, permitting operators to draw monitoring perimeters and receive automated drone suggestions and flight path grids.

### Story 2.1: Polygon Drawing on 2D Map View
**As an** Operator,  
**I want** to draw a patrol boundary directly on a 2D map,  
**So that** I can define the geographic area to be monitored.

**Acceptance Criteria:**
*   **Given** the operator is viewing the active 2D Map View.
*   **When** they select "Draw Area" and click vertices on the map.
*   **Then** the UI draws a dashed polygon boundary following their cursor points.
*   **And** closing the polygon outputs the array of WGS84 coordinates to the client state.

### Story 2.2: Automated Weather and Wind Assessment
**As an** Operator,  
**I want** the system to query real-time environmental data for the drawn area,  
**So that** we can assess weather hazards before takeoff.

**Acceptance Criteria:**
*   **Given** a drawn polygon boundary.
*   **When** the system receives the midpoint coordinates.
*   **Then** the Go gateway fetches wind speed and wind direction from the local weather API.
*   **And** stores the metrics to validate against drone limits.

### Story 2.3: Fleet State Suggestion Engine
**As an** Operator,  
**I want** the system to suggest the best drone-and-hub pair for my drawn boundary,  
**So that** I do not have to manually locate charged and available hardware.

**Acceptance Criteria:**
*   **Given** a drawn polygon and current wind speed data.
*   **When** the suggestion engine is triggered.
*   **Then** the Go gateway queries Postgres and Redis for active Drone States and Drone Hubs.
*   **And** recommends the closest docked drone that:
    1. Has a battery charge sufficient to complete the flight plus a 20% buffer.
    2. Has a wind resistance rating higher than the current wind speed.

### Story 2.4: Flight Path Grid Generator & Geofence Guard
**As an** Operator,  
**I want** the system to automatically generate a safe search-grid path,  
**So that** I can review and launch the flight without manually drawing waypoints.

**Acceptance Criteria:**
*   **Given** a recommended drone and polygon.
*   **When** the path calculator runs.
*   **Then** the Python service calculates a lawnmower grid route matching the camera field-of-view (ensuring 20% video overlap) and prepends vertical takeoff/landing corridors relative to the Drone Hub.
*   **And** cross-references local DEM terrain data to verify all waypoints are >30m above ground.
*   **And** the Go gateway validates the final trajectory against PostGIS geofences before rendering the path overlay on Sarah's map.

---

## Epic 3: Automated Launch & Real-Time Flight Monitoring

This epic covers the telemetry streaming loop, WebRTC transcoding media server setup, and sending manual overrides to the physical drone hub.

### Story 3.1: Hub Interlock & Takeoff Sequence
**As an** Operator,  
**I want** the Drone Hub to manage mechanical open/takeoff interlocks automatically,  
**So that** the drone does not launch while the hub cover is closed.

**Acceptance Criteria:**
*   **Given** the operator clicks "Confirm Mission".
*   **When** the Go gateway publishes the launch sequence command via MQTT.
*   **Then** the system first commands the Drone Hub to open its doors.
*   **And** the drone is only commanded to launch after the Drone Hub returns a "Doors Fully Open" status signal.

### Story 3.2: 1 Hz Telemetry & Operator Mutex Lease
**As an** Operator,  
**I want** to see live drone parameters refreshed at 1 Hz and maintain an exclusive control lease,  
**So that** I have full situational awareness and prevent conflicting commands from other users.

**Acceptance Criteria:**
*   **Given** a drone is flying.
*   **When** telemetry streams from the drone over MQTT.
*   **Then** the Go gateway caches the state in Redis and streams it at 1 Hz to the browser via WebSockets.
*   **And** the Go gateway secures an exclusive command mutex lock for the initiating operator in Redis.
*   **And** automatically pauses the drone if the operator's browser heartbeat ceases for >10 seconds.

### Story 3.3: WebRTC Video Stream Transcoding
**As an** Operator,  
**I want** to watch the drone's live camera feed in ultra-low latency,  
**So that** I can inspect the perimeter in real time.

**Acceptance Criteria:**
*   **Given** the drone is streaming video.
*   **When** the feed (RTSP/RTMP) reaches the local SRS Media Server.
*   **Then** SRS transcodes the stream into WebRTC packets and plays it on the client UI video overlay widget with <200ms lag.

### Story 3.4: Manual Override Long-Press Controls
**As an** Operator,  
**I want** my manual override buttons to require confirmation,  
**So that** I do not trigger an accidental landing or return-to-home during patrol.

**Acceptance Criteria:**
*   **Given** an active mission is running.
*   **When** the operator clicks "RTH" or "Land".
*   **Then** the interface requires a 1-second long-press or displays a confirmation dialog.
*   **When** confirmed, the Go gateway immediately publishes the command packet over MQTT to the drone.

### Story 3.5: Automatic Landing & Recharging
**As an** Operator,  
**I want** the drone to land and begin charging automatically,  
**So that** the hardware is prepped for the next deployment.

**Acceptance Criteria:**
*   **Given** the drone finishes its grid or executes RTH.
*   **When** it lands in the Drone Hub.
*   **Then** physical dock contact sensors confirm docking, command the hub doors to close, and activate the charging contacts, updating the dashboard status to "Charging".

### Story 3.6: Safety Alert HUD Indicator
**As an** Operator,  
**I want** to see visual alerts and countdowns when thresholds are breached,  
**So that** I can intercept the drone before an incident occurs.

**Acceptance Criteria:**
*   **Given** the drone is flying.
*   **When** wind speed exceeds 15 m/s or battery drops below safe return limits.
*   **Then** the dashboard plays a warning sound, flashes the map borders red, and presents a 10s countdown window before initiating an automated RTH.

---

## Epic 4: Post-Mission Synced Replay Engine

This epic implements database archival and the timeline scrubbing interface that links recorded video frames directly to map coordinates.

### Story 4.1: Synced Data Archiving
**As an** Operator,  
**I want** the system to save flight paths and video recordings automatically,  
**So that** past missions can be analyzed later.

**Acceptance Criteria:**
*   **Given** a mission has completed.
*   **When** the drone docks.
*   **Then** the Go gateway compiles the 1 Hz telemetry logs from Redis into a JSON log file (indexed from `t=0` relative to takeoff) and stores it in PostgreSQL alongside the MP4 video recording file.

### Story 4.2: Mission History Dashboard
**As an** Operator,  
**I want** to view a list of past missions on my left sidebar drawer,  
**So that** I can select a specific patrol log to audit.

**Acceptance Criteria:**
*   **Given** the operator clicks the "Mission Logs" tab.
*   **When** the sidebar tab is loaded.
*   **Then** the system queries PostgreSQL and lists completed missions with date, duration, and drone details.
*   **When** the operator selects a mission and clicks "View Replay", the dashboard transitions into Replay Mode.

### Story 4.3: Interactive Timeline Scrubber & Map Sync
**As an** Operator,  
**I want** my video playback to be synchronized with the map coordinates during replay,  
**So that** scrubbing the timeline shows me the exact spot where that video segment was recorded.

**Acceptance Criteria:**
*   **Given** the dashboard is in Replay Mode.
*   **When** the operator drags the bottom timeline slider.
*   **Then** the video player jumps to the matching frame, and the drone map marker moves along the drawn trajectory path to show the coordinates at that millisecond index (and vice versa).
