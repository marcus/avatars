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

func (Pebble) GenerateWithInputs(ctx context.Context, seed string, inputs Inputs) (Artwork, error) {
	if err := ctx.Err(); err != nil {
		return Artwork{}, err
	}
	color, ok := pebbleColor(inputs.Color)
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

	cx := 32 + next(3) - 1
	top := 11 + next(4)
	bottom := 52 + next(3)
	left := 10 + next(4)
	right := 51 + next(4)
	topX := cx + next(5) - 2
	bottomX := cx + next(7) - 3
	shoulderY := 18 + next(4)
	hipY := 46 + next(4)

	path := fmt.Sprintf(
		"M%d %d"+
			" C%d %d %d %d %d %d"+
			" C%d %d %d %d %d %d"+
			" C%d %d %d %d %d %d"+
			" C%d %d %d %d %d %d"+
			" C%d %d %d %d %d %d"+
			" C%d %d %d %d %d %d"+
			" C%d %d %d %d %d %d Z",
		topX, top,
		topX+7, top-2+next(3), right-2, shoulderY-3, right, shoulderY+7,
		right+1, 33, right-1, hipY-1, right-6, hipY+3,
		right-10, bottom, bottomX+7, bottom+2, bottomX, bottom,
		bottomX-8, bottom+1, left+7, bottom-1, left+4, hipY,
		left, hipY-4, left-1, 34, left+1, 29,
		left+2, 23, left+5, shoulderY, left+10, shoulderY-3,
		left+14, top+1, topX-5, top-1, topX, top,
	)

	eyeY := 31 + next(4) - 1
	spacing := 12 + next(5)
	gaze := next(3) - 1
	leftEyeX := cx - spacing/2 + gaze
	rightEyeX := cx + (spacing+1)/2 + gaze
	leftRX := 1.35 + float64(next(4))*.12
	rightRX := 1.35 + float64(next(4))*.12
	leftRY := .72 + float64(next(3))*.12
	rightRY := .72 + float64(next(3))*.12
	leftTilt := next(9) - 4
	rightTilt := next(9) - 4

	var out strings.Builder
	fmt.Fprintf(&out, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64" fill="none" aria-hidden="true"><path d="%s" fill="%s"/>`, path, color.Swatch)
	fmt.Fprintf(&out, `<ellipse cx="%d" cy="%d" rx="%.2f" ry="%.2f" fill="%s" transform="rotate(%d %d %d)"/>`, leftEyeX, eyeY, leftRX, leftRY, color.EyeColor, leftTilt, leftEyeX, eyeY)
	fmt.Fprintf(&out, `<ellipse cx="%d" cy="%d" rx="%.2f" ry="%.2f" fill="%s" transform="rotate(%d %d %d)"/>`, rightEyeX, eyeY+next(3)-1, rightRX, rightRY, color.EyeColor, rightTilt, rightEyeX, eyeY)
	out.WriteString(`</svg>`)
	return out.String()
}
