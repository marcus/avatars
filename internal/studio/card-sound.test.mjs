import assert from "node:assert/strict";
import test from "node:test";
import { createCardSound } from "./assets/card-sound.mjs";

function audioHarness(t, { suspended = false, stored = null } = {}) {
  const sources = [];
  const gains = [];
  const listeners = new Map();
  const writes = [];
  const parameter = () => ({
    value: 0,
    ramps: [],
    cancelAndHoldAtTime() {},
    cancelScheduledValues() {},
    setValueAtTime(value) { this.value = value; },
    linearRampToValueAtTime(value, time) { this.ramps.push({ value, time }); },
  });
  const node = () => ({ connect() {}, disconnect() {} });
  let context;
  let resume;
  class FakeAudioContext {
    constructor() {
      context = this;
      this.currentTime = 0;
      this.sampleRate = 1000;
      this.state = suspended ? "suspended" : "running";
      this.destination = {};
    }
    createBuffer() { return { getChannelData: () => new Float32Array(1800) }; }
    createBufferSource() {
      const source = {
        ...node(),
        starts: [],
        stops: [],
        start: (time) => source.starts.push(time ?? this.currentTime),
        stop: (time) => source.stops.push(time),
      };
      sources.push(source);
      return source;
    }
    createBiquadFilter() { return { ...node(), frequency: parameter(), Q: parameter() }; }
    createGain() {
      const gain = { ...node(), gain: parameter() };
      gains.push(gain);
      return gain;
    }
    resume() {
      return new Promise((resolve) => {
        resume = () => { this.state = "running"; resolve(); };
      });
    }
  }
  for (const [name, value] of Object.entries({
    AudioContext: FakeAudioContext,
    localStorage: {
      getItem: (key) => {
        assert.equal(key, "avatars-card-sound");
        return stored;
      },
      setItem: (key, value) => writes.push([key, value]),
    },
    document: { hidden: false, addEventListener: (name, callback) => listeners.set(name, callback) },
  })) {
    const previous = Object.getOwnPropertyDescriptor(globalThis, name);
    Object.defineProperty(globalThis, name, { configurable: true, writable: true, value });
    t.after(() => {
      if (previous) Object.defineProperty(globalThis, name, previous);
      else delete globalThis[name];
    });
  }
  return {
    sources,
    gains,
    listeners,
    writes,
    get context() { return context; },
    resume: () => resume(),
  };
}

test("movement starts on its current frame and shares one voice across a batch", async (t) => {
  const audio = audioHarness(t);
  const sound = createCardSound();
  assert.equal(audio.context, undefined, "page load must not construct audio");
  await sound.unlock();
  audio.context.currentTime = 2;
  sound.update(460);
  assert.deepEqual(audio.sources[0].starts, [2]);
  assert(audio.gains[0].gain.ramps[0].value > 0);
  assert(audio.gains[0].gain.ramps[0].time <= 2.02);
  for (let frame = 1; frame < 30; frame++) {
    audio.context.currentTime = 2 + frame / 60;
    sound.update(460 * Math.exp(-frame / 6));
  }
  assert.equal(audio.sources.length, 1, "a moving deck must not layer extra voices");
  sound.update(0);
  assert(audio.sources[0].stops[0] <= audio.context.currentTime + .03);
});

test("unlock does not replay movement that ended while audio was suspended", async (t) => {
  const audio = audioHarness(t, { suspended: true });
  const sound = createCardSound();
  const unlocking = sound.unlock();
  sound.update(460);
  sound.update(200);
  sound.update(0);
  assert.equal(audio.sources.length, 0);
  audio.resume();
  await unlocking;
  assert.equal(audio.sources.length, 0);
  sound.update(460);
  assert.equal(audio.sources.length, 1);
});

test("mute persists and hiding the page stops current and pending movement", async (t) => {
  const audio = audioHarness(t);
  const sound = createCardSound();
  await sound.unlock();
  sound.update(460);
  assert.equal(sound.toggle(), false);
  assert.deepEqual(audio.writes, [["avatars-card-sound", "false"]]);
  assert.equal(audio.sources[0].stops.length, 1);
  assert.equal(sound.toggle(), true);
  sound.update(460);
  document.hidden = true;
  audio.listeners.get("visibilitychange")();
  assert.equal(audio.sources[1].stops.length, 1);
  document.hidden = false;
  await sound.unlock();
  assert.equal(audio.sources.length, 2, "returning must not replay old movement");
});

test("missing storage and Web Audio leave preference controls usable", async (t) => {
  const previousStorage = Object.getOwnPropertyDescriptor(globalThis, "localStorage");
  const previousAudio = Object.getOwnPropertyDescriptor(globalThis, "AudioContext");
  Object.defineProperty(globalThis, "localStorage", { configurable: true, get() { throw new Error("blocked"); } });
  delete globalThis.AudioContext;
  t.after(() => {
    if (previousStorage) Object.defineProperty(globalThis, "localStorage", previousStorage);
    else delete globalThis.localStorage;
    if (previousAudio) Object.defineProperty(globalThis, "AudioContext", previousAudio);
  });
  const sound = createCardSound();
  await sound.unlock();
  assert.equal(sound.toggle(), false);
  sound.update(400);
  sound.stop();
});
