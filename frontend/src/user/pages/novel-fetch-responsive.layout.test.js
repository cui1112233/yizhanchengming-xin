import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const frontendDir = path.resolve(here, '../../..');

test('novel fetch rules editor can shrink without horizontal overflow', () => {
  const css = fs.readFileSync(path.join(frontendDir, 'public/batch-rewrite/styles.css'), 'utf8');

  assert.match(css, /\.rules-layout\s*\{[^}]*grid-template-columns:\s*minmax\(0,/s);
  assert.match(css, /\.manager-grid\s*\{[^}]*grid-template-columns:\s*repeat\(2,\s*minmax\(0,/s);
  assert.match(css, /\.rules-layout[^}]*\.manager-grid[^}]*min-width:\s*0/s);
  assert.match(css, /\.rules-layout[^}]*overflow-wrap:\s*anywhere/s);
});
