package service

import (
	"errors"
	"testing"
	"time"
	"urlcutter/internal/models"
)

type mockRepository struct {
	shortToURL    map[string]*models.URL
	originalToURL map[string]*models.URL
	incremented   []string
	createErr     error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		shortToURL:    make(map[string]*models.URL),
		originalToURL: make(map[string]*models.URL),
		incremented:   []string{},
	}
}

func (m *mockRepository) Create(u *models.URL) error {
	if m.createErr != nil {
		return m.createErr
	}
	if u == nil {
		return errors.New("nil url")
	}
	m.shortToURL[u.Short] = u
	m.originalToURL[u.Original] = u
	return nil
}

func (m *mockRepository) FindByShort(short string) (*models.URL, error) {
	if u, ok := m.shortToURL[short]; ok {
		return u, nil
	}
	return nil, nil
}

func (m *mockRepository) FindByOriginal(original string) (*models.URL, error) {
	if u, ok := m.originalToURL[original]; ok {
		return u, nil
	}
	return nil, nil
}

func (m *mockRepository) IncrementClicks(short string) error {
	m.incremented = append(m.incremented, short)
	if u, ok := m.shortToURL[short]; ok {
		u.Clicks++
	}
	return nil
}

func TestCreateShortURL_New(t *testing.T) {
	repo := newMockRepository()
	svc := NewURLService(repo)

	resp, err := svc.CreateShortURL("https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil || resp.ShortURL == "" {
		t.Fatalf("expected short url")
	}
	if len(repo.shortToURL) != 1 {
		t.Fatalf("expected 1 url in repo")
	}
}

func TestCreateShortURL_Existing(t *testing.T) {
	repo := newMockRepository()
	existing := &models.URL{Id: "abc123", Original: "https://example.com", Short: "abc123", CreatedAt: time.Now()}
	_ = repo.Create(existing)
	svc := NewURLService(repo)

	resp, err := svc.CreateShortURL("https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ShortURL != "abc123" {
		t.Fatalf("expected existing short 'abc123', got %q", resp.ShortURL)
	}
}

func TestCreateShortURL_Invalid(t *testing.T) {
	repo := newMockRepository()
	svc := NewURLService(repo)
	if _, err := svc.CreateShortURL("not-a-url"); err == nil {
		t.Fatalf("expected error for invalid URL")
	}
}

func TestGetOriginalURL(t *testing.T) {
	repo := newMockRepository()
	_ = repo.Create(&models.URL{Id: "abc123", Original: "https://example.com", Short: "abc123", CreatedAt: time.Now()})
	svc := NewURLService(repo)

	orig, err := svc.GetOriginalURL("abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if orig != "https://example.com" {
		t.Fatalf("expected original, got %q", orig)
	}
}

func TestGetOriginalURL_NotFound(t *testing.T) {
	repo := newMockRepository()
	svc := NewURLService(repo)
	if _, err := svc.GetOriginalURL("missing"); err == nil {
		t.Fatalf("expected not found error")
	}
}

func TestRedirect_IncrementsClicks(t *testing.T) {
	repo := newMockRepository()
	_ = repo.Create(&models.URL{Id: "abc123", Original: "https://example.com", Short: "abc123", CreatedAt: time.Now()})
	svc := NewURLService(repo)

	orig, err := svc.Redirect("abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if orig != "https://example.com" {
		t.Fatalf("unexpected original: %q", orig)
	}
	if repo.shortToURL["abc123"].Clicks != 1 {
		t.Fatalf("expected clicks incremented")
	}
}
