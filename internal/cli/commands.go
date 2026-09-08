package cli

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/marcus/avatars/internal/library"
	"github.com/marcus/avatars/pkg/avatar"
)

func (a *app) styles(ctx context.Context) error {
	v := struct {
		Styles  []avatar.Style `json:"styles"`
		Formats []string       `json:"formats"`
	}{a.engine.Styles(), a.engine.Formats()}
	if a.remote {
		if _, e := a.request(ctx, "GET", "/api/v1/styles", nil, &v); e != nil {
			return e
		}
	}
	if a.json {
		return a.emit(v)
	}
	for _, s := range v.Styles {
		fmt.Fprintf(a.out, "%s  %s\n  %s\n", s.ID, s.Name, s.Description)
	}
	fmt.Fprintln(a.out, "Formats:", strings.Join(v.Formats, ", "))
	return nil
}
func (a *app) generate(ctx context.Context, args []string) error {
	f := a.flags("generate")
	req := library.CreateRequest{Style: "gorey", Count: 1}
	var file, format, size, color string
	var circle bool
	f.StringVar(&req.Style, "style", "gorey", "style")
	f.IntVar(&req.Count, "count", 1, "batch size")
	f.StringVar(&req.Name, "name", "", "collection name")
	f.StringVar(&req.Seed, "seed", "", "seed")
	f.StringVar(&req.Seed, "s", "", "seed")
	f.StringVar(&color, "color", "", "style color")
	f.StringVar(&file, "out", "", "file")
	f.StringVar(&file, "o", "", "file")
	f.StringVar(&format, "format", "svg", "format")
	f.StringVar(&format, "f", "svg", "format")
	f.StringVar(&size, "size", "", "dimensions")
	f.BoolVar(&circle, "circle", false, "circle")
	f.BoolVar(&circle, "c", false, "circle")
	if e := parse(f, args); e != nil {
		return e
	}
	if f.NArg() > 1 {
		return bad("generate accepts at most one seed")
	}
	if f.NArg() == 1 {
		if req.Seed != "" {
			return bad("provide the seed once")
		}
		req.Seed = f.Arg(0)
	}
	if req.Count < 1 || req.Count > 100 {
		return bad("count must be 1..100")
	}
	o, e := dimensions(size, circle)
	if e != nil {
		return e
	}
	if format != "svg" && format != "png" {
		return bad("format must be svg or png")
	}
	if file == "" && (size != "" || circle || format != "svg") {
		return bad("generate export options require --out; use export for saved avatars")
	}
	if file != "" && req.Count != 1 {
		return bad("generate --out requires --count 1")
	}
	if color != "" {
		req.Inputs = &avatar.Inputs{Color: color}
	}
	if file == "-" && a.json {
		return bad("generate --out - cannot be combined with --json")
	}
	// Check an explicit output destination before saving a collection.
	if file != "" && file != "-" {
		if _, e := os.Lstat(file); e == nil {
			return bad("output file already exists: " + file)
		} else if !os.IsNotExist(e) {
			return e
		}
	}
	var v library.Collection
	if a.remote {
		_, e = a.request(ctx, "POST", "/api/v1/collections", req, &v)
	} else {
		e = a.local()
		if e == nil {
			v, e = a.service.Create(ctx, req)
		}
	}
	if e != nil {
		return e
	}
	if file != "" {
		b, e := a.renderSaved(ctx, v.Avatars[0].ID, format, o)
		if e != nil {
			return fmt.Errorf("collection %s was saved; export failed: %w", v.ID, e)
		}
		if e = a.writeImage(b, format, file, false); e != nil {
			return fmt.Errorf("collection %s was saved; export failed: %w", v.ID, e)
		}
	}
	v = a.linkedCollection(v)
	if a.json {
		return a.emit(v)
	}
	out := a.out
	if file == "-" {
		out = a.errOut
	}
	fmt.Fprintf(out, "%s · %d avatar(s)\n%s\n", v.Name, len(v.Avatars), v.URL)
	for _, p := range v.Avatars {
		fmt.Fprintf(out, "  %s  %s\n", p.ID, p.URL)
	}
	return nil
}
func (a *app) list(ctx context.Context) error {
	var result struct {
		Collections []library.Collection `json:"collections"`
	}
	var e error
	if a.remote {
		_, e = a.request(ctx, "GET", "/api/v1/collections", nil, &result)
	} else {
		e = a.local()
		if e == nil {
			result.Collections, e = a.service.List(ctx)
		}
	}
	if e != nil {
		return e
	}
	for i := range result.Collections {
		result.Collections[i] = a.linkedCollection(result.Collections[i])
	}
	if a.json {
		return a.emit(result)
	}
	if len(result.Collections) == 0 {
		fmt.Fprintln(a.out, "No collections yet. Run avatars generate --count 12.")
	}
	for _, v := range result.Collections {
		fmt.Fprintf(a.out, "%s  %3d  %s\n  %s\n", v.ID, len(v.Avatars), v.Name, v.URL)
	}
	return nil
}
func (a *app) show(ctx context.Context, id string) error {
	if strings.HasPrefix(id, "col_") {
		var v library.Collection
		var e error
		if a.remote {
			_, e = a.request(ctx, "GET", "/api/v1/collections/"+url.PathEscape(id), nil, &v)
		} else {
			e = a.local()
			if e == nil {
				v, e = a.service.Collection(ctx, id)
			}
		}
		if e != nil {
			return e
		}
		return a.emit(a.linkedCollection(v))
	}
	var v library.Avatar
	var e error
	if a.remote {
		_, e = a.request(ctx, "GET", "/api/v1/avatars/"+url.PathEscape(id), nil, &v)
	} else {
		e = a.local()
		if e == nil {
			v, e = a.service.Avatar(ctx, id)
		}
	}
	if e != nil {
		return e
	}
	return a.emit(a.linkedAvatar(v))
}
func dimensions(size string, circle bool) (avatar.Options, error) {
	o := avatar.Options{Circle: circle}
	if size == "" {
		return o, nil
	}
	w, h, both := strings.Cut(size, "x")
	if !both {
		h = w
	}
	var e error
	o.Width, e = strconv.Atoi(w)
	if e != nil {
		return o, bad("size must be N or WxH")
	}
	o.Height, e = strconv.Atoi(h)
	if e != nil || o.Width < 1 || o.Height < 1 || o.Width > 2048 || o.Height > 2048 {
		return o, bad("dimensions must be 1..2048")
	}
	return o, nil
}
func optionsQuery(o avatar.Options) url.Values {
	q := url.Values{}
	if o.Width != 0 {
		q.Set("width", strconv.Itoa(o.Width))
	}
	if o.Height != 0 {
		q.Set("height", strconv.Itoa(o.Height))
	}
	if o.Circle {
		q.Set("circle", "true")
	}
	return q
}
func (a *app) renderSaved(ctx context.Context, id, format string, o avatar.Options) ([]byte, error) {
	if a.remote {
		return a.request(ctx, "GET", "/api/v1/avatars/"+url.PathEscape(id)+"."+format+"?"+optionsQuery(o).Encode(), nil, nil)
	}
	if e := a.local(); e != nil {
		return nil, e
	}
	return a.service.Render(ctx, id, format, o)
}
func (a *app) image(ctx context.Context, command string, args []string) error {
	f := a.flags(command)
	var seed, style, format, size, file, color string
	var circle bool
	f.StringVar(&format, "format", "svg", "format")
	f.StringVar(&format, "f", "svg", "format")
	f.StringVar(&size, "size", "", "dimensions")
	f.StringVar(&file, "out", "", "file")
	f.StringVar(&file, "o", "", "file")
	f.BoolVar(&circle, "circle", false, "circle")
	f.BoolVar(&circle, "c", false, "circle")
	if command == "render" {
		f.StringVar(&seed, "seed", "", "seed")
		f.StringVar(&seed, "s", "", "seed")
		f.StringVar(&style, "style", "gorey", "style")
		f.StringVar(&color, "color", "", "style color")
	}
	if e := parse(f, args); e != nil {
		return e
	}
	o, e := dimensions(size, circle)
	if e != nil {
		return e
	}
	if format != "svg" && format != "png" {
		return bad("format must be svg or png")
	}
	var b []byte
	if command == "export" {
		if f.NArg() != 1 {
			return bad("export requires an avatar ID")
		}
		b, e = a.renderSaved(ctx, f.Arg(0), format, o)
	} else {
		provided := false
		f.Visit(func(v *flag.Flag) {
			if v.Name == "seed" || v.Name == "s" {
				provided = true
			}
		})
		if f.NArg() > 1 {
			return bad("render accepts one seed")
		}
		if f.NArg() == 1 {
			if provided {
				return bad("provide the seed once")
			}
			seed = f.Arg(0)
			provided = true
		}
		if !provided {
			return bad("render requires --seed TEXT or a positional seed; use generate for random avatars")
		}
		if len(seed) > 4096 {
			return bad("seed must be at most 4096 bytes")
		}
		if a.remote {
			q := optionsQuery(o)
			q.Set("seed", seed)
			q.Set("style", style)
			q.Set("format", format)
			if color != "" {
				q.Set("color", color)
			}
			b, e = a.request(ctx, "GET", "/api/v1/render?"+q.Encode(), nil, nil)
		} else {
			b, e = a.engine.RenderWithInputs(ctx, style, seed, format, avatar.Inputs{Color: color}, o)
			if e != nil {
				e = bad(e.Error())
			}
		}
	}
	if e != nil {
		return e
	}
	return a.writeImage(b, format, file, a.json)
}
func (a *app) writeImage(b []byte, format, file string, asJSON bool) error {
	if file != "" && file != "-" {
		f, e := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if e != nil {
			return e
		}
		_, e = f.Write(b)
		if e == nil {
			e = f.Sync()
		}
		closeErr := f.Close()
		if e == nil {
			e = closeErr
		}
		if e != nil {
			_ = os.Remove(file)
			return e
		}
		if asJSON {
			p, e := filepath.Abs(file)
			if e != nil {
				return e
			}
			return a.emit(map[string]any{"path": p, "format": format, "bytes": len(b)})
		}
		return nil
	}
	if asJSON {
		media := "image/svg+xml"
		if format == "png" {
			media = "image/png"
		}
		return a.emit(map[string]string{"format": format, "media_type": media, "data_base64": base64.StdEncoding.EncodeToString(b)})
	}
	_, e := a.out.Write(b)
	return e
}
