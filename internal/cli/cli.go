// Package cli provides the human and agent command-line surfaces.
package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/marcus/avatars/internal/buildinfo"
	"github.com/marcus/avatars/internal/discovery"
	"github.com/marcus/avatars/internal/library"
	"github.com/marcus/avatars/internal/store"
	"github.com/marcus/avatars/pkg/avatar"
)

type app struct {
	out, errOut      io.Writer
	json             bool
	dataDir, baseURL string
	remote           bool
	engine           *avatar.Engine
	service          *library.Service
}

func Run(ctx context.Context, args []string, out, errOut io.Writer) int {
	a := &app{out: out, errOut: errOut, engine: avatar.New(), dataDir: os.Getenv("AVATARS_DATA_DIR"), baseURL: os.Getenv("AVATARS_URL")}
	remaining, err := a.globals(args)
	if err == nil {
		err = a.run(ctx, remaining)
	}
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	if err == nil {
		return 0
	}
	d := &library.Error{Code: "internal", Message: err.Error()}
	var domain *library.Error
	if errors.As(err, &domain) {
		d.Code = domain.Code
	}
	if a.json {
		_ = json.NewEncoder(errOut).Encode(map[string]any{"error": d})
	} else {
		fmt.Fprintln(errOut, "avatars:", d.Message)
	}
	if d.Code == "invalid_request" {
		return 2
	}
	return 1
}
func bad(message string) error { return &library.Error{Code: "invalid_request", Message: message} }
func (a *app) globals(args []string) ([]string, error) {
	var remaining []string
	for i := 0; i < len(args); i++ {
		v := args[i]
		if v == "--" {
			remaining = append(remaining, args[i:]...)
			break
		}
		key, value, has := strings.Cut(v, "=")
		switch key {
		case "--json":
			a.json = true
			if has {
				b, e := strconv.ParseBool(value)
				if e != nil {
					return nil, bad("--json must be true or false")
				}
				a.json = b
			}
		case "--data-dir", "--url":
			if !has {
				i++
				if i == len(args) {
					return nil, bad(key + " requires a value")
				}
				value = args[i]
			}
			if value == "" {
				return nil, bad(key + " cannot be empty")
			}
			if key == "--data-dir" {
				a.dataDir = value
			} else {
				a.baseURL = value
			}
		default:
			remaining = append(remaining, v)
			// A global-looking string may be the value of a command flag.
			if !has && takesValue(key) && i+1 < len(args) {
				i++
				remaining = append(remaining, args[i])
			}
		}
	}
	a.remote = a.baseURL != ""
	if a.remote && a.dataDir != "" {
		return nil, bad("--data-dir / AVATARS_DATA_DIR cannot be combined with --url / AVATARS_URL; the server owns its library")
	}
	if a.remote {
		u, e := url.Parse(a.baseURL)
		if e != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.RawQuery != "" || u.Fragment != "" || u.User != nil || (u.Path != "" && u.Path != "/") {
			return nil, bad("--url must be an HTTP(S) origin, such as http://127.0.0.1:7447")
		}
		a.baseURL = strings.TrimRight(a.baseURL, "/")
	} else {
		a.baseURL = discovery.DefaultURL
	}
	return remaining, nil
}
func (a *app) run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return a.help("")
	}
	command := args[0]
	tail := args[1:]
	if command == "help" {
		if len(tail) > 1 {
			return bad("help accepts one command")
		}
		name := ""
		if len(tail) == 1 {
			name = tail[0]
		}
		return a.help(name)
	}
	if command == "--help" || command == "-h" {
		return a.help("")
	}
	if wantsHelp(tail) {
		return a.help(command)
	}
	switch command {
	case "--version", "-v", "version":
		if len(tail) != 0 {
			return bad("version takes no arguments")
		}
		if a.json {
			return a.emit(map[string]string{"version": buildinfo.Version, "commit": buildinfo.Commit})
		}
		fmt.Fprintln(a.out, buildinfo.String("avatars"))
		return nil
	case "capabilities":
		if len(tail) != 0 {
			return bad("capabilities takes no arguments")
		}
		return a.emit(discovery.Capabilities())
	case "instructions":
		if len(tail) != 0 {
			return bad("instructions takes no arguments")
		}
		if a.json {
			return a.emit(map[string]string{"instructions": discovery.Instructions})
		}
		fmt.Fprint(a.out, discovery.Instructions)
		return nil
	case "styles":
		if len(tail) != 0 {
			return bad("styles takes no arguments")
		}
		return a.styles(ctx)
	case "generate":
		return a.generate(ctx, tail)
	case "render", "export":
		return a.image(ctx, command, tail)
	case "list":
		if len(tail) != 0 {
			return bad("list takes no arguments")
		}
		return a.list(ctx)
	case "show":
		if len(tail) != 1 {
			return bad("usage: avatars show ID")
		}
		return a.show(ctx, tail[0])
	case "serve":
		return a.serve(ctx, tail)
	case "service":
		return a.serviceCommand(ctx, tail)
	default:
		return bad("unknown command " + strconv.Quote(command) + "; run avatars help")
	}
}
func (a *app) help(name string) error {
	if name != "" {
		for _, op := range discovery.Operations {
			if op.Command == name {
				fmt.Fprintf(a.out, "%s\n\nUsage: avatars %s\n\n", op.Summary, op.Usage)
				a.commonHelp()
				return nil
			}
		}
		return bad("unknown command: " + name)
	}
	fmt.Fprint(a.out, "avatars - a local avatar library and creative studio\n\nUsage: avatars COMMAND [flags]\n\n")
	for _, op := range discovery.Operations {
		fmt.Fprintf(a.out, "  %-14s %s\n", op.Command, op.Summary)
	}
	fmt.Fprintln(a.out)
	a.commonHelp()
	fmt.Fprint(a.out, "\nStart: avatars serve --open\nCreate: avatars generate --count 12 --json\nSee avatars instructions for the complete agent workflow.\n")
	return nil
}
func (a *app) commonHelp() {
	fmt.Fprint(a.out, "Global flags: --json, --data-dir PATH, --url URL, -h/--help\nDefaults: SVG; native size varies by style; random saved generation; style gorey.\n--size N is square; --size WxH sets both dimensions (1..2048).\n--out FILE creates a new file; --out - writes image bytes to stdout.\n")
}
func (a *app) emit(v any) error {
	e := json.NewEncoder(a.out)
	e.SetIndent("", "  ")
	return e.Encode(v)
}
func (a *app) local() error {
	if a.service != nil {
		return nil
	}
	if a.dataDir == "" {
		base := os.Getenv("XDG_DATA_HOME")
		if base == "" {
			home, e := os.UserHomeDir()
			if e != nil {
				return e
			}
			base = filepath.Join(home, ".local", "share")
		}
		a.dataDir = filepath.Join(base, "avatars")
	}
	s, e := store.New(filepath.Join(a.dataDir, "collections.jsonl"))
	if e != nil {
		return e
	}
	a.service = library.New(a.engine, s)
	return nil
}
func (a *app) flags(name string) *flag.FlagSet {
	f := flag.NewFlagSet(name, flag.ContinueOnError)
	f.SetOutput(io.Discard)
	return f
}

// parse allows flags before or after positionals, while preserving -- literally.
func parse(f *flag.FlagSet, args []string) error {
	var flags, positionals []string
	for i := 0; i < len(args); i++ {
		v := args[i]
		if v == "--" {
			positionals = append(positionals, args[i+1:]...)
			break
		}
		if strings.HasPrefix(v, "-") && v != "-" {
			flags = append(flags, v)
			name, _, eq := strings.Cut(strings.TrimLeft(v, "-"), "=")
			def := f.Lookup(name)
			if def == nil {
				return bad("unknown flag: " + v)
			}
			b, ok := def.Value.(interface{ IsBoolFlag() bool })
			if !eq && (!ok || !b.IsBoolFlag()) {
				i++
				if i == len(args) {
					return bad(v + " requires a value")
				}
				flags = append(flags, args[i])
			}
		} else {
			positionals = append(positionals, v)
		}
	}
	if e := f.Parse(append(append(flags, "--"), positionals...)); e != nil {
		return bad(e.Error())
	}
	return nil
}

func takesValue(name string) bool {
	switch name {
	case "--seed", "-s", "--name", "--count", "--style", "--color", "--animal", "--out", "-o", "--format", "-f", "--size", "--listen", "--public-url":
		return true
	}
	return false
}

func wantsHelp(args []string) bool {
	for i := 0; i < len(args); i++ {
		if args[i] == "--" {
			return false
		}
		if args[i] == "--help" || args[i] == "-h" {
			return true
		}
		if takesValue(args[i]) {
			i++
		}
	}
	return false
}
