import test from 'node:test';
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';

const html = await readFile(new URL('./index.html', import.meta.url), 'utf8');

test('prototype is self-contained and exposes the three approved modes', () => {
  assert.match(html, /<style>[\s\S]*<\/style>/);
  assert.match(html, /<script>[\s\S]*<\/script>/);
  assert.doesNotMatch(html, /<script[^>]+src=/);
  assert.doesNotMatch(html, /<link[^>]+stylesheet/);
  for (const label of ['一战晟铭', 'Agent代理', 'Duet']) assert.match(html, new RegExp(label));
  assert.doesNotMatch(html, />\s*内容创作\s*</);
});

test('prototype publishes state inspection and persistence contracts', () => {
  for (const token of ['cm-agent-duet-prototype-v1', 'loadState', 'saveState', 'patchState', 'renderApp', 'dispatchAction', '__CM_PROTOTYPE__']) {
    assert.match(html, new RegExp(token));
  }
});

test('prototype includes three approved themes without external assets', () => {
  for (const theme of ['dark', 'light', 'snow']) assert.match(html, new RegExp(`data-theme=["']?${theme}|theme:\\s*["']${theme}`));
  assert.doesNotMatch(html, /https?:\/\//);
});
