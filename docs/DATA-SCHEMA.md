# Mission Archive Data Schema

The mission archive (`backend/pkg/archive`) is what backs FR-17 (synced data archival) and FR-18/19 (replay). Today it's a flat JSON file at `data/missions.json` inside the gateway container/process — **not Postgres**, despite the HLD describing PostGIS as the archive store (see [CONFIGURATION.md](CONFIGURATION.md) for the full mock-vs-real inventory). This doc describes the schema as it exists now, since it's the schema Phase 3's Postgres migration needs to preserve.

## `MissionRecord` (one entry per completed/aborted flight)

```json
{
  "id": "MSN-1752460800",
  "start_time": "2026-07-14T02:00:00Z",
  "end_time": "2026-07-14T02:05:12Z",
  "duration_sec": 312.4,
  "video_path": "/videos/MSN-1752460800.mp4",
  "telemetry": [ /* TelemetryPoint[], see below */ ]
}
```

| Field | Type | Notes |
|---|---|---|
| `id` | string | `MSN-<unix seconds of takeoff>` — generated in `Archiver.SaveMission()`, not a UUID. Collision risk if two missions start in the same second; fine for one-drone MVP, revisit before multi-drone. |
| `start_time` / `end_time` | RFC3339 timestamp | `start_time` is set by `StartMission()` at launch, not by the first telemetry point. |
| `duration_sec` | float | `end_time - start_time` in seconds |
| `video_path` | string | `/videos/<id>.mp4` — **path is synthesized, not a real file.** No video recording is actually implemented yet (SRS/WebRTC archival is unbuilt) — this field is a placeholder for the eventual FR-17 video pairing. |
| `telemetry` | `TelemetryPoint[]` | Full 1 Hz log for the mission, in flight order |

## `TelemetryPoint` (one entry per second of flight)

```json
{ "t": 42, "lat": 10.762622, "lng": 106.660172, "altitude": 12.0, "battery": 87.5, "speed": 5.2 }
```

| Field | Type | Notes |
|---|---|---|
| `t` | int | Seconds elapsed since takeoff (`t=0` at first logged point) — this is what the frontend's replay scrubber indexes on, not a wall-clock timestamp |
| `lat`, `lng` | float | WGS84 |
| `altitude` | float | Meters |
| `battery` | float | Percent |
| `speed` | float | m/s |

## Storage mechanics (`backend/pkg/archive/archiver.go`)

- Single in-process `Archiver` (`DefaultArchiver`), guarded by a mutex — fine for one gateway instance, **not safe for multi-instance deployment** without moving to a real datastore.
- `StartMission()` resets the in-memory buffer; `LogPoint()` appends one point per tick; `SaveMission()` reads the whole `missions.json`, appends the new record, and rewrites the file. This means archive writes are O(total mission history) per save — acceptable at MVP scale, a real cost once the file has hundreds of missions with full telemetry embedded (related to the pagination gap noted in `GET /api/operator/missions`).
- Directory is created on demand (`os.MkdirAll`) — no migration/schema versioning exists since there's no real schema yet.

## Migration note for Phase 3 (Postgres+PostGIS)

When this moves to Postgres, the natural mapping is a `missions` table (one row per `MissionRecord`, sans `telemetry`) plus a `telemetry_points` table (one row per point, FK'd to `mission_id`, indexed on `(mission_id, t)` for scrubber lookups). `video_path` should stop being synthesized once real recording exists — track it as nullable until then rather than writing a fake path.
