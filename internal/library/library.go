// Package library owns the saved-avatar workflow shared by the CLI and HTTP API.
package library

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/marcus/avatars/pkg/avatar"
)

// Avatar stores the recipe for a portrait. The style's deterministic generator
// recreates export bytes; generated image files are not the source of truth.
type Avatar struct {
	ID           string         `json:"id"`
	CollectionID string         `json:"collection_id"`
	Style        string         `json:"style"`
	Seed         string         `json:"seed"`
	Inputs       *avatar.Inputs `json:"inputs,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	URL          string         `json:"url"`
	SVGURL       string         `json:"svg_url"`
	PNGURL       string         `json:"png_url"`
}

type Collection struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Style     string    `json:"style"`
	CreatedAt time.Time `json:"created_at"`
	URL       string    `json:"url"`
	Avatars   []Avatar  `json:"avatars"`
}

// CreateRequest creates one saved collection. Count zero means one avatar. A
// supplied seed is used verbatim for one avatar, or as "seed:index" (zero-based)
// for each avatar in a batch. An empty seed requests fresh cryptographic seeds.
type CreateRequest struct {
	Style  string         `json:"style"`
	Count  int            `json:"count"`
	Name   string         `json:"name"`
	Seed   string         `json:"seed"`
	Inputs *avatar.Inputs `json:"inputs,omitempty"`
}

// Store saves a whole collection atomically. Callers do not depend on its
// storage format, filesystem layout, or locking mechanism.
type Store interface {
	Save(context.Context, Collection) error
	List(context.Context) ([]Collection, error)
}

// Error gives all transports the same small set of failure categories.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	cause   error
}

func (e *Error) Error() string { return e.Message }

func (e *Error) Unwrap() error { return e.cause }

type Service struct {
	engine *avatar.Engine
	store  Store
}

func New(engine *avatar.Engine, store Store) *Service {
	return &Service{engine: engine, store: store}
}

func (s *Service) Create(ctx context.Context, request CreateRequest) (Collection, error) {
	if err := ctx.Err(); err != nil {
		return Collection{}, err
	}
	if request.Count == 0 {
		request.Count = 1
	}
	if request.Style == "" {
		request.Style = "gorey"
	}
	if request.Count < 1 || request.Count > 100 {
		return Collection{}, invalid("count must be between 1 and 100")
	}
	if !utf8.ValidString(request.Name) || utf8.RuneCountInString(request.Name) > 120 {
		return Collection{}, invalid("name must be valid UTF-8 and at most 120 characters")
	}
	if !utf8.ValidString(request.Seed) || len(request.Seed) > 4096 {
		return Collection{}, invalid("seed must be valid UTF-8 and at most 4096 bytes")
	}
	var selected *avatar.Style
	for _, style := range s.engine.Styles() {
		if style.ID == request.Style {
			selected = &style
			break
		}
	}
	if selected == nil {
		return Collection{}, invalid(fmt.Sprintf("unknown style %q", request.Style))
	}
	requestedInputs := avatar.Inputs{}
	if request.Inputs != nil {
		requestedInputs = *request.Inputs
	}
	resolvedInputs, err := s.engine.ResolveInputs(request.Style, requestedInputs)
	if err != nil {
		return Collection{}, invalid(err.Error())
	}
	name := strings.TrimSpace(request.Name)
	if name == "" {
		name = selected.Name + " collection"
	}
	id, err := randomID("col_")
	if err != nil {
		return Collection{}, internal("create collection ID", err)
	}
	now := time.Now().UTC()
	collection := Collection{
		ID: id, Name: name, Style: request.Style, CreatedAt: now,
		URL: "/?collection=" + id, Avatars: make([]Avatar, 0, request.Count),
	}
	for i := 0; i < request.Count; i++ {
		id, err := randomID("av_")
		if err != nil {
			return Collection{}, internal("create avatar ID", err)
		}
		seed := request.Seed
		if seed == "" {
			seed, err = randomID("")
			if err != nil {
				return Collection{}, internal("create avatar seed", err)
			}
		} else if request.Count > 1 {
			seed = fmt.Sprintf("%s:%d", seed, i)
		}
		recipeInputs, err := s.engine.ResolveRecipeInputs(request.Style, seed, resolvedInputs)
		if err != nil {
			return Collection{}, invalid(err.Error())
		}
		collection.Avatars = append(collection.Avatars, Avatar{
			ID: id, CollectionID: collection.ID, Style: request.Style, Seed: seed,
			Inputs:    generationInputs(recipeInputs),
			CreatedAt: now, URL: "/?avatar=" + id,
			SVGURL: "/api/v1/avatars/" + id + ".svg", PNGURL: "/api/v1/avatars/" + id + ".png",
		})
	}
	if err := s.store.Save(ctx, collection); err != nil {
		return Collection{}, internal("save collection", err)
	}
	return collection, nil
}

// List returns newest collections first and reads the store every time, so a
// long-running studio immediately sees collections created by a separate CLI.
func (s *Service) List(ctx context.Context) ([]Collection, error) {
	collections, err := s.store.List(ctx)
	if err != nil {
		return nil, internal("read collections", err)
	}
	if collections == nil {
		collections = []Collection{}
	}
	sort.SliceStable(collections, func(i, j int) bool {
		if collections[i].CreatedAt.Equal(collections[j].CreatedAt) {
			return collections[i].ID > collections[j].ID
		}
		return collections[i].CreatedAt.After(collections[j].CreatedAt)
	})
	return collections, nil
}

func (s *Service) Collection(ctx context.Context, id string) (Collection, error) {
	collections, err := s.List(ctx)
	if err != nil {
		return Collection{}, err
	}
	for _, collection := range collections {
		if collection.ID == id {
			return collection, nil
		}
	}
	return Collection{}, &Error{Code: "not_found", Message: "collection not found"}
}

func (s *Service) Avatar(ctx context.Context, id string) (Avatar, error) {
	collections, err := s.List(ctx)
	if err != nil {
		return Avatar{}, err
	}
	for _, collection := range collections {
		for _, item := range collection.Avatars {
			if item.ID == id {
				return item, nil
			}
		}
	}
	return Avatar{}, &Error{Code: "not_found", Message: "avatar not found"}
}

func (s *Service) Render(ctx context.Context, id, format string, options avatar.Options) ([]byte, error) {
	item, err := s.Avatar(ctx, id)
	if err != nil {
		return nil, err
	}
	inputs := avatar.Inputs{}
	if item.Inputs != nil {
		inputs = *item.Inputs
	}
	data, err := s.engine.RenderWithInputs(ctx, item.Style, item.Seed, format, inputs, options)
	if err != nil {
		if errors.Is(err, avatar.ErrInvalidOptions) || errors.Is(err, avatar.ErrInvalidInputs) || errors.Is(err, avatar.ErrUnknownFormat) {
			return nil, invalid(err.Error())
		}
		return nil, internal("render avatar", err)
	}
	return data, nil
}

func generationInputs(inputs avatar.Inputs) *avatar.Inputs {
	if inputs.Empty() {
		return nil
	}
	copy := inputs
	return &copy
}

func invalid(message string) error {
	return &Error{Code: "invalid_request", Message: message}
}

func internal(action string, err error) error {
	return &Error{Code: "internal", Message: fmt.Sprintf("%s: %v", action, err), cause: err}
}

func randomID(prefix string) (string, error) {
	var data [16]byte
	if _, err := rand.Read(data[:]); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(data[:]), nil
}
