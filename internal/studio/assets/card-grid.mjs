import { createCardPhysics } from "./card-physics.mjs";
import { createCardSound } from "./card-sound.mjs";

const DRAG_THRESHOLD_MOUSE = 6;
const DRAG_THRESHOLD_TOUCH = 4;
const THROW_SAMPLE_MS = 90;
const THROW_FRESH_MS = 100;

export function deterministicCardAngle(id) {
  let hash = 0x811c9dc5;
  for (const character of String(id)) {
    hash ^= character.codePointAt(0);
    hash = Math.imul(hash, 0x01000193);
  }
  return (((hash >>> 0) % 1601) - 800) / 10000;
}

export function reconcileKeyedChildren(container, items, keyFor, create, update) {
  const existing = new Map([...container.children].map((node) => [node.dataset.avatar, node]));
  let cursor = container.firstElementChild;
  const result = [];
  for (const item of items) {
    const key = keyFor(item);
    const node = existing.get(key) || create(item);
    existing.delete(key);
    update(node, item);
    if (node !== cursor) container.insertBefore(node, cursor);
    cursor = node.nextElementSibling;
    result.push(node);
  }
  for (const node of existing.values()) node.remove();
  return result;
}

function footprint(card) {
  const cosine = Math.abs(Math.cos(card.angle || 0));
  const sine = Math.abs(Math.sin(card.angle || 0));
  return {
    rx: (card.width * cosine + card.height * sine) / 2,
    ry: (card.height * cosine + card.width * sine) / 2,
  };
}

export function chooseOpenSlot(slots, occupied, fallback) {
  return slots.find((candidate) => {
    const size = footprint(candidate);
    return occupied.every((other) => {
      const otherSize = footprint(other);
      return Math.abs(candidate.x - other.x) >= size.rx + otherSize.rx + 4
        || Math.abs(candidate.y - other.y) >= size.ry + otherSize.ry + 4;
    });
  }) || fallback;
}

export function createCardGrid({ grid, sound = createCardSound() }) {
  const physics = createCardPhysics();
  const reducedMotion = matchMedia("(prefers-reduced-motion: reduce)");
  let cards = [];
  let poses = new Map();
  let collectionKey = null;
  let drag = null;
  let animationFrame = 0;
  let lastFrame = 0;
  let arranging = null;
  let resizeFrame = 0;
  let observedWidth = 0;
  let observedHeight = 0;

  function slot(card) {
    return {
      x: card.offsetLeft + card.offsetWidth / 2,
      y: card.offsetTop + card.offsetHeight / 2,
      angle: deterministicCardAngle(card.dataset.avatar),
    };
  }

  function bounds() {
    return { left: 0, top: 0, right: grid.clientWidth, bottom: grid.clientHeight };
  }

  function specs({ preserve = false } = {}) {
    const slots = cards.map((card) => ({ ...slot(card), width: card.offsetWidth, height: card.offsetHeight }));
    const used = preserve ? [...poses.values()].map((pose) => {
      const card = cards.find((item) => item.dataset.avatar === pose.id);
      return { ...pose, width: card?.offsetWidth || 0, height: card?.offsetHeight || 0 };
    }) : [];
    return cards.map((card, index) => {
      let resting = slots[index];
      const pose = preserve ? poses.get(card.dataset.avatar) : null;
      if (preserve && !pose) {
        resting = chooseOpenSlot(slots, used, resting);
        used.push({ ...resting, width: card.offsetWidth, height: card.offsetHeight });
      }
      return {
        id: card.dataset.avatar,
        x: pose?.x ?? resting.x,
        y: pose?.y ?? resting.y,
        width: card.offsetWidth,
        height: card.offsetHeight,
        angle: pose?.angle ?? resting.angle,
      };
    });
  }

  function apply(next) {
    poses = new Map(next.map((pose) => [pose.id, pose]));
    for (const card of cards) {
      const pose = poses.get(card.dataset.avatar);
      if (!pose) continue;
      const resting = slot(card);
      card.style.setProperty("--card-x", `${pose.x - resting.x}px`);
      card.style.setProperty("--card-y", `${pose.y - resting.y}px`);
      card.style.setProperty("--card-angle", `${pose.angle}rad`);
    }
  }

  function reset({ preserve = false } = {}) {
    if (!cards.length || grid.clientWidth <= 0 || grid.clientHeight <= 0) return;
    physics.reset(specs({ preserve }), bounds());
    apply(physics.step(0));
  }

  function point(event) {
    const rect = grid.getBoundingClientRect();
    return { x: event.clientX - rect.left, y: event.clientY - rect.top };
  }

  function stopAnimation() {
    if (animationFrame) cancelAnimationFrame(animationFrame);
    animationFrame = 0;
    lastFrame = 0;
    arranging = null;
    sound.stop();
  }

  function movementSpeed(previous, next, seconds) {
    if (!(seconds > 0)) return 0;
    let speed = 0;
    for (const pose of next) {
      const before = previous.get(pose.id);
      if (before) speed = Math.max(speed, Math.hypot(pose.x - before.x, pose.y - before.y) / seconds);
    }
    return speed;
  }

  function physicsFrame(time) {
    animationFrame = 0;
    const seconds = lastFrame ? Math.min((time - lastFrame) / 1000, .05) : 0;
    lastFrame = time;
    const previous = poses;
    const next = physics.step(seconds);
    apply(next);
    sound.update(reducedMotion.matches ? 0 : movementSpeed(previous, next, seconds));
    if (physics.isActive()) animationFrame = requestAnimationFrame(physicsFrame);
    else {
      lastFrame = 0;
      sound.stop();
    }
  }

  function animatePhysics() {
    if (!animationFrame) animationFrame = requestAnimationFrame(physicsFrame);
  }

  function finishDrag(cancelled = false) {
    if (!drag) return;
    const current = drag;
    drag = null;
    if (current.started) {
      const latest = current.samples.at(-1);
      const first = current.samples[0];
      const elapsed = (latest.time - first.time) / 1000;
      const fresh = performance.now() - latest.time < THROW_FRESH_MS;
      const throwCard = !cancelled && !reducedMotion.matches && fresh && elapsed > .008;
      physics.release(
        throwCard ? (latest.x - first.x) / elapsed : 0,
        throwCard ? (latest.y - first.y) / elapsed : 0,
        cancelled || !throwCard,
      );
      current.card.dataset.suppressClick = "true";
      current.card.classList.remove("is-dragging");
      animatePhysics();
    }
    if (current.card.hasPointerCapture?.(current.pointerId)) current.card.releasePointerCapture(current.pointerId);
  }

  function arrange({ animate = true, audible = false } = {}) {
    finishDrag(true);
    stopAnimation();
    if (!cards.length) return;
    const targets = new Map(cards.map((card) => [card.dataset.avatar, { id: card.dataset.avatar, ...slot(card) }]));
    if (!animate || reducedMotion.matches || !poses.size) {
      reset();
      return;
    }
    const starts = new Map(cards.map((card) => [card.dataset.avatar, poses.get(card.dataset.avatar) || targets.get(card.dataset.avatar)]));
    const started = performance.now();
    arranging = { starts, targets, started, audible };
    const frame = (time) => {
      animationFrame = 0;
      if (!arranging) return;
      const progress = Math.min(1, (time - started) / 540);
      const eased = 1 - Math.pow(1 - progress, 3);
      const previous = poses;
      const next = cards.map((card) => {
        const id = card.dataset.avatar;
        const from = starts.get(id);
        const to = targets.get(id);
        return {
          id,
          x: from.x + (to.x - from.x) * eased,
          y: from.y + (to.y - from.y) * eased,
          angle: from.angle + (to.angle - from.angle) * eased,
        };
      });
      apply(next);
      const seconds = lastFrame ? Math.max((time - lastFrame) / 1000, .001) : 0;
      lastFrame = time;
      sound.update(audible ? movementSpeed(previous, next, seconds) : 0);
      if (progress < 1) animationFrame = requestAnimationFrame(frame);
      else {
        arranging = null;
        lastFrame = 0;
        sound.stop();
        reset();
      }
    };
    animationFrame = requestAnimationFrame(frame);
  }

  function sync({ nextCollectionKey = "all" } = {}) {
    const nextCards = [...grid.querySelectorAll(":scope > [data-avatar]")];
    const nextIDs = nextCards.map((card) => card.dataset.avatar).join("\u0000");
    const oldIDs = cards.map((card) => card.dataset.avatar).join("\u0000");
    const oldCount = cards.length;
    const changedCollection = collectionKey !== null && collectionKey !== nextCollectionKey;
    collectionKey = nextCollectionKey;
    if (nextIDs === oldIDs && !changedCollection && poses.size === nextCards.length) return;
    finishDrag(true);
    stopAnimation();
    cards = nextCards;
    if (!cards.length) {
      poses.clear();
      return;
    }
    const preserve = !changedCollection && nextCards.length >= oldCount && Boolean(oldIDs);
    reset({ preserve });
    if (preserve && nextCards.length !== oldCount) observedHeight = grid.clientHeight;
  }

  function onPointerDown(event) {
    const card = event.target.closest("[data-avatar]");
    if (!card || card.parentElement !== grid || event.button !== 0 || !event.isPrimary || drag || reducedMotion.matches) return;
    if (arranging) {
      stopAnimation();
      reset({ preserve: true });
    }
    void sound.unlock();
    delete card.dataset.suppressClick;
    const origin = point(event);
    drag = {
      card,
      pointerId: event.pointerId,
      pointerType: event.pointerType,
      origin,
      started: false,
      samples: [{ ...origin, time: performance.now() }],
    };
  }

  function onPointerMove(event) {
    if (!drag || drag.pointerId !== event.pointerId) return;
    const position = point(event);
    const dx = position.x - drag.origin.x;
    const dy = position.y - drag.origin.y;
    if (!drag.started) {
      const threshold = drag.pointerType === "touch" ? DRAG_THRESHOLD_TOUCH : DRAG_THRESHOLD_MOUSE;
      if (Math.hypot(dx, dy) < threshold) return;
      if (drag.pointerType === "touch" && Math.abs(dy) > Math.abs(dx)) return;
      if (!physics.grab(drag.card.dataset.avatar, drag.origin.x, drag.origin.y)) {
        drag = null;
        return;
      }
      drag.started = true;
      drag.card.classList.add("is-dragging");
      drag.card.setPointerCapture?.(event.pointerId);
    }
    event.preventDefault();
    const now = performance.now();
    drag.samples.push({ ...position, time: now });
    while (drag.samples.length > 2 && now - drag.samples[0].time > THROW_SAMPLE_MS) drag.samples.shift();
    physics.move(position.x, position.y);
    animatePhysics();
  }

  function onPointerEnd(event) {
    if (drag?.pointerId === event.pointerId) finishDrag(event.type !== "pointerup");
  }

  function onClick(event) {
    const card = event.target.closest("[data-avatar]");
    if (card?.dataset.suppressClick === "true" && event.detail === 0) {
      delete card.dataset.suppressClick;
      return;
    }
    if (card?.dataset.suppressClick === "true" && event.detail !== 0) {
      delete card.dataset.suppressClick;
      event.preventDefault();
      event.stopImmediatePropagation();
    }
  }

  function onResize() {
    const width = grid.clientWidth;
    const height = grid.clientHeight;
    const widthChanged = width !== observedWidth;
    const heightChanged = height !== observedHeight;
    observedWidth = width;
    observedHeight = height;
    if (!widthChanged && !heightChanged) return;
    if (resizeFrame) cancelAnimationFrame(resizeFrame);
    resizeFrame = requestAnimationFrame(() => {
      resizeFrame = 0;
      arrange({ animate: false });
    });
  }

  function onReducedMotion() {
    sound.stop();
    arrange({ animate: false });
  }

  grid.addEventListener("pointerdown", onPointerDown);
  grid.addEventListener("pointermove", onPointerMove);
  grid.addEventListener("pointerup", onPointerEnd);
  grid.addEventListener("pointercancel", onPointerEnd);
  grid.addEventListener("lostpointercapture", onPointerEnd, true);
  grid.addEventListener("click", onClick, true);
  grid.addEventListener("dragstart", (event) => event.preventDefault());
  grid.addEventListener("selectstart", (event) => event.preventDefault());
  window.addEventListener("blur", () => finishDrag(true));
  window.addEventListener("pointerup", onPointerEnd);
  window.addEventListener("pointercancel", onPointerEnd);
  window.addEventListener("resize", onResize);
  document.addEventListener("visibilitychange", () => {
    if (document.hidden) {
      finishDrag(true);
      sound.stop();
    }
  });
  document.addEventListener("keydown", (event) => {
    if (event.key === "Escape" && drag) {
      finishDrag(true);
      event.preventDefault();
    }
  });
  reducedMotion.addEventListener("change", onReducedMotion);
  new ResizeObserver(onResize).observe(grid);

  return {
    sync,
    arrange,
    cancel: () => finishDrag(true),
    unlockSound: () => sound.unlock(),
    toggleSound: () => sound.toggle(),
    get soundEnabled() { return sound.enabled; },
  };
}
