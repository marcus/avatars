package httpapi_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/marcus/avatars/internal/library"
)

func TestCompanionHTTPRecipeAndRefusals(t *testing.T) {
	h := handler(t)
	styles := call(h, "GET", "/api/v1/styles", "", nil)
	for _, token := range []string{`"id":"companions"`, `"native_width":128`, `"default":"dog"`, `"mixed_value":"mixed"`, `"label":"Cats"`} {
		if styles.Code != 200 || !strings.Contains(styles.Body.String(), token) {
			t.Fatalf("missing discovery %s: %s", token, styles.Body.String())
		}
	}
	for _, requested := range []string{"", "dog", "cat"} {
		body := `{"style":"companions","seed":"http-companion"}`
		if requested != "" {
			body = `{"style":"companions","seed":"http-companion","inputs":{"animal":"` + requested + `"}}`
		}
		created := call(h, "POST", "/api/v1/collections", body, nil)
		var collection library.Collection
		if created.Code != 201 || json.Unmarshal(created.Body.Bytes(), &collection) != nil {
			t.Fatal(created.Code, created.Body.String())
		}
		want := requested
		if want == "" {
			want = "dog"
		}
		item := collection.Avatars[0]
		if item.Inputs == nil || item.Inputs.Animal != want || item.Inputs.Color != "" {
			t.Fatalf("wrong saved inputs: %+v", item)
		}
		for _, format := range []string{"svg", "png"} {
			saved := call(h, "GET", "/api/v1/avatars/"+item.ID+"."+format+"?width=64&circle=true", "", nil)
			stateless := call(h, "GET", "/api/v1/render?style=companions&seed=http-companion&animal="+want+"&format="+format+"&width=64&circle=true", "", nil)
			if saved.Code != 200 || stateless.Code != 200 || !bytes.Equal(saved.Body.Bytes(), stateless.Body.Bytes()) {
				t.Fatalf("saved/stateless %s %s mismatch", want, format)
			}
		}
		for _, query := range []string{"animal=dog", "color=sage"} {
			rejected := call(h, "GET", item.SVGURL+"?"+query, "", nil)
			if rejected.Code != 400 {
				t.Fatal("saved export accepted appearance override", query)
			}
		}
	}
	for _, path := range []string{
		"/api/v1/render?style=companions&seed=x&animal=fox",
		"/api/v1/render?style=companions&seed=x&animal=cat&animal=dog",
		"/api/v1/render?style=companions&seed=x&animal=cat&color=sage",
		"/api/v1/render?style=pebble&seed=x&animal=cat",
		"/api/v1/render?style=gorey&seed=x&animal=dog",
	} {
		if w := call(h, "GET", path, "", nil); w.Code != 400 {
			t.Fatalf("accepted %s: %d", path, w.Code)
		}
	}
	rejecting := handler(t)
	for _, body := range []string{
		`{"style":"companions","inputs":{"animal":"fox"}}`,
		`{"style":"companions","inputs":{"animal":["cat"]}}`,
		`{"style":"companions","inputs":{"animal":"cat","unknown":true}}`,
		`{"style":"companions","inputs":{"color":"sage"}}`,
		`{"style":"pebble","inputs":{"animal":"cat"}}`,
		`{"style":"gorey","inputs":{"animal":"dog"}}`,
	} {
		if w := call(rejecting, "POST", "/api/v1/collections", body, nil); w.Code != 400 {
			t.Fatalf("accepted %s: %d %s", body, w.Code, w.Body.String())
		}
	}
	list := call(rejecting, "GET", "/api/v1/collections", "", nil)
	if list.Body.String() != "{\"collections\":[]}\n" {
		t.Fatal("invalid request saved a record", list.Body.String())
	}
}
