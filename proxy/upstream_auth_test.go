package proxy

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/napmany/llmsnap/proxy/config"
)

// The inbound LLMSnap key is never forwarded to a managed upstream. Each
// managed model supplies its own backend key, independently of other models.
func TestManagedModelUpstreamAuthorization(t *testing.T) {
	t.Setenv("TEST_UPSTREAM_KEY_A", "backend-a-secret")
	t.Setenv("TEST_UPSTREAM_KEY_B", "backend-b-secret")

	tests := []struct {
		name     string
		envName  string
		expected string
		stream   bool
	}{
		{"backend A", "TEST_UPSTREAM_KEY_A", "Bearer backend-a-secret", false},
		{"backend B", "TEST_UPSTREAM_KEY_B", "Bearer backend-b-secret", false},
		{"no backend key", "", "", false},
		{"streaming backend A", "TEST_UPSTREAM_KEY_A", "Bearer backend-a-secret", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if got := r.Header.Get("Authorization"); got != tc.expected {
					t.Errorf("upstream Authorization = %q, want %q", got, tc.expected)
				}
				if got := r.Header.Get("x-api-key"); got != "" {
					t.Errorf("client x-api-key leaked to upstream")
				}
				if tc.stream {
					w.Header().Set("Content-Type", "text/event-stream")
					_, _ = w.Write([]byte("data: {\"ok\":true}\n\n"))
					if flusher, ok := w.(http.Flusher); ok {
						flusher.Flush()
					}
				} else {
					_, _ = w.Write([]byte("ok"))
				}
			}))
			defer upstream.Close()

			p := NewProcess("managed-model", 15, config.ModelConfig{
				Proxy: upstream.URL, UpstreamAPIKeyEnv: tc.envName,
			}, debugLogger, debugLogger)
			p.state = StateReady // Exercise forwarding without starting a subprocess.
			req := httptest.NewRequest("POST", "/v1/chat/completions", strings.NewReader("{}"))
			req.Header.Set("Authorization", "Bearer inbound-llmsnap-secret")
			req.Header.Set("x-api-key", "inbound-llmsnap-secret")
			response := httptest.NewRecorder()
			p.ProxyRequest(response, req)
			if response.Code != http.StatusOK {
				t.Fatalf("upstream status = %d, want 200; body: %s", response.Code, response.Body.String())
			}
			if tc.stream {
				if ct := response.Header().Get("Content-Type"); !strings.Contains(ct, "text/event-stream") {
					t.Errorf("streaming content type = %q", ct)
				}
				if !strings.Contains(response.Body.String(), "data:") {
					t.Errorf("streaming response body not forwarded: %s", response.Body.String())
				}
			}
		})
	}
}
