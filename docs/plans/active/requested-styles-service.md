# Requested styles and avatar delivery

Status: proposed; implementation is not part of the current change. This is the controlling plan for the next service capability. Planning tracked with `td-58936f`.

## Outcome

Marcus can enter a style request in the studio, follow an agent's progress, inspect a contact sheet, and accept a new style into the catalog. The same workflow is available through the CLI and HTTP API. People Marcus authorizes can eventually submit requests and generate avatars; they can download SVG/PNG files or use hosted links.

Start with one local worker. Keep the existing studio, generators, collection library, and export adapters. Hosting, Google sign-in, and invited access follow the local proof; this plan does not authorize public exposure or move local agent credentials to a server.

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

Today, `pkg/avatar/engine.go` registers five compiled generators, inputs are typed `Color` and `Animal` fields, and saved recipes do not pin an implementation version. These are the seams to extend, not capabilities already present.

A bundle contains a manifest (ID, immutable version, SDK/protocol version, native dimensions, input descriptors and digests), source, a build artifact for the worker platform, fixtures, preview images, check results, and reference provenance. Store requests and accepted bundles under the application data directory through `RequestStore`, `StyleCatalog`, and artifact-store interfaces. Agents can write their attempt directory; only the service can promote a candidate into the catalog.

The proposed `ProcessGenerator` speaks a small versioned JSON protocol: a validated seed and resolved inputs enter; bounded artwork bytes, media type, and dimensions leave. It implements the existing generator contract so SVG/PNG exports remain shared. Build source using a pinned toolchain and service-owned commands. Measure process startup and generation cost before choosing per-request processes or a retained worker.

This adds a deployment artifact and process supervision, but allows new styles without rebuilding the main app. Go's in-process plugin mechanism has platform and toolchain compatibility limits; its own documentation suggests IPC for some of these cases. A separate process is a compatibility boundary, not an operating-system sandbox. [Go plugin documentation](https://pkg.go.dev/plugin)

Publish an atomic catalog snapshot after acceptance, with explicit reload available through the core. Validate protocol, digests and supported platform before activation. Never overwrite a version in place. Pin saved recipes to a version/digest; define a stable legacy mapping for existing built-in recipes so old collections keep their appearance. Rollback selects a prior version for new generation while retaining versions referenced by saved avatars.

Add only the input description needed by this workflow: bounded enum choices, labels, defaults, and an explicit mixed/random choice with deterministic resolution. Introduce a style-scoped input map with compatibility translation for existing color/animal clients. The manifest describes controls; the core validates and resolves them. Arbitrary scripts or form schemas are outside this phase. This keeps a new style's input from requiring a new hardcoded studio field.

## Worker and authoring contract

`AgentRunner` starts, observes, cancels, and reconciles attempts. Its application contract contains no ACP-specific events or session IDs; the adapter maps those internally. Pin and probe the tested acpx/ACP bridge versions. acpx provides structured output, named sessions, cancellation, and custom agent commands, which fit this seam. [acpx](https://acpx.sh/)

The local CLI is installed at version **0.13.2**; help output was inspected. No real acpx job or provider authentication was exercised for this plan. Use a unique session name and absolute workspace per attempt because acpx session identity includes the agent command, working directory, and optional name. [Session behavior](https://acpx.sh/sessions.html)

Give the worker the capabilities the art task needs: public reference search, writes inside its workspace, service-owned build/check commands, and a browser or renderer for previews. Do not give it application auth secrets, unrestricted home access, deployment credentials, Git push access, or authority to activate styles. Clear inherited environment variables except an explicit allowlist, and deliver provider credentials only to the component that needs them.

acpx tool policies and noninteractive denial help bound actions, but do not prove containment of a harness's own terminal or filesystem tools. Validate the actual child-process permissions and operating-system boundary. Denied actions produce a reviewable needs-attention state; do not fall back to blanket approval. Reference browsing should not grant access to private network services. Before invited prompts can execute, demonstrate containment of build scripts, generated executables, network access, and preview content on the chosen worker platform. Until then, keep execution owner-only and document its actual trust boundary. [acpx permissions](https://acpx.sh/permissions.html)

Extend [the authoring guide](../../guides/active/creating-styles.md) and [brief template](../../guides/active/style-brief-template.md) with the bundle SDK, manifest example, permitted commands, expected result paths, mixed-choice convention, reproducibility checks, small-size contact sheets, and reference attribution. The first proof agent receives only this checked-in package plus the user's request. It must not depend on instructions held in a chat session.

## Identity, hosting, and delivery

Google identifies a person; an allowlist or invitation grants access. Keep owner, requester, and viewer permissions distinct. Require authorization for prompt submission, job events, references, candidate code, activation, and private collections. Apply request limits before exposing an expensive coding worker to invited users.

Inquiry contains Google routes under `src/routes/api/auth/` and helpers in `src/lib/server/auth.js` at `/Users/marcus/code/inquiry`. They identify the likely setup: `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, and a callback route. The Cloud project, credential ownership, and registered redirects have not been verified. Reuse setup knowledge, not the handlers unchanged.

Use a maintained OIDC library with state/PKCE and server-side token validation, including issuer, audience and expiry; use Google's stable subject identifier for accounts. Configure exact callback URLs, secure HTTP-only sessions, CSRF protection for mutations, and explicit allowlist/invite checks. Avoid requesting offline Google access when sign-in is the only need. [Google OpenID Connect guidance](https://developers.google.com/identity/openid-connect/openid-connect)

For hosting, keep a public web/API service separable from the worker. A local worker can pull authorized jobs outbound, retaining local provider configuration; a future paid-API runner can implement the same adapter. Check provider terms and credential support before selecting a server runtime. Subscription portability is not assumed.

Use one asset-store interface for immutable generated exports, initially local files and later an object-store adapter. Downloads and hosted links resolve the same saved recipe/artifact. Default collections and requests to private; grant unlisted or public asset access explicitly, define revocation, and keep public assets separate from private prompts and job logs. Add delivery caching, quotas, and retention rules when hosting is introduced.

## Work sequence and evidence

1. **Bundle proof:** package one existing style, load it through the proposed adapter, add a version, and prove existing saved SVG/PNG output remains stable. Confirm catalog reload and unsupported-bundle errors before teaching agents the format.
2. **Local request steel thread:** owner authentication, durable request/attempt core, one acpx worker, CLI/API/studio progress, preview review, and explicit activation. Prove a fresh style can be created from the committed guide and activated without editing or rebuilding the app. Exercise cancellation, denied permissions, failed checks, and restart recovery.
3. **Invited access:** confirm the Google project, implement OIDC and scoped script credentials, invitations and owner review. Prove authorization parity and worker containment before letting invited prompts dispatch.
4. **Hosted delivery:** choose the domain and worker location, introduce sharing and an asset-store adapter, and verify remote downloads, stable URLs, private-resource isolation, and limits.

Before implementation, resolve the owner bootstrap choice (local secret or Google), verify the OAuth project and domain, and choose a containment mechanism supported by the actual worker platform. Public visibility, automatic activation, arbitrary third-party plugins, AI image providers, billing, and multi-worker scheduling remain outside the first version.

## Handoff

This document records the proposed next capability only. The current delivery adds Mixed dogs/cats and the general mixed-choice authoring convention. Future work should start from the bundle proof and update this plan as the contracts are verified.
