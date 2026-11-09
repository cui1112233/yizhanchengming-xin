import test from 'node:test';
import assert from 'node:assert/strict';
import { normalizeWindowBackgroundId, windowBackgroundStyle, windowBackgrounds } from './window-background.js';

test('window background catalog matches the visible Luoshui choices', () => {
  assert.deepEqual(windowBackgrounds.map(item => item.id), ['none', 'default', 'preset2', 'preset3', 'preset4', 'preset5']);
  assert.equal(normalizeWindowBackgroundId('custom', ''), 'none');
  assert.equal(normalizeWindowBackgroundId('custom', '/user-content/backgrounds/a.png'), 'custom');
  assert.match(windowBackgroundStyle('preset3').backgroundImage, /linear-gradient/);
});
