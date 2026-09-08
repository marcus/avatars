package avatar

import (
	"context"
	"fmt"
	"strings"
)

const pebbleDefaultColor = "walnut"

var pebbleColors = []ColorChoice{
	{Value: "walnut", Label: "Walnut", Swatch: "#92744F", EyeColor: "#FFF4DC"},
	{Value: "cocoa", Label: "Cocoa", Swatch: "#655047", EyeColor: "#FFF4DC"},
	{Value: "clay", Label: "Clay", Swatch: "#B66F56", EyeColor: "#FFF4DC"},
	{Value: "apricot", Label: "Apricot", Swatch: "#E6AA78", EyeColor: "#4A3427"},
	{Value: "butter", Label: "Butter", Swatch: "#E4CA78", EyeColor: "#4A3427"},
	{Value: "moss", Label: "Moss", Swatch: "#71805A", EyeColor: "#FFF4DC"},
	{Value: "sage", Label: "Sage", Swatch: "#A5B59A", EyeColor: "#4A3427"},
	{Value: "teal", Label: "Teal", Swatch: "#4D8985", EyeColor: "#FFF4DC"},
	{Value: "sky", Label: "Sky", Swatch: "#9DBFD1", EyeColor: "#4A3427"},
	{Value: "denim", Label: "Denim", Swatch: "#607C9B", EyeColor: "#FFF4DC"},
	{Value: "lavender", Label: "Lavender", Swatch: "#AAA0C6", EyeColor: "#4A3427"},
	{Value: "mauve", Label: "Mauve", Swatch: "#A47B97", EyeColor: "#FFF4DC"},
	{Value: "rose", Label: "Rose", Swatch: "#D6A0A4", EyeColor: "#4A3427"},
	{Value: "coral", Label: "Coral", Swatch: "#D77F6D", EyeColor: "#4A3427"},
	{Value: "slate", Label: "Slate", Swatch: "#6C777D", EyeColor: "#FFF4DC"},
}

// Pebble creates small, softly irregular characters with two simple eyes.
type Pebble struct{}

func (Pebble) Style() Style {
	return Style{
		ID:          "pebble",
		NativeWidth: 64, NativeHeight: 64,
		Name:        "Pebble",
		Description: "Pudgy, softly irregular characters with two tiny eyes.",
		Inputs: &StyleInputs{Color: &ColorInput{
			Default: pebbleDefaultColor,
			Values:  append([]ColorChoice(nil), pebbleColors...),
		}},
	}
}

func (p Pebble) Generate(ctx context.Context, seed string) (Artwork, error) {
	return p.GenerateWithInputs(ctx, seed, Inputs{Color: pebbleDefaultColor})
}

func (p Pebble) GenerateWithInputs(ctx context.Context, seed string, inputs Inputs) (Artwork, error) {
	if err := ctx.Err(); err != nil {
		return Artwork{}, err
	}
	resolved, err := resolveStyleInputs(p.Style(), inputs)
	if err != nil {
		return Artwork{}, err
	}
	color, ok := pebbleColor(resolved.Color)
	if !ok {
		return Artwork{}, fmt.Errorf("%w: invalid color %q for style %q", ErrInvalidInputs, inputs.Color, "pebble")
	}
	return Artwork{Data: []byte(pebbleSVG(seed, color)), MediaType: "image/svg+xml", Width: 64, Height: 64}, nil
}

func pebbleColor(value string) (ColorChoice, bool) {
	if value == "" {
		value = pebbleDefaultColor
	}
	for _, color := range pebbleColors {
		if color.Value == value {
			return color, true
		}
	}
	return ColorChoice{}, false
}

func pebbleSVG(seed string, color ColorChoice) string {
	hash := uint32(2166136261)
	for _, char := range seed {
		hash = (hash ^ uint32(char)) * 16777619
	}
	hash += 0x9e3779b9 // Keep xorshift moving for a zero FNV result.
	next := func(max int) int {
		hash ^= hash << 13
		hash ^= hash >> 17
		hash ^= hash << 5
		return int(hash % uint32(max))
	}

	// Keep the mass pudgy, but let a seed choose a recognisable proportion.
	// Width and height move together instead of independently distorting a blob.
	proportions := [][2]float64{{24, 17}, {18, 23}, {22, 20}, {20, 22}, {23, 18}}
	proportion := proportions[next(len(proportions))]
	rx := proportion[0] + float64(next(3)-1)*.5
	ry := proportion[1] + float64(next(3)-1)*.5
	cx, cy := 32.0, 33.0
	lean := float64(next(9) - 4)
	topX, bottomX := cx+lean, cx-lean*.35
	top, bottom := cy-ry, cy+ry
	left, right := cx-rx, cx+rx
	leftY := cy + float64(next(5))
	rightY := cy + float64(next(5))
	// Unequal cheeks and a broad, softly resting base carry the body gesture.
	crown := .52 + float64(next(4))*.04
	path := fmt.Sprintf(
		"M%.2f %.2f"+
			" C%.2f %.2f %.2f %.2f %.2f %.2f"+
			" C%.2f %.2f %.2f %.2f %.2f %.2f"+
			" C%.2f %.2f %.2f %.2f %.2f %.2f"+
			" C%.2f %.2f %.2f %.2f %.2f %.2f Z",
		topX, top,
		topX+rx*crown, top, right, rightY-ry*.68, right, rightY,
		right, rightY+ry*.68, bottomX+rx*.72, bottom, bottomX, bottom,
		bottomX-rx*.72, bottom, left, leftY+ry*.62, left, leftY,
		left, leftY-ry*.65, topX-rx*crown, top, topX, top,
	)

	// Choose the whole expression first. Small paired variations keep the eyes
	// related; independent eye angles easily turn a gentle face into a scowl.
	mood := next(7)
	spacing := 12.0 + float64(next(4))
	faceX := cx + lean*.3 + float64(next(5)-2)
	faceY := cy + float64(next(5)-2)
	leftRX, rightRX, leftRY, rightRY := 1.55, 1.55, 1.95, 1.95
	leftArc, rightArc := false, false
	switch mood {
	case 0: // Quiet, close to the original tiny horizontal eyes.
		leftRX, rightRX, leftRY, rightRY = 1.8, 1.8, .95, .95
	case 1: // Curious: one eye opens a little more and the face looks up.
		leftRY, rightRY = 2.5, 1.65
		faceY -= 2
	case 2: // Sleepy, with both soft eye marks still present.
		leftRX, rightRX, leftRY, rightRY = 2.1, 2.1, .85, .85
		faceY += 1
	case 3: // Shy: a lower, closer-set face.
		spacing = 9.5 + float64(next(3))
		leftRX, rightRX, leftRY, rightRY = 1.35, 1.35, 1.7, 1.7
		faceY += 3
	case 4: // Contented soft squints, drawn as two filled arches.
		leftArc, rightArc = true, true
	case 5: // A friendly wink, never opposing angry eye angles.
		if next(2) == 0 {
			leftArc = true
		} else {
			rightArc = true
		}
	case 6: // Wide-eyed, with a little more space to breathe.
		spacing += 1
		leftRY, rightRY = 2.35, 2.35
	}
	faceTilt := float64(next(5)-2) * .45
	leftEyeX, rightEyeX := faceX-spacing/2, faceX+spacing/2
	var out strings.Builder
	fmt.Fprintf(&out, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64" fill="none" aria-hidden="true"><path d="%s" fill="%s"/>`, path, color.Swatch)
	pebbleEye(&out, leftEyeX, faceY-faceTilt, leftRX, leftRY, leftArc, color.EyeColor)
	pebbleEye(&out, rightEyeX, faceY+faceTilt, rightRX, rightRY, rightArc, color.EyeColor)
	out.WriteString(`</svg>`)
	return out.String()
}

func pebbleEye(out *strings.Builder, x, y, rx, ry float64, arc bool, fill string) {
	if arc {
		// A filled arch avoids stroke/rasterizer differences at tiny export sizes.
		fmt.Fprintf(out, `<path d="M%.2f %.2f C%.2f %.2f %.2f %.2f %.2f %.2f C%.2f %.2f %.2f %.2f %.2f %.2f Z" fill="%s"/>`,
			x-2.1, y+.65, x-2.3, y-2.5, x+2.3, y-2.5, x+2.1, y+.65,
			x+.9, y-.8, x-.9, y-.8, x-2.1, y+.65, fill)
		return
	}
	fmt.Fprintf(out, `<ellipse cx="%.2f" cy="%.2f" rx="%.2f" ry="%.2f" fill="%s"/>`, x, y, rx, ry, fill)
}
