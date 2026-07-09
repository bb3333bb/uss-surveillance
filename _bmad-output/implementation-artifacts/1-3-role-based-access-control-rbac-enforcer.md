# Story 1.3: Role-Based Access Control (RBAC) Enforcer

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an Administrator or Operator,  
I want the system to enforce role scopes,  
so that unprivileged users cannot issue commands or alter flight plans.

## Acceptance Criteria

1. **Given** an operator has logged in.
2. **When** their JWT claims contain the role `viewer`.
3. **Then** the right sidebar command buttons (Pause, RTH, Land) are disabled on the UI and HTTP request validation on the Go gateway blocks command execution.
4. **When** their JWT contains `operator` (Sarah), all flight planning and launch override actions are enabled.

## Tasks / Subtasks

- [x] Task 1: Go Backend JWT Role Verification Middleware (AC: 3, 4)
  - [x] Subtask 1.1: Implement a role-enforcing middleware helper `RequireRole(allowedRoles ...string) func(http.Handler) http.Handler` that checks `UserClaims` from the request context.
  - [x] Subtask 1.2: If the claims role matches none of the allowed roles, return an HTTP 403 Forbidden response with error envelope: `{"code": "FORBIDDEN", "message": "Access denied for this role"}`.
  - [x] Subtask 1.3: Apply `RequireRole("operator", "admin")` to protected command mock endpoints on the Go gateway.
- [x] Task 2: Frontend Dashboard UI Role Locking (AC: 3, 4)
  - [x] Subtask 2.1: Read the current `role` parameter inside `App.jsx` using `useAuth()`.
  - [x] Subtask 2.2: If `role === 'viewer'`, set `disabled={true}` on the Right Sidebar override buttons (Pause Flight, Return-To-Home).
  - [x] Subtask 2.3: If `role === 'viewer'`, hide the "Toggle 3D View" and polygon drawing control buttons on the Center Map panel.
- [x] Task 3: Unit Testing Go Role Checks (AC: 3)
  - [x] Subtask 3.1: Extend `backend/pkg/auth/middleware_test.go` to include test cases verifying HTTP 403 Forbidden responses when `viewer` tokens attempt to access `RequireRole("operator")` routes.

## Dev Notes

- **Language & Frameworks:** Go (Gateway API) and JavaScript (React UI with MUI).
- **Source Paths to Modify:**
  - Modify: `backend/pkg/auth/middleware.go` (Add role validation middleware helper)
  - Modify: `backend/pkg/auth/middleware_test.go` (Add unit test cases for role checks)
  - Modify: `backend/cmd/gateway/main.go` (Apply role enforcement middleware to write routes)
  - Modify: `frontend/src/App.jsx` (Add disabled/hidden properties depending on user role)
- **Previous Learnings Integration:**
  - Story 1.1 successfully integrated OIDC AuthContext, exposing `role` to React clients.
  - Story 1.2 completed the visual overrides layout; these will now be disabled depending on `role`.

### References

- [Source: _bmad-output/planning-artifacts/prds/prd-uss-surveillance-2026-07-08/prd.md#FR-16]
- [Source: _bmad-output/planning-artifacts/architecture/architecture-uss-surveillance-2026-07-09/ARCHITECTURE-SPINE.md#AD-8]

## Dev Agent Record

### Agent Model Used

gemini-2.5-pro

### Debug Log References
- /home/bangnq/.gemini/antigravity-cli/brain/b29dd0a5-8fa1-4047-b14d-171d279485af/.system_generated/logs/transcript.jsonl

### Completion Notes List
- Implemented `RequireRole` middleware helper in Go backend.
- Added test cases verifying successful access by `operator` and HTTP 403 Forbidden blocking for `viewer` roles.
- Applied `RequireRole` to `/api/operator/command` endpoint on Go gateway.
- Wired React App.jsx UI elements to check `role` from useAuth.
- Configured conditional rendering to hide the Toggle 3D View button and set disabled state on Pause and RTH override buttons for users with `viewer` role.
- Verified successful client compile using production Vite build script.

### File List
- backend/pkg/auth/middleware.go
- backend/pkg/auth/middleware_test.go
- backend/cmd/gateway/main.go
- frontend/src/App.jsx
