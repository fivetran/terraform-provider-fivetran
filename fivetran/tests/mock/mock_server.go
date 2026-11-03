package mock

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// NewMockServer serves the handlers registered on MockClient over HTTP, for provider binaries
// that run outside the test process, such as released versions in ExternalProviders. Configure
// them with api_url = "<server URL>/v1": the handlers match on method and path only.
func NewMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	var mu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		resp, err := mockClient.Do(req)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotImplemented)
			fmt.Fprintf(w, `{"code": "NotImplemented", "message": %q}`, err.Error())
			return
		}
		defer resp.Body.Close()

		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		w.WriteHeader(resp.StatusCode)
		if _, err := io.Copy(w, resp.Body); err != nil {
			t.Errorf("mock server: writing response: %v", err)
		}
	}))
	t.Cleanup(server.Close)
	return server
}
