import assert from "node:assert/strict";
import test from "node:test";
import { generationInputsForStyle, inputChoiceForRecipe, inputFor, inputValueForStyle, readExportView, writeExportView } from "./assets/view.mjs";

const styles = [
  { id: "companions", native_width: 128, native_height: 128, inputs: { animal: { default: "dog", values: [
    { value: "dog", label: "Dogs" }, { value: "cat", label: "Cats" },
  ] } } },
  { id: "gorey", name: "Gorey" },
  { id: "pebble", name: "Pebble", inputs: { color: { default: "walnut", values: [
    { value: "walnut", label: "Walnut", swatch: "#92744F" },
    { value: "sage", label: "Sage", swatch: "#A5B59A" },
  ] } } },
];

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

test("style color metadata resolves defaults without leaking into other styles", () => {
  assert.equal(inputFor(styles, "gorey", "color"), null);
  assert.equal(inputValueForStyle(styles, "gorey", "color", "sage"), null);
  assert.equal(inputValueForStyle(styles, "pebble", "color", ""), "walnut");
  assert.equal(inputValueForStyle(styles, "pebble", "color", "sage"), "sage");
  assert.equal(inputValueForStyle(styles, "pebble", "color", "missing"), "walnut");
});

test("saved colors restore a readable inspector choice, including old default recipes", () => {
  assert.deepEqual(inputChoiceForRecipe(styles, { style: "pebble", inputs: { color: "sage" } }, "color"), { value: "sage", label: "Sage", swatch: "#A5B59A" });
  assert.deepEqual(inputChoiceForRecipe(styles, { style: "pebble" }, "color"), { value: "walnut", label: "Walnut", swatch: "#92744F" });
  assert.equal(inputChoiceForRecipe(styles, { style: "gorey", inputs: { color: "sage" } }, "color"), null);
});

test("animal choices restore saved recipes and exclude unsupported stale inputs", () => {
  assert.equal(inputFor(styles, "pebble", "animal"), null);
  assert.deepEqual(inputChoiceForRecipe(styles, { style: "companions", inputs: { animal: "cat" } }, "animal"), { value: "cat", label: "Cats" });
  assert.deepEqual(inputChoiceForRecipe(styles, { style: "companions" }, "animal"), { value: "dog", label: "Dogs" });
  const stale = { color: "sage", animal: "cat" };
  assert.deepEqual(generationInputsForStyle(styles, "companions", stale), { animal: "cat" });
  assert.deepEqual(generationInputsForStyle(styles, "pebble", stale), { color: "sage" });
  assert.deepEqual(generationInputsForStyle(styles, "gorey", stale), {});
  assert.deepEqual(generationInputsForStyle(styles, "companions", { animal: "missing" }), { animal: "dog" });
});

test("native size drives default framing while copied dimensions take precedence", () => {
  const style = styles[0];
  assert.deepEqual(readExportView(new URLSearchParams(), style), { shape: "portrait", width: 256, height: 256 });
  assert.deepEqual(readExportView(new URLSearchParams("width=512"), style), { shape: "portrait", width: 512, height: 512 });
  assert.deepEqual(readExportView(new URLSearchParams("width=512&height=288"), style), { shape: "portrait", width: 512, height: 288 });
});
