// Package httpapi exposes the library to browser and programmatic clients.
package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/marcus/avatars/internal/buildinfo"
	"github.com/marcus/avatars/internal/discovery"
	"github.com/marcus/avatars/internal/library"
	"github.com/marcus/avatars/internal/lifecycle"
	"github.com/marcus/avatars/pkg/avatar"
)

type API struct {
	library *library.Service
	engine  *avatar.Engine
}

// Config declares the external origin of a trusted proxy such as Tailscale Serve.
// The listener remains loopback-only; forwarded headers never grant access.
type Config struct {
	PublicURL string
	Status    *lifecycle.Status
}

func New(service *library.Service, engine *avatar.Engine, studio http.Handler) http.Handler {
	return NewWithConfig(service, engine, studio, Config{})
}

func NewWithConfig(service *library.Service, engine *avatar.Engine, studio http.Handler, config Config) http.Handler {
	publicOrigin, _ := NormalizePublicURL(config.PublicURL)

	a := &API{service, engine}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		if config.Status != nil {
			JSON(w, 200, config.Status)
			return
		}
		JSON(w, 200, lifecycle.Status{Service: "avatars", APIVersion: "v1", Version: buildinfo.Version, Commit: buildinfo.Commit})
	})
	mux.HandleFunc("GET /api/v1/styles", func(w http.ResponseWriter, r *http.Request) {
		JSON(w, 200, map[string]any{"styles": engine.Styles(), "formats": engine.Formats()})
	})
	mux.HandleFunc("GET /api/v1/capabilities", func(w http.ResponseWriter, r *http.Request) { JSON(w, 200, discovery.Capabilities()) })
	mux.HandleFunc("GET /api/v1/instructions", func(w http.ResponseWriter, r *http.Request) {
		JSON(w, 200, map[string]string{"instructions": discovery.Instructions})
	})
	mux.HandleFunc("GET /api/v1/collections", a.list)
	mux.HandleFunc("POST /api/v1/collections", a.create)
	mux.HandleFunc("GET /api/v1/collections/{id}", a.collection)
	mux.HandleFunc("GET /api/v1/avatars/{id}", a.portrait)
	mux.HandleFunc("GET /api/v1/render", a.render)
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		WriteError(w, &library.Error{Code: "not_found", Message: "API route not found"})
	})
	mux.Handle("/", studio)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self'; script-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'")
		w.Header().Set("Cache-Control", "no-store")
		host := strings.ToLower(r.Host)
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}
		localHost := host == "localhost" || (net.ParseIP(strings.Trim(host, "[]")) != nil && net.ParseIP(strings.Trim(host, "[]")).IsLoopback())
		requestOrigin, _ := NormalizePublicURL("https://" + r.Host)
		if !localHost && (publicOrigin == "" || requestOrigin != publicOrigin) {
			JSON(w, 403, map[string]any{"error": map[string]string{"code": "invalid_request", "message": "the service accepts only loopback hosts or its configured public host"}})
			return
		}
		if r.Method != "GET" && r.Method != "HEAD" {
			origin := r.Header.Get("Origin")
			sameOrigin := origin == "http://"+r.Host && localHost
			normalizedOrigin, _ := NormalizePublicURL(origin)
			trustedProxyOrigin := publicOrigin != "" && normalizedOrigin == publicOrigin
			if r.Header.Get("Sec-Fetch-Site") == "cross-site" || (origin != "" && !sameOrigin && !trustedProxyOrigin) {
				JSON(w, 403, map[string]any{"error": map[string]string{"code": "invalid_request", "message": "cross-origin mutations are not allowed"}})
				return
			}
		}
		mux.ServeHTTP(w, r)
	})
}

func JSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func WriteError(w http.ResponseWriter, err error) {
	status := 500
	d := &library.Error{Code: "internal", Message: "operation failed"}
	if errors.As(err, &d) {
		switch d.Code {
		case "invalid_request":
			status = 400
		case "not_found":
			status = 404
		}
	}
	JSON(w, status, map[string]any{"error": d})
}
func invalid(message string) error { return &library.Error{Code: "invalid_request", Message: message} }
func (a *API) list(w http.ResponseWriter, r *http.Request) {
	v, err := a.library.List(r.Context())
	if err != nil {
		WriteError(w, err)
		return
	}
	JSON(w, 200, map[string]any{"collections": v})
}
func (a *API) collection(w http.ResponseWriter, r *http.Request) {
	v, err := a.library.Collection(r.Context(), r.PathValue("id"))
	if err != nil {
		WriteError(w, err)
		return
	}
	JSON(w, 200, v)
}
func (a *API) create(w http.ResponseWriter, r *http.Request) {
	media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || media != "application/json" {
		WriteError(w, invalid("Content-Type must be application/json"))
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	var body *struct {
		library.CreateRequest
		Count *int `json:"count"`
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&body); err != nil {
		WriteError(w, invalid("invalid JSON request: "+err.Error()))
		return
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		WriteError(w, invalid("request must contain one JSON object"))
		return
	}
	if body == nil {
		WriteError(w, invalid("request must be a JSON object"))
		return
	}
	req := body.CreateRequest
	if body.Count != nil {
		if *body.Count < 1 || *body.Count > 100 {
			WriteError(w, invalid("count must be 1..100"))
			return
		}
		req.Count = *body.Count
	}
	v, err := a.library.Create(r.Context(), req)
	if err != nil {
		WriteError(w, err)
		return
	}
	w.Header().Set("Location", "/api/v1/collections/"+v.ID)
	JSON(w, 201, v)
}
func Options(q url.Values) (avatar.Options, error) {
	var o avatar.Options
	for key := range q {
		if key != "width" && key != "height" && key != "circle" && key != "seed" && key != "style" && key != "format" && key != "color" && key != "animal" {
			return o, invalid("unknown query parameter: " + key)
		}
		if len(q[key]) != 1 {
			return o, invalid("query parameter must appear once: " + key)
		}
	}
	for name, target := range map[string]*int{"width": &o.Width, "height": &o.Height} {
		if q.Has(name) {
			n, err := strconv.Atoi(q.Get(name))
			if err != nil || n < 1 || n > 2048 {
				return o, invalid(name + " must be 1..2048")
			}
			*target = n
		}
	}
	if q.Has("circle") {
		v, err := strconv.ParseBool(q.Get("circle"))
		if err != nil {
			return o, invalid("circle must be true or false")
		}
		o.Circle = v
	}
	return o, nil
}
func (a *API) portrait(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	format := ""
	if i := strings.LastIndexByte(id, '.'); i >= 0 {
		format = id[i+1:]
		id = id[:i]
		if format != "svg" && format != "png" {
			WriteError(w, invalid("format must be svg or png"))
			return
		}
	}
	if format == "" {
		v, err := a.library.Avatar(r.Context(), id)
		if err != nil {
			WriteError(w, err)
			return
		}
		JSON(w, 200, v)
		return
	}
	for key := range r.URL.Query() {
		if key != "width" && key != "height" && key != "circle" {
			WriteError(w, invalid("unknown export query parameter: "+key))
			return
		}
	}
	o, err := Options(r.URL.Query())
	if err != nil {
		WriteError(w, err)
		return
	}
	b, err := a.library.Render(r.Context(), id, format, o)
	if err != nil {
		WriteError(w, err)
		return
	}
	sendImage(w, b, format, id)
}
func (a *API) render(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	o, err := Options(q)
	if err != nil {
		WriteError(w, err)
		return
	}
	if !q.Has("seed") || len(q.Get("seed")) > 4096 {
		WriteError(w, invalid("seed is required and must be at most 4096 bytes"))
		return
	}
	style := q.Get("style")
	if style == "" {
		style = "gorey"
	}
	format := q.Get("format")
	if format == "" {
		format = "svg"
	}
	b, err := a.engine.RenderWithInputs(r.Context(), style, q.Get("seed"), format, avatar.Inputs{Color: q.Get("color"), Animal: q.Get("animal")}, o)
	if err != nil {
		WriteError(w, invalid(err.Error()))
		return
	}
	sendImage(w, b, format, "avatar")
}
func sendImage(w http.ResponseWriter, b []byte, format, id string) {
	media := "image/svg+xml"
	if format == "png" {
		media = "image/png"
	}
	w.Header().Set("Content-Type", media)
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", id+"."+format))
	w.WriteHeader(200)
	_, _ = w.Write(b)
}
