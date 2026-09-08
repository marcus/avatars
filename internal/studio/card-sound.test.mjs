import test from 'node:test';
import assert from 'node:assert/strict';
import { createCardSound } from './assets/card-sound.mjs';

function audioHarness(t, muted = false) {
  const sources = [], gains = [];
  const parameter = () => ({ value: 0, ramps: [], cancelAndHoldAtTime() {}, linearRampToValueAtTime(value) { this.ramps.push(value); } });
  const node = () => ({ connect() {}, disconnect() {} });
  class AudioContext {
    state = 'running';
    sampleRate = 1000;
    currentTime = 0;
    destination = {};
    createBuffer(channels, size) { return { getChannelData: () => new Float32Array(size) }; }
    createBufferSource() {
      const source = { ...node(), starts: 0, stops: 0, start() { this.starts++; }, stop() { this.stops++; } };
      sources.push(source);
      return source;
    }
    createBiquadFilter() { return { ...node(), frequency: parameter(), Q: { value: 0 } }; }
    createGain() { const gain = { ...node(), gain: parameter() }; gains.push(gain); return gain; }
  }
  const replacements = {
    AudioContext,
    localStorage: { getItem: () => String(!muted), setItem() {} },
  };
  for (const [key, value] of Object.entries(replacements)) {
    const previous = Object.getOwnPropertyDescriptor(globalThis, key);
    Object.defineProperty(globalThis, key, { configurable: true, value });
    t.after(() => previous ? Object.defineProperty(globalThis, key, previous) : delete globalThis[key]);
  }
  return { sources, gains };
}

test('many motion updates use one slide voice and deceleration fades all the way to zero', async t => {
  const { sources, gains } = audioHarness(t);
  const sound = createCardSound();
  await sound.unlock();
  for (const speed of [420, 340, 220, 80, 10, 1]) sound.update(speed);
  assert.equal(sources.length, 1);
  assert.equal(sources[0].starts, 1);
  const levels = gains[0].gain.ramps;
  assert.ok(levels.every((level, index) => level > 0 && (!index || level < levels[index - 1])));
  sound.update(0);
  assert.equal(gains[0].gain.ramps.at(-1), 0);
  assert.equal(sources[0].stops, 1);
});

test('muted motion stays silent and muting an active slide stops its voice', async t => {
  const { sources } = audioHarness(t, true);
  const sound = createCardSound();
  await sound.unlock();
  sound.update(420);
  assert.equal(sources.length, 0);
  sound.toggle();
  sound.update(200);
  assert.equal(sources.length, 1);
  sound.toggle();
  assert.equal(sources[0].stops, 1);
  sound.update(420);
  assert.equal(sources.length, 1);
});
