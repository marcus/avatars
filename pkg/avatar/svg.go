package avatar

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
)

// SVGExporter preserves native SVG verbatim and frames explicitly sized output.
type SVGExporter struct{}

func (SVGExporter) Format() string { return "svg" }
func (SVGExporter) Export(ctx context.Context, a Artwork, o Options) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if a.MediaType != "image/svg+xml" {
		return nil, fmt.Errorf("%w: SVG exporter requires image/svg+xml", ErrUnsupportedArtwork)
	}
	var err error
	o, err = o.Normalize(a.Width, a.Height)
	if err != nil {
		return nil, err
	}
	if o.Width == a.Width && o.Height == a.Height && !o.Circle {
		return bytes.Clone(a.Data), nil
	}
	// Read the root separately so existing width/height attributes can be
	// replaced without altering the artwork's presentation attributes.
	decoder := xml.NewDecoder(bytes.NewReader(a.Data))
	var root xml.StartElement
	for {
		token, decodeErr := decoder.Token()
		if decodeErr != nil {
			return nil, fmt.Errorf("%w: expected a standalone SVG element: %v", ErrUnsupportedArtwork, decodeErr)
		}
		if start, ok := token.(xml.StartElement); ok {
			root = start
			break
		}
	}
	if root.Name.Local != "svg" {
		return nil, fmt.Errorf("%w: expected an SVG root", ErrUnsupportedArtwork)
	}
	rootEnd := decoder.InputOffset()
	// Validate before nesting; malformed native artwork must not escape the wrapper.
	for {
		_, decodeErr := decoder.Token()
		if decodeErr == io.EOF {
			break
		}
		if decodeErr != nil {
			return nil, fmt.Errorf("%w: invalid SVG: %v", ErrUnsupportedArtwork, decodeErr)
		}
	}
	scale, x, y := placement(a.Width, a.Height, o)
	clip := ""
	if o.Circle {
		clip = fmt.Sprintf(`<defs><clipPath id="avatar-circle"><circle cx="%g" cy="%g" r="%g"/></clipPath></defs><g clip-path="url(#avatar-circle)">`, float64(o.Width)/2, float64(o.Height)/2, float64(min(o.Width, o.Height))/2)
	}
	var out bytes.Buffer
	fmt.Fprintf(&out, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" fill="none" aria-hidden="true">%s<g transform="translate(%g %g) scale(%g)">`, o.Width, o.Height, o.Width, o.Height, clip, x, y, scale)
	// Keep native root presentation attributes and native viewport clipping.
	fmt.Fprintf(&out, `<svg width="%d" height="%d" `, a.Width, a.Height)
	for _, attribute := range root.Attr {
		if attribute.Name.Local == "width" || attribute.Name.Local == "height" {
			continue
		}
		name := attribute.Name.Local
		if attribute.Name.Space == "http://www.w3.org/1999/xlink" {
			name = "xlink:" + name
		}
		if attribute.Name.Space == "xmlns" {
			name = "xmlns:" + name
		}
		fmt.Fprintf(&out, `%s="`, name)
		_ = xml.EscapeText(&out, []byte(attribute.Value))
		out.WriteString(`" `)
	}
	out.WriteByte('>')
	// XML tokenization expands a self-closing root into start/end tokens,
	// but its original byte suffix contains no closing tag to copy.
	if bytes.HasSuffix(a.Data[:rootEnd], []byte("/>")) {
		out.WriteString(`</svg>`)
	}
	out.Write(a.Data[rootEnd:])
	out.WriteString(`</g>`)
	if o.Circle {
		out.WriteString(`</g>`)
	}
	out.WriteString(`</svg>`)
	return out.Bytes(), nil
}

func placement(width, height int, o Options) (scale, x, y float64) {
	sx, sy := float64(o.Width)/float64(width), float64(o.Height)/float64(height)
	scale = min(sx, sy)
	if o.Circle {
		scale = max(sx, sy)
	}
	return scale, (float64(o.Width) - float64(width)*scale) / 2, (float64(o.Height) - float64(height)*scale) / 2
}
