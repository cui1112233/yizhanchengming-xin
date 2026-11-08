import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const dir = path.dirname(fileURLToPath(import.meta.url));
const sourcePath = path.join(dir, 'ProductUpdateModal.jsx');
const cssPath = path.join(dir, '..', 'styles', 'product-update-modal.css');
const layoutPath = path.join(dir, '..', 'layouts', 'UserLayout.jsx');

test('waits for the first paint before fetching the latest update', () => {
  const source = fs.readFileSync(sourcePath, 'utf8');

  assert.match(source, /const SHOW_DELAY_MS = 1500/);
  assert.match(source, /setTimeout\(/);
  assert.match(source, /fetchLatestProductUpdate\(\)/);
  assert.match(source, /normalizeProductUpdate\(payload\)/);
  assert.match(source, /shouldShowProductUpdate\(normalized\.releaseId, resolveProductUpdateStorage\(\)\)/);
});

test('closes through the confirm button, the top-right control, ESC and the backdrop', () => {
  const source = fs.readFileSync(sourcePath, 'utf8');

  assert.match(source, /我知道了/);
  assert.match(source, /aria-label="关闭更新说明"/);
  assert.match(source, /event\.key === 'Escape'/);
  assert.match(source, /markProductUpdateSeen\(resolveProductUpdateStorage\(\), note\.releaseId\)/);
});

test('renders grouped sections and hides the experience link when absent', () => {
  const source = fs.readFileSync(sourcePath, 'utf8');

  assert.match(source, /data-type=\{section\.type\}/);
  assert.match(source, /section\.label/);
  assert.match(source, /item\.tags\.map/);
  assert.match(source, /\{item\.link \? \(/);
  assert.match(source, /体验 →/);
  assert.match(source, /aria-modal="true"/);
});

test('the modal is mounted in the global shell for signed-in users', () => {
  const layout = fs.readFileSync(layoutPath, 'utf8');

  assert.match(layout, /import \{ ProductUpdateModal \} from '\.\.\/components\/ProductUpdateModal'/);
  assert.equal(layout.match(/<ProductUpdateModal enabled=\{isLoggedIn\} \/>/g)?.length, 3);
});

test('dark card styles reuse shell tokens and adapt on mobile', () => {
  const css = fs.readFileSync(cssPath, 'utf8');

  assert.match(css, /\.product-update-modal\s*\{/);
  assert.match(css, /var\(--sh-panel,/);
  assert.match(css, /var\(--sh-panel-strong,/);
  assert.match(css, /var\(--sh-border,/);
  assert.match(css, /var\(--sh-muted,/);
  assert.match(css, /var\(--text-success,/);
  assert.match(css, /@media \(max-width: 640px\)/);
  assert.match(css, /@media \(prefers-reduced-motion: reduce\)/);
});
