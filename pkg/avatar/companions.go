package avatar

import (
	"context"
	"fmt"
	"hash/fnv"
	"strings"
)

// Companions draws original comic pets from a seed. Its random state and drawing
// helpers belong to a single call, so rendering is deterministic and concurrent.
type Companions struct{}

func (Companions) Style() Style {
	return Style{ID: "companions", Name: "Companions", Description: "Playful dogs and cats in lively ink and warm color.", NativeWidth: 128, NativeHeight: 128,
		Inputs: &StyleInputs{Animal: &AnimalInput{Default: "dog", MixedValue: "mixed", Values: []AnimalChoice{{Value: "dog", Label: "Dogs"}, {Value: "cat", Label: "Cats"}, {Value: "mixed", Label: "Mixed"}}}},
	}
}

func (c Companions) Generate(ctx context.Context, seed string) (Artwork, error) {
	return c.GenerateWithInputs(ctx, seed, Inputs{})
}

func (c Companions) GenerateWithInputs(ctx context.Context, seed string, inputs Inputs) (Artwork, error) {
	if err := ctx.Err(); err != nil {
		return Artwork{}, err
	}
	resolved, err := c.ResolveRecipeInputs(seed, inputs)
	if err != nil {
		return Artwork{}, err
	}
	return generateCompanion(ctx, seed, resolved.Animal)
}

// ResolveRecipeInputs selects one species for Mixed without changing artwork
// randomness. A saved recipe always holds dog or cat, never a batch instruction.
func (c Companions) ResolveRecipeInputs(seed string, inputs Inputs) (Inputs, error) {
	resolved, err := resolveStyleInputs(c.Style(), inputs)
	if err != nil {
		return Inputs{}, err
	}
	if resolved.Animal == "mixed" {
		hash := fnv.New32a()
		_, _ = hash.Write([]byte("companions:animal:" + seed))
		resolved.Animal = []string{"dog", "cat"}[hash.Sum32()%2]
	}
	return resolved, nil
}

func generateCompanion(ctx context.Context, seed, animal string) (Artwork, error) {
	if err := ctx.Err(); err != nil {
		return Artwork{}, err
	}
	return Artwork{Data: []byte(companionSVG(seed, animal)), MediaType: "image/svg+xml", Width: 128, Height: 128}, nil
}

type companionPalette struct{ coat, shade, light, accent string }

type companionPen struct{ strings.Builder }

func (p *companionPen) path(d, fill string, width float64) {
	stroke := "#342e2a"
	if width == 0 {
		stroke = "none"
	}
	fmt.Fprintf(&p.Builder, `<path d="%s" fill="%s" stroke="%s" stroke-width="%g" stroke-linecap="round" stroke-linejoin="round"/>`, d, fill, stroke, width)
}
func (p *companionPen) line(d string, width float64) { p.path(d, "none", width) }
func (p *companionPen) ellipse(x, y, rx, ry float64, fill string, width float64) {
	stroke := "#342e2a"
	if width == 0 {
		stroke = "none"
	}
	fmt.Fprintf(&p.Builder, `<ellipse cx="%g" cy="%g" rx="%g" ry="%g" fill="%s" stroke="%s" stroke-width="%g"/>`, x, y, rx, ry, fill, stroke, width)
}

func companionSVG(seed, animal string) string {
	// FNV-1a on UTF-8 bytes followed by xorshift32. Never copy seed text to SVG.
	state := uint32(2166136261)
	for i := 0; i < len(seed); i++ {
		state = (state ^ uint32(seed[i])) * 16777619
	}
	state ^= 0x517cc1b7
	if state == 0 {
		state = 0x9e3779b9
	}
	next := func(n int) int {
		state ^= state << 13
		state ^= state >> 17
		state ^= state << 5
		return int(state % uint32(n))
	}
	palettes := [...]companionPalette{
		{"#cf9654", "#956440", "#f8e9cc", "#7a9a91"},
		{"#efe3cb", "#a9764b", "#fff5df", "#b56b51"},
		{"#a66c48", "#72503b", "#f1d9b6", "#78968c"},
		{"#a4a9a3", "#697876", "#f3ead6", "#b77759"},
		{"#d6b888", "#a28359", "#fff0d5", "#708b99"},
		{"#6c6961", "#4f504c", "#e8ddc5", "#bf865a"},
		{"#d8945d", "#a45e3e", "#ffecd0", "#6e9290"},
		{"#f1ebd9", "#7b817b", "#fff7e6", "#b76d54"},
	}
	palette := palettes[next(len(palettes))]
	kind, expression := next(6), next(6)
	tilt, gaze, marking := next(15)-7, next(5)-2, next(5)
	var p companionPen
	p.Grow(7500)
	p.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 128 128" fill="none" aria-hidden="true">`)
	p.path("M0 0h128v128H0z", "#f5efdf", 0)
	// Quiet paper flecks use fixed geometry, so decoration never consumes identity.
	p.WriteString(`<path d="M19 42l2-.5M104 52l2 .4M30 100l1.5-.6M95 110l2-.3M65 14l2 .2M15 74l1.5 .5" stroke="#ded4bf" stroke-width=".7" stroke-linecap="round"/>`)
	// The shoulders sit inside the square; the face and ears fit its inscribed circle.
	p.ellipse(64, 119, 29, 2.4, "#e5dbc5", 0)
	p.path("M37 118Q37 103 47 94L51 85H77L81 94Q91 102 93 118Q71 122 37 118Z", palette.coat, 2)
	p.path("M51 93Q58 101 65 97Q70 101 77 92L81 117Q65 121 48 117Z", palette.light, 0)
	p.line("M44 109q-2 4-1 8M85 109q2 4 2 8", 1.2)
	if next(3) == 0 {
		p.path("M47 91Q63 102 81 91L78 100L66 112L52 101Z", palette.accent, 1.6)
		p.line("M52 98q10 6 21 0M65 103l1 5", 1)
	} else {
		p.path("M48 92Q64 101 80 92L80 98Q65 107 48 98Z", palette.accent, 1.5)
		p.ellipse(65, 103, 3, 3.5, "#deb86f", 1.1)
	}
	fmt.Fprintf(&p.Builder, `<g transform="rotate(%d 64 69)">`, tilt)
	if animal == "cat" {
		companionCat(&p, palette, kind%3, expression, gaze, marking)
	} else {
		companionDog(&p, palette, kind, expression, gaze, marking, next(3))
	}
	p.WriteString(`</g></svg>`)
	return p.String()
}

func companionDog(p *companionPen, c companionPalette, kind, expression, gaze, marking, muzzle int) {
	// Ears are separate silhouettes behind the cheek contour. Each recipe has a
	// compatible face width and muzzle, rather than independently shuffled parts.
	switch kind {
	case 0: // Soft hound ears and a longer, drooping cheek.
		p.path("M42 41C32 34 26 45 25 58C23 74 28 89 34 88Q40 85 43 66Z", c.shade, 2.2)
		p.path("M85 40C99 35 104 48 103 63C104 77 99 88 93 86Q88 82 86 65Z", c.shade, 2.2)
		p.line("M32 48Q27 64 32 79M96 48Q101 63 97 77", 1.1)
	case 1: // Shepherd-like upright ears, softened at the tips.
		p.path("M38 53Q28 43 34 22Q36 18 39 23L53 40Z", c.coat, 2.2)
		p.path("M76 40L90 23Q94 18 96 25Q100 42 90 55Z", c.coat, 2.2)
		p.path("M37 28L47 43L37 46Z", c.shade, 0)
		p.path("M91 28L82 44L92 47Z", c.shade, 0)
	case 2: // Folded terrier ears.
		p.path("M37 54L27 37Q31 28 47 38L53 44Z", c.shade, 2.2)
		p.path("M77 42Q93 27 102 37L91 57Z", c.shade, 2.2)
		p.path("M29 37L37 48L42 39M99 37L91 49L87 39", c.coat, 1.4)
	case 3: // Broad spaniel lobes.
		p.path("M44 42Q27 32 24 52L25 73L28 71Q26 90 35 86L38 88L45 68Z", c.shade, 2.2)
		p.path("M83 42Q101 33 105 53L104 73L101 71Q104 89 94 86L91 88L83 67Z", c.shade, 2.2)
		p.line("M30 51q-3 9 0 16M34 73l-1 6M97 50q4 9 1 17M96 74l1 6", 1.1)
	case 4: // Scruffy triangular folds.
		p.path("M38 55L29 42L31 40L26 33Q39 32 51 41M78 40Q91 31 102 35L97 40L99 43L90 56", c.shade, 2.2)
	default: // One ear up, one folded, a deliberately lopsided mutt.
		p.path("M36 54Q30 44 35 24Q37 19 40 25L54 43Z", c.shade, 2.2)
		p.path("M78 41Q97 30 104 46Q107 61 96 70L88 58Z", c.shade, 2.2)
		p.line("M96 44Q103 56 96 62", 1.1)
	}
	face := "M39 48C40 38 52 34 64 36C78 33 89 40 91 52L93 70Q95 84 82 91Q66 100 50 93Q35 88 36 74Z"
	if kind == 0 {
		face = "M40 48Q41 35 63 36Q86 32 89 50L90 72Q93 92 75 96Q56 102 43 90Q34 84 38 67Z"
	}
	if kind == 4 {
		face = "M38 51L35 46L41 46L42 39L49 41L54 34L58 38Q69 34 79 39L86 38L86 44L92 45L90 53L93 65L98 75L93 75L96 82L89 82L87 90L81 89Q66 100 51 92L43 93L42 87L34 85L37 79L31 76L36 69Z"
	}
	if kind == 2 {
		face = "M38 50Q40 37 63 37Q86 35 90 50L90 70L96 78L90 78L92 84Q81 96 64 96Q47 96 36 85L39 79L33 79L38 70Z"
	}
	p.path(face, c.coat, 2.2)
	switch marking {
	case 0:
		p.path("M40 49Q43 39 53 41Q64 46 59 62Q53 72 41 66Z", c.shade, 0)
	case 1:
		p.path("M71 42Q84 38 88 52L89 67Q77 72 70 63Q66 53 71 42Z", c.shade, 0)
	case 2:
		p.path("M60 38Q65 36 70 39L67 58L71 73L56 73L61 57Z", c.light, 0)
	case 3:
		p.path("M40 50Q46 37 56 41L58 57L51 66L40 64Z", c.shade, 0)
		p.path("M73 41Q85 40 88 52L89 65L77 65L70 56Z", c.shade, 0)
	}
	// A few bent tufts break the clean outline without becoming tiny fur noise.
	p.line("M54 40q4-6 8-5M63 40q4-4 7-3", 1.25)
	companionEyes(p, expression, gaze, false)
	// Muzzle widths are linked to the head recipe; all stay below the eye whites.
	y, wide := 78.0, 18.0
	if kind == 0 {
		y, wide = 81, 16
	}
	if kind == 2 || kind == 4 {
		wide = 21
	}
	if muzzle == 2 {
		wide -= 2
	}
	p.path(fmt.Sprintf("M%g %gQ%g %g 64 %gQ%g %g %g %gQ%g %g 64 %gQ%g %g %g %gZ", 64-wide, y, 64-wide, y-11, y-8, 64+wide, y-12, 64+wide, y, 64+wide, y+13, y+12, 64-wide, y+12, 64-wide, y), c.light, 1.25)
	if expression == 4 {
		p.path(fmt.Sprintf("M54 %gQ64 %g 75 %gQ72 %g 63 %gQ57 %g 54 %gZ", y+3, y+9, y+2, y+16, y+15, y+14, y+3), "#4a352c", 1.4)
		p.path(fmt.Sprintf("M62 %gQ69 %g 73 %gL71 %gQ64 %g 62 %gZ", y+10, y+7, y+10, y+16, y+21, y+10), "#d89483", 1.1)
		p.line(fmt.Sprintf("M68 %gl-1 4", y+11), .8)
	} else {
		p.line(fmt.Sprintf("M64 %gv5Q58 %g 53 %gM64 %gQ71 %g 77 %g", y, y+10, y+4, y+5, y+11, y+3), 1.55)
	}
	p.path(fmt.Sprintf("M56 %gQ63 %g 72 %gQ74 %g 67 %gQ63 %g 57 %gQ54 %g 56 %gZ", y-9, y-12, y-9, y-5, y-2, y+1, y-3, y-7, y-9), "#342e2a", .7)
	p.ellipse(60, y-7, 2, .7, c.light, 0)
	p.ellipse(51, y, .75, .75, "#7f6a51", 0)
	p.ellipse(76, y, .75, .75, "#7f6a51", 0)
	if kind == 4 {
		p.line("M39 68l4 3M87 70l4-3M47 88l3-2M78 88l3 1", 1.1)
	}
}

func companionCat(p *companionPen, c companionPalette, kind, expression, gaze, marking int) {
	// The tall ears and tiny divided muzzle keep every silhouette distinctly feline.
	face := "M36 57L33 29Q33 25 37 28L54 41Q64 37 75 41L92 27Q96 25 95 31L92 58Q100 73 91 83Q80 96 64 96Q47 96 37 84Q28 74 36 57Z"
	if kind == 1 {
		face = "M35 55L32 28Q32 25 36 27L53 41Q63 36 76 41L92 27Q96 25 95 30L92 55L98 63L94 65L101 75L95 75L97 82L89 83L86 91L80 89Q65 100 49 91L42 93L40 86L33 85L35 79L28 76L33 70L30 66Z"
	}
	if kind == 2 {
		face = "M39 56L34 27Q34 23 38 26L55 41Q65 38 75 41L92 25Q95 23 95 28L90 57Q95 72 86 84Q78 95 64 97Q52 96 42 86Q33 74 39 56Z"
	}
	p.path(face, c.coat, 2.2)
	p.path("M38 33L48 44L40 51Z", "#d9a594", 0)
	p.path("M90 33L80 44L89 51Z", "#d9a594", 0)
	p.line("M39 36l4 7M88 36l-4 8", 1)
	switch marking {
	case 0, 1: // Tabby caps follow the forehead; cheek bars don't touch whiskers.
		p.path("M52 43L57 46L59 57L54 51ZM63 41L67 41L67 54L63 56ZM74 43L78 44L73 55L70 56Z", c.shade, 0)
		p.path("M36 62L44 64L45 68L36 66ZM36 73L45 74L45 78L38 77ZM92 62L85 64L83 68L92 66ZM92 74L84 74L83 78L90 78Z", c.shade, 0)
	case 2: // Mask with a paper-colored nose bridge.
		p.path("M39 52L37 32L54 44L60 43L59 66L49 71L38 67Z", c.shade, 0)
		p.path("M70 43L77 44L92 31L89 54L92 65L79 71L68 66Z", c.shade, 0)
		p.path("M63 48L67 59L70 80H58L62 63Z", c.light, 0)
	case 3: // One large patch, deliberately clear of the contour.
		p.path("M72 44L80 44L91 32L88 52L91 65Q84 73 74 69Q65 63 72 44Z", c.shade, 0)
	}
	if kind == 1 {
		p.line("M52 43l3-5M60 43l4-5M39 81l5-1M85 83l4-2", 1.15)
	}
	companionEyes(p, expression, gaze, true)
	p.path("M64 76Q54 70 50 78Q46 85 55 88Q61 91 64 87Q70 92 77 86Q82 80 75 76Q70 73 64 76Z", c.light, 0)
	p.path("M60 76Q64 74 68 76Q69 78 64 81Q59 78 60 76Z", "#a56558", 1.1)
	p.line("M64 81v3Q60 91 55 85M64 84Q69 91 74 84", 1.45)
	p.line("M53 80Q40 75 29 77M52 84Q40 82 29 85M76 80Q88 75 100 77M77 84Q91 82 100 85", 1.15)
	p.line("M58 93q5 2 10 0", .9)
}

func companionEyes(p *companionPen, expression, gaze int, cat bool) {
	y, spread := 61.0, 13.0
	if cat {
		y, spread = 64, 13
	}
	for side := -1; side <= 1; side += 2 {
		x := 64 + float64(side)*spread
		if expression == 2 || expression == 3 && side == 1 {
			// Happy closed eyes curve upward, with relaxed raised brows.
			p.line(fmt.Sprintf("M%g %gQ%g %g %g %g", x-5, y+1, x, y-5, x+5, y+1), 1.8)
			continue
		}
		if expression == 1 { // Drowsy, with horizontal lids instead of angry angles.
			p.path(fmt.Sprintf("M%g %gQ%g %g %g %gQ%g %g %g %gZ", x-5, y-1, x, y-2, x+5, y-1, x, y+8, x-5, y-1), "#fff8e8", 1.3)
			p.ellipse(x+float64(gaze)*.6, y+1.5, 1.8, 2.3, "#342e2a", 0)
			p.line(fmt.Sprintf("M%g %gQ%g %g %g %g", x-5.5, y-1, x, y-2, x+5.5, y-1), 1.6)
		} else {
			rx, ry := 5.0, 6.1
			if cat {
				rx, ry = 5.8, 5.3
			}
			if expression == 5 && side == -1 {
				ry += 1.3
			}
			p.ellipse(x, y, rx, ry, "#fff8e8", 1.3)
			p.ellipse(x+float64(gaze)*.8, y+.4, 2.05, 3.1, "#342e2a", 0)
			p.ellipse(x+float64(gaze)*.8-.6, y-.7, .65, .85, "#fff8e8", 0)
		}
		brow := y - 10
		if expression == 5 && side == -1 {
			brow -= 2
		}
		p.line(fmt.Sprintf("M%g %gQ%g %g %g %g", x-4, brow+1, x, brow-2, x+4, brow), 1.35)
	}
}
