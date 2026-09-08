// Package avatar provides deterministic portrait generation and format adapters.
package avatar

import (
	"context"
	"fmt"
	"strings"
)

// Gorey generates the original pen-and-ink portrait style.
type Gorey struct{}

func (Gorey) Style() Style {
	return Style{NativeWidth: 64, NativeHeight: 72, ID: "gorey", Name: "Gorey", Description: "Engraved pen-and-ink portraits on warm paper."}
}

func (Gorey) Generate(ctx context.Context, seed string) (Artwork, error) {
	if err := ctx.Err(); err != nil {
		return Artwork{}, err
	}
	return Artwork{Data: []byte(goreySVG(seed)), MediaType: "image/svg+xml", Width: 64, Height: 72}, nil
}

// goreySVG preserves the reference's arithmetic, sampling order, and markup.
// FNV-1a intentionally hashes Unicode code points, not UTF-8 bytes.
func goreySVG(seed string) string {
	hash := uint32(2166136261)
	for _, char := range seed {
		hash = (hash ^ uint32(char)) * 16777619
	}
	initial := hash
	next := func(max int) int {
		hash ^= hash << 13
		hash ^= hash >> 17
		hash ^= hash << 5
		return int(hash % uint32(max))
	}
	ink := "#292820"
	paper := []string{"#d6d0bb", "#ded7c5", "#cfcbb8", "#d8d0be"}[initial%4]
	wardrobe := (initial ^ uint32(0x6d2b79f5)) * uint32(0x45d9f3b)
	outfit := wardrobe % 6
	backdrop := []string{"#d6d0bb", "#c9b68f", "#c5a296", "#b9a69b", "#b8b69b", "#c3abab"}[(wardrobe>>8)%6]
	cloth := []string{"#3c3a30", "#484234", "#454638", "#51423a"}[(wardrobe>>16)%4]
	faceWidth, chin, eyeY, nose := 11+next(4), 44+next(6), 27+next(3), 34+next(4)
	hair, accessory, coat := next(5), next(6), next(3)
	left, right := 32-faceWidth, 32+faceWidth
	strokes := make([]string, 0, 96)
	path := func(d string, args ...any) string {
		fill, width, color := any("none"), any(0.8), any(ink)
		if len(args) > 0 {
			fill = args[0]
		}
		if len(args) > 1 {
			width = args[1]
		}
		if len(args) > 2 {
			color = args[2]
		}
		return fmt.Sprintf(`<path d="%s" fill="%v" stroke="%v" stroke-width="%v" stroke-linecap="round" stroke-linejoin="round"/>`, d, fill, color, width)
	}
	hatchSlope := -2
	if coat == 1 {
		hatchSlope = 4
	}
	hatchInk := paper
	if outfit == 3 {
		hatchInk = "#8b8471"
	}
	// Uneven parallel strokes give the backdrop and cloth a dry engraved finish.
	for i := 0; i < 18; i++ {
		x, y := 4+next(57), 4+next(62)
		strokes = append(strokes, path(fmt.Sprintf(`M%v %vl%v %v`, x, y, 1+next(3), next(3)-1), `none`, .25, `#9b9584`))
	}
	// Consume the original hatch samples in the original order so head generation never changes.
	hatchLengths := make([]int, 16)
	for i := range hatchLengths {
		hatchLengths[i] = 7 + next(6)
	}
	shirtFill := cloth
	if outfit == 3 {
		shirtFill = paper
	} else if outfit == 0 {
		shirtFill = ink
	}
	strokes = append(strokes, path(`M5 72L9 58Q16 51 26 49L38 49Q49 53 55 59L61 72Z`, shirtFill))
	for i := 0; i < 16; i++ {
		x := 9 + i*3
		strokes = append(strokes, path(fmt.Sprintf(`M%v 71l%v -%v`, x, hatchSlope, hatchLengths[i]), `none`, .45, hatchInk))
	}
	strokes = append(strokes, path(`M27 43L26 51L32 58L39 50L37 43`, paper))
	if outfit == 0 {
		// The original evening shirt and bow tie remain in the wardrobe.
		strokes = append(strokes, path(`M25 49L18 54L25 65L31 56M39 49L46 54L38 64L33 56`, paper, .9))
		strokes = append(strokes, path(`M29 56l-5 2 5 4 3-3 3 3 5-4-6-2Z`, ink))
		strokes = append(strokes, path(`M32 61v11`, `none`, .6, paper))
	} else if outfit == 1 {
		// A narrow tie and buttoned waistcoat.
		strokes = append(strokes, path(`M26 49L19 53L28 65L32 59L36 65L45 53L38 49L32 54Z`, paper))
		strokes = append(strokes, path(`M30 54h4l-1 5 2 11-3 2-3-2 2-11Z`, ink, .55))
		strokes = append(strokes, path(`M18 54l7 9-4 2 9 7M46 54l-7 9 4 2-9 7`, `none`, .6, paper))
	} else if outfit == 2 {
		// A high ribbed pullover, without a collar or tie.
		strokes = append(strokes, path(`M25 49Q32 52 39 49L40 57Q32 61 24 57Z`, cloth))
		for x := 26; x <= 38; x += 2 {
			strokes = append(strokes, path(fmt.Sprintf(`M%v 51v6`, x), `none`, .4, paper))
		}
		strokes = append(strokes, path(`M16 57q-2 6-1 14M48 57q2 6 1 14`, `none`, .7, paper))
	} else if outfit == 3 {
		// An open ivory shirt with dark braces.
		strokes = append(strokes, path(`M26 49L22 52L27 60L32 54L37 60L42 52L38 49L32 54Z`, paper))
		strokes = append(strokes, path(`M17 54L14 72h4l3-20M43 52l3 20h4l-3-18`, ink, .6))
		strokes = append(strokes, path(`M32 56v16`, `none`, .55))
	} else if outfit == 4 {
		// A round-neck knit with a lightly crosshatched chest.
		strokes = append(strokes, path(`M23 51Q32 64 41 51L44 53Q32 69 20 53Z`, cloth))
		strokes = append(strokes, path(`M23 54q9 12 18 0`, `none`, .65, paper))
		for y := 63; y <= 71; y += 3 {
			strokes = append(strokes, path(fmt.Sprintf(`M13 %vh38`, y), `none`, .35, paper))
		}
	} else {
		// A soft shawl collar and double-breasted cardigan.
		strokes = append(strokes, path(`M26 49Q18 50 19 57L31 69L30 58Z`, paper))
		strokes = append(strokes, path(`M38 49Q46 50 45 57L33 69L34 58Z`, paper))
		strokes = append(strokes, path(`M31 64v8M33 64v8`, `none`, .6, paper))
	}
	if outfit == 0 || outfit == 3 || outfit == 5 {
		buttonInk := paper
		if outfit == 3 {
			buttonInk = ink
		}
		strokes = append(strokes, fmt.Sprintf(`<circle cx="35" cy="66" r=".8" fill="%v"/><circle cx="35" cy="71" r=".8" fill="%v"/>`, buttonInk, buttonInk))
	}
	strokes = append(strokes, path(fmt.Sprintf(`M%v 28Q%v 23 %v 34L%v 38M%v 28Q%v 24 %v 34L%v 38`, left, left-5, left-3, left+1, right, right+5, right+3, right-1), paper))
	strokes = append(strokes, path(fmt.Sprintf(`M%v 24Q%v 10 32 10Q%v 11 %v 25L%v 37Q%v %v 32 %vQ%v %v %v 36Z`, left, left-1, right+2, right, right-1, right-3, chin, chin+1, left+3, chin-1, left+1), paper, 1))
	for i := 0; i < 5; i++ {
		strokes = append(strokes, path(fmt.Sprintf(`M%v %vl2 4`, float64(left+1)+float64(i)*.6, 32+i), `none`, .35))
	}
	for i := 0; i < 4; i++ {
		strokes = append(strokes, path(fmt.Sprintf(`M%v %vl-1 5`, float64(right-3)+float64(i)*.6, 30+i), `none`, .3))
	}
	// Small hooded eyes, angular noses and restrained expressions, rather than emoji faces.
	strokes = append(strokes, path(fmt.Sprintf(`M%v %vl6 -1M%v %vl6 2`, left+3, eyeY-3, right-9, eyeY-4), `none`, 1))
	strokes = append(strokes, path(fmt.Sprintf(`M%v %vq3 -2 6 0M%v %vq3 -2 6 0`, left+3, eyeY, right-9, eyeY)))
	strokes = append(strokes, fmt.Sprintf(`<ellipse cx="%v" cy="%v" rx=".8" ry="1.1" fill="%v"/><ellipse cx="%v" cy="%v" rx=".8" ry="1.1" fill="%v"/>`, left+6, eyeY, ink, right-6, eyeY, ink))
	strokes = append(strokes, path(fmt.Sprintf(`M32 %vL29 %vQ31 %v 35 %vM29 %vq4 -1 %v 0M31 %vh3`, eyeY-2, nose, nose+2, nose, nose+5, 5+next(3), nose+7), `none`, .7))
	if accessory == 0 {
		strokes = append(strokes, fmt.Sprintf(`<g fill="none" stroke="%v" stroke-width=".7"><circle cx="%v" cy="%v" r="4.2"/><circle cx="%v" cy="%v" r="4.2"/></g>`, ink, left+6, eyeY, right-6, eyeY))
		strokes = append(strokes, path(fmt.Sprintf(`M%v %vh%v`, left+10, eyeY, 2*faceWidth-20)))
	} else if accessory == 1 {
		strokes = append(strokes, path(fmt.Sprintf(`M32 %vq-4 -3 -9 3 6 1 9-1 3 3 9 0-5-5-9-2Z`, nose+2), ink, .5))
	} else if accessory == 2 {
		strokes = append(strokes, path(fmt.Sprintf(`M27 %vl5 7 5-7-5 2Z`, chin-2), ink, .5))
	} else if accessory == 3 {
		strokes = append(strokes, fmt.Sprintf(`<circle cx="%v" cy="%v" r="4.5" fill="none" stroke="%v" stroke-width=".8"/>`, right-6, eyeY, ink))
		strokes = append(strokes, path(fmt.Sprintf(`M%v %vq9 16 4 25`, right-2, eyeY+2), `none`, .5))
	}
	if hair == 0 {
		strokes = append(strokes, path(fmt.Sprintf(`M%v 20L%v 4L%v 5L%v 20Z`, left-1, left+1, right-1, right+1), ink))
		strokes = append(strokes, path(fmt.Sprintf(`M%v 21Q32 17 %v 22Q31 25 %v 21Z`, left-6, right+6, left-6), ink))
		for i := 0; i < 5; i++ {
			strokes = append(strokes, path(fmt.Sprintf(`M%v 7l-1 10`, left+3+i*2), `none`, .35, paper))
		}
		strokes = append(strokes, path(fmt.Sprintf(`M%v 17L%v 18`, left+1, right), `none`, 1.3, paper))
	} else if hair == 1 {
		strokes = append(strokes, path(fmt.Sprintf(`M%v 26Q%v 8 26 9Q34 3 %v 15L%v 26L%v 18Q31 16 27 13L%v 24Z`, left-1, left-8, right+2, right+1, right-3, left+4), ink))
		for i := 0; i < 7; i++ {
			strokes = append(strokes, path(fmt.Sprintf(`M%v 20q-3 -6 2 -9`, left+3+i*2), `none`, .4, paper))
		}
	} else if hair == 2 {
		strokes = append(strokes, path(fmt.Sprintf(`M%v 30L%v 25L%v 21L%v 17L%v 16L%v 9L%v 12L%v 6L32 11L38 7L41 12L%v 11L%v 18L%v 23L%v 30L%v 19Q31 21 25 15L%v 25Z`, left, left-3, left-1, left-5, left+1, left, left+5, left+9, right+4, right+1, right+5, right, right-2, left+2), ink, .6))
	} else if hair == 3 {
		strokes = append(strokes, path(fmt.Sprintf(`M%v 29Q%v 10 28 9Q42 5 %v 22L%v 34L%v 24Q37 16 29 13Q23 20 %v 25Z`, left-1, left-7, right+3, right, right-2, left+1), ink))
		for i := 0; i < 6; i++ {
			strokes = append(strokes, path(fmt.Sprintf(`M%v %vQ30 %v %v %v`, left+1, 13+i, 6+i, right, 17+i), `none`, .35, paper))
		}
	} else {
		strokes = append(strokes, path(fmt.Sprintf(`M%v 33Q%v 18 %v 14L%v 28L%v 35M%v 33Q%v 19 %v 14L%v 28L%v 35`, left-1, left-6, left+4, left+2, left+1, right, right+6, right-3, right-2, right-1), ink))
		strokes = append(strokes, path(`M29 13q3 -4 5 0M27 15q3 -3 8 0`, `none`, .5))
	}
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 72" fill="none" aria-hidden="true"><path fill="%s" d="M0 0h64v72H0z"/>%s</svg>`, backdrop, strings.Join(strokes, ""))
}
