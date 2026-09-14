#!/usr/bin/env node
// Regenerate the README screenshots in docs/screenshots.
//
// Seeds a throwaway library with one collection per style, runs the studio
// from ./bin/avatars on a spare loopback port, captures the studio with
// Playwright, and renders one strip of eight CLI-rendered avatars per style.
// Every seed is fixed, so reruns only change when artwork or the UI changes.
//
//   make screenshots
//
// Playwright must be resolvable from this file (`npm i -D playwright` in a
// scratch directory and set PLAYWRIGHT_MODULE to its path, or install it
// globally). No Node dependencies are added to the repository.

import { spawn, execFileSync } from 'node:child_process';
import { mkdtempSync, mkdirSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join, resolve } from 'node:path';

const root = resolve(new URL('..', import.meta.url).pathname);
const bin = join(root, 'bin', 'avatars');
const out = join(root, 'docs', 'screenshots');
const port = process.env.AVATARS_SCREENSHOT_PORT ?? '7457';
const base = `http://127.0.0.1:${port}`;
const { chromium } = await import(process.env.PLAYWRIGHT_MODULE ?? 'playwright');

const styles = [
  { style: 'gorey', name: 'Gorey cast', seed: 'cast', count: 24 },
  { style: 'gorey-expanded', name: 'Expanded cast', seed: 'expanded', count: 24 },
  { style: 'picasso', name: 'Picasso studies', seed: 'studies', count: 24 },
  { style: 'pebble', name: 'Mixed pebbles', seed: 'pebbles', count: 24, flags: ['--color', 'random'] },
  { style: 'companions', name: 'Companions', seed: 'friends', count: 24, flags: ['--animal', 'mixed'] },
  { style: 'field-birds', name: 'Birds of Elsewhere', seed: 'field-notes', count: 48 },
];
// Studio views worth a full screenshot: the hero plus the styles whose
// controls differ (Animal, Color) and the 48-card grid.
const studioViews = ['companions', 'pebble', 'field-birds'];

const dataDir = mkdtempSync(join(tmpdir(), 'avatars-screenshots-'));
const cli = (...args) => execFileSync(bin, ['--data-dir', dataDir, ...args], { encoding: 'utf8' });

const collections = {};
// Collections are listed newest-first in the sidebar; create in reverse so
// the sidebar reads in the order the README lists the styles.
for (const s of [...styles].reverse()) {
  const json = JSON.parse(cli('generate', '--style', s.style, ...(s.flags ?? []),
    '--seed', s.seed, '--count', String(s.count), '--name', s.name, '--json'));
  collections[s.style] = json.id;
}

const server = spawn(bin, ['--data-dir', dataDir, 'serve', '--listen', `127.0.0.1:${port}`], { stdio: 'ignore' });
try {
  for (let i = 0; i < 50; i++) {
    try { if ((await fetch(base + '/')).ok) break; } catch {}
    await new Promise(r => setTimeout(r, 100));
  }
  mkdirSync(out, { recursive: true });
  const browser = await chromium.launch();
  // The hero is captured at 2x; the supporting views at 1.25x keep the
  // repository small while still reading crisply in the README.
  const hero = await browser.newPage({ viewport: { width: 1440, height: 900 }, deviceScaleFactor: 2 });
  const views = await browser.newPage({ viewport: { width: 1440, height: 900 }, deviceScaleFactor: 1.25 });
  const open = async (page, style, card) => {
    await page.goto(`${base}/?collection=${collections[style]}`, { waitUntil: 'networkidle' });
    await page.waitForTimeout(900);
    await page.locator('button.portrait-card').nth(card).click();
    await page.waitForTimeout(700);
  };
  await open(hero, 'gorey', 2);
  await hero.screenshot({ path: join(out, 'studio-hero.png') });
  for (const style of studioViews) {
    await open(views, style, style === 'field-birds' ? 4 : 1);
    await views.screenshot({ path: join(out, `studio-${style}.png`) });
  }
  await browser.close();

  // Style strips: eight avatars at 160 px with a 12 px gutter, composed in the
  // browser so no image library is needed.
  const stripBrowser = await chromium.launch();
  const stripPage = await stripBrowser.newPage({ viewport: { width: 1364, height: 160 }, deviceScaleFactor: 1 });
  for (const s of styles) {
    const images = [];
    for (let i = 1; i <= 8; i++) {
      const png = execFileSync(bin, ['render', '--style', s.style, ...(s.flags ?? []),
        '--seed', `readme-${s.style}:${i}`, '--format', 'png', '--size', '160', '--out', '-']);
      images.push(`data:image/png;base64,${png.toString('base64')}`);
    }
    const html = `<body style="margin:0;background:transparent;display:flex;gap:12px;width:1364px;height:160px">${
      images.map(src => `<img src="${src}" width="160" height="160">`).join('')}</body>`;
    await stripPage.setContent(html);
    await stripPage.screenshot({ path: join(out, `styles-${s.style}.png`), omitBackground: true });
  }
  await stripBrowser.close();
} finally {
  server.kill();
  rmSync(dataDir, { recursive: true, force: true });
}
console.log(`wrote screenshots to ${out}`);
