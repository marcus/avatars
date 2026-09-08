import assert from "node:assert/strict";
import test from "node:test";
import { chooseOpenSlot, deterministicCardAngle, reconcileKeyedChildren } from "./assets/card-grid.mjs";

class FakeNode {
  constructor(id) {
    this.dataset = { avatar: id };
    this.parentElement = null;
    this.updated = 0;
  }

  get nextElementSibling() {
    if (!this.parentElement) return null;
    const index = this.parentElement.children.indexOf(this);
    return this.parentElement.children[index + 1] || null;
  }

  remove() {
    if (!this.parentElement) return;
    const index = this.parentElement.children.indexOf(this);
    this.parentElement.children.splice(index, 1);
    this.parentElement = null;
  }
}

class FakeContainer {
  constructor(nodes) {
    this.children = nodes;
    for (const node of nodes) node.parentElement = this;
  }

  get firstElementChild() { return this.children[0] || null; }

  insertBefore(node, cursor) {
    if (node.parentElement) {
      const oldIndex = node.parentElement.children.indexOf(node);
      node.parentElement.children.splice(oldIndex, 1);
    }
    const index = cursor ? this.children.indexOf(cursor) : this.children.length;
    this.children.splice(index, 0, node);
    node.parentElement = this;
  }
}

test("keyed refreshes preserve existing card elements while adding and removing avatars", () => {
  const first = new FakeNode("first");
  const second = new FakeNode("second");
  const container = new FakeContainer([first, second]);
  const created = [];
  const items = [{ id: "new" }, { id: "second" }];
  const result = reconcileKeyedChildren(
    container,
    items,
    (item) => item.id,
    (item) => {
      const node = new FakeNode(item.id);
      created.push(node);
      return node;
    },
    (node) => { node.updated += 1; },
  );

  assert.deepEqual(container.children.map((node) => node.dataset.avatar), ["new", "second"]);
  assert.equal(result[1], second, "the unchanged avatar keeps its DOM identity and physical state");
  assert.equal(second.updated, 1);
  assert.equal(first.parentElement, null);
  assert.equal(created.length, 1);
});

test("card angles are deterministic, subtle, and varied", () => {
  const ids = ["av_one", "av_two", "av_three", "av_four"];
  const angles = ids.map(deterministicCardAngle);
  assert.deepEqual(ids.map(deterministicCardAngle), angles);
  assert(angles.every((angle) => Math.abs(angle) <= .08));
  assert(new Set(angles).size > 1);
});

test("new cards choose a slot clear of complete displaced card footprints", () => {
  const slots = [
    { x: 80, y: 110, width: 140, height: 200, angle: 0 },
    { x: 240, y: 110, width: 140, height: 200, angle: 0 },
    { x: 400, y: 110, width: 140, height: 200, angle: 0 },
  ];
  const occupied = [
    { id: "first", x: 80, y: 110, width: 140, height: 200, angle: 0 },
    { id: "second", x: 236, y: 116, width: 140, height: 200, angle: .08 },
  ];
  assert.equal(chooseOpenSlot(slots, occupied, slots[0]), slots[2]);
});
