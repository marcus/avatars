import { generationInputsForStyle, inputChoiceForRecipe, inputFor, nativeRatio, readExportView, readPreviewBackground, writeExportView } from "/view.mjs";

const $ = (id) => document.getElementById(id);
const state = {
  collections: [], styles: [], formats: [], collectionId: null, avatarId: null,
  selected: null, loaded: false, generating: false, refreshing: false,
  librarySignature: "", gridSignature: "", request: 0, routeRequest: 0,
  inputsByStyle: {}, previewBackground: "dark",
};
let toastTimer;
let lastSelected = null;
const date = new Intl.DateTimeFormat(undefined, { month: "short", day: "numeric" });
const dateTime = new Intl.DateTimeFormat(undefined, { month: "short", day: "numeric", hour: "numeric", minute: "2-digit" });

function formatDate(value, full = false) {
  const parsed = new Date(value);
  return Number.isNaN(parsed.valueOf()) ? "" : (full ? dateTime : date).format(parsed);
}

function icon(name) {
  const svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
  svg.classList.add("icon");
  svg.setAttribute("aria-hidden", "true");
  const use = document.createElementNS("http://www.w3.org/2000/svg", "use");
  use.setAttribute("href", `#i-${name}`);
  svg.append(use);
  return svg;
}

function element(tag, className, text) {
  const node = document.createElement(tag);
  if (className) node.className = className;
  if (text !== undefined) node.textContent = text;
  return node;
}

async function api(path, options = {}) {
  const response = await fetch(path, {
    cache: "no-store", credentials: "same-origin", signal: AbortSignal.timeout(30000),
    ...options, headers: { Accept: "application/json", ...options.headers },
  });
  const body = await response.json();
  if (!response.ok) throw new Error(body.error?.message || `Request failed (${response.status}).`);
  return body;
}

function explainError(error) {
  if (error.name === "TimeoutError") return "The studio took too long to respond. Try again.";
  if (error instanceof TypeError) return "Couldn't reach the studio. Check that the server is running and try again.";
  return error.message || "Something went wrong. Try again.";
}

function showNotice(message, source = "general") {
  $("notice-text").textContent = message;
  $("notice").dataset.source = source;
  $("notice").hidden = false;
}

function toast(message) {
  clearTimeout(toastTimer);
  $("toast-text").textContent = message;
  $("toast").hidden = false;
  toastTimer = setTimeout(() => { $("toast").hidden = true; }, 3000);
}

function connection(connected) {
  const status = $("connection-status");
  status.classList.toggle("offline", !connected);
  status.lastElementChild.textContent = connected ? "Local workspace" : "Connection interrupted";
  status.title = connected ? "Collections refresh automatically" : "Trying to reconnect. You can also refresh the library.";
}

function routeURL(collectionId = null, avatarId = null) {
  let params = new URLSearchParams();
  if (collectionId) params.set("collection", collectionId);
  if (avatarId) {
    params.set("avatar", avatarId);
    const style = state.styles.find((item) => item.id === avatarFor(avatarId)?.style);
    const view = state.selected ? readExportControls() : null;
    params = writeExportView(params, view || readExportView(new URLSearchParams(), style));
    params.set("background", state.previewBackground);
  }
  return params.size ? `/?${params}` : "/";
}

function collectionFor(id) {
  return state.collections.find((collection) => collection.id === id);
}

function avatarFor(id) {
  for (const collection of state.collections) {
    const avatar = collection.avatars.find((item) => item.id === id);
    if (avatar) return avatar;
  }
  return null;
}

function styleName(id) {
  return state.styles.find((style) => style.id === id)?.name || id;
}

function portraitNumber(avatar) {
  const index = collectionFor(avatar.collection_id)?.avatars.findIndex((item) => item.id === avatar.id);
  return index >= 0 ? String(index + 1).padStart(2, "0") : "";
}

function visibleAvatars() {
  const collections = state.collectionId ? [collectionFor(state.collectionId)].filter(Boolean) : state.collections;
  return collections.flatMap((collection) => collection.avatars);
}

function imageURL(avatar, format = "svg", options = {}) {
  const url = new URL(avatar[`${format}_url`], location.origin);
  if (url.origin !== location.origin) throw new Error("The portrait has an invalid image URL.");
  for (const [key, value] of Object.entries(options)) url.searchParams.set(key, String(value));
  return url.href;
}

function watchImage(image) {
  image.addEventListener("error", () => { image.dataset.failed = "true"; });
  image.addEventListener("load", () => {
    delete image.dataset.failed;
    delete image.dataset.retries;
  });
}

function retryFailedImages(manual = false) {
  for (const image of document.querySelectorAll("img[data-failed]")) {
    const attempts = manual ? 0 : Number(image.dataset.retries || 0);
    if (attempts >= 3) continue;
    image.dataset.retries = String(attempts + 1);
    delete image.dataset.failed;
    const source = image.src;
    image.removeAttribute("src");
    image.src = source;
  }
}

function setLibraryOpen(open) {
  $("library").classList.toggle("open", open);
  $("sidebar-scrim").hidden = !open;
  $("menu-toggle").setAttribute("aria-expanded", String(open));
  $("menu-toggle").setAttribute("aria-label", open ? "Close library" : "Show library");
}

function renderLibrary() {
  const signature = JSON.stringify(state.collections.map((collection) => [collection.id, collection.name, collection.avatars.length, collection.avatars[0]?.id]));
  if (signature !== state.librarySignature) {
    const focused = document.activeElement?.dataset.collection;
    const fragment = document.createDocumentFragment();
    for (const collection of state.collections) {
      const link = element("a", "nav-item collection-item");
      link.href = routeURL(collection.id);
      link.dataset.nav = "";
      link.dataset.collection = collection.id;
      link.title = collection.name;
      if (collection.avatars[0]) {
        const thumbnail = element("img", "collection-thumbnail");
        watchImage(thumbnail);
        thumbnail.src = imageURL(collection.avatars[0]);
        thumbnail.alt = "";
        thumbnail.loading = "lazy";
        link.append(thumbnail);
      } else link.append(icon("folder"));
      const copy = element("span", "collection-copy");
      copy.append(element("span", "collection-name", collection.name));
      copy.append(element("span", "collection-meta", `${collection.avatars.length} ${collection.avatars.length === 1 ? "portrait" : "portraits"} · ${formatDate(collection.created_at)}`));
      link.append(copy);
      fragment.append(link);
    }
    if (!state.collections.length) fragment.append(element("p", "sidebar-empty", "Your collections will appear here."));
    $("collections").replaceChildren(fragment);
    state.librarySignature = signature;
    if (focused) [...$("collections").querySelectorAll("a")].find((link) => link.dataset.collection === focused)?.focus();
  }
  for (const link of $("collections").querySelectorAll("a")) {
    const active = link.dataset.collection === state.collectionId;
    link.classList.toggle("active", active);
    if (active) link.setAttribute("aria-current", "page");
    else link.removeAttribute("aria-current");
  }
  const all = document.querySelector(".library-all");
  all.classList.toggle("active", !state.collectionId);
  if (!state.collectionId) all.setAttribute("aria-current", "page");
  else all.removeAttribute("aria-current");
  $("library-count").textContent = state.collections.reduce((sum, collection) => sum + collection.avatars.length, 0);
}

function renderGrid() {
  const avatars = visibleAvatars();
  const collection = collectionFor(state.collectionId);
  const signature = JSON.stringify([state.collectionId, avatars.map((avatar) => [avatar.id, collectionFor(avatar.collection_id)?.name])]);
  if (signature !== state.gridSignature) {
    const focused = document.activeElement?.dataset.avatar;
    const fragment = document.createDocumentFragment();
    for (const avatar of avatars) {
      const number = portraitNumber(avatar);
      const card = element("button", "portrait-card");
      card.type = "button";
      card.dataset.avatar = avatar.id;
      card.setAttribute("aria-label", `Portrait ${number}, ${collectionFor(avatar.collection_id)?.name || styleName(avatar.style)}`);
      const mat = element("span", "portrait-mat");
      const image = element("img", "portrait-image");
      watchImage(image);
      image.src = imageURL(avatar);
      image.alt = "";
      image.width = 192;
      image.height = 216;
      image.loading = "lazy";
      image.decoding = "async";
      const mark = element("span", "selection-mark");
      mark.append(icon("check"));
      mat.append(image, mark);
      const label = element("span", "portrait-label");
      label.append(element("span", "", `Portrait ${number}`), element("span", "portrait-number", avatar.id.slice(-5).toUpperCase()));
      card.append(mat, label);
      if (!state.collectionId) card.append(element("span", "portrait-source", collectionFor(avatar.collection_id)?.name || styleName(avatar.style)));
      fragment.append(card);
    }
    $("portrait-grid").replaceChildren(fragment);
    state.gridSignature = signature;
    if (focused) [...$("portrait-grid").children].find((card) => card.dataset.avatar === focused)?.focus({ preventScroll: true });
  }
  for (const card of $("portrait-grid").children) card.setAttribute("aria-pressed", String(card.dataset.avatar === state.avatarId));
  $("view-eyebrow").textContent = state.collectionId ? "COLLECTION" : "YOUR WORKSPACE";
  $("view-title").textContent = collection?.name || (state.collectionId ? "Collection unavailable" : "All portraits");
  $("view-description").textContent = collection
    ? `${avatars.length} ${avatars.length === 1 ? "portrait" : "portraits"} · ${styleName(collection.style)} · ${formatDate(collection.created_at)}`
    : `${state.collections.length} ${state.collections.length === 1 ? "collection" : "collections"} · Endless character`;
  $("copy-collection").hidden = !collection;
  $("footer-count").textContent = `${avatars.length} ${avatars.length === 1 ? "portrait" : "portraits"}`;
  $("portrait-grid").setAttribute("aria-busy", "false");
  $("load-state").hidden = true;
  $("empty-state").hidden = avatars.length > 0 || Boolean(state.collectionId);
  $("portrait-grid").hidden = !avatars.length;
  document.title = state.selected ? `Portrait ${portraitNumber(state.selected)} · Avatars Studio` : `${collection?.name || "Avatars"} · Studio`;
}

function renderInspector() {
  const avatar = state.selected;
  const inspector = $("inspector");
  inspector.classList.toggle("has-selection", Boolean(avatar));
  $("inspector-empty").hidden = Boolean(avatar);
  $("inspector-content").hidden = !avatar;
  $("close-inspector").hidden = !avatar;
  if (!avatar) { lastSelected = null; return; }
  const collection = collectionFor(avatar.collection_id);
  $("portrait-title").textContent = `Portrait ${portraitNumber(avatar)}`;
  $("portrait-collection").textContent = collection?.name || "View collection";
  $("portrait-collection").href = routeURL(avatar.collection_id);
  $("portrait-style").textContent = styleName(avatar.style);
  for (const name of ["color", "animal"]) {
    const choice = inputChoiceForRecipe(state.styles, avatar, name);
    $(`portrait-${name}-row`).hidden = !choice;
    $(`portrait-${name}`).textContent = choice?.label || "";
  }
  $("portrait-date").textContent = formatDate(avatar.created_at, true);
  $("portrait-date").title = new Date(avatar.created_at).toLocaleString();
  $("preview-caption").textContent = `${styleName(avatar.style).toUpperCase()} / ${["gorey", "gorey-expanded"].includes(avatar.style) ? "PEN & INK" : "PORTRAIT"}`;
  $("preview").alt = `${styleName(avatar.style)} portrait ${portraitNumber(avatar)}`;
  if (lastSelected !== avatar.id) {
    $("export-error").hidden = true;
    lastSelected = avatar.id;
    inspector.scrollTop = 0;
  }
  updatePreview();
  for (const button of $("export-form").querySelectorAll("button[name=format]")) button.disabled = !state.formats.includes(button.value);
}

function render() {
  renderLibrary();
  renderGrid();
  renderInspector();
}

async function resolveRoute({ scroll = false, restoreView = true } = {}) {
  const sequence = ++state.routeRequest;
  const params = new URLSearchParams(location.search);
  state.collectionId = params.get("collection");
  state.avatarId = params.get("avatar");
  state.selected = null;
  try {
    if (state.collectionId && !collectionFor(state.collectionId)) {
      const collection = await api(`/api/v1/collections/${encodeURIComponent(state.collectionId)}`);
      if (sequence !== state.routeRequest) return;
      state.collections.push(collection);
    }
    if (state.avatarId) {
      const avatar = avatarFor(state.avatarId) || await api(`/api/v1/avatars/${encodeURIComponent(state.avatarId)}`);
      if (sequence !== state.routeRequest) return;
      state.selected = avatar;
      if (!collectionFor(avatar.collection_id)) {
        const collection = await api(`/api/v1/collections/${encodeURIComponent(avatar.collection_id)}`);
        if (sequence !== state.routeRequest) return;
        state.collections.push(collection);
      }
    }
  } catch (error) {
    if (sequence !== state.routeRequest) return;
    showNotice(explainError(error));
  }
  if (sequence !== state.routeRequest) return;
  if (restoreView) {
    const style = state.selected?.style || collectionFor(state.collectionId)?.style;
    if (state.styles.some((item) => item.id === style)) $("style").value = style;
    const recipe = state.selected || collectionFor(state.collectionId)?.avatars[0];
    renderGenerationInputs(recipe ? generationInputsForStyle(state.styles, style, recipe.inputs) : undefined);
    if (state.avatarId) {
      const view = readExportView(params, state.styles.find((item) => item.id === state.selected?.style));
      $("export-form").elements.shape.value = view.shape;
      $("export-width").value = view.width;
      $("export-height").value = view.height;
      setPreviewBackground(readPreviewBackground(params));
    }
  }
  render();
  if (scroll && state.avatarId) [...$("portrait-grid").children].find((card) => card.dataset.avatar === state.avatarId)?.scrollIntoView({ block: "nearest" });
}

function navigate(url) {
  const target = new URL(url, location.origin);
  if (target.origin !== location.origin) return;
  const previousCollection = state.collectionId;
  history.pushState(null, "", target.pathname + target.search);
  setLibraryOpen(false);
  $("notice").hidden = true;
  void resolveRoute();
  if (previousCollection !== new URLSearchParams(target.search).get("collection")) $("canvas-scroll").scrollTop = 0;
}

function renderStyles() {
  const previous = $("style").value;
  $("style").replaceChildren(...state.styles.map((style) => {
    const option = element("option", "", style.name);
    option.value = style.id;
    option.title = style.description;
    return option;
  }));
  if (state.styles.some((style) => style.id === previous)) $("style").value = previous;
  else if (state.styles.some((style) => style.id === "gorey")) $("style").value = "gorey";
  $("style").disabled = !state.styles.length;
  renderGenerationInputs();
  setGenerating(state.generating);
}

function renderGenerationInputs(preferred = {}) {
  const styleId = $("style").value;
  const selected = generationInputsForStyle(state.styles, styleId, {
    ...state.inputsByStyle[styleId], ...preferred,
  });
  for (const name of ["color", "animal"]) {
    const input = inputFor(state.styles, styleId, name);
    $(`${name}-field`).hidden = !input;
    $(name).replaceChildren(...(input?.values || []).map((choice) => {
      const option = element("option", "", choice.label);
      option.value = choice.value;
      return option;
    }));
    if (input) $(name).value = selected[name];
  }
  state.inputsByStyle[styleId] = selected;
  updateColorSwatch();
}

function updateColorSwatch() {
  const input = inputFor(state.styles, $("style").value, "color");
  const choice = input?.values.find((item) => item.value === $("color").value);
  if (choice) $("color-swatch").style.setProperty("--swatch", choice.swatch);
  else $("color-swatch").style.removeProperty("--swatch");
}

function setGenerating(generating) {
  state.generating = generating;
  $("generate").disabled = generating || !state.styles.length;
  $("empty-generate").disabled = generating || !state.styles.length;
  $("generate").classList.toggle("is-generating", generating);
  $("generate").lastElementChild.textContent = generating ? "Creating…" : "Generate";
  $("generate-form").setAttribute("aria-busy", String(generating));
}

async function refresh({ loud = false, initial = false } = {}) {
  if (state.refreshing || state.generating) return;
  state.refreshing = true;
  $("refresh").disabled = true;
  const sequence = ++state.request;
  try {
    const requests = [api("/api/v1/collections")];
    if (!state.styles.length) requests.push(api("/api/v1/styles"));
    const results = await Promise.allSettled(requests);
    if (sequence !== state.request) return;
    if (results[1]?.status === "fulfilled") {
      state.styles = results[1].value.styles;
      state.formats = results[1].value.formats;
      renderStyles();
      if ($("notice").dataset.source === "styles") $("notice").hidden = true;
    } else if (results[1]?.status === "rejected") showNotice(explainError(results[1].reason), "styles");
    if (results[0].status === "rejected") throw results[0].reason;
    const previous = state.collections.reduce((sum, collection) => sum + collection.avatars.length, 0);
    state.collections = results[0].value.collections;
    const next = state.collections.reduce((sum, collection) => sum + collection.avatars.length, 0);
    state.loaded = true;
    connection(true);
    if ($("notice").dataset.source === "connection") $("notice").hidden = true;
    await resolveRoute({ scroll: initial, restoreView: initial });
    retryFailedImages(loud);
    if (!initial && next > previous) toast(`${next - previous} new ${next - previous === 1 ? "portrait" : "portraits"} in your library`);
    else if (loud) toast("Library is up to date");
  } catch (error) {
    connection(false);
    if (initial || loud) showNotice(explainError(error), "connection");
    if (!state.loaded) {
      const retry = element("button", "button", "Try again");
      retry.addEventListener("click", () => refresh({ loud: true }));
      $("load-state").replaceChildren(element("p", "", "Couldn't open your library."), retry);
      $("portrait-grid").setAttribute("aria-busy", "false");
    }
  } finally {
    state.refreshing = false;
    $("refresh").disabled = false;
  }
}

async function generate(event) {
  event?.preventDefault();
  if (state.generating || !state.styles.length || !$("generate-form").reportValidity()) return;
  setGenerating(true);
  $("notice").hidden = true;
  ++state.request; // An earlier library read must not replace this new collection.
  const body = { style: $("style").value, count: Number($("count").value) };
  const inputs = generationInputsForStyle(state.styles, body.style, { color: $("color").value, animal: $("animal").value });
  if (Object.keys(inputs).length) body.inputs = inputs;
  const name = $("collection-name").value.trim();
  if (name) body.name = name;
  try {
    const collection = await api("/api/v1/collections", {
      method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(body),
    });
    state.collections = [collection, ...state.collections.filter((item) => item.id !== collection.id)];
    state.loaded = true;
    connection(true);
    $("collection-name").value = "";
    navigate(routeURL(collection.id));
    toast(`${collection.avatars.length} ${collection.avatars.length === 1 ? "portrait" : "portraits"} created`);
  } catch (error) {
    showNotice(explainError(error));
  } finally {
    setGenerating(false);
  }
}

async function copyLink(url, label) {
  const absolute = new URL(url, location.origin).href;
  try {
    if (!navigator.clipboard?.writeText) throw new Error("Clipboard unavailable");
    await navigator.clipboard.writeText(absolute);
    toast(`${label} link copied`);
  } catch {
    showNotice(`Couldn't copy automatically. Copy this link: ${absolute}`);
  }
}

function circleSelected() {
  return $("export-form").elements.shape.value === "circle";
}

function readExportControls() {
  const width = Number($("export-width").value);
  const height = Number($("export-height").value);
  if (![width, height].every((value) => Number.isInteger(value) && value >= 1 && value <= 2048)) return null;
  return { shape: circleSelected() ? "circle" : "portrait", width, height };
}

function syncExportURL() {
  const view = readExportControls();
  if (!state.avatarId || !view) return;
  const url = new URL(location.href);
  const params = writeExportView(url.searchParams, view);
  params.set("background", state.previewBackground);
  url.search = params.toString();
  history.replaceState(history.state, "", url.pathname + url.search + url.hash);
}

function setPreviewBackground(value) {
  state.previewBackground = readPreviewBackground(new URLSearchParams({ background: value }));
  $("preview-stage").dataset.background = state.previewBackground;
  for (const input of $("preview-background").querySelectorAll("input")) {
    input.checked = input.value === state.previewBackground;
  }
}

function updatePreview() {
  const view = readExportControls();
  if (!state.selected || !view) return;
  const { width, height } = view;
  const circle = view.shape === "circle";
  $("preview-stage").classList.toggle("circle", circle);
  const source = imageURL(state.selected, "svg", { width, height, circle });
  if ($("preview").src !== source) $("preview").src = source;
  $("export-note").textContent = circle ? "A circular portrait with a transparent surround." : "Vector detail at any size.";
}

async function exportAvatar(event) {
  event.preventDefault();
  if (!state.selected || !$("export-form").reportValidity()) return;
  const avatar = state.selected;
  const format = event.submitter?.value || "png";
  if (!state.formats.includes(format)) return;
  const button = event.submitter;
  $("export-error").hidden = true;
  delete $("export-error").dataset.source;
  if (button) button.disabled = true;
  try {
    const url = imageURL(avatar, format, {
      width: Number($("export-width").value), height: Number($("export-height").value), circle: circleSelected(),
    });
    const response = await fetch(url, { signal: AbortSignal.timeout(30000), credentials: "same-origin" });
    if (!response.ok) {
      const body = await response.json();
      throw new Error(body.error?.message || "Couldn't export this portrait. Try again.");
    }
    const blob = URL.createObjectURL(await response.blob());
    const download = element("a");
    download.href = blob;
    download.download = `${avatar.style}-${avatar.id}.${format}`;
    document.body.append(download);
    download.click();
    download.remove();
    setTimeout(() => URL.revokeObjectURL(blob), 60000);
    toast(`${format.toUpperCase()} export ready`);
  } catch (error) {
    $("export-error").textContent = explainError(error);
    $("export-error").dataset.source = "export";
    $("export-error").hidden = false;
  } finally {
    if (button) button.disabled = false;
  }
}

document.addEventListener("click", (event) => {
  const link = event.target.closest("a[data-nav]");
  if (!link || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey || event.button !== 0) return;
  event.preventDefault();
  navigate(link.href);
});
$("portrait-grid").addEventListener("click", (event) => {
  const card = event.target.closest("[data-avatar]");
  if (card) navigate(routeURL(state.collectionId, card.dataset.avatar));
});
$("portrait-grid").addEventListener("keydown", (event) => {
  if (!["ArrowLeft", "ArrowRight", "ArrowUp", "ArrowDown", "Home", "End"].includes(event.key)) return;
  const cards = [...$("portrait-grid").children];
  const index = cards.indexOf(document.activeElement);
  if (index < 0) return;
  const columns = getComputedStyle($("portrait-grid")).gridTemplateColumns.split(" ").length;
  const next = { ArrowLeft: index - 1, ArrowRight: index + 1, ArrowUp: index - columns, ArrowDown: index + columns, Home: 0, End: cards.length - 1 }[event.key];
  event.preventDefault();
  cards[Math.min(cards.length - 1, Math.max(0, next))]?.focus();
});
$("generate-form").addEventListener("submit", generate);
$("style").addEventListener("change", () => renderGenerationInputs());
for (const name of ["color", "animal"]) $(name).addEventListener("change", () => {
  state.inputsByStyle[$("style").value][name] = $(name).value;
  if (name === "color") updateColorSwatch();
});
$("empty-generate").addEventListener("click", generate);
$("refresh").addEventListener("click", () => refresh({ loud: true }));
$("dismiss-notice").addEventListener("click", () => { $("notice").hidden = true; });
$("close-inspector").addEventListener("click", () => {
  const id = state.avatarId;
  navigate(routeURL(state.collectionId));
  [...$("portrait-grid").children].find((card) => card.dataset.avatar === id)?.focus({ preventScroll: true });
});
$("copy-avatar").addEventListener("click", () => copyLink(routeURL(null, state.avatarId), "Portrait"));
$("copy-collection").addEventListener("click", () => copyLink(routeURL(state.collectionId), "Collection"));
$("thumbnail-size").addEventListener("input", (event) => { $("portrait-grid").dataset.size = event.target.value; });
$("menu-toggle").addEventListener("click", () => setLibraryOpen(!$("library").classList.contains("open")));
$("sidebar-scrim").addEventListener("click", () => setLibraryOpen(false));
$("preview-background").addEventListener("change", (event) => {
  setPreviewBackground(event.target.value);
  syncExportURL();
});
$("export-form").addEventListener("submit", exportAvatar);
$("export-form").addEventListener("change", (event) => {
  if (event.target.name === "shape") {
    const width = Number($("export-width").value);
    if (width >= 1 && width <= 2048) $("export-height").value = Math.min(2048, circleSelected() ? width : Math.round(width * nativeRatio(state.styles.find((style) => style.id === state.selected?.style))));
  }
  syncExportURL();
  updatePreview();
});
$("preview").addEventListener("error", () => {
  $("export-error").textContent = "Couldn't load the portrait preview. Check the connection, then refresh the library.";
  $("export-error").dataset.source = "preview";
  $("export-error").hidden = false;
});
$("preview").addEventListener("load", () => {
  if ($("export-error").dataset.source === "preview") {
    $("export-error").hidden = true;
    delete $("export-error").dataset.source;
  }
});
watchImage($("preview"));
document.addEventListener("keydown", (event) => {
  if (event.key !== "Escape") return;
  if ($("library").classList.contains("open")) { setLibraryOpen(false); $("menu-toggle").focus(); }
  else if (state.avatarId) $("close-inspector").click();
});
window.addEventListener("popstate", () => {
  setLibraryOpen(false);
  $("notice").hidden = true;
  void resolveRoute({ scroll: true });
});
window.addEventListener("focus", () => refresh());
document.addEventListener("visibilitychange", () => { if (!document.hidden) void refresh(); });
setInterval(() => { if (!document.hidden) void refresh(); }, 5000);
void refresh({ initial: true });
