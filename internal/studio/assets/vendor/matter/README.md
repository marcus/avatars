# Matter.js 0.20.0

Source: https://github.com/liabru/matter-js/tree/0.20.0

Distribution: https://raw.githubusercontent.com/liabru/matter-js/0.20.0/build/matter.min.js

License: MIT; the upstream license is included in `LICENSE`.

`matter.mjs` wraps the unmodified upstream UMD bundle in module-local `module` and `exports` bindings, then exports `module.exports`. This makes the same vendored engine importable from browsers and Node without globals, a CDN, or a build step.

Upstream distribution SHA-256: `72d30be0f579eb02ce1e0b6f9d359a4f392e6837e5a26ba8be5dbee7f88e24ae`.

The card adapter uses Bodies, Body, Composite, Constraint, Engine and Sleeping. It supplies its own animation clock; Matter's DOM mouse and rendering modules are not used.
