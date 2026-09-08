package avatar

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"math"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

// PNGExporter rasterizes self-contained SVG using pure Go, preserving SVG
// round stroke caps and joins. It does not fetch resources or execute scripts.
type PNGExporter struct{}

func (PNGExporter) Format() string { return "png" }
func (PNGExporter) Export(ctx context.Context, a Artwork, o Options) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if a.MediaType != "image/svg+xml" {
		return nil, fmt.Errorf("%w: PNG exporter requires image/svg+xml", ErrUnsupportedArtwork)
	}
	var err error
	o, err = o.Normalize(a.Width, a.Height)
	if err != nil {
		return nil, err
	}
	icon, err := oksvg.ReadIconStream(bytes.NewReader(a.Data), oksvg.StrictErrorMode)
	if err != nil {
		return nil, fmt.Errorf("parse SVG: %w", err)
	}
	canvas := image.NewRGBA(image.Rect(0, 0, o.Width, o.Height))
	scale, x, y := placement(a.Width, a.Height, o)
	icon.SetTarget(x, y, float64(a.Width)*scale, float64(a.Height)*scale)
	scanner := rasterx.NewScannerGV(o.Width, o.Height, canvas, canvas.Bounds())
	icon.Draw(rasterx.NewDasher(o.Width, o.Height, scanner), 1)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if o.Circle {
		maskCircle(canvas)
	}
	var out bytes.Buffer
	if err := png.Encode(&out, canvas); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func maskCircle(img *image.RGBA) {
	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	cx, cy, radius := float64(width)/2, float64(height)/2, float64(min(width, height))/2
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// One-pixel analytic coverage keeps the mask smooth at icon sizes.
			coverage := max(0, min(1, radius+.5-math.Hypot(float64(x)+.5-cx, float64(y)+.5-cy)))
			if coverage == 1 {
				continue
			}
			i := img.PixOffset(x, y)
			for c := 0; c < 4; c++ {
				img.Pix[i+c] = uint8(math.Round(float64(img.Pix[i+c]) * coverage))
			}
		}
	}
}
