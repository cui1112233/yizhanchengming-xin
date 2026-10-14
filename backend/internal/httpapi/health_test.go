package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthReturnsOKWhenMySQLIsReachable(t *testing.T) {
	handler := NewRouter(RouterOptions{
		Health: func(context.Context) error { return nil },
	})

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
}

func TestHealthReturnsServiceUnavailableWhenMySQLIsUnreachable(t *testing.T) {
	handler := NewRouter(RouterOptions{
		Health: func(context.Context) error { return errors.New("mysql is unavailable") },
	})

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}
