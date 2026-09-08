import Matter from './vendor/matter/matter.mjs';

const { Bodies, Body, Composite, Constraint, Engine, Sleeping } = Matter;
const FIXED_SECONDS = 1 / 120;
const BASE_SECONDS = 1 / 60; // Matter velocity units are pixels per base (60Hz) update.
const MAX_THROW = 1600;
const MAX_DRAG = 2600;
const clamp = (value, min, max) => Math.max(min, Math.min(max, value));

/** Headless tabletop adapter. CSS pixels, y down, clockwise radians; no owned clock or DOM. */
export function createCardPhysics() {
  const engine = Engine.create({
    enableSleeping: true,
    gravity: { x: 0, y: 0, scale: 0 },
    positionIterations: 8,
    velocityIterations: 6,
    constraintIterations: 3
  });
  const cards = new Map();
  let bounds = null;
  let dragging = null;
  let accumulator = 0;

  function snapshot() {
    return [...cards].map(([id, { body }]) => ({ id, x: body.position.x, y: body.position.y, angle: body.angle }));
  }

  function reset(specs, tableBounds) {
    dragging = null;
    accumulator = 0;
    Composite.clear(engine.world, false);
    Engine.clear(engine);
    cards.clear();
    bounds = { ...tableBounds };
    const width = bounds.right - bounds.left;
    const height = bounds.bottom - bounds.top;
    if (!(width > 0 && height > 0)) throw new RangeError('Card table bounds must have positive dimensions');
    const thickness = Math.max(200, ...specs.map(card => Math.max(card.width, card.height)));
    const cx = (bounds.left + bounds.right) / 2;
    const cy = (bounds.top + bounds.bottom) / 2;
    const wall = { isStatic: true, restitution: .12, friction: .22 };
    Composite.add(engine.world, [
      Bodies.rectangle(cx, bounds.top - thickness / 2, width + thickness * 2, thickness, wall),
      Bodies.rectangle(cx, bounds.bottom + thickness / 2, width + thickness * 2, thickness, wall),
      Bodies.rectangle(bounds.left - thickness / 2, cy, thickness, height, wall),
      Bodies.rectangle(bounds.right + thickness / 2, cy, thickness, height, wall)
    ]);
    for (const spec of specs) {
      const body = Bodies.rectangle(spec.x, spec.y, spec.width, spec.height, {
        angle: spec.angle || 0,
        restitution: .16,
        friction: .24,
        frictionStatic: .4,
        frictionAir: .042,
        sleepThreshold: 40,
        slop: .025
      });
      Body.setInertia(body, body.inertia * 2.5);
      cards.set(spec.id, { body, width: spec.width, height: spec.height });
      Composite.add(engine.world, body);
      // A freshly dealt table stays exactly where the layout placed it until touched.
      Sleeping.set(body, true);
    }
  }

  function grab(id, x, y) {
    if (dragging) release(0, 0, true);
    const entry = cards.get(id);
    if (!entry) return false;
    const { body } = entry;
    Sleeping.set(body, false);
    Body.setVelocity(body, { x: 0, y: 0 });
    Body.setAngularVelocity(body, 0);
    const constraint = Constraint.create({
      pointA: { x, y },
      bodyB: body,
      // Matter stores this offset at the body's current angle and rotates it as the body turns.
      pointB: { x: x - body.position.x, y: y - body.position.y },
      length: 0,
      stiffness: .48,
      damping: .18,
      angularStiffness: .35
    });
    dragging = { entry, constraint, pointer: { x, y } };
    Composite.add(engine.world, constraint);
    return true;
  }

  function move(x, y) {
    if (!dragging || !Number.isFinite(x) || !Number.isFinite(y)) return;
    dragging.pointer = { x, y };
    Sleeping.set(dragging.entry.body, false);
  }

  function release(vx = 0, vy = 0, cancelled = false) {
    if (!dragging) return;
    const { entry: { body }, constraint } = dragging;
    Composite.remove(engine.world, constraint);
    // Matter caches the final constraint impulse for the next solver pass. Once the
    // hand releases, that warm-start impulse must not reintroduce motion or spin.
    body.constraintImpulse.x = 0;
    body.constraintImpulse.y = 0;
    body.constraintImpulse.angle = 0;
    dragging = null;
    if (cancelled) {
      Body.setVelocity(body, { x: 0, y: 0 });
      Body.setAngularVelocity(body, 0);
      Sleeping.set(body, true);
      return;
    }
    vx = Number.isFinite(vx) ? vx : 0;
    vy = Number.isFinite(vy) ? vy : 0;
    const speed = Math.hypot(vx, vy);
    const scale = Math.min(1, MAX_THROW / (speed || 1));
    Body.setVelocity(body, { x: vx * scale * BASE_SECONDS, y: vy * scale * BASE_SECONDS });
    Body.setAngularVelocity(body, speed > 20 ? clamp(Body.getAngularVelocity(body), -.024, .024) : 0);
    Sleeping.set(body, speed < 1);
  }

  function advancePointer() {
    if (!dragging) return;
    const { entry, constraint, pointer } = dragging;
    const { body, width, height } = entry;
    Sleeping.set(body, false);
    const c = Math.abs(Math.cos(body.angle)), s = Math.abs(Math.sin(body.angle));
    const rx = Math.min((width * c + height * s) / 2, (bounds.right - bounds.left) / 2);
    const ry = Math.min((height * c + width * s) / 2, (bounds.bottom - bounds.top) / 2);
    // Clamp the whole current footprint, rather than only the pointer, to the table.
    const targetX = clamp(pointer.x - constraint.pointB.x, bounds.left + rx, bounds.right - rx) + constraint.pointB.x;
    const targetY = clamp(pointer.y - constraint.pointB.y, bounds.top + ry, bounds.bottom - ry) + constraint.pointB.y;
    const dx = targetX - constraint.pointA.x, dy = targetY - constraint.pointA.y;
    const fraction = Math.min(1, MAX_DRAG * FIXED_SECONDS / (Math.hypot(dx, dy) || 1));
    constraint.pointA.x += dx * fraction;
    constraint.pointA.y += dy * fraction;
  }

  function containActiveCards() {
    for (const entry of cards.values()) {
      const { body, width, height } = entry;
      if (body.isSleeping) continue;
      const c = Math.abs(Math.cos(body.angle)), s = Math.abs(Math.sin(body.angle));
      const rx = Math.min((width * c + height * s) / 2, (bounds.right - bounds.left) / 2);
      const ry = Math.min((height * c + width * s) / 2, (bounds.bottom - bounds.top) / 2);
      const x = clamp(body.position.x, bounds.left + rx, bounds.right - rx);
      const y = clamp(body.position.y, bounds.top + ry, bounds.bottom - ry);
      const dx = x - body.position.x, dy = y - body.position.y;
      if (Math.abs(dx) < 1e-7 && Math.abs(dy) < 1e-7) continue;
      // The final pointer-constraint pass can leave a few pixels of wall penetration.
      // Project only that numerical excess back inside; contacts remain Matter's job.
      const velocity = Body.getVelocity(body);
      Body.setPosition(body, { x, y });
      const bounce = dragging?.entry === entry ? 0 : .12;
      if (dx * velocity.x < 0) velocity.x *= -bounce;
      if (dy * velocity.y < 0) velocity.y *= -bounce;
      Body.setVelocity(body, velocity);
    }
  }

  function step(deltaSeconds = 0) {
    if (!Number.isFinite(deltaSeconds) || deltaSeconds <= 0) return snapshot();
    accumulator += Math.min(deltaSeconds, .05);
    while (accumulator + 1e-10 >= FIXED_SECONDS) {
      advancePointer();
      Engine.update(engine, FIXED_SECONDS * 1000);
      containActiveCards();
      accumulator -= FIXED_SECONDS;
    }
    return snapshot();
  }

  function isActive() {
    return dragging !== null || [...cards.values()].some(({ body }) => !body.isSleeping);
  }

  function turn(id, angle = .012) {
    const entry = cards.get(id);
    if (!entry || !Number.isFinite(angle)) return;
    const { body } = entry;
    Body.setVelocity(body, { x: 0, y: 0 });
    Body.setAngularVelocity(body, 0);
    Body.setAngle(body, body.angle + clamp(angle, -.025, .025));
    Sleeping.set(body, false);
  }

  return { reset, grab, move, release, turn, step, isActive };
}
