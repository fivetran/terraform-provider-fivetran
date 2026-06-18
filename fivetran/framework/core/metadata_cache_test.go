package core

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	fivetran "github.com/fivetran/go-fivetran"
)

const metadataDetailsBody = `{
	"code": "Success",
	"data": {
		"id": "google_sheets",
		"name": "Google Sheets",
		"type": "File",
		"config": {"properties": {"spreadsheet_id": {"type": "string", "fieldStatus": "private_preview"}}},
		"auth":   {"properties": {}}
	}
}`

func newMetadataServer(t *testing.T, service string, callCount *atomic.Int32, body string) (*httptest.Server, *fivetran.Client) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metadata/connector-types/"+service {
			if callCount != nil {
				callCount.Add(1)
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(body)) //nolint:errcheck
			return
		}
		http.NotFound(w, r)
	}))
	client := fivetran.New("key", "secret")
	// BaseURL replaces the full base (including /v1), so we point directly at the test server.
	client.BaseURL(srv.URL)
	t.Cleanup(srv.Close)
	return srv, client
}

func TestMetadataCache_HitReturnsCachedValue(t *testing.T) {
	t.Parallel()
	var callCount atomic.Int32
	_, client := newMetadataServer(t, "google_sheets", &callCount, metadataDetailsBody)
	cache := &sync.Map{}

	m1, err := GetCachedConnectorMetadata(context.Background(), client, cache, "google_sheets")
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	m2, err := GetCachedConnectorMetadata(context.Background(), client, cache, "google_sheets")
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if callCount.Load() != 1 {
		t.Errorf("expected exactly 1 API call, got %d", callCount.Load())
	}
	if m1 != m2 {
		t.Error("cached calls should return the same pointer")
	}
	if m1.ID != "google_sheets" {
		t.Errorf("unexpected ID: %v", m1.ID)
	}
	if got := m1.Config.Properties["spreadsheet_id"].FieldStatus; got != "private_preview" {
		t.Errorf("unexpected field status: got %q, want %q", got, "private_preview")
	}
}

func TestMetadataCache_TwoDistinctServicesMakeTwoCalls(t *testing.T) {
	t.Parallel()
	var count1, count2 atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/metadata/connector-types/google_sheets":
			count1.Add(1)
			w.Write([]byte(`{"code":"Success","data":{"id":"google_sheets","config":{"properties":{}},"auth":{"properties":{}}}}`)) //nolint:errcheck
		case "/metadata/connector-types/postgres":
			count2.Add(1)
			w.Write([]byte(`{"code":"Success","data":{"id":"postgres","config":{"properties":{}},"auth":{"properties":{}}}}`)) //nolint:errcheck
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	client := fivetran.New("key", "secret")
	client.BaseURL(srv.URL)
	cache := &sync.Map{}

	for i := 0; i < 3; i++ {
		if _, err := GetCachedConnectorMetadata(context.Background(), client, cache, "google_sheets"); err != nil {
			t.Fatalf("google_sheets call %d: %v", i, err)
		}
		if _, err := GetCachedConnectorMetadata(context.Background(), client, cache, "postgres"); err != nil {
			t.Fatalf("postgres call %d: %v", i, err)
		}
	}

	if count1.Load() != 1 {
		t.Errorf("google_sheets: expected 1 API call, got %d", count1.Load())
	}
	if count2.Load() != 1 {
		t.Errorf("postgres: expected 1 API call, got %d", count2.Load())
	}
}

func TestMetadataCache_ErrorNotCached(t *testing.T) {
	t.Parallel()
	var callCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := callCount.Add(1)
		if n == 1 {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(metadataDetailsBody)) //nolint:errcheck
	}))
	t.Cleanup(srv.Close)
	client := fivetran.New("key", "secret")
	client.BaseURL(srv.URL)
	cache := &sync.Map{}

	_, err := GetCachedConnectorMetadata(context.Background(), client, cache, "google_sheets")
	if err == nil {
		t.Fatal("expected error on first (failing) call")
	}

	m, err := GetCachedConnectorMetadata(context.Background(), client, cache, "google_sheets")
	if err != nil {
		t.Fatalf("second call should succeed after transient failure: %v", err)
	}
	if m == nil {
		t.Error("expected metadata on retry")
	}
	if callCount.Load() != 2 {
		t.Errorf("expected 2 API calls (error + retry), got %d", callCount.Load())
	}
}

func TestMetadataCache_ConcurrentSameService_SingleFetch(t *testing.T) {
	t.Parallel()
	var callCount atomic.Int32
	_, client := newMetadataServer(t, "google_sheets", &callCount, metadataDetailsBody)
	cache := &sync.Map{}

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)
	results := make([]error, goroutines)
	for i := 0; i < goroutines; i++ {
		i := i
		go func() {
			defer wg.Done()
			_, results[i] = GetCachedConnectorMetadata(context.Background(), client, cache, "google_sheets")
		}()
	}
	wg.Wait()

	for i, err := range results {
		if err != nil {
			t.Errorf("goroutine %d: %v", i, err)
		}
	}
	if callCount.Load() == 0 {
		t.Error("expected at least one API call")
	}
}

func TestMetadataCache_FreshPerProviderInstance(t *testing.T) {
	t.Parallel()
	var callCount atomic.Int32
	_, client := newMetadataServer(t, "google_sheets", &callCount, metadataDetailsBody)

	cache1 := &sync.Map{}
	cache2 := &sync.Map{}

	GetCachedConnectorMetadata(context.Background(), client, cache1, "google_sheets") //nolint:errcheck
	GetCachedConnectorMetadata(context.Background(), client, cache2, "google_sheets") //nolint:errcheck

	if callCount.Load() != 2 {
		t.Errorf("separate caches should each make one API call, got %d total", callCount.Load())
	}
}
