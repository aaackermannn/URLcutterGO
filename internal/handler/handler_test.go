package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"urlcutter/internal/models"
)

type mockService struct {
	createResp       *models.CreateURLResponse
	createErr        error
	original         string
	getErr           error
	redirectOriginal string
	redirectErr      error
}

func (m *mockService) CreateShortURL(original string) (*models.CreateURLResponse, error) {
	return m.createResp, m.createErr
}
func (m *mockService) GetOriginalURL(short string) (string, error) { return m.original, m.getErr }
func (m *mockService) Redirect(short string) (string, error) {
	return m.redirectOriginal, m.redirectErr
}

func TestCreateShortURL_OK(t *testing.T) {
	svc := &mockService{createResp: &models.CreateURLResponse{ShortURL: "abc123"}}
	h := NewHandler(svc)

	body, _ := json.Marshal(models.CreateURLRequest{URL: "https://example.com"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/shorten", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.CreateShortURL(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
}

func TestCreateShortURL_BadRequest(t *testing.T) {
	svc := &mockService{createErr: http.ErrNotSupported}
	h := NewHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/shorten", bytes.NewReader([]byte("{")))
	rr := httptest.NewRecorder()
	h.CreateShortURL(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestGetURLInfo_OK(t *testing.T) {
	svc := &mockService{original: "https://example.com"}
	h := NewHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/url/abc123", nil)
	rr := httptest.NewRecorder()

	// simulate mux vars
	req = muxSetVar(req, "short", "abc123")
	h.GetURLInfo(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestRedirect_NotFound(t *testing.T) {
	svc := &mockService{redirectErr: http.ErrMissingFile}
	h := NewHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	rr := httptest.NewRecorder()
	req = muxSetVar(req, "short", "abc123")
	h.Redirect(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// helper to inject mux vars without importing mux in test
func muxSetVar(r *http.Request, k, v string) *http.Request {
	ctx := r.Context()
	type muxKey struct{}
	return r.WithContext(context.WithValue(ctx, muxKey{}, map[string]string{k: v}))
}
