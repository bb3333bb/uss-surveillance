# Story 3.5: Automatic Landing & Recharging

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an Operator,  
I want the system to coordinate automatic landing sequences and cover closures at Dock Alpha,  
so that the drone lands safely, begins charging, and is protected from outdoor elements.

## Acceptance Criteria

1. **Given** the drone completes its patrol flight path or is commanded RTH.
2. **When** the drone approaches the designated Dock Alpha.
3. **Then** the Go gateway verifies the precise GPS coordinates match the Dock Alpha coordinates.
4. **And** commands the Drone Hub to open doors if they are closed.
5. **And** publishes the landing and recharge command.
6. **And** updates the Hub status to "Closed & Recharging" upon landing confirmation.

## Tasks / Subtasks

- [x] Task 1: Go Gateway Landing & Recharge Orchestrator (AC: 3, 4, 5)
  - [x] Subtask 1.1: Implement precise coordinate verification in the background simulation ticker: verify final coordinates match Dock Alpha center (`10.762622, 106.660172`) with 7-decimal precision (AD-6).
  - [x] Subtask 1.2: Upon approaching dock, check hub doors status. If closed, publish `open_doors` via MQTT, verify status feedback, and then publish `land`.
- [x] Task 2: Go Gateway Recharging State Machine (AC: 6)
  - [x] Subtask 2.1: Upon landing confirmation, set global drone state `IsFlying = false` and trigger automatic battery recharge increments (from e.g. active battery percentage back to 100%).
  - [x] Subtask 2.2: Publish mechanical MQTT cover command `close_doors` and update Hub status state to `Closed & Recharging`.
- [x] Task 3: React UI Active Recharging Feed (AC: 6)
  - [x] Subtask 3.1: Update Left Sidebar "Active Hubs" panels in `App.jsx` to dynamically render: door status (`Closed`, `Opening`, `Open`, `Closed & Recharging`) and battery charging loops.
  - [x] Subtask 3.2: Render pulsing charging bolt icons and linear progress battery animations when docking contacts are confirmed active.
- [x] Task 4: Landing Sequence Unit Testing (AC: 3, 4)
  - [x] Subtask 4.1: Write Go unit tests verifying coordinate validation tolerance boundaries and sequential interlock commands.

## Dev Notes

- **Language & Frameworks:** Go (Gateway), JavaScript (React UI with MUI).
- **Source Paths to Create/Modify:**
  - Modify: `backend/cmd/gateway/main.go` (Coordinate check logic, battery charge ticks, and MQTT handlers)
  - Modify: `frontend/src/App.jsx` (Dynamic hub widgets and active battery charging visual updates)
- **Previous Learnings Integration:**
  - Verify coordinate projection scales (WGS84) are serialized exactly as floats with 7 decimals.

### References

- [Source: _bmad-output/planning-artifacts/prds/prd-uss-surveillance-2026-07-08/prd.md#FR-7]
- [Source: _bmad-output/planning-artifacts/ux-designs/ux-uss-surveillance-2026-07-08/DESIGN.md#StatusHUD]

## Dev Agent Record

### Agent Model Used

gemini-2.5-pro

### Debug Log References
- /home/bangnq/.gemini/antigravity-cli/brain/b29dd0a5-8fa1-4047-b14d-171d279485af/.system_generated/logs/transcript.jsonl

### Completion Notes List
- Coded WGS84 coordinates matching verification comparing latitude/longitude coordinates (AD-6).
- Programmed mechanical door simulation state transitions: opening, open, closing, recharging.
- Integrated automatic battery replenishment calculations, ticking battery charging levels up to 100% when docked.
- Updated WS telemetry payloads to broadcast `hub_doors` state variables.
- Overwrote React `App.jsx` to render dynamic hubs cover status and animated pulsing battery charging bars.

### File List
- backend/cmd/gateway/main.go
- frontend/src/App.jsx
