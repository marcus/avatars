/**
 * A quiet, procedural paper-slide voice. No audio or clock is created at import.
 *
 * Call unlock() directly from pointerdown or another trusted gesture. Call
 * update(speedInPixelsPerSecond) from the table's existing physics frames;
 * update(0) / stop() fades out. toggle() returns the new preference immediately.
 * The getter reflects that preference, including in browsers without Web Audio.
 */
export function createCardSound() {
  const preferenceKey = 'avatars-card-sound';
  let enabled = true;
  try { enabled = localStorage.getItem(preferenceKey) !== 'false'; } catch { /* Private storage is optional. */ }
  let context = null;
  let buffer = null;
  let voice = null;
  let desiredSpeed = 0;

  function hold(parameter, time) {
    if (typeof parameter.cancelAndHoldAtTime === 'function') parameter.cancelAndHoldAtTime(time);
    else {
      parameter.cancelScheduledValues(time);
      parameter.setValueAtTime(parameter.value, time);
    }
  }

  function noiseBuffer() {
    const samples = Math.round(context.sampleRate * 1.8);
    const result = context.createBuffer(1, samples, context.sampleRate);
    const data = result.getChannelData(0);
    let softNoise = 0;
    let grainEnvelope = 0;
    let seed = 0x49c8a571;
    const noise = () => {
      seed ^= seed << 13; seed ^= seed >>> 17; seed ^= seed << 5;
      return (seed >>> 0) / 2147483648 - 1;
    };
    for (let i = 0; i < samples; i++) {
      const white = noise();
      softNoise += .15 * (white - softNoise);
      grainEnvelope += .0018 * (noise() - grainEnvelope);
      // Broad rubbing noise, softened by fibers; no clicks, chimes, or synthetic tone.
      data[i] = (white * .72 + softNoise * .28) * (.8 + grainEnvelope * .7);
    }
    return result;
  }

  function beginVoice() {
    if (voice || !context || context.state !== 'running') return;
    buffer ||= noiseBuffer();
    const source = context.createBufferSource();
    const highpass = context.createBiquadFilter();
    const lowpass = context.createBiquadFilter();
    const gain = context.createGain();
    source.buffer = buffer;
    source.loop = true;
    highpass.type = 'highpass';
    highpass.frequency.value = 220;
    highpass.Q.value = .5;
    lowpass.type = 'lowpass';
    lowpass.frequency.value = 1300;
    lowpass.Q.value = .55;
    gain.gain.value = 0;
    source.connect(highpass);
    highpass.connect(lowpass);
    lowpass.connect(gain);
    gain.connect(context.destination);
    source.onended = () => {
      source.disconnect();
      highpass.disconnect();
      lowpass.disconnect();
      gain.disconnect();
    };
    voice = { source, lowpass, gain };
    source.start();
  }

  function fadeOut() {
    if (!voice || !context) return;
    const fading = voice;
    voice = null;
    const now = context.currentTime;
    try {
      hold(fading.gain.gain, now);
      fading.gain.gain.linearRampToValueAtTime(0, now + .020);
      fading.source.stop(now + .028);
    } catch { /* A closed or interrupted audio context is already silent. */ }
  }

  function applySpeed() {
    if (!enabled || desiredSpeed < .5 || (typeof document !== 'undefined' && document.hidden)) {
      fadeOut();
      return;
    }
    if (!context || context.state !== 'running') return;
    try {
      beginVoice();
      if (!voice) return;
      const intensity = Math.min(1, desiredSpeed / 850);
      const gain = .11 * intensity;
      const now = context.currentTime;
      hold(voice.gain.gain, now);
      // Follow deceleration all the way down, without a minimum-volume tail.
      voice.gain.gain.linearRampToValueAtTime(gain, now + .014);
      hold(voice.lowpass.frequency, now);
      voice.lowpass.frequency.linearRampToValueAtTime(850 + intensity * 1950, now + .016);
    } catch {
      fadeOut();
    }
  }

  async function unlock() {
    if (!enabled) return;
    try {
      if (!context) {
        const AudioContextClass = globalThis.AudioContext || globalThis.webkitAudioContext;
        if (!AudioContextClass) return;
        // Construction and resume both happen before the first await, within the gesture.
        context = new AudioContextClass({ latencyHint: 'interactive' });
      }
      if (context.state !== 'running') await context.resume();
      applySpeed();
    } catch { /* Audio permission or device failure must never affect the table. */ }
  }

  function stop() {
    desiredSpeed = 0;
    fadeOut();
  }

  function toggle() {
    enabled = !enabled;
    try { localStorage.setItem(preferenceKey, String(enabled)); } catch { /* Session preference still works. */ }
    if (enabled) void unlock();
    else stop();
    return enabled;
  }

  function update(speedPixelsPerSecond) {
    desiredSpeed = Number.isFinite(speedPixelsPerSecond) ? Math.max(0, speedPixelsPerSecond) : 0;
    applySpeed();
  }

  if (typeof document !== 'undefined') {
    document.addEventListener('visibilitychange', () => { if (document.hidden) stop(); });
  }

  return { get enabled() { return enabled; }, toggle, unlock, update, stop };
}
