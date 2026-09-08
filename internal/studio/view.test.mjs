import assert from "node:assert/strict";
import test from "node:test";
import { readExportView, writeExportView } from "./assets/view.mjs";

test("links without a shape and invalid shapes open a portrait", () => {
  for (const query of ["", "shape=square", "shape=CIRCLE"]) {
    assert.deepEqual(readExportView(new URLSearchParams(query)), { shape: "portrait", width: 256, height: 288 });
  }
});

test("circle links use square defaults and retain explicit export dimensions", () => {
  assert.deepEqual(readExportView(new URLSearchParams("shape=circle")), { shape: "circle", width: 256, height: 256 });
  assert.deepEqual(readExportView(new URLSearchParams("shape=circle&width=512&height=384")), { shape: "circle", width: 512, height: 384 });
  assert.deepEqual(readExportView(new URLSearchParams("shape=circle&width=1024")), { shape: "circle", width: 1024, height: 1024 });
});

test("invalid dimensions fall back to usable dimensions", () => {
  for (const value of ["", "0", "-1", "2049", "4.5", "NaN", "1e3"]) {
    assert.deepEqual(readExportView(new URLSearchParams({ shape: "circle", width: value, height: value })), { shape: "circle", width: 256, height: 256 });
  }
});

test("copied views round-trip both shapes and preserve collection context", () => {
  const original = new URLSearchParams("collection=col_example&avatar=av_example");
  for (const shape of ["portrait", "circle"]) {
    const view = { shape, width: 1024, height: 768 };
    const share = writeExportView(original, view);
    assert.equal(share.get("collection"), "col_example");
    assert.equal(share.get("avatar"), "av_example");
    assert.equal(share.get("shape"), shape);
    assert.deepEqual(readExportView(share), view);
  }
  assert.equal(original.has("shape"), false, "serializing a view does not mutate its source parameters");
});
