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
	NativeWidth  int          `json:"native_width,omitempty"`
	NativeHeight int          `json:"native_height,omitempty"`
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	Inputs       *StyleInputs `json:"inputs,omitempty"`
}

// Inputs identifies the appearance choices that are part of a generated
// avatar's recipe. Export dimensions and cropping remain in Options.
type Inputs struct {
	Color  string `json:"color,omitempty"`
	Animal string `json:"animal,omitempty"`
}

func (i Inputs) Empty() bool { return i.Color == "" && i.Animal == "" }

// StyleInputs publishes the small set of generation inputs supported by a
// style. The descriptors are shared by validation, clients, and renderers.
type StyleInputs struct {
	Color  *ColorInput  `json:"color,omitempty"`
	Animal *AnimalInput `json:"animal,omitempty"`
}

type AnimalInput struct {
	// MixedValue names the request choice to suggest for a collection with varied values.
	MixedValue string         `json:"mixed_value,omitempty"`
	Default    string         `json:"default"`
	Values     []AnimalChoice `json:"values"`
}

type AnimalChoice struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type ColorInput struct {
	MixedValue string        `json:"mixed_value,omitempty"`
	Default    string        `json:"default"`
	Values     []ColorChoice `json:"values"`
}

type ColorChoice struct {
	Value    string `json:"value"`
	Label    string `json:"label"`
	Swatch   string `json:"swatch,omitempty"`
	EyeColor string `json:"-"`
}

// Generator implementations must support concurrent calls.
type Generator interface {
	Style() Style
	Generate(context.Context, string) (Artwork, error)
}

// InputGenerator is the optional extension for generators with appearance
// inputs. Generate remains the stable input-free path and uses style defaults.
type InputGenerator interface {
	Generator
	GenerateWithInputs(context.Context, string, Inputs) (Artwork, error)
}

// RecipeInputResolver optionally turns seed-dependent requests into concrete
// saved appearance inputs. Implementations must be deterministic and concurrent-safe.
type RecipeInputResolver interface {
	ResolveRecipeInputs(seed string, inputs Inputs) (Inputs, error)
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
	ErrInvalidInputs      = errors.New("invalid generation inputs")
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
	_ = e.RegisterGenerator(GoreyExpanded{})
	_ = e.RegisterGenerator(Picasso{})
	_ = e.RegisterGenerator(Pebble{})
	_ = e.RegisterGenerator(Companions{})
	_ = e.RegisterGenerator(FieldBirds{})
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
	if g.Style().Inputs != nil {
		if _, ok := g.(InputGenerator); !ok {
			return fmt.Errorf("style %q publishes inputs but does not implement InputGenerator", id)
		}
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

// ResolveInputs applies a style's defaults and validates supported values.
// It has no style-specific branches: each generator publishes its own typed
// descriptor through Style.
func (e *Engine) ResolveInputs(style string, inputs Inputs) (Inputs, error) {
	if style == "" {
		style = "gorey"
	}
	e.mu.RLock()
	g, ok := e.generators[style]
	e.mu.RUnlock()
	if !ok {
		return Inputs{}, fmt.Errorf("%w: %q", ErrUnknownStyle, style)
	}
	return resolveStyleInputs(g.Style(), inputs)
}

// ResolveRecipeInputs resolves appearance for one avatar's final seed. Unlike
// ResolveInputs, it replaces request choices such as Random with concrete values.
func (e *Engine) ResolveRecipeInputs(style, seed string, inputs Inputs) (Inputs, error) {
	if style == "" {
		style = "gorey"
	}
	e.mu.RLock()
	g, ok := e.generators[style]
	e.mu.RUnlock()
	if !ok {
		return Inputs{}, fmt.Errorf("%w: %q", ErrUnknownStyle, style)
	}
	if resolver, ok := g.(RecipeInputResolver); ok {
		return resolver.ResolveRecipeInputs(seed, inputs)
	}
	return resolveStyleInputs(g.Style(), inputs)
}

// resolveStyleInputs also serves direct generator calls, keeping defaults and
// unsupported-input refusal the same with or without an Engine.
func resolveStyleInputs(style Style, inputs Inputs) (Inputs, error) {
	spec := style.Inputs
	if spec == nil {
		spec = &StyleInputs{}
	}
	if spec.Color == nil {
		if inputs.Color != "" {
			return Inputs{}, fmt.Errorf("%w: style %q does not support color", ErrInvalidInputs, style.ID)
		}
	} else {
		if inputs.Color == "" {
			inputs.Color = spec.Color.Default
		}
		valid := false
		for _, choice := range spec.Color.Values {
			if choice.Value == inputs.Color {
				valid = true
				break
			}
		}
		if !valid {
			return Inputs{}, fmt.Errorf("%w: invalid color %q for style %q", ErrInvalidInputs, inputs.Color, style.ID)
		}
	}
	if spec.Animal == nil {
		if inputs.Animal != "" {
			return Inputs{}, fmt.Errorf("%w: style %q does not support animal", ErrInvalidInputs, style.ID)
		}
	} else {
		if inputs.Animal == "" {
			inputs.Animal = spec.Animal.Default
		}
		valid := false
		for _, choice := range spec.Animal.Values {
			if choice.Value == inputs.Animal {
				valid = true
				break
			}
		}
		if !valid {
			return Inputs{}, fmt.Errorf("%w: invalid animal %q for style %q", ErrInvalidInputs, inputs.Animal, style.ID)
		}
	}
	return inputs, nil
}

func (e *Engine) Render(ctx context.Context, style, seed, format string, options Options) ([]byte, error) {
	return e.RenderWithInputs(ctx, style, seed, format, Inputs{}, options)
}

// RenderWithInputs renders a recipe with validated generation inputs.
func (e *Engine) RenderWithInputs(ctx context.Context, style, seed, format string, inputs Inputs, options Options) ([]byte, error) {
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
	resolved, err := e.ResolveRecipeInputs(style, seed, inputs)
	if err != nil {
		return nil, err
	}
	// Reject excessive requested dimensions before asking a potentially costly generator.
	if options.Width < 0 || options.Height < 0 || options.Width > MaxDimension || options.Height > MaxDimension {
		return nil, fmt.Errorf("%w: dimensions must be between 1 and %d", ErrInvalidOptions, MaxDimension)
	}
	var artwork Artwork
	if inputGenerator, ok := g.(InputGenerator); ok {
		artwork, err = inputGenerator.GenerateWithInputs(ctx, seed, resolved)
	} else {
		artwork, err = g.Generate(ctx, seed)
	}
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
