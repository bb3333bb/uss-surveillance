# Adversarial Architecture Review: USS Surveillance

## Overall Verdict
PASS WITH CAUTION. The core invariants prevent major data collisions, but two potential gaps exist around Redis session expiration and coordinate precision between Go and Python.

## Findings

### Finding 1: Redis Lease Expiration Handling (AD-5)
- **Severity: High**
- **Location:** `ARCHITECTURE-SPINE.md` (§ Invariants, AD-5)
- **Obeying the Letter:** The system enforces that only the leaseholder can send commands. If the operator's browser loses power, the Redis lease will eventually time out and expire.
- **The Divergence:** During the period between the browser disconnect and the Redis lease timeout, the drone is flying but has no operator. Conversely, when it expires, the drone could be left running without any default fail-safe triggering.
- **Fix:** If the lease holder fails to send a WebSocket keepalive ping for more than 10 seconds, the Go gateway must automatically trigger a "Pause" hover command and flag the mission for lease hijacking by other operators.

### Finding 2: GIS Spatial Coordinate Drift (AD-6, AD-2)
- **Severity: Medium**
- **Location:** `ARCHITECTURE-SPINE.md` (§ Invariants, AD-2 & AD-6)
- **Obeying the Letter:** Python computes path waypoints; Go validates them against PostGIS.
- **The Divergence:** If Python uses a GIS library that rounds floating-point lat/lng values to 5 decimal places (~1.1 meter accuracy) and PostGIS uses high-precision doubles, a waypoint calculated by Python to be 0.5 meters outside a No-Fly Zone could be read by PostGIS as overlapping the boundary. Go would reject the path, failing the mission launch without explanation.
- **Fix:** Enforce a strict precision convention: all coordinates exchanged between services must use standard EPSG:4326 projection, serialized as floats with exactly 7 decimal places (millimeter accuracy).
