// Package discovery describes the supported agent and HTTP surfaces.
package discovery

const DefaultAddress = "127.0.0.1:7447"
const DefaultURL = "http://" + DefaultAddress

type Operation struct {
	Command string `json:"command"`
	Summary string `json:"summary"`
	Usage   string `json:"usage"`
	Method  string `json:"method,omitempty"`
	Path    string `json:"path,omitempty"`
}

var Operations = []Operation{
	{"generate", "Generate and save a collection of random avatars", "generate [SEED] [--count 12] [--style gorey] [--color COLOR|random] [--animal dog|cat|mixed] [--name TEXT] [--seed TEXT] [--out FILE] [--format svg|png] [--size WxH] [--circle]", "POST", "/api/v1/collections"},
	{"list", "List saved collections, newest first", "list", "GET", "/api/v1/collections"},
	{"show", "Inspect a saved avatar or collection", "show ID", "GET", "/api/v1/collections/{id} or /api/v1/avatars/{id}"},
	{"export", "Export a saved avatar as SVG or PNG", "export AVATAR_ID [--format svg|png] [--size 256x288] [--circle] [--out FILE|-]", "GET", "/api/v1/avatars/{id}.{format}"},
	{"render", "Render a reproducible avatar without saving it", "render [SEED] [--seed TEXT] [--style gorey] [--color COLOR|random] [--animal dog|cat|mixed] [--format svg|png] [--size WxH] [--circle] [--out FILE|-]", "GET", "/api/v1/render"},
	{"styles", "List installed styles and export formats", "styles", "GET", "/api/v1/styles"},
	{"serve", "Run the HTTP API and studio on loopback", "serve [--listen 127.0.0.1:7447] [--public-url HTTPS_ORIGIN] [--open]", "", ""},
	{"service", "Ensure, inspect, or stop the on-demand local service", "service ensure|status|stop [--listen 127.0.0.1:7447] [--public-url HTTPS_ORIGIN] [--timeout 5s] [--no-auto-start]", "", ""},
	{"instructions", "Print operational guidance for agents", "instructions", "GET", "/api/v1/instructions"},
	{"capabilities", "Print machine-readable command and API discovery", "capabilities", "GET", "/api/v1/capabilities"},
	{"version", "Print build information", "version", "GET", "/api/v1/health"},
}

const Instructions = `Avatars generates portraits locally and saves each batch as a collection.

Start the studio: avatars serve --open
Generate for review: avatars generate --count 12 --name "First portraits" --json
The result contains stable avatar and collection URLs. Open those links in the
running studio. Local CLI calls and the server share the same library directory.
Use --data-dir PATH on both to select an isolated library, or --url URL to make
CLI calls through a running HTTP service. AVATARS_DATA_DIR and AVATARS_URL set
these defaults. With --url, the server owns the library; --data-dir is refused.

No seed is required. Omit --seed to use fresh random seeds. For reproducibility,
use --seed TEXT; a single avatar uses that exact seed, while batches use
TEXT:0, TEXT:1, and so on. Pebble supports --color with a Walnut default.
Use --color random to choose independently from all 15 palette colors for each
avatar. A fixed seed repeats its color; saved recipes retain the concrete value.
Companions supports --animal dog|cat|mixed, defaulting to dog. Mixed selects
one species per final seed and saves dog or cat; it does not promise batch balance. Use avatars styles
--json for each style's accepted inputs, defaults, and native dimensions.
Field Birds invents bird species from the seed alone and accepts no appearance inputs.
Saved avatars retain these choices; export refuses appearance overrides.

List: avatars list --json
Inspect: avatars show ID --json
Export: avatars export AVATAR_ID --format png --size 256 --circle --out icon.png
Companions: avatars generate --style companions --animal cat --count 12 --json
Field Birds: avatars generate --style field-birds --seed field-notes --count 48 --json
Pebble: avatars generate --style pebble --color random --count 12 --json
Stateless: avatars render --style pebble --color sage --seed agent-42 --format svg --out icon.svg
SVG is the default format. --size N requests an N by N canvas; --size WxH sets
both dimensions. Native dimensions depend on the style: Gorey portraits are
64 by 72, Pebble is 64 by 64, Companions is 128 by 128, and Field Birds is 160 by 160. A circle crops the center and
leaves transparent corners. Each dimension must be 1..2048. Output files are
created exclusively: an existing file is never overwritten. Without --out,
render and export write image bytes to stdout. With --json they instead return
base64 image data, format, and media_type, or an output path when writing a file.
Generate --out supports one avatar and still saves its collection.

HTTP: GET /api/v1/styles, GET/POST /api/v1/collections,
GET /api/v1/collections/ID, GET /api/v1/avatars/ID,
GET /api/v1/avatars/ID.svg or .png, and GET /api/v1/render?seed=TEXT.
Export queries: format, style, seed, color, and animal (render only), plus width,
height, and circle. POST JSON: {"style":"pebble","inputs":{"color":"sage"},"count":12}.
Companions POST: {"style":"companions","inputs":{"animal":"cat"},"count":12}.
Unknown fields, duplicate scalar query values, and invalid inputs are refused. Errors use
{"error":{"code":"invalid_request|not_found|internal","message":"..."}}.
CLI exit codes: 0 success, 1 operation failure, 2 invalid invocation.

Use avatars help COMMAND for syntax and avatars capabilities for JSON discovery.
The service runs in the foreground until interrupted; it does not manage tmux
or start a daemon. For access from your other devices, place Tailscale Serve
in front of the loopback listener and start serve with --public-url HTTPS_ORIGIN
(or AVATARS_PUBLIC_URL). This permits that exact proxy host and browser origin.
Use --url HTTPS_ORIGIN for CLI calls that return shareable links.
`

func Capabilities() any {
	return struct {
		APIVersion  string         `json:"api_version"`
		Operations  []Operation    `json:"operations"`
		GlobalFlags []string       `json:"global_flags"`
		DefaultURL  string         `json:"default_url"`
		Limits      map[string]int `json:"limits"`
	}{"v1", Operations, []string{"--json", "--data-dir PATH", "--url URL", "--help"}, DefaultURL,
		map[string]int{"count": 100, "dimension": 2048, "name_characters": 120, "seed_bytes": 4096}}
}
