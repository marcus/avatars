package httpapi_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/marcus/avatars/internal/httpapi"
	"github.com/marcus/avatars/internal/library"
	"github.com/marcus/avatars/internal/store"
	"github.com/marcus/avatars/internal/studio"
	"github.com/marcus/avatars/pkg/avatar"
)

func handler(t *testing.T) http.Handler {
	t.Helper()
	s, e := store.New(filepath.Join(t.TempDir(), "collections.jsonl"))
	if e != nil {
		t.Fatal(e)
	}
	engine := avatar.New()
	return httpapi.New(library.New(engine, s), engine, studio.Handler())
}
func call(h http.Handler, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, "http://127.0.0.1:7447"+path, strings.NewReader(body))
	if body != "" {
		r.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		if k == "Host" {
			r.Host = v
		} else {
			r.Header.Set(k, v)
		}
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}
func TestAPIContractAndStudio(t *testing.T) {
	h := handler(t)
	w := call(h, "POST", "/api/v1/collections", `{"style":"gorey","count":3,"seed":"agent"}`, nil)
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body)
	}
	var c library.Collection
	if e := json.Unmarshal(w.Body.Bytes(), &c); e != nil {
		t.Fatal(e)
	}
	if len(c.Avatars) != 3 {
		t.Fatal(c)
	}
	for _, path := range []string{"/?collection=" + c.ID, "/?avatar=" + c.Avatars[0].ID, "/api/v1/styles", "/api/v1/capabilities", "/api/v1/instructions", "/api/v1/collections/" + c.ID, "/api/v1/avatars/" + c.Avatars[0].ID} {
		w := call(h, "GET", path, "", nil)
		if w.Code != 200 {
			t.Fatal(path, w.Code, w.Body)
		}
	}
	w = call(h, "GET", c.Avatars[0].SVGURL, "", nil)
	expected := call(h, "GET", "/api/v1/render?seed=agent%3A0", "", nil)
	if !bytes.Equal(w.Body.Bytes(), expected.Body.Bytes()) || w.Header().Get("Content-Type") != "image/svg+xml" {
		t.Fatal("rendering parity failure")
	}
	for _, query := range []string{"format=png", "style=missing", "seed=ignored"} {
		w = call(h, "GET", c.Avatars[0].SVGURL+"?"+query, "", nil)
		if w.Code != 400 {
			t.Fatal("ignored invalid export query", query, w.Code, w.Body)
		}
	}
	for _, path := range []string{"/api/v1/collections/nope", "/api/v1/avatars/nope", "/api/v1/nope"} {
		w = call(h, "GET", path, "", nil)
		if w.Code != 404 || !strings.Contains(w.Body.String(), `"code":"not_found"`) {
			t.Fatal(path, w.Code, w.Body)
		}
	}
}
func TestRejectInvalidAndCrossOrigin(t *testing.T) {
	h := handler(t)
	for _, body := range []string{`null`, `[]`, `{"count":0}`, `{"count":101}`, `{"count":-1}`, `{"count":1,"typo":true}`, `{} {}`, `{"style":"absent"}`, `{"name":"` + strings.Repeat("x", 121) + `"}`} {
		w := call(h, "POST", "/api/v1/collections", body, nil)
		if w.Code != 400 {
			t.Fatal(body, w.Code, w.Body)
		}
	}
	for _, path := range []string{"/api/v1/render", "/api/v1/render?seed=x&width=0", "/api/v1/render?seed=x&width=2049", "/api/v1/render?seed=x&width=4&width=9", "/api/v1/render?seed=x&circle=maybe", "/api/v1/render?seed=x&formt=png"} {
		w := call(h, "GET", path, "", nil)
		if w.Code != 400 {
			t.Fatal(path, w.Code, w.Body)
		}
	}
	for _, headers := range []map[string]string{{"Origin": "https://evil.example"}, {"Sec-Fetch-Site": "cross-site"}, {"Host": "attacker.example:7447"}} {
		w := call(h, "POST", "/api/v1/collections", `{}`, headers)
		if w.Code != 403 {
			t.Fatal(headers, w.Code)
		}
	}
	w := call(h, "POST", "/api/v1/collections", `{}`, map[string]string{"Origin": "http://127.0.0.1:7447"})
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body)
	}
	server := httptest.NewServer(h)
	defer server.Close()
	res, e := http.Get(server.URL + "/api/v1/health")
	if e != nil {
		t.Fatal(e)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	if res.StatusCode != 200 || !bytes.Contains(b, []byte(`"service":"avatars"`)) {
		t.Fatal(res.StatusCode, string(b))
	}
}

func TestTrustedProxyOrigin(t *testing.T) {
	s, e := store.New(filepath.Join(t.TempDir(), "collections.jsonl"))
	if e != nil {
		t.Fatal(e)
	}
	engine := avatar.New()
	h := httpapi.NewWithConfig(library.New(engine, s), engine, studio.Handler(), httpapi.Config{PublicURL: "https://STUDIO.example:7447"})
	for _, host := range []string{"127.0.0.1:7447", "studio.example:7447"} {
		w := call(h, "POST", "/api/v1/collections", `{}`, map[string]string{"Host": host, "Origin": "https://studio.example:7447", "Sec-Fetch-Site": "same-origin"})
		if w.Code != 201 {
			t.Fatal(host, w.Code, w.Body)
		}
	}
	for _, headers := range []map[string]string{
		{"Host": "evil.example:7447"},
		{"Host": "studio.example:7447", "Origin": "http://studio.example:7447"},
		{"Host": "studio.example:7447", "Origin": "https://evil.example:7447"},
		{"Host": "studio.example:7447", "Origin": "https://studio.example:7447", "Sec-Fetch-Site": "cross-site"},
		{"Host": "127.0.0.1:7447", "Origin": "https://evil.example", "X-Forwarded-Host": "evil.example", "X-Forwarded-Proto": "https"},
	} {
		w := call(h, "POST", "/api/v1/collections", `{}`, headers)
		if w.Code != 403 {
			t.Fatal(headers, w.Code, w.Body)
		}
	}
}

func TestPublicOriginCanonicalization(t *testing.T) {
	for _, input := range []string{"https://Studio.Example:443", "https://studio.example/", "https://STUDIO.example"} {
		got, err := httpapi.NormalizePublicURL(input)
		if err != nil || got != "https://studio.example" {
			t.Fatal(input, got, err)
		}
	}
	for _, input := range []string{"http://studio.example", "https://studio.example:0", "https://studio.example:65536", "https://user@studio.example", "https://studio.example/path", "https://studio.example?x=1", "https://studio.example#x", "https://*.example"} {
		if _, err := httpapi.NormalizePublicURL(input); err == nil {
			t.Fatal("accepted invalid origin", input)
		}
	}
	s, e := store.New(filepath.Join(t.TempDir(), "collections.jsonl"))
	if e != nil {
		t.Fatal(e)
	}
	engine := avatar.New()
	h := httpapi.NewWithConfig(library.New(engine, s), engine, studio.Handler(), httpapi.Config{PublicURL: "https://Studio.Example:443/"})
	for _, host := range []string{"studio.example", "STUDIO.example:443", "127.0.0.1:7447"} {
		w := call(h, "POST", "/api/v1/collections", `{}`, map[string]string{"Host": host, "Origin": "https://studio.example"})
		if w.Code != 201 {
			t.Fatal(host, w.Code, w.Body)
		}
	}
}
