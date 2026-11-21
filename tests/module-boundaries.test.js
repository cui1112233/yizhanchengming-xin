import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import test from 'node:test';
import { checkModuleBoundaries } from '../scripts/check-module-boundaries.mjs';

function fixture() {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'module-boundary-'));
  for (const dir of ['modules/a', 'modules/b', 'service']) {
    fs.mkdirSync(path.join(root, dir), { recursive: true });
  }
  fs.writeFileSync(path.join(root, 'modules/b/internal.js'), 'export const hidden = true;');
  fs.writeFileSync(path.join(root, 'service/media.js'), 'export const media = true;');
  return root;
}

const config = {
  moduleRoots: ['modules'],
  sharedRoots: ['service'],
  goModulePrefix: 'example/',
};

test('reports JavaScript imports that cross feature-module boundaries', () => {
  const root = fixture();
  fs.writeFileSync(path.join(root, 'modules/a/index.js'), "import '../b/internal.js';\n");

  const result = checkModuleBoundaries({ root, config });

  assert.deepEqual(result.violations.map(item => [item.owner, item.targetOwner]), [['a', 'b']]);
});

test('reports Go imports that cross feature-module boundaries', () => {
  const root = fixture();
  fs.writeFileSync(path.join(root, 'modules/a/main.go'), 'package a\nimport "example/modules/b"\n');

  const result = checkModuleBoundaries({ root, config });

  assert.equal(result.violations.length, 1);
  assert.equal(result.violations[0].importPath, 'example/modules/b');
});

test('allows every feature module to import shared service code', () => {
  const root = fixture();
  fs.writeFileSync(path.join(root, 'modules/a/index.js'), "import '../../service/media.js';\n");

  const result = checkModuleBoundaries({ root, config });

  assert.deepEqual(result.violations, []);
});
