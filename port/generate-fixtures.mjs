// Run from any directory with Node 22.18+ to refresh Go parity fixtures.
import { agentPortrait } from './agent-portrait.ts';
import { createHash } from 'node:crypto';
import { writeFileSync } from 'node:fs';
const seeds = [...Array.from({ length: 100 }, (_, i) => `agent-${i}`), '', 'marcus', 'é', 'e\u0301', '你好', '🦉🧑🏽‍🎨', 'a\0b', '<script>alert("x")</script><svg onload="evil()">'];
function features(seed) {
  let hash = 2166136261;
  for (const char of seed) hash = Math.imul(hash ^ char.codePointAt(0), 16777619) >>> 0;
  const initial = hash;
  const next = max => { hash ^= hash << 13; hash ^= hash >>> 17; hash ^= hash << 5; return (hash >>> 0) % max; };
  const outfit = (Math.imul(initial ^ 0x6d2b79f5, 0x45d9f3b) >>> 0) % 6;
  next(4); next(6); next(3); next(4);
  return { outfit, hair: next(5), accessory: next(6) };
}
const fixtures = seeds.map(seed => ({ seed, sha256: createHash('sha256').update(agentPortrait(seed)).digest('hex'), ...features(seed) }));
writeFileSync(new URL('../pkg/avatar/testdata/reference.json', import.meta.url), JSON.stringify(fixtures, null, 2) + '\n');
