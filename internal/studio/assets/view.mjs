// Shareable presentation state. These values never change the saved portrait.
export function readExportView(params) {
  const shape = params.get("shape") === "circle" ? "circle" : "portrait";
  const dimension = (value, fallback) => {
    if (!/^\d+$/.test(value || "")) return fallback;
    const number = Number(value);
    return Number.isInteger(number) && number >= 1 && number <= 2048 ? number : fallback;
  };
  const width = dimension(params.get("width"), 256);
  const height = dimension(params.get("height"), shape === "circle" ? width : Math.min(2048, Math.round(width * 9 / 8)));
  return { shape, width, height };
}

export function writeExportView(params, view) {
  const result = new URLSearchParams(params);
  const normalized = readExportView(new URLSearchParams(view));
  for (const [key, value] of Object.entries(normalized)) result.set(key, String(value));
  return result;
}
