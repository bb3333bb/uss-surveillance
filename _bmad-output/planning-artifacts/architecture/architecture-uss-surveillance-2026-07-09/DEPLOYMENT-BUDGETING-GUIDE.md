# Deployment & Budgeting Sizing Guide (On-Premises)

This guide provides stakeholders with hardware sizing, networking requirements, and licensing structures needed to calculate the capital expenditure (CapEx) and operational expenditure (OpEx) for deploying USS Surveillance.

---

## 1. On-Premises Server Hardware Sizing
Because USS Surveillance is deployed on-premises, we require physical server hardware. The primary resource driver is **real-time video transcoding** (converting drone RTSP/RTMP streams to WebRTC playout via SRS).

### 1.1 Sizing Tiers (1 to 3 Active Drones/Hubs)
For a pilot deployment supporting up to 3 active Drone Hubs running simultaneous missions:

| Component | Minimum Specification | Recommended Specification | Purpose |
| --- | --- | --- | --- |
| **CPU** | Intel Xeon (8 Cores, 16 Threads) | AMD EPYC (16 Cores, 32 Threads) | Handles Go server concurrency, Python calculations, and database indexing. |
| **RAM** | 32 GB DDR4 ECC | 64 GB DDR5 ECC | Caching GIS telemetry data in Redis and executing PostGIS spatial joins. |
| **Storage** | 1 TB Enterprise SSD (SATA) | 2 TB NVMe SSD (RAID 1) | Databases and archival storage for mission MP4 video logs (estimated 1.5 GB per 30-min flight). |
| **GPU** | None (CPU Transcoding) | NVIDIA T4 or NVIDIA A2 (8GB VRAM) | **Highly Recommended.** Offloads video transcoding from the CPU, supporting up to 8 concurrent WebRTC streams. |
| **Network** | 1 Gbps Ethernet Port | 2x 10 Gbps Ethernet Ports (LACP) | Video stream ingestion and WebSocket dashboard connectivity. |

*   **Estimated Server Hardware Cost (One-time CapEx):** **$3,500 – $6,500** depending on GPU and storage redundancy.

---

## 2. On-Premises Network Sizing
Drones require reliable network bandwidth to stream telemetry and high-definition video back to the server.

*   **Video Ingestion:** An active HD (1080p, 30fps) stream from a drone payload consumes **4 to 6 Mbps** constant upload bandwidth.
*   **Telemetry Ingestion:** 1 Hz MQTT telemetry consumes **<10 Kbps** per drone.
*   **Total Local LAN Bandwidth (3 Drones active):** **12 to 18 Mbps** total.
*   **Network Constraint:** The local server must reside on a subnet accessible by both the Drone Hub controllers (over local mesh Wi-Fi or private LTE) and the operator workstations.

---

## 3. Software Licensing Cost Model
To keep recurring operational costs minimal, the platform is designed around a fully open-source and permissively-licensed software stack:

| Component | Technology | License Type | Annual License Cost |
| --- | --- | --- | --- |
| **Operating System** | Ubuntu Server 22.04 LTS | Open Source (GPL) | **$0** |
| **Database** | PostgreSQL + PostGIS | Open Source (PostgreSQL License) | **$0** |
| **Telemetry Cache** | Redis Community Edition | Open Source (RSALv2/SSPLv1) | **$0** |
| **WebRTC Media Server**| SRS (Simple Realtime Server) | Open Source (MIT) | **$0** |
| **MQTT Broker** | Eclipse Mosquitto | Open Source (EPL/EDL) | **$0** |
| **Mapping Engine** | MapLibre GL / Leaflet | Open Source (BSD/MIT) | **$0** |
| **Map Base Tiles** | OpenStreetMap (OSM) | Open Database License (ODbL) | **$0** (if self-hosted) |

### 3.1 Base Map Sizing Note
To operate 100% offline or within a secure LAN, we must host our own map tiles. 
*   **Self-hosted Tile Server:** Using open-source tile servers (e.g. Tegola or MapTiler Server) using free OSM vector maps yields **$0** licensing.
*   **Setup Overhead:** Requires a one-time import of region-specific GIS map layers (e.g. the country or facility boundary map, ~10 GB - 50 GB storage).

---

## 4. Operational Expenditure (OpEx) Estimate
Unlike cloud deployments where costs scale with data processing and bandwidth, the on-premises operational model is highly predictable:

1.  **Server Power & Cooling:** ~$20 – $40 per month.
2.  **External Weather API Subscription:** (Optional, used by the Suggestion Engine)
    *   OpenWeatherMap (Free tier up to 1,000 calls/day): **$0 / month**.
    *   Professional Weather API: **$40 – $120 / month** (if higher update frequencies or radar overlays are required).
3.  **Physical Hardware Maintenance:** Internal IT support overhead.
