---
title: USS Surveillance - Drone Monitoring Platform
status: final
created: 2026-07-08
updated: 2026-07-08
---

# PRD: USS Surveillance - Drone Monitoring Platform

## 0. Document Purpose
This document outlines the product requirements for USS Surveillance, a commercial drone fleet management and monitoring platform. It is designed to align product stakeholders, system architects, and development teams. The document structures user journeys, glossary terms, functional and non-functional requirements, and success metrics for the commercial launch of the platform.

## 1. Vision
USS Surveillance is a full-stack commercial solution (comprising web clients, backend services, and data pipelines) that empowers organizations to automate area monitoring and surveillance using drone fleets. By automating drone allocation from physical Drone Hubs, route calculation, and real-time telemetry/video streaming, the platform minimizes human operator overhead while maximizing safety and situational awareness on an open-source 2D/3D map engine.

## 2. Target User

### 2.1 Jobs To Be Done
- **Automated Operations**: Ensure security patrol areas are monitored consistently without manual waypoint-by-waypoint flight path creation.
- **Fleet Allocation**: Automatically allocate the most suitable drone based on state, location, and capability, avoiding manual inventory scheduling.
- **Real-Time Supervision**: Monitor the active state of drone missions via live video and telemetry overlay, intervening only when manual overrides are required.
- **Post-Mission Auditing**: Replay completed drone flights with synchronized video and telemetry to audit facility status or investigate security incidents.

### 2.2 Non-Users (v1)
- **Individual Hobbyists**: The platform is built for organizational multi-drone management, not standalone recreation.
- **Fully Manual Joyriders**: While manual intervention is supported for safety, the primary mode of operation is plan-based automation.

### 2.3 Key User Journeys

- **UJ-1. Sarah schedules and monitors an automated area surveillance mission.**
  - **Persona + context:** Sarah, Operations Coordinator for an industrial facility, needs to ensure daily boundary patrol is completed without managing low-level flight plans.
  - **Entry state:** Opens the web application and logs in using organizational Single Sign-On (SSO). She views the 2D mission dashboard.
  - **Path:**
    1. Sarah defines the target monitoring area by drawing a polygon on the 2D map.
    2. The system fetches real-time local weather/wind data, checks active Drone States (battery, payload capabilities, location) and Drone Hub availability, and selects the most suitable drone-and-hub pair (e.g., Drone-3 at Drone Hub South).
    3. The system automatically calculates an optimized search-grid flight path starting and ending at the designated Drone Hub.
    4. Sarah reviews the drone suggestion and the calculated path. Once she clicks "Confirm Mission", the system uploads the mission plan to the selected Drone Hub and drone.
    5. The drone launches automatically from its Drone Hub; Sarah monitors the live telemetry (altitude, battery, speed) and video stream overlay on the 2D map (switching to the 3D map view to visualize terrain and building clearance relative to the drone's altitude).
  - **Climax:** The drone streams live video showing the patrol area is clear and safe, and automatically returns to its Drone Hub when the path is completed.
  - **Resolution:** The drone lands inside the Drone Hub, the hub closes and begins automatically recharging the drone, and video streaming ends. Sarah accesses the mission history dashboard to play back the entire mission, scrubbing through the video timeline to see the drone's position marker move in sync along its historical flight path on the 2D Map View.
  - **Edge case:** If weather conditions deteriorate (e.g., wind speed exceeds the drone's safety threshold) or the battery drops below the safe return limit, the system alerts Sarah and automatically triggers a safe landing or Return-to-Home (RTH) to the nearest available Drone Hub.

## 3. Glossary
*Downstream workflows and readers must use these terms exactly. FRs, UJs, and SMs use Glossary terms verbatim; introducing a synonym anywhere in the PRD is a discipline violation. If §4 introduces a new domain noun, add it to the Glossary in the same pass.*

- **Monitoring Area** — A geographic polygon drawn on the map that defines the bounds for a drone surveillance patrol.
- **Drone Hub** — A physical station or docking nest containing a drone. It handles automatic drone launch, recovery, housing/protection, and automatic recharging when the drone returns from a mission.
- **Drone State** — The real-time operational status of a physical drone (e.g., battery charge, GPS lock, current location, payloads, readiness).
- **Environmental Factors** — Real-time weather parameters (wind speed, wind direction, precipitation, temperature) that affect flight safety and viability.
- **Drone Suggestion Engine** — The system logic that analyzes Drone States, Drone Hub availability, and Environmental Factors to recommend the best drone for a Monitoring Area.
- **Flight Plan** — The set of calculated waypoints, speed, altitude, and action commands generated by the system and uploaded to a drone.
- **Telemetry Data** — Real-time drone status data (e.g., GPS coordinates, altitude, battery percentage, heading, speed) streamed back to the server.
- **Live Video Stream** — The real-time video feed captured by the drone's camera and streamed to the operator's dashboard.
- **Return-to-Home (RTH)** — A fail-safe protocol where the drone stops its mission and automatically flies back to its takeoff/base point or its originating Drone Hub.
- **2D Map View** — The flat, OpenStreetMap-based map interface (built on Leaflet or MapLibre) where Sarah conducts flight planning and views live status.
- **3D Map View** — The 3D view of the map (built on MapLibre or other) used for visualizing terrain and building clearance relative to the drone's altitude.
- **SSO (Single Sign-On)** — The organization's centralized identity provider used to authenticate operators.

## 4. Features
*Each subsection is a coherent feature: behavioral description first, FRs nested under it, optional feature-specific NFRs and notes. FRs are numbered globally (FR-1 through FR-N) so downstream artifacts have stable references even if features get reorganized. Reference user journeys by ID inline ("realizes UJ-1") where the chain matters.*

### 4.1 Interactive Mission Planner (Map UI)
**Description:** This is the primary operational workspace for the coordinator. Built on Leaflet/MapLibre using OpenStreetMap (OSM) tile sources, it allows drawing, scaling, and deleting polygon bounds for patrolling. Operators can toggle to a 3D Map View to visualize terrain and building structures relative to the calculated flight altitude. Realizes UJ-1.

**Functional Requirements:**

#### FR-1: Polygon Area Drawing
The authenticated operator can draw, edit vertices of, and delete a single closed polygon representing the Monitoring Area on the 2D Map View. Realizes UJ-1.
**Consequences (testable):**
- The map UI must output the coordinates (array of GPS lat/lng) of the drawn polygon to the Suggestion Engine upon completion of drawing.

#### FR-2: 2D/3D View Toggle
The operator can toggle between the 2D Map View and 3D Map View at any time. Realizes UJ-1.
**Consequences (testable):**
- The 3D view must load digital elevation model (DEM) tiles and show the 3D trajectory (X, Y, Z waypoints) of the Flight Plan.

#### FR-3: Drone & Hub Suggestion Display
The map must display an overlay indicating the suggested Drone Hub location, the suggested drone, and the generated Flight Plan overlay. Realizes UJ-1.

---

### 4.2 Automated Drone & Hub Suggestion Engine
**Description:** A backend service that determines the feasibility of a mission and selects the best asset. It checks the capability of the drone (e.g., wind limit) against real-time local conditions. Realizes UJ-1.

**Functional Requirements:**

#### FR-4: Environmental Check
The system must retrieve current wind speed, wind direction, and weather alerts for the target Monitoring Area coordinates from a verified weather service. Realizes UJ-1.
**Consequences (testable):**
- If wind speed is above 15 m/s (or the specific drone model's threshold), the system must block mission confirmation and display a warning.

#### FR-5: State-Based Recommendation
The engine must query active Drone States and Drone Hub states to recommend a drone that:
1. Is fully docked and ready in an active Drone Hub.
2. Has a battery level sufficient to complete the Flight Plan plus a 20% safety margin.
3. Meets the wind resistance required for the current wind speed.
Realizes UJ-1.
**Consequences (testable):**
- If no drone in the fleet meets the conditions, the system must suggest a "Bring Drone to Location" task, identifying the nearest suitable drone that is currently offline/charging.

---

### 4.3 Flight Path Generator
**Description:** A backend service that takes the operator’s drawn polygon coordinates and generates an optimized, safe Flight Plan. The path must cover the target area efficiently while ensuring collision avoidance relative to terrain. Realizes UJ-1.

**Functional Requirements:**

#### FR-6: Grid Route Optimization
The system must automatically generate a grid-style (lawnmower pattern) flight path covering the Monitoring Area. The lane spacing must be calculated automatically based on the drone camera's Field of View (FOV) and the selected flight altitude to ensure 20% video overlap. Realizes UJ-1.
**Consequences (testable):**
- The output must be a JSON array of waypoints containing Latitude, Longitude, Altitude, and speed limits.

#### FR-7: Hub Handoff Pathing
The Flight Plan must automatically prepend a safe vertical ascent corridor from the originating Drone Hub, and append a safe descent corridor back to the landing dock of the originating Drone Hub. Realizes UJ-1.

#### FR-8: Terrain & Obstacle Clearance
The path generator must cross-reference digital elevation model (DEM) terrain data and known obstacle databases. Realizes UJ-1.
**Consequences (testable):**
- Every waypoint altitude must maintain a minimum clearance of 30 meters above the highest terrain point or building structure in that coordinate sector.

---

### 4.4 Real-time Mission Monitor & Telemetry Dashboard
**Description:** The dashboard view during active flights. Sarah uses this to observe the drone's position on the 2D/3D map, watch the live camera feed, and view telemetry metrics. It also hosts the emergency intervention buttons. Realizes UJ-1.

**Functional Requirements:**

#### FR-9: Telemetry Visualization
The dashboard must display real-time Telemetry Data (GPS coordinates, altitude, battery percentage, heading, speed, and signal strength) streamed from the drone, refreshed at a rate of at least 1 Hz. Realizes UJ-1.

#### FR-10: Live Video Overlay
The dashboard must overlay a low-latency Live Video Stream widget on the Map View, which can be maximized to full screen. Realizes UJ-1.

#### FR-11: Manual Override Controls
The dashboard must display explicit control buttons to issue instantaneous commands to the drone:
1. Pause Mission: Commands the drone to hover in place.
2. Return-to-Home (RTH): Commands the drone to immediately abort the mission and fly back to the Drone Hub.
3. Emergency Land: Commands the drone to immediately land vertically at its current position.
Realizes UJ-1.
**Consequences (testable):**
- Commands sent from the dashboard must be dispatched to the drone within 200ms.

---

### 4.5 Drone Hub Automated Controller
**Description:** The integration layer that communicates with the physical Drone Hub hardware. It coordinates the mechanical actions of the hub (opening/closing the dome) with the takeoff and landing sequence of the drone, and manages the automatic charging cycle. Realizes UJ-1.

**Functional Requirements:**

#### FR-12: Mechanical Handoff & Safety Interlock
The system must command the Drone Hub to open its doors prior to takeoff. The drone must not be powered up or launched until the Drone Hub returns a "Doors Fully Open" status. Realizes UJ-1.
**Consequences (testable):**
- Flight launch sequence is aborted and an alert is shown to Sarah if "Doors Fully Open" is not received within 15 seconds of the command.

#### FR-13: Automatic Charging Management
Once the drone lands and docking is confirmed by physical sensors, the system must command the Drone Hub to close its doors and activate the charging contacts. Realizes UJ-1.
**Consequences (testable):**
- The system dashboard must display the charging current and estimated time to full charge.

#### FR-14: Hub Telemetry
The Drone Hub must stream its own telemetry (internal temperature, power source status, heating/cooling unit status, door open/close sensors) to the database. Realizes UJ-1.

---

### 4.6 Authentication & Access Control (SSO)
**Description:** Ensures security and accountability for commercial drone operations by integrating with organizational authentication systems and enforcing role-based permissions. Realizes UJ-1.

**Functional Requirements:**

#### FR-15: SSO Integration
The web application must authenticate users via the organization's standard Single Sign-On (SSO) provider (e.g., OpenID Connect or OAuth2). Realizes UJ-1.

#### FR-16: Role-Based Access Control (RBAC)
The system must restrict access based on three roles:
1. Viewer: Can view the maps, active drones, and live video streams, but cannot interact or change anything.
2. Operator (Sarah): Can draw monitoring areas, approve suggestions, launch missions, and send manual overrides (Pause, RTH, Land).
3. Administrator: Can manage user permissions, register new Drone Hubs, and update drone parameters. Realizes UJ-1.

---

### 4.7 Mission Replay & Video Synchronization Engine
**Description:** Provides a synchronized playback workspace for analyzing completed missions. It links the recorded video payload timestamp with the logged telemetry timeline, allowing operators to see the exact location and orientation of the drone on the map for any frame of the video. Realizes UJ-1.

**Functional Requirements:**

#### FR-17: Synced Data Archival
The system must save the full telemetry log (GPS coordinates, altitude, orientation, time) and the recorded high-definition video file from the mission, aligning their timelines in the database. Realizes UJ-1.

#### FR-18: Interactive Replay Map
The system must allow loading a completed mission, drawing its historical flight trajectory on the map, and loading the video player. Realizes UJ-1.

#### FR-19: Time-Sync Scrubbing
Scrubbing the video player timeline must move the drone's position icon on the map to the coordinates at that timestamp, and clicking any point on the map trajectory must jump the video player to that timestamp. Realizes UJ-1.

---

## 5. Non-Goals (Explicit)
*   **No custom hardware/firmware:** We do not build physical drones or Drone Hubs. We write software that integrates with existing hardware APIs (e.g., DJI SDK/API).
*   **No browser-based manual joystick steering:** For security, latency, and safety reasons, Sarah cannot steer the drone using a virtual joystick in her browser. She can only issue high-level commands (Pause, RTH, Land).
*   **No drone swarming (v1):** We will not coordinate multiple drones flying in the same Monitoring Area at the same time. Only one active drone is supported per mission.
*   **No AI object detection/threat analysis (v1):** The system streams the live video feed but does not perform computer vision anomaly detection (e.g., detecting intruders automatically). A human operator must watch the video feed.

## 6. MVP Scope

### 6.1 In Scope
- SSO Authentication & 3 Roles (Viewer, Operator, Admin).
- Interactive 2D Map View (drawing polygon coordinates) & 3D Map View (displaying flight clearance relative to elevation).
- Suggestion Engine using local weather (wind, precipitation) and fleet battery/readiness states.
- Automatic grid-path calculation, including takeoff and landing corridors.
- Real-time telemetry streaming (1 Hz) and live video streaming overlay.
- High-level override commands: Pause, RTH, Land.
- Automatic Drone Hub integration (commanding doors, starting charging).
- Mission Replay Engine (synchronized telemetry history and recorded video playback).
- Support for one target drone model and its compatible dock (e.g., DJI Dock).

### 6.2 Out of Scope for MVP
- Multiple drones executing a single mission (swarming).
- Offline capability (the system requires active internet connection for weather data and OSM tiles).
- Automated AI video analytics.
- Support for multiple brands of drones or custom-built drone platforms.

## 7. Success Metrics
- **SM-1 (Deployment Success Rate):** Percentage of planned missions that successfully launch, execute, and land back in their originating Drone Hub without manual operator intervention. **Target: >95%**. Validates FR-12, FR-13.
- **SM-2 (System Latency):** Command propagation latency from Sarah clicking "RTH" or "Pause" to the physical drone initiating the command. **Target: <500ms**. Validates FR-11.
- **SM-3 (Suggestion Accuracy):** Percentage of times the suggested drone successfully completes the mission without encountering battery or environmental fail-safes. **Target: >90%**. Validates FR-4, FR-5.
- **SM-C1 (Safety Incident Rate - Counter-Metric):** Number of drone collisions or manual emergency landings away from the hub. **Must be 0.** Counterbalances SM-1.

## 8. Open Questions
1. Which specific video streaming protocols (e.g. WebRTC, HLS, RTSP) will the target Drone Hub and drone models use for live streaming vs. recorded archival?
2. Will the 3D Map View require offline terrain/DEM tile loading capability, or can it rely entirely on internet-sourced map servers?
3. What is the maximum wind speed tolerance of the target drone model, and does the local weather provider API update frequently enough (e.g. every 5 minutes) to prevent safety issues?

## 9. Assumptions Index
- `[ASSUMPTION: 2D-Primary]` - The operator primarily operates on the 2D map; the 3D map is for terrain and obstacle reference only.
- `[ASSUMPTION: SSO]` - The organization has an existing SSO system supporting OpenID Connect or OAuth2.
- `[ASSUMPTION: Drone API]` - The target physical Drone Hub and drone support API endpoints for remote path upload, takeoff command, and telemetry streaming.
