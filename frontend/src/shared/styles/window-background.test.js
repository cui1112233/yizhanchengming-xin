import test from 'node:test';
import assert from 'node:assert/strict';
import { normalizeWindowBackgroundId, windowBackgroundStyle, windowBackgrounds, windowBackgroundChoices, windowVideoDefinition } from './window-background.js';

test('settings only offers the four cinematic wallpapers and automatic rotation', () => {
  assert.deepEqual(windowBackgroundChoices.map(item => item.id), ['cinema-auto', 'cinema-0', 'cinema-1', 'cinema-2', 'cinema-3']);
  assert.ok(windowBackgroundChoices.every(item => item.video));
});

test('window background catalog matches the visible Luoshui choices', () => {
  assert.deepEqual(windowBackgrounds.slice(0, 10).map(item => item.id), ['none', 'default', 'preset2', 'preset3', 'preset4', 'preset5', 'preset6', 'preset7', 'preset8', 'preset9']);
  assert.equal(normalizeWindowBackgroundId('custom', ''), 'none');
  assert.equal(normalizeWindowBackgroundId('custom', '/user-content/backgrounds/a.png'), 'custom');
  assert.equal(normalizeWindowBackgroundId('preset9'), 'preset9');
  assert.match(windowBackgroundStyle('preset3').backgroundImage, /linear-gradient/);
});
test('cinematic wallpapers resolve independently from theme and keep video out of background-image', () => {
  for (let index = 0; index < 4; index++) {
    const id = `cinema-${index}`;
    assert.equal(normalizeWindowBackgroundId(id), id);
    assert.equal(windowVideoDefinition(id).src, `/assets/cinematic/scene-${index}.mp4`);
    assert.match(windowBackgroundStyle(id).backgroundImage, /poster/);
  }
  assert.equal(windowVideoDefinition('cinema-auto', 3).src, '/assets/cinematic/scene-3.mp4');
  assert.equal(windowVideoDefinition('preset2'), null);
});

test('image-feel presets keep their own layer sizing', () => {
  assert.match(windowBackgroundStyle('preset7', '', 'light').backgroundImage, /feTurbulence/);
  assert.equal(windowBackgroundStyle('preset7', '', 'light').backgroundSize, '320px 320px, cover, cover, cover');
  assert.equal(windowBackgroundStyle('preset8', '', 'light').backgroundSize, '18px 18px, cover, cover, cover');
  assert.equal(windowBackgroundStyle('preset6', '', 'light').backgroundSize, undefined);
});

test('preset recipes switch between light and dark bases', () => {
  const light = windowBackgroundStyle('preset7', '', 'light').backgroundImage;
  const dark = windowBackgroundStyle('preset7', '', 'dark').backgroundImage;
  assert.notEqual(light, dark);
  assert.match(light, /#ffffff/);
  assert.doesNotMatch(light, /#070b11/);
  assert.match(dark, /#070b11/);
  assert.match(windowBackgroundStyle('default', '', 'light').backgroundImage, /#ffffff/);
  assert.match(windowBackgroundStyle('default', '', 'dark').backgroundImage, /#070b11/);
});
