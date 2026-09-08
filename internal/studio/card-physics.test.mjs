import { test } from 'node:test';
import assert from 'node:assert/strict';
import { createCardPhysics } from './assets/card-physics.mjs';

const bounds = { left: 0, top: 0, right: 1200, bottom: 800 };
const card = (id, x, y = 300, angle = 0) => ({ id, x, y, width: 100, height: 140, angle });
const byId = (physics, id) => physics.step(0).find(item => item.id === id);
function advance(physics, seconds, delta = 1 / 60) {
  for (let elapsed = 0; elapsed < seconds - 1e-9; elapsed += delta) physics.step(delta);
}

test('turning a card stops its slide at the current center and leaves a small physical angle change', () => {
  const physics = createCardPhysics();
  physics.reset([card('turning', 300), card('neighbor', 800)], bounds);
  physics.grab('turning', 300, 300);
  physics.release(400, 0);
  advance(physics, .15);
  const before = byId(physics, 'turning');
  const neighbor = byId(physics, 'neighbor');
  physics.turn('turning', .012);
  advance(physics, 1);
  const after = byId(physics, 'turning');
  assert(Math.abs(after.x - before.x) < .01);
  assert(Math.abs(after.y - before.y) < .01);
  assert(Math.abs(after.angle - before.angle - .012) < .0001);
  assert.deepEqual(byId(physics, 'neighbor'), neighbor);
  assert.equal(physics.isActive(), false);
});

test('deals rest exactly; a throw transfers motion to a neighboring sleeping card', () => {
  const physics = createCardPhysics();
  physics.reset([card('first', 200), card('second', 410)], bounds);
  assert.equal(physics.isActive(), false);
  assert.equal(physics.grab('missing', 0, 0), false);
  physics.grab('first', 200, 300);
  physics.release(1000, 0);
  advance(physics, .6);
  assert(byId(physics, 'second').x > 440, 'the collision must move the neighboring card');
  assert(byId(physics, 'first').x < byId(physics, 'second').x, 'cards should not pass through each other');
});

test('a sliding card decelerates and eventually sleeps without a perpetual animation loop', () => {
  const physics = createCardPhysics();
  physics.reset([card('card', 200)], bounds);
  physics.grab('card', 200, 300);
  physics.release(600, 0);
  advance(physics, .2);
  const firstDistance = byId(physics, 'card').x - 200;
  const before = byId(physics, 'card').x;
  advance(physics, .2);
  assert(byId(physics, 'card').x - before < firstDistance);
  advance(physics, 8);
  assert.equal(physics.isActive(), false);
  const resting = byId(physics, 'card');
  advance(physics, 1);
  assert.deepEqual(byId(physics, 'card'), resting);
});

test('bounded high-speed throws remain inside enclosing walls at maximum caller delta', () => {
  const physics = createCardPhysics();
  physics.reset([card('card', 1000, 640, .1)], bounds);
  physics.grab('card', 1000, 640);
  physics.release(100000, 70000);
  let bounced = false;
  let lastX = 1000;
  for (let i = 0; i < 100; i++) {
    const pose = physics.step(.05)[0];
    const rx = (100 * Math.abs(Math.cos(pose.angle)) + 140 * Math.abs(Math.sin(pose.angle))) / 2;
    const ry = (140 * Math.abs(Math.cos(pose.angle)) + 100 * Math.abs(Math.sin(pose.angle))) / 2;
    assert(pose.x - rx >= -.5 && pose.x + rx <= bounds.right + .5, 'horizontal wall containment');
    assert(pose.y - ry >= -.5 && pose.y + ry <= bounds.bottom + .5, 'vertical wall containment');
    if (pose.x < lastX) bounced = true;
    lastX = pose.x;
  }
  assert(bounced, 'the card should respond to the wall, not cross it');
});

test('off-center dragging follows the grab point and produces restrained rotation', () => {
  const physics = createCardPhysics();
  physics.reset([card('card', 300)], bounds);
  physics.grab('card', 270, 250);
  assert.equal(physics.isActive(), true);
  physics.move(510, 310);
  advance(physics, .5);
  const pose = byId(physics, 'card');
  assert(pose.x > 450);
  assert(Math.abs(pose.angle) > .01 && Math.abs(pose.angle) < Math.PI);
});

test('cancellation stops the held card; a zero-speed release does not invent a throw', () => {
  for (const cancelled of [true, false]) {
    const physics = createCardPhysics();
    physics.reset([card('card', 300)], bounds);
    physics.grab('card', 300, 300);
    physics.move(450, 330);
    advance(physics, .05);
    physics.release(cancelled ? 900 : 0, cancelled ? 100 : 0, cancelled);
    const released = byId(physics, 'card');
    advance(physics, .5);
    assert.deepEqual(byId(physics, 'card'), released);
    assert.equal(physics.isActive(), false);
  }
});

test('resize reset discards old bodies, active drag, and all momentum', () => {
  const physics = createCardPhysics();
  physics.reset([card('old', 200)], bounds);
  physics.grab('old', 200, 300);
  physics.move(700, 400);
  advance(physics, .1);
  physics.reset([card('new', 150, 150, .05)], { left: 0, top: 0, right: 400, bottom: 400 });
  physics.release(1200, 0);
  advance(physics, 1);
  assert.deepEqual(physics.step(0), [{ id: 'new', x: 150, y: 150, angle: .05 }]);
  assert.equal(physics.isActive(), false);
});

test('fixed internal stepping gives the same motion for 20Hz and 60Hz callers', () => {
  const trajectories = [1 / 60, .05].map(delta => {
    const physics = createCardPhysics();
    physics.reset([card('card', 200)], bounds);
    physics.grab('card', 200, 300);
    physics.release(700, 0);
    advance(physics, .6, delta);
    return byId(physics, 'card');
  });
  assert(Math.abs(trajectories[0].x - trajectories[1].x) < 1e-7);
});


test('dragging an off-center grab far outside the table keeps the whole card enclosed', () => {
  const physics = createCardPhysics();
  physics.reset([card('card', 200, 300, .1)], bounds);
  physics.grab('card', 160, 250);
  physics.move(5000, -5000);
  for (let i = 0; i < 120; i++) {
    const pose = physics.step(1 / 60)[0];
    const rx = (100 * Math.abs(Math.cos(pose.angle)) + 140 * Math.abs(Math.sin(pose.angle))) / 2;
    const ry = (140 * Math.abs(Math.cos(pose.angle)) + 100 * Math.abs(Math.sin(pose.angle))) / 2;
    assert(pose.x - rx >= -1e-5 && pose.x + rx <= bounds.right + 1e-5);
    assert(pose.y - ry >= -1e-5 && pose.y + ry <= bounds.bottom + 1e-5);
  }
});
