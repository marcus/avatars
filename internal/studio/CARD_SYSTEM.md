# Tactile preview card provenance

The studio preview grid adapts the smaller OpenTangle card system inspected at
commit `78b0a86862770b695183f589f32b85d10b93cff9`:

- `assets/card-physics.mjs` retains OpenTangle's headless Matter.js adapter,
  including its fixed stepping, restrained throw limits, damped collisions,
  bounded dragging, cancellation, and sleep behavior.
- `assets/card-sound.mjs` retains OpenTangle's single procedural paper-rubbing
  voice. The only product-specific change is the persisted preference key,
  `avatars-card-sound`.
- `assets/vendor/matter/` retains OpenTangle's pinned Matter.js 0.20.0 ESM
  bridge, upstream license, and distribution checksum notes. The bundle is
  local and requires no CDN or build pipeline.

OpenTangle's `card-table.js`, `card-skin.css`, and `cards.css` informed the
pointer threshold, movement-to-sound coupling, stock layers, grain, and quiet
shadows. Their full-page scene and Three.js renderer were deliberately omitted:
the avatar studio needs a scrolling preview panel, and direct DOM transforms
keep that boundary small.

Marcus Pictures was inspected at commit `7a5142d`. Its shared-voice timing tests
confirmed that motion sound should begin from the current animation frame and
must never replay a movement that finished while audio permission was pending.
The studio keeps the OpenTangle voice because preview cards do not need the
gallery transition whoosh added by Marcus Pictures.
