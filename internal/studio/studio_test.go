package studio

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEmbeddedStudioRoutes(t *testing.T) {
	for _, target := range []string{"/", "/?collection=col_example", "/?avatar=av_example", "/app.js", "/styles.css"} {
		t.Run(target, func(t *testing.T) {
			response := httptest.NewRecorder()
			Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, target, nil))
			if response.Code != http.StatusOK || response.Body.Len() == 0 {
				t.Fatalf("embedded studio %s: status %d, %d bytes", target, response.Code, response.Body.Len())
			}
			if response.Header().Get("Content-Security-Policy") == "" {
				t.Fatal("studio must constrain embedded resources to this origin")
			}
			if strings.HasPrefix(target, "/?") && !strings.Contains(response.Body.String(), "<title>Avatars Studio</title>") {
				t.Fatal("deep links must open the studio document")
			}
		})
	}
}

func TestStudioDoesNotSwallowUnknownRoutes(t *testing.T) {
	for _, target := range []string{"/missing", "/assets/", "/studio.go", "/api/v1/missing"} {
		response := httptest.NewRecorder()
		Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, target, nil))
		if response.Code != http.StatusNotFound {
			t.Errorf("%s: got status %d, want 404", target, response.Code)
		}
	}
	response := httptest.NewRecorder()
	Handler().ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/", nil))
	if response.Code != http.StatusMethodNotAllowed || response.Header().Get("Allow") != "GET, HEAD" {
		t.Fatal("studio must reject writes and advertise its supported methods")
	}
}
