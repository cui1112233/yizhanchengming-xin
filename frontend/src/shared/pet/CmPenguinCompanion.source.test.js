import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'CmPenguinCompanion.jsx'), 'utf8');
const styles = fs.readFileSync(path.join(here, 'cm-penguin-companion.css'), 'utf8');

test('Little Companion click is a tease and chat has a separate trigger', () => {
  assert.match(source, /pet\.id === 'little-companion'[\s\S]*?showSpeech\(/);
  assert.match(source, /pet\.id === 'little-companion'[\s\S]*?return;/);
  assert.match(source, /kind: 'dance', pose: 'dance', durationMs: 1_220/);
  assert.match(source, /aria-label=\{`和 \$\{pet\.displayName\} 聊天`\}/);
});

test('Little Companion click keeps its current facing direction', () => {
  assert.doesNotMatch(
    source,
    /if \(pet\.id === 'little-companion'\) \{[\s\S]*?setLittleDirection\(current => current \* -1\);/
  );
});

test('Little Companion uses its source motion without the shared vertical bob', () => {
  assert.match(
    styles,
    /\.cm-penguin-pet\.is-little-companion\s*\{[^}]*animation:\s*none;/
  );
});

test('Little Companion clips every sprite-sheet row to one visible frame', () => {
  assert.match(
    styles,
    /\.little-companion-sprite\s*\{[^}]*clip-path:\s*inset\(0\);/
  );
});
