package avatar

import (
	"context"
	"fmt"
	"strings"
)

// Picasso makes recognizable portraits with restrained cubist planes. Each
// seed selects a complete numeric recipe; no seed text enters the artwork.
type Picasso struct{}

func (Picasso) Style() Style {
	return Style{NativeWidth: 64, NativeHeight: 72, ID: "picasso", Name: "Picasso", Description: "Expressive portraits in bold ink, muted color, and asymmetric planes."}
}

func (Picasso) Generate(ctx context.Context, seed string) (Artwork, error) {
	if err := ctx.Err(); err != nil {
		return Artwork{}, err
	}
	return Artwork{Data: []byte(picassoSVG(seed)), MediaType: "image/svg+xml", Width: 64, Height: 72}, nil
}

type picassoPalette struct {
	paper, ground, blue, red, ochre, face, shadow, light, hair string
}

func picassoSVG(seed string) string {
	hash := uint32(2166136261)
	for _, char := range seed {
		hash = (hash ^ uint32(char)) * 16777619
	}
	// The nonzero offset avoids xorshift's absorbing state for a zero hash.
	hash ^= 0x9e3779b9
	if hash == 0 {
		hash = 0x85ebca6b
	}
	next := func(n int) int {
		hash ^= hash << 13
		hash ^= hash >> 17
		hash ^= hash << 5
		return int(hash % uint32(n))
	}
	palettes := [...]picassoPalette{
		{"#e7dfc9", "#d6b677", "#527e89", "#ac5144", "#c49548", "#ddbf91", "#b88663", "#efdbc0", "#353b39"},
		{"#e2dfca", "#a9b5a7", "#456879", "#a94e40", "#cb9d59", "#cda37c", "#a47559", "#e8cdaa", "#34372f"},
		{"#e6d9c6", "#bb7662", "#628d98", "#98453c", "#c5a25a", "#e0c5a3", "#c09470", "#efddba", "#3b3835"},
		{"#ded7c1", "#789a9c", "#456772", "#b0614d", "#ba8d48", "#b98b65", "#956b51", "#d8b98f", "#313a39"},
		{"#e6ddca", "#cab17b", "#537781", "#b96249", "#d3a75d", "#d5b291", "#af805e", "#edd7b6", "#554239"},
		{"#e3dfcf", "#a7b5af", "#43697f", "#ab5d4a", "#b3975a", "#ac7c59", "#805a46", "#cba582", "#303631"},
	}
	p := palettes[next(len(palettes))]
	ink := "#30332e"
	cx := 31 + next(3)
	left, right := cx-12-next(3), cx+12+next(3)
	top, chin := 11+next(3), 48+next(4)
	hair, attire := next(8), next(5)
	eyeY, tilt := 28+next(3), next(3)-1
	feature, backdrop := next(7), next(4)
	noseTip, noseY := cx+3+next(4), 35+next(3)
	facePlane := p.shadow
	if next(4) == 0 {
		facePlane = p.blue
	}
	var out strings.Builder
	out.Grow(8000)
	fmt.Fprintf(&out, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 72" fill="none" aria-hidden="true"><path fill="%s" d="M0 0h64v72H0z"/>`, p.paper)
	path := func(d, fill string, width float64) {
		stroke := ink
		if width == 0 {
			stroke = "none"
		}
		fmt.Fprintf(&out, `<path d="%s" fill="%s" stroke="%s" stroke-width="%g" stroke-linecap="round" stroke-linejoin="round"/>`, d, fill, stroke, width)
	}
	shape := func(d, fill string) { path(d, fill, 0) }
	line := func(d string, width float64) { path(d, "none", width) }
	// Large quiet fields carry color at icon sizes, without fragmenting the face.
	switch backdrop {
	case 0:
		shape("M0 0h21l-8 72H0Z", p.ground)
		shape("M45 0h19v48L53 40Z", p.blue)
	case 1:
		shape("M0 0h64v23L0 39Z", p.ground)
		shape("M0 39l13-3 2 36H0Z", p.red)
	case 2:
		shape("M0 0h64v72H48L53 17 0 29Z", p.ground)
		shape("M0 0h12l-2 47H0Z", p.blue)
	case 3:
		shape("M0 0h64v72H0Z", p.ground)
		shape("M9 8h43l-4 55H16Z", p.paper)
		shape("M51 8h13v64H56Z", p.red)
	}
	// A handful of fine dry marks retain a drawn, printed surface.
	for i := 0; i < 5; i++ {
		x, y := 3+next(6), 5+next(42)
		fmt.Fprintf(&out, `<path d="M%d %dl%d -1" stroke="%s" stroke-width=".35" opacity=".42"/>`, x, y, 2+next(4), p.paper)
	}
	// Clothing is deliberately adult: broad lapels, knit necklines, and scarves.
	coat := []string{p.blue, p.red, p.ochre, p.hair, p.blue}[attire]
	path(fmt.Sprintf("M2 72L7 59L%d 51L%d 50L56 58L63 72Z", cx-10, cx+10), coat, 1.2)
	shape(fmt.Sprintf("M%d 53L%d 72H63L55 58Z", cx+5, cx+12), p.ground)
	path(fmt.Sprintf("M%d 42L%d 54L%d 61L%d 53L%d 42Z", cx-8, cx-9, cx, cx+9, cx+7), p.shadow, 1)
	shape(fmt.Sprintf("M%d 45L%d 54L%d 58L%d 46Z", cx-5, cx-6, cx+1, cx+2), p.face)
	switch attire {
	case 0:
		path(fmt.Sprintf("M%d 51L16 55L25 68L%d 60L39 68L48 54L%d 51L%d 57Z", cx-9, cx, cx+9, cx), p.paper, .9)
		path(fmt.Sprintf("M%d 57l3 3-2 12h-5l2-12Z", cx), p.red, .75)
		line("M11 61l7 11M51 62l-6 10", .7)
	case 1:
		path(fmt.Sprintf("M%d 52Q%d 64 %d 52L%d 58Q%d 68 %d 58Z", cx-9, cx, cx+9, cx+10, cx, cx-11), p.blue, .9)
		line("M14 63l-2 8M21 67l-1 5M45 65l2 7M51 63l3 9", .6)
	case 2:
		path(fmt.Sprintf("M%d 51L%d 55L%d 64L%d 57L%d 51", cx-9, cx-13, cx-3, cx, cx+8), p.paper, 1)
		path(fmt.Sprintf("M%d 56L%d 63L%d 72H%dL%d 61Z", cx+4, cx+8, cx+5, cx-3, cx+1), p.red, .8)
	case 3:
		path(fmt.Sprintf("M%d 51L%d 58L%d 65L%d 57L%d 51L%d 62Z", cx-8, cx-13, cx-4, cx+11, cx+8, cx), p.ochre, .9)
		line(fmt.Sprintf("M%d 64v8M12 61l3 11M52 60l-3 12", cx), .8)
	case 4:
		path(fmt.Sprintf("M%d 52Q%d 66 %d 52L%d 55Q%d 72 %d 55Z", cx-10, cx, cx+10, cx+12, cx, cx-12), p.red, .9)
		line("M10 63l44 5M9 68l25 4", .7)
	}
	// One cheek turns into profile; the silhouette and both eyes remain legible.
	path(fmt.Sprintf("M%d 27Q%d 23 %d 33L%d 37M%d 27Q%d 24 %d 34L%d 38", left, left-6, left-4, left+1, right, right+6, right+3, right-1), p.shadow, 1)
	face := fmt.Sprintf("M%d 24L%d %dQ%d %d %d %dL%d 25L%d 40L%d %dL%d %dL%d %dL%d 37Z", left, left+3, top+3, cx-1, top-5, right-3, top+1, right+1, right-2, cx+7, chin-3, cx, chin+1, left+5, chin-4, left-1)
	path(face, p.face, 1.25)
	shape(fmt.Sprintf("M%d %dL%d 27L%d %dL%d 40L%d %dL%d %dL%d 40L%d 25L%d %dZ", cx+1, top-1, cx-1, noseTip, noseY, cx, cx+2, chin-1, cx+7, chin-3, right-2, right+1, right-3, top+1), facePlane)
	shape(fmt.Sprintf("M%d 34L%d 37L%d 45L%d 41Z", left+2, cx-5, cx-3, left+4), p.light)
	line(fmt.Sprintf("M%d %dL%d 25", cx+1, top+4, cx-1), .65)
	// Deliberate mismatched angles suggest two viewpoints rather than a mask.
	lx, rx := cx-7, cx+7
	path(fmt.Sprintf("M%d %dQ%d %d %d %dQ%d %d %d %dZ", lx-4, eyeY, lx, eyeY-4, lx+4, eyeY, lx, eyeY+3, lx-4, eyeY), p.light, .85)
	path(fmt.Sprintf("M%d %dL%d %dQ%d %d %d %dZ", rx-4, eyeY+tilt, rx+3, eyeY-2+tilt, rx+5, eyeY+2+tilt, rx-4, eyeY+tilt), p.light, .85)
	fmt.Fprintf(&out, `<ellipse cx="%d" cy="%g" rx="1.15" ry="1.7" fill="%s"/><ellipse cx="%d" cy="%g" rx="1.1" ry="1.45" fill="%s"/>`, lx+next(2), float64(eyeY)-.3, ink, rx, float64(eyeY+tilt)-.3, ink)
	line(fmt.Sprintf("M%d %dL%d %dM%d %dQ%d %d %d %d", lx-4, eyeY-5, lx+3, eyeY-4, rx-3, eyeY-4+tilt, rx+1, eyeY-7+tilt, rx+5, eyeY-4+tilt), 1.3)
	path(fmt.Sprintf("M%d 27L%d %dL%d %dL%d %dL%d %d", cx, cx-2, noseY-3, noseTip, noseY, cx-1, noseY+2, cx-3, noseY), p.ochre, 1.05)
	line(fmt.Sprintf("M%d %dl2 -1", cx+1, noseY+1), .75)
	mouth := chin - 7
	shape(fmt.Sprintf("M%d %dL%d %dL%d %dL%d %dZ", cx-5, mouth, cx+1, mouth-1, cx+5, mouth+1, cx-1, mouth+2), p.red)
	line(fmt.Sprintf("M%d %dQ%d %d %d %d", cx-5, mouth, cx, mouth+1, cx+5, mouth), .85)
	line(fmt.Sprintf("M%d %dl4 1", cx-2, chin-2), .5)
	// Hair is a second large silhouette, with eight distinct adult cuts.
	switch hair {
	case 0:
		path(fmt.Sprintf("M%d 30L%d 18Q%d 5 %d 10L%d 19L%d 29L%d 20L%d 16L%d 23Z", left-1, left-2, cx-7, cx+6, right+1, right, right-4, cx-2, left+4), p.hair, 1.2)
		line(fmt.Sprintf("M%d 14Q%d 11 %d 16", left+3, cx-3, cx+6), .5)
		shape(fmt.Sprintf("M%d 14L%d 12L%d 18L%d 17Z", cx-3, cx+5, right-1, cx+2), p.ochre)
	case 1:
		path(fmt.Sprintf("M%d 26L%d 13L%d 8L%d 11L%d 17L%d 26L%d 20L%d 17L%d 20L%d 17L%d 25Z", left-1, left, cx-4, cx+7, right+1, right, right-4, cx+2, cx-4, cx-7, left+3), p.hair, 1.1)
		shape(fmt.Sprintf("M%d 13l10 2-5 2Z", cx-7), p.ground)
	case 2:
		path(fmt.Sprintf("M%d 28L%d 47L%d 49L%d 24L%d 18L%d 18L%d 24L%d 49L%d 47L%d 20Q%d 3 %d 10Q%d 9 %d 28Z", left-1, left-5, left+2, left+3, cx-5, cx+5, right-3, right-1, right+5, right+3, cx+7, cx-2, left-5, left-1), p.hair, 1.2)
		shape(fmt.Sprintf("M%d 20L%d 40L%d 42L%d 25Z", right+1, right+2, right-1, right-1), p.blue)
	case 3:
		path(fmt.Sprintf("M%d 31Q%d 19 %d 16L%d 21L%d 31ZM%d 29L%d 18Q%d 20 %d 33Z", left, left-3, left+3, left+2, left+2, right-1, right-1, right+5, right+1), p.hair, .9)
		line(fmt.Sprintf("M%d 14Q%d 10 %d 15", cx-4, cx+1, cx+5), .65)
	case 4:
		path(fmt.Sprintf("M%d 32Q%d 16 %d 13Q%d 5 %d 11Q%d 7 %d 18L%d 35L%d 29L%d 18L%d 21L%d 17L%d 26Z", left-2, left-8, left, cx-4, cx+1, right+2, right+3, right+3, right-1, right-4, cx-3, cx-7, left+3), p.hair, 1.2)
		shape(fmt.Sprintf("M%d 14q6 -3 9 1l-6 2Z", cx-5), p.ochre)
	case 5:
		path(fmt.Sprintf("M%d 15Q%d 1 %d 3Q%d 5 %d 15Z", cx+3, cx+2, cx+11, cx+20, cx+10), p.hair, 1.1)
		path(fmt.Sprintf("M%d 29L%d 19Q%d 5 %d 11L%d 21L%d 29L%d 19L%d 16L%d 23Z", left, left-1, cx-4, cx+7, right+1, right, right-4, cx-1, left+3), p.hair, 1.1)
		line(fmt.Sprintf("M%d 10l5 2", cx+8), .7)
	case 6:
		path(fmt.Sprintf("M%d 20Q%d 7 %d 8L%d 5Q%d 9 %d 19L%d 22L%d 18Z", left-5, left-7, cx-2, cx+6, right+8, right+5, right-1, cx-4), p.blue, 1.25)
		line(fmt.Sprintf("M%d 8l2 -4M%d 20Q%d 17 %d 20", cx+2, left-2, cx, right), 1)
		path(fmt.Sprintf("M%d 22l-1 9 3 -7M%d 23l-2 10 3 -5", left, right), p.hair, .9)
	case 7:
		path(fmt.Sprintf("M%d 33L%d 45L%d 48L%d 23L%d 16L%d 19L%d 28L%d 45L%d 43L%d 21Q%d 6 %d 9Q%d 8 %d 33Z", left-2, left-5, left+2, left+3, cx-5, cx+5, right-3, right-1, right+4, right+3, cx+7, cx-3, left-6, left-2), p.hair, 1.15)
		shape(fmt.Sprintf("M%d 14L%d 18L%d 22L%d 18Z", cx-4, cx+6, right-1, cx+4), p.red)
	}
	// Small accents vary age and expression without obscuring the face.
	if feature == 0 && (hair == 0 || hair == 1 || hair == 3 || hair == 6) {
		path(fmt.Sprintf("M%d %dQ%d %d %d %dL%d %dL%d %dZ", cx-5, mouth-2, cx-1, mouth-5, cx+5, mouth-1, cx+1, mouth-1, cx-1, mouth-2), p.hair, .55)
	} else if feature == 1 && hair == 3 {
		path(fmt.Sprintf("M%d %dl4 3 3 -3-3 7Z", cx-4, chin-4), p.hair, .6)
	} else if feature == 2 {
		line(fmt.Sprintf("M%d %dl-2 3M%d %dl2 3", lx-5, eyeY+3, rx+5, eyeY+3), .5)
	} else if feature == 3 {
		path(fmt.Sprintf("M%d %dh9v6h-9ZM%d %dh8v6h-8ZM%d %dh4", lx-5, eyeY-3, rx-4, eyeY-3+tilt, lx+4, eyeY), "none", .7)
	} else if feature == 4 && (hair == 2 || hair == 5 || hair == 7) {
		path(fmt.Sprintf("M%d 37l-2 5 3 2 2 -5Z", right+1), p.ochre, .7)
	}
	out.WriteString(`</svg>`)
	return out.String()
}
