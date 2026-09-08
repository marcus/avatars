// Command avatars is the CLI for generating deterministic pen-and-ink avatars and profile icons.
package main

import (
	"fmt"
	"os"

	"github.com/marcus/avatars/internal/buildinfo"
)

const usage = `avatars - deterministic pen-and-ink avatar and profile icon generator

Usage:
  avatars <command> [flags]

Commands:
  generate    Generate an avatar for a given seed/identity
  version     Print version information

Flags:
  -h, --help       Show help
  -v, --version    Show version

Use "avatars <command> --help" for more information about a command.
`

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(argv []string) int {
	if len(argv) == 0 {
		fmt.Fprint(os.Stderr, usage)
		return 1
	}

	cmd := argv[0]
	switch cmd {
	case "-v", "--version", "version":
		fmt.Println(buildinfo.String("avatars"))
		return 0

	case "-h", "--help", "help":
		fmt.Print(usage)
		return 0

	case "generate":
		fmt.Fprintln(os.Stderr, "avatars: generation engine port in progress (see port/ and docs/plans/active/)")
		return 1

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %q\n\n%s", cmd, usage)
		return 1
	}
}
