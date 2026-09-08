package avatar

import (
	"context"
	"fmt"
	"math"
	"strings"
)

// FieldBirds draws invented species with field-guide anatomy and plumage.
// All geometry is derived locally from the seed; the artwork uses no resources.
type FieldBirds struct{}

func (FieldBirds) Style() Style {
	return Style{ID: "field-birds", Name: "Field Birds", Description: "Invented birds in quiet field-guide color, with varied silhouettes and fine feather detail.", NativeWidth: 160, NativeHeight: 160}
}
func (FieldBirds) Generate(ctx context.Context, seed string) (Artwork, error) {
	if err := ctx.Err(); err != nil {
		return Artwork{}, err
	}
	return Artwork{Data: []byte(fieldBirdSVG(seed)), MediaType: "image/svg+xml", Width: 160, Height: 160}, nil
}

type birdPalette struct{ back, wing, edge, breast, belly, accent string }

var birdPalettes = [...]birdPalette{
	{"#7d8870", "#505e53", "#aeb89a", "#d5c277", "#f0e8c9", "#b47642"},
	{"#708699", "#455d73", "#acbdc6", "#be704c", "#efddbc", "#dad2b4"},
	{"#9d7755", "#675944", "#c8ad7b", "#d4ab6c", "#eee2c6", "#544f48"},
	{"#b59b4c", "#6d6e44", "#d5cb8a", "#e0c567", "#f5e8b9", "#737e71"},
	{"#647e78", "#365c5c", "#92afa1", "#c39565", "#e7dfc2", "#c78454"},
	{"#a0614c", "#684e46", "#cfa384", "#c4a182", "#f0e4cb", "#d7bd79"},
	{"#8b8191", "#585461", "#b4a5b4", "#c99b8e", "#eadbd1", "#808f8b"},
	{"#aaa99b", "#72796e", "#d5d4bf", "#ded8c1", "#f7f0dc", "#c19850"},
	{"#4f5f67", "#34434b", "#889b9d", "#abb5ae", "#e6e6d5", "#bc7a55"},
	{"#99905e", "#625f45", "#c2b889", "#d5d3b1", "#eeecd7", "#947052"},
	{"#799298", "#4a6a79", "#b6c8c8", "#d7b878", "#f0e4c5", "#9f6658"},
	{"#b49280", "#765c58", "#d4b6a0", "#ddd0b1", "#f0e9d6", "#6b8186"},
}

// The continuous body contours include the head and neck. Local wing frames
// let the feather construction follow each family's anatomy without a clipPath.
type birdForm struct {
	name, body, tail, legs          string
	hx, hy, wx, wy, wsx, wsy, angle float64
	bill                            int
}

var birdForms = [...]birdForm{
	{"songbird", "M39 88Q57 76 81 67Q91 64 96 52C98 37 119 35 124 49Q130 53 123 61Q118 68 115 82C112 102 93 112 73 105Q52 100 39 88Z", "M51 83L20 101L22 106L58 95Z", "M79 103L83 118L77 126M83 118L94 124M77 126L72 125M95 103L99 119L109 122M99 119L96 125", 113, 50, 73, 74, 1, 1, -7, 0},
	{"finch", "M37 89Q49 66 79 65Q86 59 87 51C87 31 113 28 122 45L124 58Q113 66 116 82C119 107 93 119 70 109Q48 105 37 89Z", "M47 87L24 99L28 109L62 100Z", "M76 108L80 122L73 128M80 122L91 127M96 109L99 123L111 126M99 123L96 129", 110, 47, 65, 73, 1.08, 1.05, -11, 1},
	{"longtail", "M54 79Q69 66 84 65Q92 60 95 49C100 34 119 38 122 51L123 60Q116 65 113 78Q105 98 87 101Q65 94 54 79Z", "M67 82Q43 99 25 121L32 123Q58 107 79 92Z", "M87 98L86 111L97 118M86 111L80 119M99 94L102 108L111 111M102 108L101 117", 112, 49, 77, 70, .92, .95, -12, 0},
	{"groundbird", "M33 88Q36 64 61 65Q76 69 90 62L96 50C102 37 120 42 121 56Q127 60 119 66L121 86C120 112 98 121 75 117Q42 116 33 88Z", "M42 77L26 69L30 91L52 99Z", "M69 115L69 128L59 131M69 128L81 132M98 116L102 128L113 129M102 128L99 134", 112, 54, 63, 79, 1.1, 1.02, 9, 1},
	{"wader", "M36 79Q48 62 73 68Q86 72 93 67Q98 64 93 53Q88 44 94 35Q101 25 114 32Q123 36 120 44Q116 48 108 46Q102 46 106 57Q118 75 108 90Q99 101 73 97Q50 95 36 79Z", "M48 74L25 80L30 86L59 88Z", "M71 95L74 113L71 132L81 134M71 132L65 134M91 96L94 113L99 131L110 133M99 131L95 134", 110, 38, 62, 72, 1.08, .67, 10, 2},
	{"shorebird", "M30 86Q39 72 67 75Q90 79 96 64L99 56Q105 43 119 49Q130 55 123 64Q119 69 115 84C111 104 89 110 67 105Q43 99 30 86Z", "M45 82L23 87L29 94L55 94Z", "M70 104L73 120L68 130L80 132M68 130L63 132M94 102L94 118L105 130L114 131M105 130L100 133", 114, 56, 61, 80, 1.05, .7, -8, 3},
	{"raptor", "M47 95Q48 76 69 67L81 52Q83 35 100 34Q118 30 122 46L119 61Q115 71 117 87Q114 110 96 115Q67 115 47 95Z", "M64 94L38 114L45 120L83 108Z", "M88 113L87 128L76 132M87 128L96 132M103 111L105 125L116 129M105 125L102 132", 109, 46, 73, 74, .94, 1.12, 3, 4},
	{"nectarbird", "M46 91Q64 77 80 68Q93 63 94 52Q97 38 110 40Q124 42 120 56Q113 63 111 75Q110 93 96 102Q80 111 63 101Z", "M65 93Q46 111 29 116L32 122Q61 112 80 100Z", "M89 103L90 115L100 120M90 115L84 121M99 98L104 111L113 115", 111, 49, 77, 75, .85, .91, -12, 5},
}

type birdPen struct{ strings.Builder }

func (p *birdPen) path(d, fill, stroke string, width float64) {
	fmt.Fprintf(&p.Builder, `<path d="%s" fill="%s" stroke="%s" stroke-width="%.2f" stroke-linecap="round" stroke-linejoin="round"/>`, d, fill, stroke, width)
}
func (p *birdPen) line(d, color string, width float64) { p.path(d, "none", color, width) }
func (p *birdPen) ellipse(x, y, rx, ry float64, fill string) {
	fmt.Fprintf(&p.Builder, `<ellipse cx="%.2f" cy="%.2f" rx="%.2f" ry="%.2f" fill="%s"/>`, x, y, rx, ry, fill)
}

func fieldBirdRecipe(seed string) (func(int) int, int, int) {
	state := uint32(2166136261)
	for i := 0; i < len(seed); i++ {
		state = (state ^ uint32(seed[i])) * 16777619
	}
	state ^= 0xa511e9b3
	if state == 0 {
		state = 0x9e3779b9
	}
	next := func(n int) int {
		state ^= state << 13
		state ^= state >> 17
		state ^= state << 5
		return int(state % uint32(n))
	}
	return next, next(len(birdForms)), next(len(birdPalettes))
}
func fieldBirdSVG(seed string) string {
	next, family, palette := fieldBirdRecipe(seed)
	f, c := birdForms[family], birdPalettes[palette]
	// Coherent markings vary independently of the construction family.
	markings, face, crest, flip := next(4), next(4), next(5), next(2)
	var p birdPen
	p.Grow(24000)
	p.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 160 160" fill="none" aria-hidden="true"><defs>`)
	fmt.Fprintf(&p.Builder, `<linearGradient id="bird-body" x1="15%%" y1="0%%" x2="70%%" y2="100%%"><stop stop-color="%s"/><stop offset=".48" stop-color="%s"/><stop offset=".78" stop-color="%s"/><stop offset="1" stop-color="%s"/></linearGradient>`, c.back, c.back, c.breast, c.belly)
	fmt.Fprintf(&p.Builder, `<linearGradient id="bird-wing" x1="20%%" y1="0%%" x2="70%%" y2="100%%"><stop stop-color="%s"/><stop offset="1" stop-color="%s"/></linearGradient></defs>`, c.back, c.wing)
	p.path("M0 0h160v160H0Z", "#faf8ef", "none", 0)
	if flip == 1 {
		p.WriteString(`<g transform="translate(160 0) scale(-1 1)">`)
	} else {
		p.WriteString(`<g>`)
	}
	p.ellipse(80, 136, 30, 1.1, "#e6e1d3")
	p.path(f.tail, c.wing, c.wing, .7)
	// Tail detail stays inside the family-specific contour.
	if family == 2 || family == 7 {
		for i := 0; i < 3; i++ {
			p.line(fmt.Sprintf("M%d %dQ49 106 %d %d", 68+i*2, 91+i, 31+i*2, 118), c.edge, .4)
		}
	}
	legs := f.legs
	p.line(legs, "#7b7160", 1.65)
	p.line(legs, "#b9a687", .65)
	p.path(f.body, "url(#bird-body)", c.wing, .65)
	// The throat wash and face are attached to the head, not to screen coordinates.
	// A feathered cheek patch replaces an airbrushed spot with a field mark.
	if face == 2 {
		p.ellipse(f.hx-5, f.hy+7, 5.5, 3.5, c.accent)
		for i := 0; i < 9; i++ {
			x := f.hx - 9 + float64(i)
			p.line(fmt.Sprintf("M%.1f %.1fl-.5 1.4", x, f.hy+8), c.back, .4)
		}
	}
	if crest == 0 && family != 4 && family != 5 {
		p.path(fmt.Sprintf("M%.1f %.1fQ%.1f %.1f %.1f %.1fL%.1f %.1fQ%.1f %.1f %.1f %.1fZ", f.hx-16, f.hy-6, f.hx-15, f.hy-19, f.hx-5, f.hy-22, f.hx-7, f.hy-13, f.hx, f.hy-17, f.hx+6, f.hy-9), c.back, c.wing, .5)
	}
	// Face masks, eyebrows, and cheeks remain small and anatomical.
	switch face {
	case 0:
		p.path(fmt.Sprintf("M%.1f %.1fQ%.1f %.1f %.1f %.1fL%.1f %.1fQ%.1f %.1f %.1f %.1fZ", f.hx-14, f.hy-4, f.hx, f.hy-9, f.hx+9, f.hy-2, f.hx+7, f.hy+1, f.hx-2, f.hy-4, f.hx-14, f.hy-1), c.belly, "none", 0)
	case 1:
		p.path(fmt.Sprintf("M%.1f %.1fQ%.1f %.1f %.1f %.1fL%.1f %.1fQ%.1f %.1f %.1f %.1fZ", f.hx-13, f.hy-1, f.hx-1, f.hy-4, f.hx+9, f.hy+2, f.hx+5, f.hy+5, f.hx-5, f.hy+2, f.hx-13, f.hy+3), c.wing, "none", 0)
	case 2:
		p.line(fmt.Sprintf("M%.1f %.1fl4 1", f.hx-9, f.hy+5), c.edge, .7)
	case 3:
		p.line(fmt.Sprintf("M%.1f %.1fQ%.1f %.1f %.1f %.1f", f.hx-12, f.hy+2, f.hx-6, f.hy+9, f.hx+2, f.hy+7), c.belly, 1.8)
	}
	// Fine, broken contour feathers give the exposed body a painted surface.
	// This ellipse sits inside each continuous body contour; the wing occludes
	// its back half, preserving natural feather layering without SVG clipping.
	p.WriteString(`<g opacity=".38">`)
	for i := 0; i < 320; i++ {
		a := float64(next(6283)) / 1000
		r := math.Sqrt(float64(next(1000)) / 1000)
		x := f.wx + 13 + math.Cos(a)*r*18
		y := f.wy + 15 + math.Sin(a)*r*9*f.wsy
		color := c.belly
		if i%3 == 0 {
			color = c.back
		}
		p.line(fmt.Sprintf("M%.1f %.1fq.8 .8 .2 %.1f", x, y, .8+float64(next(15))/10), color, .18+float64(next(12))/100)
	}
	p.WriteString(`</g>`)
	// Fine breast streaks sit in a bounded patch below the folded wing's leading edge.
	for row := 0; row < 5; row++ {
		for col := 0; col < 3; col++ {
			x := f.wx + 19 + float64(col*3) - float64(row)*1.4
			y := f.wy + 11 + float64(row)*3
			if family == 4 {
				x -= 4
				y -= 2
			}
			if markings == 0 {
				p.line(fmt.Sprintf("M%.1f %.1fq1 1.5 .4 3", x, y), c.back, .65)
			}
			if markings == 1 {
				p.path(fmt.Sprintf("M%.1f %.1fl1.3 2.1 -2 .3Z", x, y), c.back, "none", 0)
			}
			if markings == 2 {
				p.line(fmt.Sprintf("M%.1f %.1fq2 1.5 4 0", x, y), c.back, .5)
			}
		}
	}
	// Folded wing: long primaries, overlapping secondaries, and scalloped coverts.
	fmt.Fprintf(&p.Builder, `<g transform="translate(%.1f %.1f) rotate(%.1f) scale(%.2f %.2f)">`, f.wx, f.wy, f.angle, f.wsx, f.wsy)
	p.path("M-18 7Q-5-8 16-3Q29 1 20 14Q10 28-31 35Q-22 23-18 7Z", "url(#bird-wing)", c.wing, .6)
	for i := 0; i < 8; i++ {
		x := float64(i)*3.5 - 13
		endx := -32 + float64(i)*4.3
		endy := 35 - float64(i)*1.65
		p.path(fmt.Sprintf("M%.1f 5Q%.1f 16 %.1f %.1fQ%.1f %.1f %.1f %.1fQ%.1f 15 %.1f 4Z", x, x-2, endx, endy, endx+2, endy+1, endx+4, endy-1, x+7, x+4), c.wing, c.edge, .55)
		p.line(fmt.Sprintf("M%.1f 10L%.1f %.1f", x+2, endx+2, endy-2), c.back, .35)
	}
	p.path("M-18 7Q-6-9 16-3Q25 0 23 6Q13 17-6 21L-18 17Z", "url(#bird-wing)", c.wing, .4)
	for row := 0; row < 3; row++ {
		for col := 0; col < 7-row; col++ {
			x := -14 + float64(col)*4.8 + float64(row)*2
			y := 3 + float64(row)*4.4 + math.Sin(float64(col)*.7)*2
			p.line(fmt.Sprintf("M%.1f %.1fq1.3 4 4 1", x, y), c.edge, .5)
		}
	}
	bars := 1 + next(3)
	if markings != 3 {
		for i := 0; i < bars; i++ {
			p.line(fmt.Sprintf("M%d %dQ2 %d %d %d", -16+i*2, 9+i*5, 18+i*4, 18-i*3, 5+i*6), c.belly, 1.5)
		}
	}
	p.WriteString(`<g opacity=".4">`)
	for i := 0; i < 90; i++ {
		x := -12 + float64(next(290))/10
		y := 2 + float64(next(110))/10
		p.line(fmt.Sprintf("M%.1f %.1fl-1.2 1.7", x, y), c.edge, .22)
	}
	p.WriteString(`</g>`)
	p.WriteString(`</g>`)
	// Small dry-brush marks on the crown and nape break up the smooth washes.
	for i := 0; i < 17; i++ {
		a := float64(i) * .31
		x := f.hx - 6 + math.Cos(a)*6
		y := f.hy - 4 + math.Sin(a)*1.5
		p.line(fmt.Sprintf("M%.1f %.1fl-1.7 -.5", x, y), c.edge, .3)
	}
	bx, by := f.hx+9, f.hy+2
	length := float64(8 + next(5))
	switch f.bill {
	case 1:
		p.path(fmt.Sprintf("M%.1f %.1fl%.1f 5 -%.1f 5Z", bx, by-4, length, length+1), "#c6ae85", c.wing, .6)
	case 2:
		p.path(fmt.Sprintf("M%.1f %.1fl24 3 -24 2Z", bx, by-2), "#b7a173", c.wing, .5)
	case 3, 5:
		p.path(fmt.Sprintf("M%.1f %.1fQ%.1f %.1f 143 %.1fQ136 %.1f %.1f %.1fZ", bx, by-1, bx+11, by, by+8, by+3, bx, by+2), c.wing, c.wing, .4)
	case 4:
		p.path(fmt.Sprintf("M%.1f %.1fq13 -3 11 7l-5 4 1-6 -8-1Z", bx-2, by-2), "#b5a16b", c.wing, .6)
	default:
		p.path(fmt.Sprintf("M%.1f %.1fl%.1f 4 -%.1f 1Z", bx, by-2, length, length+1), c.wing, c.wing, .4)
	}
	p.line(fmt.Sprintf("M%.1f %.1fl6 1", bx, by+1), "#494b43", .4)
	p.ellipse(f.hx+7, f.hy+1, .6, .4, "#48483e")
	p.ellipse(f.hx, f.hy, 2.55, 2.35, c.belly)
	p.ellipse(f.hx+.15, f.hy, 1.8, 1.9, "#252d2b")
	p.ellipse(f.hx+.65, f.hy-.65, .55, .55, "#faf8ef")
	p.WriteString(`</g></svg>`)
	return p.String()
}
