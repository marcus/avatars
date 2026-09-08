package studio

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEmbeddedStudioRoutes(t *testing.T) {
	for _, target := range []string{"/", "/?collection=col_example", "/?avatar=av_example&shape=circle", "/app.js", "/view.mjs", "/styles.css", "/card-grid.css", "/card-grid.mjs", "/card-physics.mjs", "/card-sound.mjs", "/vendor/matter/matter.mjs"} {
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

func TestStudioCardAssetsUseBrowserMIMETypes(t *testing.T) {
	for target, expected := range map[string]string{
		"/card-grid.css":             "text/css",
		"/card-grid.mjs":             "text/javascript",
		"/card-physics.mjs":          "text/javascript",
		"/card-sound.mjs":            "text/javascript",
		"/vendor/matter/matter.mjs": "text/javascript",
	} {
		response := httptest.NewRecorder()
		Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, target, nil))
		if contentType := response.Header().Get("Content-Type"); !strings.HasPrefix(contentType, expected) {
			t.Errorf("%s: content type %q, want %q", target, contentType, expected)
		}
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
