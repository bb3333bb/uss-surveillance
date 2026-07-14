package cors

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddlewareAnswersPreflightWithoutInvokingHandler(t *testing.T) {
	handlerCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	req := httptest.NewRequest(http.MethodOptions, "/api/operator/missions", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rr := httptest.NewRecorder()

	Middleware("http://localhost:5173")(next).ServeHTTP(rr, req)

	if handlerCalled {
		t.Error("expected the wrapped handler not to be invoked for an OPTIONS preflight")
	}
	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content, got %d", rr.Code)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Errorf("expected Access-Control-Allow-Origin header, got %q", got)
	}
	if got := rr.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Error("expected Access-Control-Allow-Headers to be set")
	}
}

func TestMiddlewarePassesThroughNonOptionsRequests(t *testing.T) {
	handlerCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusUnauthorized) // simulates AuthMiddleware still running for the real request
	})

	req := httptest.NewRequest(http.MethodGet, "/api/operator/missions", nil)
	rr := httptest.NewRecorder()

	Middleware("http://localhost:5173")(next).ServeHTTP(rr, req)

	if !handlerCalled {
		t.Error("expected the wrapped handler to be invoked for a non-OPTIONS request")
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Errorf("expected Access-Control-Allow-Origin header on the real response too, got %q", got)
	}
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected the downstream handler's status to pass through, got %d", rr.Code)
	}
}
