import { test } from 'node:test';
import assert from 'node:assert/strict';
import { agentPortrait } from './agent-portrait.ts';

test('the same agent identity always receives the same static portrait', () => {
	const original = agentPortrait('agent-constant');
	for (let i = 0; i < 10; i++) agentPortrait(`unrelated-${i}`);
	assert.equal(agentPortrait('agent-constant'), original);
	assert.match(original, /viewBox="0 0 64 72"/);
	assert.doesNotMatch(original, /<animate|<script|<foreignObject|\bon\w+=|https?:\/\/(?!www\.w3\.org)/i);
});

test('different identities produce a varied portrait collection', () => {
	const portraits = Array.from({ length: 100 }, (_, i) => agentPortrait(`agent-${i}`));
	assert.equal(new Set(portraits).size, 100);
});

test('agent identity text never enters the SVG markup', () => {
	const portrait = agentPortrait('\"><script>alert(1)</script><image href="https://example.com/">');
	assert.doesNotMatch(portrait, /script|alert|image|example\.com/);
	assert.match(portrait, /^<svg /);
	assert.match(portrait, /<\/svg>$/);
});
