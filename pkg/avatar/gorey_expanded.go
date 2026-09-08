package avatar

import (
	"context"
	"fmt"
	"strings"
)

// GoreyExpanded adds a larger cast while leaving the original Gorey recipes intact.
type GoreyExpanded struct{}

func (GoreyExpanded) Style() Style {
	return Style{NativeWidth: 64, NativeHeight: 72, ID: "gorey-expanded", Name: "Gorey Expanded", Description: "An extended cast of engraved portraits, with varied faces, hair, hats, and dress."}
}
func (GoreyExpanded) Generate(ctx context.Context, seed string) (Artwork, error) {
	if err := ctx.Err(); err != nil {
		return Artwork{}, err
	}
	return Artwork{Data: []byte(expandedSVG(seed)), MediaType: "image/svg+xml", Width: 64, Height: 72}, nil
}

type portraitPen struct{ strings.Builder }

func (p *portraitPen) path(d, fill string, width float64, color string) {
	if fill != "none" && width >= .65 {
		width = max(width, .9)
	}
	fmt.Fprintf(&p.Builder, `<path d="%s" fill="%s" stroke="%s" stroke-width="%g" stroke-linecap="round" stroke-linejoin="round"/>`, d, fill, color, width)
}
func (p *portraitPen) ellipse(x, y, rx, ry float64, fill string) {
	fmt.Fprintf(&p.Builder, `<ellipse cx="%g" cy="%g" rx="%g" ry="%g" fill="%s"/>`, x, y, rx, ry, fill)
}
func expandedSVG(seed string) string {
	hash := uint32(2166136261)
	for _, char := range seed {
		hash = (hash ^ uint32(char)) * 16777619
	}
	hash ^= 0xa67f31d9
	next := func(max int) int {
		hash ^= hash << 13
		hash ^= hash >> 17
		hash ^= hash << 5
		return int(hash % uint32(max))
	}
	ink := "#292820"
	paper := []string{"#ddd5bf", "#d1c9b6", "#e0d6c0", "#d3ccba", "#d7cfb9"}[next(5)]
	backdrop := []string{"#c4b69b", "#b7b49a", "#c2a89d", "#c2b8a8", "#b4b7a8", "#c8b4a9", "#b4a8a2", "#cbbc9f"}[next(8)]
	cloth := []string{"#3c3a32", "#50483e", "#41453d", "#53433d", "#484543", "#555246"}[next(6)]
	hairColor := []string{ink, ink, "#49463b", "#6a6251", "#bbb29d"}[next(5)]
	hairLine := paper
	if hairColor == "#bbb29d" {
		hairLine = "#686252"
	}
	feminine := next(2) == 0
	hair := next(9)
	if feminine {
		hair = []int{1, 2, 3, 4, 5, 7, 8}[next(7)]
	} else {
		hair = []int{0, 1, 2, 5, 6, 7, 8}[next(7)]
	}
	hat := next(12) // Four headwear silhouettes; most portraits leave the hair visible.
	outfit := next(9)
	if feminine {
		outfit = []int{1, 2, 3, 4, 5, 6, 8}[next(7)]
	}
	width := 10 + next(5)
	left, right := 32-width, 32+width
	chin := 43 + next(7)
	shape := next(4)
	age := next(4)
	eyeY := 27 + next(3)
	nose := 34 + next(4)
	expression := next(5)
	accessory := next(8)
	shoulder := next(3)
	pen := &portraitPen{}
	fmt.Fprintf(&pen.Builder, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 72" fill="none" aria-hidden="true"><path fill="%s" d="M0 0h64v72H0z"/>`, backdrop)
	p := pen.path
	f := fmt.Sprintf
	// Light, uneven engraving in the paper, never a decorative prop.
	for i := 0; i < 27; i++ {
		x, y := 3+next(58), 3+next(66)
		p(f("M%d %dl%d %d", x, y, 1+next(3), next(3)-1), "none", .22, "#8e897a")
	}
	// Long hair, braids and the bun sit behind the neck and clothing.
	switch hair {
	case 3:
		p(f("M%d 19Q%d 5 32 9Q%d 8 %d 22L%d 58Q47 62 39 55L25 55Q16 62 %d 58Z", left-2, left-7, right+8, right+3, right+5, left-5), hairColor, .8, ink)
		for i := 0; i < 5; i++ {
			p(f("M%g 23q-3 15 -1 32M%g 23q3 15 1 32", float64(left-2)+float64(i)*.8, float64(right+2)-float64(i)*.8), "none", .3, hairLine)
		}
	case 4:
		p(f("M%d 27Q%d 43 %d 56L%d 59L%d 42M%d 27Q%d 43 %d 56L%d 59L%d 42", left, left-5, left-7, left-2, left+1, right, right+5, right+7, right+2, right-1), hairColor, .7, ink)
		for i := 0; i < 7; i++ {
			y := 34 + i*3
			p(f("M%d %dl4 3-4 2M%d %dl-4 3 4 2", left-2-i/3, y, right+2+i/3, y), "none", .4, hairLine)
		}
	case 8:
		pen.ellipse(float64(right+2), 18, 5, 7, hairColor)
		for i := 0; i < 4; i++ {
			p(f("M%d %dq6 -4 7 3", right-1, 14+i*2), "none", .35, hairLine)
		}
	}
	// The body has a different shoulder fall and a single considered garment.
	p(f("M%d 72L%d 59Q17 %d 26 49L38 49Q49 %d %d 59L%d 72Z", 4+shoulder, 8+shoulder, 51+shoulder, 51+shoulder, 56-shoulder, 60-shoulder), cloth, .85, ink)
	for i := 0; i < 17; i++ {
		x := 8 + i*3
		p(f("M%d 71l%d -%d", x, next(3)-1, 6+next(8)), "none", .42, paper)
	}
	p("M27 41L26 52L32 58L39 52L37 41", paper, .8, ink)
	switch outfit {
	case 0: // Narrow tie and waistcoat.
		p("M25 49L18 54L27 64L32 56L37 64L46 54L39 49L32 53Z", paper, .8, ink)
		p("M30 54h4l-1 5 2 11-3 2-3-2 2-11Z", ink, .5, ink)
		p("M15 56l7 9-3 2 8 5M49 56l-7 9 3 2-8 5", "none", .6, paper)
	case 1: // High-neck buttoned dress with slightly puffed shoulders.
		p("M25 48L26 54Q32 57 38 54L39 48Q32 51 25 48Z", cloth, .7, ink)
		for x := 27; x <= 37; x += 2 {
			p(f("M%d 50v4", x), "none", .35, paper)
		}
		p("M15 54Q8 57 12 66M49 54Q56 57 52 66M32 57v15", "none", .65, paper)
		for y := 60; y <= 70; y += 5 {
			pen.ellipse(34, float64(y), .8, .8, paper)
		}
	case 2: // Ivory blouse, a soft pointed collar, and fine gathers.
		p("M8 72L11 59Q17 53 26 51L32 56L38 51Q48 54 53 59L57 72Z", paper, .7, ink)
		p("M26 49L22 53L27 60L32 55L37 60L42 53L38 49L32 54Z", paper, .7, ink)
		p("M32 57v15M16 57q3 6 2 14M48 57q-3 6-2 14", "none", .55, ink)
		for i := 0; i < 6; i++ {
			p(f("M%d 60l1 %dM%d 60l-1 %d", 20+i, 7+next(5), 44-i, 7+next(5)), "none", .24, "#837c69")
		}
		for y := 63; y < 72; y += 5 {
			pen.ellipse(34, float64(y), .65, .65, ink)
		}
	case 3: // Wrapped wool scarf above a simple coat.
		p("M24 48Q32 53 40 48L42 55Q32 61 22 55Z", paper, .75, ink)
		p("M32 56L38 55L43 70L37 72Z", paper, .65, ink)
		for i := 0; i < 4; i++ {
			p(f("M24 %dq8 3 16-1", 50+i), "none", .35, "#7b7463")
		}
		for i := 0; i < 5; i++ {
			p(f("M%d 59l4 11", 33+i), "none", .28, "#7b7463")
		}
		p("M17 55l7 9-3 4M48 56l-4 13", "none", .75, paper)
	case 4: // Broad shawl over a round-neck dress.
		p("M25 49Q32 58 39 49L43 53Q32 64 21 53Z", paper, .65, ink)
		p("M23 51L12 55L27 72L32 66Z", cloth, .7, ink)
		p("M41 51L52 55L37 72L32 66Z", cloth, .7, ink)
		for i := 0; i < 6; i++ {
			p(f("M%d %dl12 13M%d %dl-12 13", 13+i, 56-i/2, 51-i, 56-i/2), "none", .35, paper)
		}
		pen.ellipse(32, 65, 1.5, 1.9, paper)
	case 5: // Round-neck knit with crosshatching.
		p("M23 50Q32 64 41 50L44 53Q32 70 20 53Z", cloth, .8, ink)
		p("M23 53Q32 65 41 53", "none", .65, paper)
		for y := 62; y <= 71; y += 3 {
			p(f("M12 %dh40", y), "none", .33, paper)
		}
	case 6: // A fitted dark dress, ivory collar and an oval brooch.
		p("M26 49L21 53Q24 62 31 57L32 54L33 57Q40 62 43 53L38 49L32 53Z", paper, .8, ink)
		pen.ellipse(32, 56, 1.5, 2, ink)
		pen.ellipse(32, 55.5, .55, .8, paper)
		p("M32 60v12M17 59l5 13M47 59l-5 13", "none", .5, paper)
	case 7: // Evening shirt and modest bow tie.
		p("M25 49L18 54L26 65L32 56L38 65L46 54L39 49L32 54Z", paper, .8, ink)
		p("M30 55l-6 2 5 4 3-3 3 3 5-4-6-2Z", ink, .6, ink)
		p("M32 61v11", "none", .6, paper)
	case 8: // Double-breasted tweed coat.
		p("M26 49L18 53L27 67L31 59Z", paper, .7, ink)
		p("M38 49L46 53L37 67L33 59Z", paper, .7, ink)
		for y := 62; y < 73; y += 4 {
			p(f("M11 %dh42", y), "none", .28, paper)
		}
		for y := 65; y <= 70; y += 5 {
			pen.ellipse(30, float64(y), .9, .9, ink)
			pen.ellipse(36, float64(y), .9, .9, ink)
		}
	}
	// Ear contours and four distinct jaw geometries.
	p(f("M%d 27Q%d 22 %d 34L%d 37M%d 27Q%d 23 %d 34L%d 37", left, left-5, left-3, left+1, right, right+5, right+3, right-1), paper, .75, ink)
	switch shape {
	case 0:
		p(f("M%d 24Q%d 10 32 10Q%d 10 %d 24L%d 36Q%d %d 32 %dQ%d %d %d 36Z", left, left-1, right+1, right, right-1, right-4, chin-1, chin+1, left+4, chin-1, left+1), paper, .95, ink)
	case 1:
		p(f("M%d 24Q%d 10 32 10Q%d 10 %d 24L%d 39L37 %dL27 %dL%d 39Z", left, left-2, right+2, right, right-1, chin, chin, left+1), paper, .95, ink)
	case 2:
		p(f("M%d 24Q%d 9 32 10Q%d 9 %d 24Q%d 40 37 %dQ32 %d 27 %dQ%d 40 %d 24Z", left, left-2, right+2, right, right+2, chin-1, chin+3, chin-1, left-2, left), paper, .95, ink)
	case 3:
		p(f("M%d 24Q%d 11 32 11Q%d 11 %d 24L%d 34L%d 41L35 %dL29 %dL%d 41L%d 34Z", left, left-1, right+1, right, right-2, right-1, chin, chin+1, left+1, left+2), paper, .95, ink)
	}
	// Fine cheek shading belongs to the facial planes, with heavier age engraving.
	for i := 0; i < 4+age; i++ {
		p(f("M%g %dl%g 4", float64(left+1)+float64(i)*.55, 32+i, 1.4+float64(shape)*.2), "none", .27, ink)
	}
	for i := 0; i < 3+age; i++ {
		p(f("M%g %dl-1 4", float64(right-3)+float64(i)*.45, 31+i), "none", .25, ink)
	}
	if age > 1 {
		p(f("M27 18q5 -2 10 0M28 21q4 -1 8 0M%d %dl-2 2M%d %dl2 2", left+3, eyeY+2, right-3, eyeY+2), "none", .3, "#706a5b")
		p(f("M28 %dq-2 1-2 4M36 %dq2 1 2 4", nose+1, nose+1), "none", .35, ink)
	}
	browLeft, browRight := eyeY-3, eyeY-3
	if expression == 1 {
		browLeft--
		browRight++
	}
	if expression == 3 {
		browLeft++
		browRight--
	}
	p(f("M%d %dl6 -1M%d %dl6 2", left+3, browLeft, right-9, browRight-1), "none", 1, ink)
	p(f("M%d %dq3 -2 6 0M%d %dq3 -2 6 0", left+3, eyeY, right-9, eyeY), "none", .8, ink)
	pen.ellipse(float64(left+6), float64(eyeY), .7, 1, ink)
	pen.ellipse(float64(right-6), float64(eyeY), .7, 1, ink)
	p(f("M32 %dL%d %dQ31 %d 35 %d", eyeY-2, 29-next(2), nose, nose+2, nose), "none", .8, ink)
	mouthY := nose + 5
	mouth := []string{f("M29 %dq3 -1 6 0", mouthY), f("M29 %dq3 .5 7-1", mouthY), f("M29 %dq3 -2 6 0", mouthY), f("M29 %dl6 1", mouthY), f("M28 %dq4 .3 8-1", mouthY)}[expression]
	p(mouth, "none", .65, ink)
	p(f("M31 %dh3", mouthY+2), "none", .3, ink)
	// Accessories stay on the face; no novelty props compete with the portrait.
	switch accessory {
	case 0, 1:
		radius := 3.7
		if accessory == 1 {
			radius = 4.4
		}
		fmt.Fprintf(&pen.Builder, `<g fill="none" stroke="%s" stroke-width=".65"><ellipse cx="%d" cy="%d" rx="%g" ry="%g"/><ellipse cx="%d" cy="%d" rx="%g" ry="%g"/></g>`, ink, left+6, eyeY, radius, radius*.9, right-6, eyeY, radius, radius*.9)
		p(f("M%g %dh%gM%d %dl-3-1M%d %dl3-1", float64(left+6)+radius, eyeY, float64(2*width-12)-2*radius, left+2, eyeY, right-2, eyeY), "none", .6, ink)
	case 2:
		if !feminine {
			p(f("M32 %dq-4 -3-8 2 5 1 8-1 3 2 8 0-4-4-8-1Z", nose+3), hairColor, .5, ink)
		}
	case 3:
		if !feminine {
			p(f("M26 %dq6 4 12 0L35 %dL29 %dZ", chin-5, chin+3, chin+3), hairColor, .5, ink)
			for x := 29; x <= 35; x += 2 {
				p(f("M%d %dv4", x, chin-2), "none", .28, hairLine)
			}
		}
	case 4:
		if feminine {
			pen.ellipse(float64(left-1), 37, 1, 1.4, ink)
			pen.ellipse(float64(right+1), 37, 1, 1.4, ink)
		} else {
			p(f("M26 %dq6 2 12 0", chin-3), "none", .3, ink)
		}
	case 5:
		fmt.Fprintf(&pen.Builder, `<circle cx="%d" cy="%d" r="4.3" fill="none" stroke="%s" stroke-width=".65"/>`, right-6, eyeY, ink)
		p(f("M%d %dq7 13 3 22", right-2, eyeY+2), "none", .4, ink)
	}
	// Hair architecture is independent of age and clothes but carefully layered.
	switch hair {
	case 0: // Close-cropped, with a receding temple.
		p(f("M%d 29L%d 16Q25 8 32 9Q40 8 %d 17L%d 29L%d 20Q32 13 %d 20Z", left, left-1, right+1, right, right-2, left+2), hairColor, .7, ink)
		for i := 0; i < 10; i++ {
			p(f("M%d %dl1-3", left+2+i*2, 16+next(3)), "none", .3, hairLine)
		}
	case 1: // Side part, swept fringe.
		p(f("M%d 27Q%d 9 27 9Q35 5 %d 16L%d 27L%d 19Q31 18 27 13L%d 24Z", left-1, left-6, right+2, right+1, right-3, left+3), hairColor, .8, ink)
		for i := 0; i < 7; i++ {
			p(f("M%d 20q-3-6 2-9", left+3+i*2), "none", .35, hairLine)
		}
	case 2: // Waved bob with an engraved curved hem.
		p(f("M%d 35Q%d 9 27 9Q40 3 %d 23L%d 43L%d 43L%d 24Q37 17 29 14Q23 21 %d 25L%d 43L%d 43Z", left-4, left-7, right+4, right+4, right, right-2, left+1, left, left-4), hairColor, .8, ink)
		for i := 0; i < 5; i++ {
			p(f("M%d %dQ29 %d %d %d", left, 15+i, 7+i, right, 19+i), "none", .3, hairLine)
		}
		p(f("M%d 27q-1 5 0 9M%d 27q1 5 0 9", left-1, right+1), "none", .35, hairLine)
	case 3: // Long center part.
		p(f("M%d 29Q%d 10 32 9Q%d 10 %d 29L%d 22Q35 18 32 12Q29 19 %d 23Z", left-1, left-4, right+4, right+1, right-2, left+2), hairColor, .8, ink)
		for i := 0; i < 5; i++ {
			p(f("M31 %dq-7 1-11 10M33 %dq7 1 11 10", 12+i, 12+i), "none", .32, hairLine)
		}
		p(f("M%d 27Q%d 40 %d 55L%d 57Q%d 38 %d 23M%d 27Q%d 40 %d 55L%d 57Q%d 38 %d 23", left, left-2, left-3, left-6, left-4, left-1, right, right+2, right+3, right+6, right+4, right+1), hairColor, .7, ink)
		for i := 0; i < 3; i++ {
			p(f("M%g 29q-1 13-3 24M%g 29q1 13 3 24", float64(left-1)-float64(i)*.7, float64(right+1)+float64(i)*.7), "none", .28, hairLine)
		}
	case 4: // Pinned center part over braided lengths.
		p(f("M%d 27Q%d 9 32 9Q%d 9 %d 27L%d 20Q35 15 32 11Q29 16 %d 21Z", left-1, left-3, right+3, right+1, right-2, left+2), hairColor, .7, ink)
		for i := 0; i < 4; i++ {
			p(f("M31 %dq-5 1-10 7M33 %dq5 1 10 7", 12+i, 12+i), "none", .3, hairLine)
		}
	case 5: // Tight curls, drawn as a coherent scalloped silhouette.
		p(f("M%d 28Q%d 24 %d 21Q%d 16 %d 14Q%d 8 27 10Q31 4 35 9Q41 6 43 12Q%d 11 %d 18Q%d 23 %d 28L%d 21Q34 17 28 15L%d 24Z", left, left-5, left-3, left-7, left-2, left-2, right+6, right+3, right+7, right, right-2, left+2), hairColor, .8, ink)
		for i := 0; i < 11; i++ {
			x := left + next(width*2)
			y := 11 + next(8)
			p(f("M%d %dq-2-3 1-3t2 3", x, y), "none", .32, hairLine)
		}
	case 6: // Bald crown and sparse combed side hair.
		p(f("M%d 33Q%d 19 %d 15L%d 26L%d 34M%d 33Q%d 19 %d 15L%d 26L%d 34", left, left-4, left+3, left+2, left+1, right, right+4, right-3, right-2, right-1), hairColor, .7, ink)
		p("M28 14q3-4 6-1M27 16q4-3 9-1", "none", .4, ink)
	case 7: // Short, soft waves, with a higher forehead.
		p(f("M%d 27Q%d 15 %d 12Q25 7 30 10Q35 5 40 10Q%d 9 %d 21L%d 28L%d 20Q36 16 30 14Q26 20 %d 21Z", left, left-5, left, left+width+16, right+2, right, right-2, left+2), hairColor, .8, ink)
		for i := 0; i < 6; i++ {
			p(f("M%d 17q-2-3 2-5", left+3+i*3), "none", .35, hairLine)
		}
	case 8: // Drawn-back hair with a side bun.
		p(f("M%d 28Q%d 11 28 9Q42 5 %d 21L%d 29L%d 23Q36 17 29 13Q24 21 %d 25Z", left, left-5, right+2, right, right-2, left+1), hairColor, .8, ink)
		for i := 0; i < 6; i++ {
			p(f("M%d %dQ30 %d %d %d", left+1, 14+i, 7+i, right, 18+i), "none", .3, hairLine)
		}
	}
	switch hat {
	case 0: // Narrow-brim felt hat.
		p(f("M%d 16L%d 5Q32 8 %d 5L%d 17Z", left-2, left+2, right-2, right+2), cloth, .8, ink)
		p(f("M%d 15L%d 16L%d 19L%d 18Z", left-1, right+1, right+2, left-2), ink, .5, ink)
		p(f("M%d 19Q32 15 %d 20Q32 24 %d 19Z", left-7, right+7, left-7), cloth, .8, ink)
		for i := 0; i < 6; i++ {
			p(f("M%d 8l-1 6", left+4+i*2), "none", .3, paper)
		}
	case 1: // Soft beret.
		p(f("M%d 18Q%d 9 27 7Q42 3 %d 14L%d 20Q31 17 %d 21Z", left-2, left-5, right+5, right, left), cloth, .8, ink)
		p(f("M%d 19Q32 15 %d 18", left, right), "none", .7, paper)
		for i := 0; i < 5; i++ {
			p(f("M%d 10q-2 3-2 5", 26+i*3), "none", .32, paper)
		}
	case 2: // Rounded cloche, descending gently over the forehead.
		p(f("M%d 24Q%d 5 32 5Q%d 5 %d 24Q32 19 %d 24Z", left-3, left-3, right+3, right+3, left-3), cloth, .8, ink)
		p(f("M%d 20Q32 15 %d 20", left-2, right+2), "none", 1.6, ink)
		for i := 0; i < 7; i++ {
			p(f("M%d 9q-2 4-1 8", left+3+i*3), "none", .3, paper)
		}
	case 3: // Flat cap with a short peak.
		p(f("M%d 18Q%d 7 32 8Q%d 8 %d 20L%d 22Q31 17 %d 21Z", left-3, left-2, right+4, right+2, right+6, left-3), cloth, .85, ink)
		p(f("M%d 18Q32 14 %d 19M32 9q-4 3-5 7", left-1, right+1), "none", .45, paper)
	}
	pen.WriteString("</svg>")
	return pen.String()
}
