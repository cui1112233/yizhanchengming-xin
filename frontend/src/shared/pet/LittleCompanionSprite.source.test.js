import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'LittleCompanionSprite.jsx'), 'utf8');

test('keeps the closest idle frame when Little Companion finishes walking', () => {
  assert.match(source, /useLayoutEffect/);
  assert.match(source, /littleCompanionIdleFrameAfterWalk\(lastWalkFrameRef\.current\)/);
});

test('renders one fixed background frame instead of moving the full visible atlas', () => {
  assert.match(source, /className="little-companion-frame"/);
  assert.match(source, /backgroundImage:\s*`url\("\$\{displayedFrame\.src\}"\)`/);
  assert.match(source, /backgroundSize:\s*`\$\{displayedFrame\.columns \* 120\}px \$\{displayedFrame\.rows \* 120\}px`/);
  assert.match(source, /backgroundPosition:\s*`\$\{-displayedFrame\.column \* 120\}px \$\{-displayedFrame\.row \* 120\}px`/);
  assert.doesNotMatch(source, /<img[\s\S]*?style=\{\{[\s\S]*?transform:/);
});

test('preloads every pose sheet so walking never blanks while a sheet loads', () => {
  assert.match(source, /PRELOAD_POSE_ORDER = \['idle', 'walk'/);
  assert.match(source, /preloadLittleCompanionPoses/);
  assert.match(source, /loadLittleCompanionImage\(definition\.src\)/);
  assert.match(source, /IMAGE_RETRY_DELAY_MS/);
  assert.match(source, /image\.onerror/);
});

test('never renders a name-only state when a sprite sheet fails', () => {
  assert.match(source, /displayedFrame = frameStatus === 'loaded'/);
  assert.match(source, /lastGoodFrame/);
  assert.match(source, /className="little-companion-placeholder-art"/);
  assert.ok(
    source.includes('{spriteUnavailable ? <span className="little-companion-placeholder-name">{label}</span> : null}')
  );
  assert.doesNotMatch(source, /little-companion-fallback/);
  assert.doesNotMatch(source, /little-companion-preload/);
});
