# Requested styles and avatar delivery

Status: Part 1 is implemented, independently reviewed, and installed locally under `td-e3efc3`, with Avatars lifecycle in `td-5ed8eb` and comms-web integration in `td-3543ba`. Later phases remain proposed. This is the controlling plan for the service capability; original planning is tracked with `td-58936f`.

## Outcome

Marcus can enter a style request in the studio, follow an agent's progress, inspect a contact sheet, and accept a new style into the catalog. The same workflow is available through the CLI and HTTP API. People Marcus authorizes can eventually submit requests and generate avatars; they can download SVG/PNG files or use hosted links.

Start with an on-demand local service, then one local agent worker. Keep the existing studio, generators, collection library, and export adapters. Hosting, Google sign-in, and invited access follow the local proof; this plan does not authorize public exposure or move local agent credentials to a server.

## On-demand service and the first consuming app

Part 1 is delivered. `avatars service ensure|status|stop --json` manages an on-demand local service without requiring the agent worker or authentication phases. The process and locking patterns are adapted from Comms with source attribution. Offline `avatars render` remains available.

`ensure` returns a compatible service or takes an endpoint-specific lifecycle lock, rechecks ownership and library identity, starts one detached child, and checks HTTP plus the private control socket. Its result includes `endpoint`, `pid`, `instance_id`, `launch_mode`, `version`, `commit`, `data_dir`, `lifecycle_socket`, and `log_path`. `status` never starts a service. `stop` requires the matching auto-started instance and refuses foreign, foreground, or supervised ownership. Shutdown is available only through a mode-0600 Unix socket, not the HTTP API.

Configure the loopback address with `--listen` or `AVATARS_LISTEN`, the trusted studio proxy with `--public-url` or `AVATARS_PUBLIC_URL`, and the startup deadline with `--timeout`. `--no-auto-start` or `AVATARS_AUTO_START=0` disables startup. The existing `--data-dir` / `AVATARS_DATA_DIR` selects the library; lifecycle paths are partitioned by endpoint under the runtime/cache directory. Automatic upgrade handoff remains deferred.

Comms-web calls `avatars service ensure --json` on the server and renders through the Avatars HTTP API. The browser uses same-origin `/api/avatars/image`; it never targets the laptop's localhost. `AVATARS_BIN` selects the executable. An explicit `AVATARS_ENDPOINT` uses that service without spawning locally. The adapter coalesces startup and identical renders, caches up to 256 images by full recipe and generator version/commit, refreshes service metadata after 30 seconds, and uses bounded timeouts with a two-second failure cooldown. The original TypeScript portrait generator remains the fallback for an unavailable installation; Avatars is the only selectable provider.

The header offers a compact default picker; the reader portrait opens the scoped picker. Both flyouts escape pane clipping and show every style and its advertised appearance inputs, including random Pebble colors and mixed Companions animals. Choices save immediately and invalidate portraits in the current view. The UI has no collection concepts.

Comms-web owns an append-only JSONL preference store at `~/.local/state/comms-web/avatar-preferences.jsonl`, configurable through `COMMS_WEB_AVATAR_PREFERENCES`. Its shared resolution order is session override, agent override, shared default, then Gorey. `GET/PUT/DELETE /api/avatars` exposes the same choices to headless callers. Session keys include agent identity and the recorded session reference; historical messages never borrow a newer session. The session control appears only when Comms supplies a `session_ref`. The current live messages have no such references, so session precedence was proved through the API and focused tests. Clearing an override restores inheritance; changing the default leaves explicit overrides intact.

### Part 1 verification and local installation

- Avatars implementation was independently reviewed at `a3568d7` and landed/installed from `4963e32`. `make fmt-check vet test test-race build` passed, with focused CLI/lifecycle race and vet checks after the final ownership patch. Linux and macOS CI passed for the landed implementation.
- Twelve concurrent real CLI callers returned one PID and instance; the socket was mode 0600. Status/opt-out, render, reuse, and successful stop followed by stopped status were verified on isolated port 17447.
- Comms-web is on `main` at `b41bc7a`. All 16 Node tests pass, `pnpm check` reports no errors or warnings, and the production build succeeds. The final flyout adjustment received independent review and browser verification.
- Twenty image requests through an isolated comms-web server used the installed on-demand service without fallback. The missing-install proof returned the legacy SVG in 10 ms, then 1 ms during the failure cooldown. The server image matched the direct Avatars SVG bytes.
- API proof covered session/agent/default precedence, same session reference on different agents, historical-session isolation, default changes preserving overrides, and clearing overrides. Browser proof covered live default and agent changes, six style previews, circular crops, and unclipped global/profile flyouts. The production default was restored to Gorey after verification.
- The verified tailnet URLs remain [Avatars Studio](https://aerie.tail53fd54.ts.net:7447) and [Comms Web](https://aerie.tail53fd54.ts.net:9111). The old studio foreground process was explicitly migrated to the managed service. Comms-web's existing launchd service carries `AVATARS_PUBLIC_URL` so future cold starts preserve the studio proxy. Other Tailscale routes were not changed. Both deployed services survived the user-reported tmux crash.
- Machine-local rollback and proof artifacts are under `~/.local/state/avatars/rollbacks/20260907-220801-part1/`: prior binary path, launcher, production build archive, original untracked comms-web file, and JSON proof results. The pre-existing comms-web state-module refactor was preserved byte-for-byte.

## Decisions and proposals

The requested direction is settled: local first, agents use the committed authoring guide, the harness and model are replaceable, saved styles are reusable, and users get downloads or hosting. Offer a mixed/random choice for finite appearance options where it makes sense, resolving and saving each avatar's actual choice.

The following implementation choices are proposals to validate in the first slice:

- Use **acpx** as the first `AgentRunner` adapter. Keep harness, model, and permission profile in trusted configuration. Website prompts cannot supply executable commands.
- Store durable requests, attempts, and catalog records behind narrow repository interfaces with JSONL as the first adapter. Start with one worker; no external queue is needed.
- Load immutable, versioned **style bundles** through a generator process adapter. An agent produces a candidate; accepting it activates a catalog version. Existing built-in generators keep working.
- Let Marcus's requests dispatch directly within the configured worker policy. Invited requests initially need owner acceptance before starting an agent. Completing a build does not automatically publish or activate its output.
- Use Google OIDC for invited access. An owner-only local prototype can begin with a configured access secret and a signed browser session; choose Google immediately if reusing the OAuth project is straightforward.

## First user journey

1. The owner opens Request a style, enters the art direction, and optionally attaches a reference. Submission creates a durable request and returns its URL immediately.
2. The worker creates an isolated attempt workspace containing a generated brief, the authoring guide, a pinned SDK/template, and fixtures. It invokes the configured agent and streams concise progress through the request API.
3. The agent writes a candidate bundle and previews. The service runs the required checks through its own build wrapper and records results. An agent saying it finished is insufficient.
4. The owner reviews contact sheets at avatar sizes, input controls, SVG/PNG output, and any failed checks. Accept activates the immutable version; reject preserves the attempt for revision.
5. The new style appears in the existing style picker and `styles --json`. Generating, saving, linking, and downloading use the existing application core.

Use a separate request screen and review panel; the avatar grid remains a place to explore generated avatars. Reference attachments and generated previews are treated as untrusted content and are served with appropriate media validation and isolation.

## Shared core and records

| Owned capability | Proposed CLI | HTTP / studio use |
| --- | --- | --- |
| Submit, inspect, list requests | `requests create/show/list --json` | Submit prompt, open request URL |
| Observe and control attempts | `requests events/cancel/retry --json` | Progress, cancel, retry |
| Accept an invited request for execution | `requests authorize` | Owner request queue |
| Inspect and activate a candidate | `styles candidates/show/activate --json` | Review and accept |
| Generate, save, export, share | Existing commands, extended with version/share metadata | Existing studio and asset links |

Command and route names are provisional. Authentication and authorization apply to the same application operations across transports. A script needs a scoped, revocable credential; browser cookies are not its only path.

Keep records small: `StyleRequest` owns prompt, references and requester; `Attempt` owns runner identity, state, events and artifacts; `StyleVersion` owns manifest, digest, source and build provenance; an avatar recipe pins its style version and resolved inputs. Ownership and share visibility belong to application records, not URL conventions.

Persist submission before dispatch. Use one active attempt per request, attempt IDs for retries, explicit cancellation, bounded execution, and states such as queued, running, needs-attention, failed, ready-for-review, accepted and canceled. On restart, reconcile the recorded runner session or mark the attempt interrupted; never blindly launch it again. Store normalized progress separately from restricted raw logs, which may contain private prompt or tool data.

## Saving and loading styles

Today, `pkg/avatar/engine.go` registers six compiled generators, inputs are typed `Color` and `Animal` fields, and saved recipes do not pin an implementation version. These are the seams to extend, not capabilities already present.

A bundle contains a manifest (ID, immutable version, SDK/protocol version, native dimensions, input descriptors and digests), source, a build artifact for the worker platform, fixtures, preview images, check results, and reference provenance. Store requests and accepted bundles under the application data directory through `RequestStore`, `StyleCatalog`, and artifact-store interfaces. Agents can write their attempt directory; only the service can promote a candidate into the catalog.

The proposed `ProcessGenerator` speaks a small versioned JSON protocol: a validated seed and resolved inputs enter; bounded artwork bytes, media type, and dimensions leave. It implements the existing generator contract so SVG/PNG exports remain shared. Build source using a pinned toolchain and service-owned commands. Measure process startup and generation cost before choosing per-request processes or a retained worker.

This adds a deployment artifact and process supervision, but allows new styles without rebuilding the main app. Go's in-process plugin mechanism has platform and toolchain compatibility limits; its own documentation suggests IPC for some of these cases. A separate process is a compatibility boundary, not an operating-system sandbox. [Go plugin documentation](https://pkg.go.dev/plugin)

Publish an atomic catalog snapshot after acceptance, with explicit reload available through the core. Acquire one immutable version handle for validation, input resolution, generation, and the persisted recipe identity; concurrent activation must not switch implementations halfway through an operation. Validate protocol, digests and supported platform before activation. Never overwrite a version in place. Pin saved recipes to a version/digest; define a stable legacy mapping backed by retained built-in implementations so old collections keep their appearance. Rollback selects a prior version for new generation while retaining versions referenced by saved avatars.

Add only the input description needed by this workflow: bounded enum choices, labels, defaults, and an explicit mixed/random choice with deterministic resolution. Introduce a style-scoped input map with compatibility translation for existing color/animal clients. The manifest describes controls; the core validates and resolves them. Arbitrary scripts or form schemas are outside this phase. This keeps a new style's input from requiring a new hardcoded studio field.

## Worker and authoring contract

`AgentRunner` starts, observes, cancels, and reconciles attempts. Its application contract contains no ACP-specific events or session IDs; the adapter maps those internally. Pin and probe the tested acpx/ACP bridge versions. acpx provides structured output, named sessions, cancellation, and custom agent commands, which fit this seam. [acpx](https://acpx.sh/)

The local CLI is installed at version **0.13.2**; help output was inspected. No real acpx job or provider authentication was exercised for this plan. Use a unique session name and absolute workspace per attempt because acpx session identity includes the agent command, working directory, and optional name. [Session behavior](https://acpx.sh/sessions.html)

Give the worker the capabilities the art task needs: public reference search, writes inside its workspace, service-owned build/check commands, and a browser or renderer for previews. Do not give it application auth secrets, unrestricted home access, deployment credentials, Git push access, or authority to activate styles. Clear inherited environment variables except an explicit allowlist, and deliver provider credentials only to the component that needs them.

acpx tool policies and noninteractive denial help bound actions, but do not prove containment of a harness's own terminal or filesystem tools. Validate the actual child-process permissions and operating-system boundary. Denied actions produce a reviewable needs-attention state; do not fall back to blanket approval. Reference browsing should not grant access to private network services. Before invited prompts can execute, demonstrate containment of build scripts, generated executables, network access, and preview content on the chosen worker platform. Until then, keep execution owner-only and document its actual trust boundary. [acpx permissions](https://acpx.sh/permissions.html)

Apply containment to every activated bundle render as well as authoring and build attempts: accepting artwork does not authorize arbitrary host access. Bound CPU, memory, time and output size; deny network and writes outside disposable scratch space, and keep catalog artifacts read-only. Validate generated SVG/media in service-owned code before rasterizing, storing or serving it, including native-size SVG which the existing exporter passes through. Prove the chosen boundary before enabling generated bundles for untrusted callers.

Extend [the authoring guide](../../guides/active/creating-styles.md) and [brief template](../../guides/active/style-brief-template.md) with the bundle SDK, manifest example, permitted commands, expected result paths, mixed-choice convention, reproducibility checks, small-size contact sheets, and reference attribution. The first proof agent receives only this checked-in package plus the user's request. It must not depend on instructions held in a chat session.

## Identity, hosting, and delivery

Google identifies a person; an allowlist or invitation grants access. Keep owner, requester, and viewer permissions distinct. Require authorization for prompt submission, job events, references, candidate code, activation, and private collections. Apply request limits before exposing an expensive coding worker to invited users.

Inquiry contains Google routes under `src/routes/api/auth/` and helpers in `src/lib/server/auth.js` at `/Users/marcus/code/inquiry`. They identify the likely setup: `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, and a callback route. The Cloud project, credential ownership, and registered redirects have not been verified. Reuse setup knowledge, not the handlers unchanged.

Use a maintained OIDC library with state/PKCE and server-side token validation, including issuer, audience and expiry; use Google's stable subject identifier for accounts. Configure exact callback URLs, secure HTTP-only sessions, CSRF protection for mutations, and explicit allowlist/invite checks. Avoid requesting offline Google access when sign-in is the only need. [Google OpenID Connect guidance](https://developers.google.com/identity/openid-connect/openid-connect)

For hosting, keep a public web/API service separable from the worker. A local worker can pull authorized jobs outbound, retaining local provider configuration; a future paid-API runner can implement the same adapter. Check provider terms and credential support before selecting a server runtime. Subscription portability is not assumed.

Use one asset-store interface for immutable generated exports, initially local files and later an object-store adapter. Downloads and hosted links resolve the same saved recipe/artifact. Default collections and requests to private; grant unlisted or public asset access explicitly, define revocation, and keep public assets separate from private prompts and job logs. Add delivery caching, quotas, and retention rules when hosting is introduced.

## Work sequence and evidence

1. **On-demand local service — complete:** local lifecycle and comms-web HTTP integration are installed and verified; see Part 1 evidence above.
2. **Bundle proof:** package one existing style, load it through the proposed adapter, add a version, and prove existing saved SVG/PNG output remains stable. Confirm catalog reload and unsupported-bundle errors before teaching agents the format.
3. **Local request steel thread:** owner authentication, durable request/attempt core, one acpx worker, CLI/API/studio progress, preview review, and explicit activation. Prove a fresh style can be created from the committed guide and activated without editing or rebuilding the app. Exercise cancellation, denied permissions, failed checks, and restart recovery.
4. **Invited access:** confirm the Google project, implement OIDC and scoped script credentials, invitations and owner review. Prove authorization parity and worker containment before letting invited prompts dispatch.
5. **Hosted delivery:** choose the domain and worker location, introduce sharing and an asset-store adapter, and verify remote downloads, stable URLs, private-resource isolation, and limits.

Before agent-request implementation, resolve the owner bootstrap choice (local secret or Google) and choose a containment mechanism supported by the actual worker platform. OAuth-project verification is a prerequisite for Google sign-in, and domain selection belongs to hosted delivery; neither blocks the local service or bundle proof. Public visibility, automatic activation, arbitrary third-party plugins, AI image providers, billing, and multi-worker scheduling remain outside the first version.

## Handoff

Part 1 is complete under `td-e3efc3`, `td-5ed8eb`, and `td-3543ba`. Both repositories are landed on `main`; Avatars is installed locally and comms-web is served by its existing launchd service. Ten pre-existing clean merged worktrees were removed using pinned Sidecar deletion plans. Temporary proof services are stopped and implementation worktrees are cleaned up after landing.

Continue with Part 2, the bundle proof, only when requested. Bundle loading, agent workers, authentication, invited access, and hosted delivery remain proposed. Keep this plan active while those later phases remain outstanding.
