// Shareable presentation state. These values never change the saved portrait.
export function readExportView(params, style) {
  const shape = params.get("shape") === "circle" ? "circle" : "portrait";
  const dimension = (value, fallback) => {
    if (!/^\d+$/.test(value || "")) return fallback;
    const number = Number(value);
    return Number.isInteger(number) && number >= 1 && number <= 2048 ? number : fallback;
  };
  const width = dimension(params.get("width"), 256);
  const height = dimension(params.get("height"), shape === "circle" ? width : Math.min(2048, Math.round(width * nativeRatio(style))));
  return { shape, width, height };
}

export function writeExportView(params, view) {
  const result = new URLSearchParams(params);
  const normalized = readExportView(new URLSearchParams(view));
  for (const [key, value] of Object.entries(normalized)) result.set(key, String(value));
  return result;
}

// Native artwork dimensions drive default presentation; saved explicit dimensions
// take precedence. Older/custom descriptors keep the historical portrait ratio.
export function nativeRatio(style) {
  const { native_width: width, native_height: height } = style || {};
  return width > 0 && height > 0 ? height / width : 9 / 8;
}

export function inputFor(styles, styleId, name) {
  const input = styles.find((style) => style.id === styleId)?.inputs?.[name];
  if (!input || !Array.isArray(input.values) || !input.values.length) return null;
  return input;
}

export function inputValueForStyle(styles, styleId, name, requested) {
  const input = inputFor(styles, styleId, name);
  if (!input) return null;
  return input.values.some((choice) => choice.value === requested) ? requested : input.default;
}

export function inputChoiceForRecipe(styles, avatar, name) {
  const input = inputFor(styles, avatar?.style, name);
  if (!input) return null;
  const value = inputValueForStyle(styles, avatar.style, name, avatar.inputs?.[name]);
  return input.values.find((choice) => choice.value === value) || null;
}

// This is the same narrow pair of typed inputs as the API. Descriptors provide
// values and defaults; unsupported controls can never leak stale request fields.
export function generationInputsForStyle(styles, styleId, requested = {}) {
  const inputs = {};
  for (const name of ["color", "animal"]) {
    const value = inputValueForStyle(styles, styleId, name, requested[name]);
    if (value) inputs[name] = value;
  }
  return inputs;
}

// Preview surrounds are presentation state, never renderer inputs.
export function readPreviewBackground(params) {
  const value = params.get("background");
  return ["dark", "light", "gray"].includes(value) ? value : "dark";
}

// A mixed saved collection suggests Random when that input supports it; a
// uniform collection restores its concrete value, including older defaults.
export function generationInputsForCollection(styles, collection) {
  if (!collection) return {};
  const inputs = generationInputsForStyle(styles, collection.style, collection.avatars?.[0]?.inputs);
  for (const name of Object.keys(inputs)) {
    const values = new Set((collection.avatars || []).map((avatar) => inputValueForStyle(styles, collection.style, name, avatar.inputs?.[name])));
    if (values.size > 1) inputs[name] = inputValueForStyle(styles, collection.style, name, "random");
  }
  return inputs;
}
