# Story 3.4: Manual Override Long-Press Controls

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an Operator,  
I want to trigger manual drone override commands (Pause/RTH) only via long-press controls,  
so that I prevent accidental, dangerous physical drone moves during active operations.

## Acceptance Criteria

1. **Given** a drone is flying.
2. **When** an operator triggers a manual command (Pause/RTH).
3. **Then** the UI requires a long-press (>= 1.5 seconds) on the button with a circular progress overlay before sending the command.
4. **And** the Go gateway validates operator roles and command leases in Redis prior to publishing the MQTT message.

## Tasks / Subtasks

- [x] Task 1: Go Backend Command Validation Handler (AC: 4)
  - [x] Subtask 1.1: Register `POST /api/operator/command` on the Go gateway (wrapped under SSO auth roles middleware).
  - [x] Subtask 1.2: Implement validator checking: operator holds the exclusive lease for the drone (or if lease is expired/unheld, grant it to them), and user owns "operator" or "admin" roles.
  - [x] Subtask 1.3: Upon validation success, publish the command (`hover` or `rth`) to MQTT topic `drone/hub/command`.
- [x] Task 2: React long-press gesture listener hook (AC: 3)
  - [x] Subtask 2.1: Implement long-press event handlers in `App.jsx` capturing `onMouseDown`/`onMouseUp`, `onTouchStart`/`onTouchEnd`, and `onMouseLeave`.
  - [x] Subtask 2.2: Ensure the timer requires exactly 1500ms of continuous press to fire the action. Reset calculations on early release.
- [x] Task 3: Right Sidebar Button HUD Progression (AC: 3)
  - [x] Subtask 3.1: Style the "Pause Flight" and "Return-To-Home" buttons to render inline MUI circular progress indicators or linear progress bars filling up during hold gestures.
  - [x] Subtask 3.2: Trigger Axios POST `/api/operator/command` request upon long-press completion, showing notification alerts for operator verification.
- [x] Task 4: Command Override Unit Testing (AC: 4)
  - [x] Subtask 4.1: Write Go unit tests verifying command validation rules, lease blocks (returns 403 when locked by other operator), and MQTT publish triggers.

## Dev Notes

- **Language & Frameworks:** Go (Gateway), JavaScript (React UI with MUI).
- **Source Paths to Create/Modify:**
  - Modify: `backend/cmd/gateway/main.go` (Add `/api/operator/command` handler with validations)
  - Modify: `frontend/src/App.jsx` (Integrate long-press event loops and progress states on override buttons)
- **Previous Learnings Integration:**
  - Verify role scopes using `RequireRole("operator", "admin")`.

### References

- [Source: _bmad-output/planning-artifacts/prds/prd-uss-surveillance-2026-07-08/prd.md#FR-4]
- [Source: _bmad-output/planning-artifacts/ux-designs/ux-uss-surveillance-2026-07-08/DESIGN.md#Controls]

## Dev Agent Record

### Agent Model Used

gemini-2.5-pro

### Debug Log References
- /home/bangnq/.gemini/antigravity-cli/brain/b29dd0a5-8fa1-4047-b14d-171d279485af/.system_generated/logs/transcript.jsonl

### Completion Notes List
- Programmed exclusive lease-validated `/api/operator/command` endpoint on Go gateway with `operator`/`admin` role verification check rules.
- Wired standard command publishing hooks to MQTT `drone/hub/command`.
- Added long-press mouse and touch gesture hooks in React `App.jsx` counting continuous press durations.
- Integrated determinate visual progress linear feedback indicators filling inline on buttons while holding.

### File List
- backend/cmd/gateway/main.go
- frontend/src/App.jsx
