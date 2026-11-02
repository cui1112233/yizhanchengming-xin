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
  assert.match(source, /backgroundImage:\s*`url\("\$\{frame\.src\}"\)`/);
  assert.match(source, /backgroundSize:\s*`\$\{frame\.columns \* 120\}px \$\{frame\.rows \* 120\}px`/);
  assert.match(source, /backgroundPosition:\s*`\$\{-frame\.column \* 120\}px \$\{-frame\.row \* 120\}px`/);
  assert.doesNotMatch(source, /<img[\s\S]*?style=\{\{[\s\S]*?transform:/);
});
