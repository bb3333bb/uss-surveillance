# WebSocket & Command API Reference

Documents the live telemetry WebSocket and the manual-override command endpoint exposed by the Go gateway (`backend/cmd/gateway/main.go`). Written from the current implementation, not the aspirational HLD — see [CONFIGURATION.md](CONFIGURATION.md) for the gap between the two.

## `GET /api/operator/telemetry` (WebSocket upgrade)

Auth: JWT required (same bearer token as REST calls, passed as `?token=` on the upgrade request per `frontend/src/App.jsx`). Role: any authenticated user may connect; only `operator`/`admin` can issue commands via the separate REST endpoint below.

On connect, the server attempts to acquire the exclusive control lease for `Drone-01` on behalf of the connecting operator (10s TTL, auto-renewed by client pings — see below). Losing the race for the lease does not close the connection; the client still receives read-only telemetry, with a `LEASE_CONFLICT` alert.

### Server → client (1 Hz)

```json
{
  "lat": 10.762622,
  "lng": 106.660172,
  "altitude": 12.0,
  "battery": 87.5,
  "speed": 5.2,
  "is_flying": true,
  "leaseholder": "operator-dev",
  "hub_doors": "open",
  "alerts": ["RESTRICTED_AIRSPACE"]
}
```

| Field | Type | Notes |
|---|---|---|
| `lat`, `lng` | float | WGS84, current drone position (or hub position at 10.762622/106.660172 when docked) |
| `altitude` | float | Meters |
| `battery` | float | Percent, 0-100 |
| `speed` | float | m/s |
| `is_flying` | bool | |
| `leaseholder` | string | Username currently holding the exclusive control lease, `""` if unheld |
| `hub_doors` | string | One of `closed`, `opening`, `open`, `closing`, `recharging` |
| `alerts` | string[] | Zero or more of: `LEASE_CONFLICT`, `WEATHER_BREACH_WIND`, `WEATHER_BREACH_RAIN`, `RESTRICTED_AIRSPACE` |

### Client → server

```json
{ "type": "ping" }
```

The only inbound message type today. Renews the sender's control lease for another 10s. **Safety watchdog:** if no ping arrives within 10s of the last one, the server releases the lease, publishes a `hover` MQTT command, and marks the drone as not-flying — this is the heartbeat failsafe backing the "operator disconnected mid-flight" case, not just a keepalive nicety.

## `POST /api/operator/command`

Auth: JWT required, role `operator` or `admin`.

```json
{ "command": "pause" }
```

| `command` | MQTT topic `drone/hub/command` payload | Notes |
|---|---|---|
| `pause` | `hover` | |
| `rth` | `rth` | Also stops the simulated flight, archives the mission, and runs the door close/recharge sequence |

Anything else → `400 BAD_REQUEST`. Note there is currently no `land` command wired here despite FR-11 requiring Pause/RTH/Land — `land` is a documented gap, not an oversight in this doc.

Responses:
- `200` — `{"status": "success", "message": "..."}`
- `403 FORBIDDEN` — another operator holds the lease
- `400 BAD_REQUEST` — malformed body or unsupported command
- `500 INTERNAL_SERVER_ERROR` — MQTT publish failed

## Related REST endpoints

- `POST /api/operator/suggestion` `{lat, lng}` → drone/dock recommendation (proxies to the Python suggestion engine; falls back to a hardcoded response if that service is unreachable).
- `POST /api/operator/plan` `{vertices: [{lat, lng}, ...]}` → generated flight path; also sets the server's active flight path for the simulator.
- `POST /api/operator/launch` → runs the mechanical takeoff interlock sequence (FR-12) and starts the simulated flight.
- `GET /api/operator/missions?offset=&limit=` → paginated mission history, see [DATA-SCHEMA.md](DATA-SCHEMA.md).
