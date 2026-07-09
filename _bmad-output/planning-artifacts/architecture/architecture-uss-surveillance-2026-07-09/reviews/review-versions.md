# Technology & Version Verification Review: USS Surveillance

## Overall Verdict
PASS. All selected technology versions are current, stable, and highly compatible with on-premises Docker orchestration.

## Findings

### Finding 1: SRS WebRTC Playout Version Check
- **Severity: Low**
- **Location:** `ARCHITECTURE-SPINE.md` (§ Stack, SRS)
- **Status:** Verified that SRS v6 is the latest major release branch (in production as of 2024-2026). It includes natively built-in WebRTC, RTMP, and RTSP forwarding, matching our on-premises low-latency requirements.
- **Fix:** Ensure Docker images use the pinned tag `ossrs/srs:6` to avoid sliding tag updates in production environments.

### Finding 2: Mosquitto MQTT v2 Configuration
- **Severity: Medium**
- **Location:** `ARCHITECTURE-SPINE.md` (§ Stack, Mosquitto)
- **Status:** Mosquitto 2.0+ enforces strict authentication by default (it rejects anonymous external connections out-of-the-box).
- **Fix:** Specify that the default on-premises Mosquitto configuration must explicitly mount a passwd file and disable anonymous connections for physical Drone Hub API authentication.
