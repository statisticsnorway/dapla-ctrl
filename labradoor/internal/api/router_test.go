package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBearerAuth(t *testing.T) {
	const token = "expected-token"

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
		wantNext   bool
	}{
		{
			name:       "missing authorization header",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "malformed authorization header",
			authHeader: "Bearer",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "wrong auth scheme",
			authHeader: "Basic expected-token",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "wrong token",
			authHeader: "Bearer wrong-token",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "too many fields",
			authHeader: "Bearer expected-token extra",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "valid bearer token",
			authHeader: "Bearer expected-token",
			wantStatus: http.StatusNoContent,
			wantNext:   true,
		},
		{
			name:       "valid bearer token with spongebob casing",
			authHeader: "bEArEr expected-token",
			wantStatus: http.StatusNoContent,
			wantNext:   true,
		},
		{
			name:       "valid bearer token with extra whitespace",
			authHeader: "Bearer   expected-token",
			wantStatus: http.StatusNoContent,
			wantNext:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusNoContent)
			})

			handler := BearerAuth(token)(next)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if nextCalled != tt.wantNext {
				t.Fatalf("nextCalled = %v, want %v", nextCalled, tt.wantNext)
			}
		})
	}
}
