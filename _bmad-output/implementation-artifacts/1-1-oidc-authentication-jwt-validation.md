# Story 1.1: OIDC Authentication & JWT Validation

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an Operator,  
I want to authenticate using my organization's Single Sign-On,  
so that I can securely access the dashboard console.

## Acceptance Criteria

1. **Given** the user navigates to the application root URL.
2. **When** they click the login trigger button.
3. **Then** the application redirects them to the OpenID Connect (OIDC) identity provider login flow.
4. **When** authentication completes successfully and they return to the app.
5. **Then** the Go gateway signs a JWT session token containing user details and role scope, saving it securely in browser session storage.

## Tasks / Subtasks

- [x] Task 1: Go Backend OIDC Client Setup (AC: 1, 2, 3)
  - [x] Subtask 1.1: Configure OIDC environment variables in the Go gateway (`OIDC_ISSUER_URL`, `OIDC_CLIENT_ID`, `OIDC_CLIENT_SECRET`, `OIDC_REDIRECT_URI`).
  - [x] Subtask 1.2: Implement `/api/auth/login` endpoint that redirects the user's browser to the OIDC provider's authorization page.
  - [x] Subtask 1.3: Implement `/api/auth/callback` callback handler that exchanges the auth code for an OIDC ID token.
- [x] Task 2: JWT Generation & Verification Middleware (AC: 4, 5)
  - [x] Subtask 2.1: Initialize HS256/RS256 JWT key signing logic in Go.
  - [x] Subtask 2.2: Sign and serialize a secure JWT token on successful callback containing user identification, OIDC token claims, and user roles (`viewer`, `operator`, `admin`).
  - [x] Subtask 2.3: Implement `AuthMiddleware` in Go that extracts the JWT token from the `Authorization: Bearer <token>` header and verifies its signature and expiration.
- [x] Task 3: Frontend SSO Redirection & JWT Session Caching (AC: 1, 3, 5)
  - [x] Subtask 3.1: Build `AuthContext` on the React/MUI dashboard client checking for the presence of a valid JWT token.
  - [x] Subtask 3.2: Cache the validated JWT token in browser `sessionStorage` upon successful callback redirection.
  - [x] Subtask 3.3: Configure global axios/fetch client to inject the JWT Bearer header for all subsequent API requests.

## Dev Notes

- **Language & Frameworks:** Go (Gateway API) and JavaScript (React UI with MUI).
- **Libraries Recommended:**
  - Go OIDC: Use `github.com/coreos/go-oidc/v3/oidc` or standard `golang.org/x/oauth2`.
  - Go JWT: Use `github.com/golang-jwt/jwt/v5`.
- **Database/Entity Impact:** In line with the Just-In-Time database rule, no database tables are required for this story since authentication is handled statelessly via JWT signatures verified by the Go gateway.
- **Source Paths to Create/Modify:**
  - Create: `backend/pkg/auth/oidc.go` (OIDC client configuration and handlers)
  - Create: `backend/pkg/auth/jwt.go` (JWT signature keys, signing, and claims parsing)
  - Create: `backend/pkg/auth/middleware.go` (HTTP authentication middleware)
  - Create: `frontend/src/context/AuthContext.js` (React auth provider)

### Project Structure Notes

- Keep all configuration structures driven by standard environment variables.
- The Go backend should listen on port `8080` by default.

### References

- [Source: _bmad-output/planning-artifacts/prds/prd-uss-surveillance-2026-07-08/prd.md#FR-15]
- [Source: _bmad-output/planning-artifacts/architecture/architecture-uss-surveillance-2026-07-09/ARCHITECTURE-SPINE.md#AD-8]

## Dev Agent Record

### Agent Model Used

gemini-2.5-pro

### Debug Log References
- /home/bangnq/.gemini/antigravity-cli/brain/b29dd0a5-8fa1-4047-b14d-171d279485af/.system_generated/logs/transcript.jsonl

### Completion Notes List
- Initialized greenfield Go backend and Vite React frontend projects.
- Implemented OIDC provider connection verification and oauth2 callbacks in Go gateway.
- Setup JWT HS256 creation and auth middleware logic, securing backend API routes.
- Wrote full unit test coverage for Go authentication and HTTP middleware. Tests passed with 100% green status.
- Implemented React AuthContext provider and wrapped the React virtual DOM tree.
- Built active Operator Portal login landing viewport and dashboard status overlays using MUI.

### File List
- backend/go.mod
- backend/go.sum
- backend/cmd/gateway/main.go
- backend/pkg/auth/oidc.go
- backend/pkg/auth/jwt.go
- backend/pkg/auth/jwt_test.go
- backend/pkg/auth/middleware.go
- backend/pkg/auth/middleware_test.go
- backend/pkg/auth/handlers.go
- frontend/package.json
- frontend/src/main.jsx
- frontend/src/App.jsx
- frontend/src/context/AuthContext.jsx
