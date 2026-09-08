package avatar

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"image/png"
	"io"
	"strings"
	"testing"
)

func TestCompanionArtwork(t *testing.T) {
	for _, animal := range []string{"dog", "cat"} {
		t.Run(animal, func(t *testing.T) {
			seen := map[string]bool{}
			for i := 0; i < 100; i++ {
				seed := fmt.Sprintf("companion-%02d", i)
				art, err := generateCompanion(context.Background(), seed, animal)
				if err != nil {
					t.Fatal(err)
				}
				again, err := generateCompanion(context.Background(), seed, animal)
				if err != nil || !bytes.Equal(art.Data, again.Data) {
					t.Fatalf("unstable %s %s: %v", animal, seed, err)
				}
				if art.Width != 128 || art.Height != 128 || art.MediaType != "image/svg+xml" {
					t.Fatalf("bad native dimensions: %+v", art)
				}
				seen[string(art.Data)] = true
				decoder := xml.NewDecoder(bytes.NewReader(art.Data))
				for {
					_, err := decoder.Token()
					if err == io.EOF {
						break
					}
					if err != nil {
						t.Fatal(err)
					}
				}
			}
			if len(seen) < 95 {
				t.Fatalf("too few distinct recipes: %d of 100", len(seen))
			}
			for _, seed := range []string{"", "🦊你好", `<script>evil()</script><svg onload="attack()">`} {
				art, err := generateCompanion(context.Background(), seed, animal)
				if err != nil {
					t.Fatal(err)
				}
				for _, token := range []string{"script", "onload", "evil", "NaN", "%!", "href=", "<image", "<text"} {
					if strings.Contains(string(art.Data), token) {
						t.Fatalf("unsafe or invalid SVG token %q", token)
					}
				}
			}
			for i := 0; i < 24; i++ {
				art, err := generateCompanion(context.Background(), fmt.Sprintf("companion-%02d", i), animal)
				if err != nil {
					t.Fatal(err)
				}
				options := Options{Width: 64, Height: 64, Circle: i%2 == 0}
				rendered, err := (PNGExporter{}).Export(context.Background(), art, options)
				if err != nil {
					t.Fatalf("%s %d: %v", animal, i, err)
				}
				img, err := png.Decode(bytes.NewReader(rendered))
				if err != nil || img.Bounds().Dx() != 64 || img.Bounds().Dy() != 64 {
					t.Fatalf("bad png: %v", err)
				}
				_, _, _, center := img.At(32, 32).RGBA()
				_, _, _, corner := img.At(0, 0).RGBA()
				if center != 65535 || options.Circle && corner != 0 {
					t.Fatal("invalid portrait opacity or circle mask")
				}
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := (Companions{}).Generate(ctx, "seed"); !errors.Is(err, context.Canceled) {
		t.Fatalf("want canceled, got %v", err)
	}
	dog, _ := generateCompanion(context.Background(), "same-seed", "dog")
	cat, _ := generateCompanion(context.Background(), "same-seed", "cat")
	if bytes.Equal(dog.Data, cat.Data) {
		t.Fatal("species must have distinct artwork")
	}
}
