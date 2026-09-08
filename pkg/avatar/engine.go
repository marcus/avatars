package avatar

import (
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"sync"
)

// Artwork is the generator's native output, independent of export format.
// Data must be self-contained; adapters must not resolve external resources.
type Artwork struct {
	Data          []byte
	MediaType     string
	Width, Height int
}

type Style struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Generator implementations must support concurrent calls.
type Generator interface {
	Style() Style
	Generate(context.Context, string) (Artwork, error)
}

// Exporter implementations must support concurrent calls and reject unsupported
// Artwork media types. Options passed by Engine are normalized.
type Exporter interface {
	Format() string
	Export(context.Context, Artwork, Options) ([]byte, error)
}

// Options controls output dimensions. Zero dimensions preserve the native aspect
// ratio. Circle defaults to a square and center-crops before applying its mask.
type Options struct {
	Width  int  `json:"width,omitempty"`
	Height int  `json:"height,omitempty"`
	Circle bool `json:"circle,omitempty"`
}

const MaxDimension = 2048

var (
	ErrInvalidOptions     = errors.New("invalid output options")
	ErrUnknownStyle       = errors.New("unknown avatar style")
	ErrUnknownFormat      = errors.New("unknown export format")
	ErrUnsupportedArtwork = errors.New("unsupported artwork")
)

// Normalize resolves missing dimensions using the native artwork dimensions.
// Noncircular images fit within the requested rectangle without distortion.
func (o Options) Normalize(nativeWidth, nativeHeight int) (Options, error) {
	if nativeWidth <= 0 || nativeHeight <= 0 {
		return o, fmt.Errorf("%w: invalid native dimensions", ErrUnsupportedArtwork)
	}
	if o.Width < 0 || o.Height < 0 {
		return o, fmt.Errorf("%w: dimensions must be positive", ErrInvalidOptions)
	}
	if o.Circle {
		if o.Width == 0 && o.Height == 0 {
			o.Width = max(nativeWidth, nativeHeight)
			o.Height = o.Width
		} else if o.Width == 0 {
			o.Width = o.Height
		} else if o.Height == 0 {
			o.Height = o.Width
		}
	} else {
		if o.Width == 0 && o.Height == 0 {
			o.Width, o.Height = nativeWidth, nativeHeight
		} else if o.Width == 0 {
			o.Width = max(1, int(math.Round(float64(o.Height)*float64(nativeWidth)/float64(nativeHeight))))
		} else if o.Height == 0 {
			o.Height = max(1, int(math.Round(float64(o.Width)*float64(nativeHeight)/float64(nativeWidth))))
		}
	}
	if o.Width < 1 || o.Height < 1 || o.Width > MaxDimension || o.Height > MaxDimension {
		return o, fmt.Errorf("%w: dimensions must be between 1 and %d", ErrInvalidOptions, MaxDimension)
	}
	return o, nil
}

// Engine connects style generators to independent format exporters.
// Its zero value is an empty registry; New installs the built-in adapters.
type Engine struct {
	mu         sync.RWMutex
	generators map[string]Generator
	exporters  map[string]Exporter
}

func New() *Engine {
	e := &Engine{}
	_ = e.RegisterGenerator(Gorey{})
	_ = e.RegisterExporter(SVGExporter{})
	_ = e.RegisterExporter(PNGExporter{})
	return e
}

var adapterID = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// RegisterGenerator adds a style without replacing an existing registration.
func (e *Engine) RegisterGenerator(g Generator) error {
	if g == nil {
		return errors.New("generator is required")
	}
	id := g.Style().ID
	if !adapterID.MatchString(id) {
		return fmt.Errorf("invalid style ID %q", id)
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.generators == nil {
		e.generators = make(map[string]Generator)
	}
	if _, ok := e.generators[id]; ok {
		return fmt.Errorf("style %q already registered", id)
	}
	e.generators[id] = g
	return nil
}

// RegisterExporter adds an output format without replacing an existing one.
func (e *Engine) RegisterExporter(x Exporter) error {
	if x == nil {
		return errors.New("exporter is required")
	}
	id := x.Format()
	if !adapterID.MatchString(id) {
		return fmt.Errorf("invalid format ID %q", id)
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.exporters == nil {
		e.exporters = make(map[string]Exporter)
	}
	if _, ok := e.exporters[id]; ok {
		return fmt.Errorf("format %q already registered", id)
	}
	e.exporters[id] = x
	return nil
}

func (e *Engine) Styles() []Style {
	e.mu.RLock()
	defer e.mu.RUnlock()
	styles := make([]Style, 0, len(e.generators))
	for _, g := range e.generators {
		styles = append(styles, g.Style())
	}
	sort.Slice(styles, func(i, j int) bool { return styles[i].ID < styles[j].ID })
	return styles
}

func (e *Engine) Formats() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	formats := make([]string, 0, len(e.exporters))
	for f := range e.exporters {
		formats = append(formats, f)
	}
	sort.Strings(formats)
	return formats
}

func (e *Engine) Render(ctx context.Context, style, seed, format string, options Options) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if style == "" {
		style = "gorey"
	}
	if format == "" {
		format = "svg"
	}
	e.mu.RLock()
	g, gok := e.generators[style]
	x, xok := e.exporters[format]
	e.mu.RUnlock()
	if !gok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownStyle, style)
	}
	if !xok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownFormat, format)
	}
	// Reject excessive requested dimensions before asking a potentially costly generator.
	if options.Width < 0 || options.Height < 0 || options.Width > MaxDimension || options.Height > MaxDimension {
		return nil, fmt.Errorf("%w: dimensions must be between 1 and %d", ErrInvalidOptions, MaxDimension)
	}
	artwork, err := g.Generate(ctx, seed)
	if err != nil {
		return nil, err
	}
	options, err = options.Normalize(artwork.Width, artwork.Height)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return x.Export(ctx, artwork, options)
}
