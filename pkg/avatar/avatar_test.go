package avatar

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"image/png"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
)

func TestTypeScriptParity(t *testing.T) {
	raw, err := os.ReadFile("testdata/reference.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []struct {
		Seed                    string
		SHA256                  string
		Outfit, Hair, Accessory int
	}
	if err := json.Unmarshal(raw, &fixtures); err != nil {
		t.Fatal(err)
	}
	outfits, hairs, accessories := map[int]bool{}, map[int]bool{}, map[int]bool{}
	for _, fixture := range fixtures {
		t.Run(fmt.Sprintf("%q", fixture.Seed), func(t *testing.T) {
			svg := goreySVG(fixture.Seed)
			got := fmt.Sprintf("%x", sha256.Sum256([]byte(svg)))
			if got != fixture.SHA256 {
				t.Fatalf("reference mismatch: got %s, want %s", got, fixture.SHA256)
			}
		})
		outfits[fixture.Outfit], hairs[fixture.Hair], accessories[fixture.Accessory] = true, true, true
	}
	if len(outfits) != 6 || len(hairs) != 5 || len(accessories) != 6 {
		t.Fatalf("incomplete fixture coverage: %d outfits, %d hair, %d accessories", len(outfits), len(hairs), len(accessories))
	}
}

func TestDeterminismVarietyAndInjection(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		seed := fmt.Sprintf("seed-%d", i)
		svg := goreySVG(seed)
		if svg != goreySVG(seed) {
			t.Fatal("nondeterministic output")
		}
		if seen[svg] {
			t.Fatal("duplicate portrait")
		}
		seen[svg] = true
	}
	seed := `<script>alert(1)</script><svg onload="alert(2)"><foreignObject>attack</foreignObject>`
	svg := goreySVG(seed)
	for _, token := range []string{"script", "onload", "foreignObject", "attack", "alert"} {
		if strings.Contains(svg, token) {
			t.Fatalf("seed leaked: %s", token)
		}
	}
	decoder := xml.NewDecoder(strings.NewReader(svg))
	for {
		_, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("invalid XML: %v", err)
		}
	}
}

func TestPNGDimensionsAndMask(t *testing.T) {
	engine := New()
	for _, o := range []Options{{}, {Width: 256}, {Height: 144}, {Width: 128, Height: 128}, {Circle: true}, {Width: 128, Height: 128, Circle: true}, {Width: 200, Height: 128, Circle: true}} {
		t.Run(fmt.Sprintf("%+v", o), func(t *testing.T) {
			data, err := engine.Render(context.Background(), "gorey", "marcus", "png", o)
			if err != nil {
				t.Fatal(err)
			}
			img, err := png.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatal(err)
			}
			normalized, _ := o.Normalize(64, 72)
			if img.Bounds().Dx() != normalized.Width || img.Bounds().Dy() != normalized.Height {
				t.Fatalf("dimensions %v", img.Bounds())
			}
			_, _, _, alpha := img.At(normalized.Width/2, normalized.Height/2).RGBA()
			if alpha != 65535 {
				t.Fatal("center should be opaque")
			}
			_, _, _, alpha = img.At(0, 0).RGBA()
			if o.Circle && alpha != 0 {
				t.Fatalf("circle corner is opaque: %d", alpha)
			}
			colors := map[string]bool{}
			for y := 0; y < normalized.Height; y++ {
				for x := 0; x < normalized.Width; x++ {
					r, g, b, a := img.At(x, y).RGBA()
					colors[fmt.Sprintf("%d-%d-%d-%d", r, g, b, a)] = true
				}
			}
			if len(colors) < 50 {
				t.Fatalf("missing artwork or antialiasing: only %d colors", len(colors))
			}
		})
	}
}

func TestSVGNativeAndSizing(t *testing.T) {
	engine := New()
	native, err := engine.Render(context.Background(), "", "marcus", "", Options{})
	if err != nil {
		t.Fatal(err)
	}
	if string(native) != goreySVG("marcus") {
		t.Fatal("native SVG changed")
	}
	scaled, err := engine.Render(context.Background(), "", "marcus", "svg", Options{Width: 128, Height: 128, Circle: true})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(scaled, []byte(`width="128" height="128"`)) || !bytes.Contains(scaled, []byte(`<circle cx="64" cy="64" r="64"/>`)) {
		t.Fatal("missing size or circular mask")
	}
	dec := xml.NewDecoder(bytes.NewReader(scaled))
	for {
		_, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestRenderErrorsAndConcurrentUse(t *testing.T) {
	e := New()
	ctx := context.Background()
	for _, tt := range []struct {
		style, format string
		options       Options
		want          error
	}{
		{"missing", "svg", Options{}, ErrUnknownStyle}, {"gorey", "missing", Options{}, ErrUnknownFormat},
		{"gorey", "svg", Options{Width: -1}, ErrInvalidOptions}, {"gorey", "png", Options{Height: 2049}, ErrInvalidOptions},
	} {
		_, err := e.Render(ctx, tt.style, "seed", tt.format, tt.options)
		if !errors.Is(err, tt.want) {
			t.Fatalf("got %v, want %v", err, tt.want)
		}
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := e.Render(canceled, "", "", "", Options{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Go(func() {
			if _, err := e.Render(ctx, "", "seed", "png", Options{}); err != nil {
				t.Error(err)
			}
			_ = e.Styles()
			_ = e.Formats()
		})
	}
	wg.Wait()
}

type sampleGenerator struct{}

func (sampleGenerator) Style() Style { return Style{ID: "sample", Name: "Sample"} }
func (sampleGenerator) Generate(context.Context, string) (Artwork, error) {
	return Artwork{Data: []byte("native pixels"), MediaType: "image/example", Width: 10, Height: 20}, nil
}

type sampleExporter struct{}

func (sampleExporter) Format() string { return "example" }
func (sampleExporter) Export(_ context.Context, a Artwork, o Options) ([]byte, error) {
	return []byte(fmt.Sprintf("%s:%s:%dx%d", a.MediaType, a.Data, o.Width, o.Height)), nil
}
func TestExtensionAdapters(t *testing.T) {
	e := New()
	if err := e.RegisterGenerator(sampleGenerator{}); err != nil {
		t.Fatal(err)
	}
	if err := e.RegisterExporter(sampleExporter{}); err != nil {
		t.Fatal(err)
	}
	out, err := e.Render(context.Background(), "sample", "", "example", Options{Width: 20})
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "image/example:native pixels:20x40" {
		t.Fatalf("adapter output: %q", out)
	}
	if err := e.RegisterGenerator(sampleGenerator{}); err == nil {
		t.Fatal("duplicate generator accepted")
	}
	if err := e.RegisterExporter(sampleExporter{}); err == nil {
		t.Fatal("duplicate exporter accepted")
	}
	if _, err := e.Render(context.Background(), "sample", "", "png", Options{}); !errors.Is(err, ErrUnsupportedArtwork) {
		t.Fatal(err)
	}
}

func TestSVGAdapterReplacesNativeDimensions(t *testing.T) {
	a := Artwork{Data: []byte(`<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg" width="64" height="72" viewBox="0 0 64 72" fill="#292820"><circle cx="32" cy="36" r="20"/></svg>`), MediaType: "image/svg+xml", Width: 64, Height: 72}
	out, err := (SVGExporter{}).Export(context.Background(), a, Options{Width: 128})
	if err != nil {
		t.Fatal(err)
	}
	decoder := xml.NewDecoder(bytes.NewReader(out))
	roots := 0
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "svg" {
			continue
		}
		roots++
		attrs := map[string]bool{}
		for _, attr := range start.Attr {
			if attrs[attr.Name.Local] {
				t.Fatalf("duplicate attribute %s", attr.Name.Local)
			}
			attrs[attr.Name.Local] = true
		}
	}
	if roots != 2 {
		t.Fatalf("got %d SVG roots", roots)
	}
}

func TestSVGAdapterFramesSelfClosingRoot(t *testing.T) {
	for _, source := range []string{
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 72"/>`,
		`<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg" width="64" height="72" viewBox="0 0 64 72" />` + "\n<!-- trailing comment -->",
	} {
		for _, options := range []Options{{Width: 128}, {Width: 128, Circle: true}} {
			artwork := Artwork{Data: []byte(source), MediaType: "image/svg+xml", Width: 64, Height: 72}
			output, err := (SVGExporter{}).Export(context.Background(), artwork, options)
			if err != nil {
				t.Fatal(err)
			}
			decoder := xml.NewDecoder(bytes.NewReader(output))
			var document struct {
				XMLName xml.Name `xml:"svg"`
			}
			if err := decoder.Decode(&document); err != nil {
				t.Fatalf("invalid framed SVG for %q: %v\n%s", source, err, output)
			}
			for {
				token, err := decoder.Token()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatal(err)
				}
				if _, ok := token.(xml.StartElement); ok {
					t.Fatal("unexpected additional document root")
				}
			}
		}
	}
}
