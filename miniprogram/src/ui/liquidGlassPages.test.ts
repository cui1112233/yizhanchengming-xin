import { readFileSync } from 'node:fs';
import { expect, test } from 'vitest';

const read = (path: string) => readFileSync(new URL(path, import.meta.url), 'utf8');

test('functions uses a large feature card and two compact cards for every group', () => {
  const source = read('../pages/functions/index.tsx');
  expect(source).toContain('gallery-feature-layout gallery-mosaic');
  expect(source).toContain('is-feature');
  expect(source).toContain('is-compact-top');
  expect(source).toContain('is-compact-bottom');
});

test('tasks uses one glass list container instead of a glass rim on every row', () => {
  const source = read('../pages/tasks/index.tsx');
  expect(source).toContain('mp-task-list liquid-glass--list');
  expect(source).not.toContain('mp-task-card mp-task-card-new liquid-glass');
});

test('workspace uses the approved execution cards and explicit glass surfaces', () => {
  const page = read('../pages/workspace/index.tsx');
  expect(page).toContain('workbench-current-card liquid-glass--secondary');
  expect(page).toContain('workbench-project-card liquid-glass--secondary');
  expect(page).toContain('workbench-detail-card liquid-glass--secondary');
  expect(page).not.toContain('studio-entry-grid');
  expect(page).not.toContain('<Companion parked />');
  expect(page).toContain('data-glass-role="secondary"');
});

test('functions leaves card optics to the shared liquid material stack', () => {
  const style = read('../pages/functions/functions-glass.scss');
  expect(style).not.toMatch(/\.tools-entry::before/);
  expect(style).not.toMatch(/\.tools-entry\.liquid-glass--secondary::before/);
  expect(style).not.toMatch(/\.tools-entry[^\{]*\{[^}]*background:var\(--glass-fill\)/);
  expect(style).not.toMatch(/\.tools-entry[^\{]*\{[^}]*border:1rpx solid var\(--glass-edge\)/);
});

test('functions keeps card glow restrained like the approved reference', () => {
  const style = read('../pages/functions/gallery.scss');
  expect(style).not.toContain('0 0 16PX rgba(66,226,207,.08)');
  expect(style).not.toContain('0 0 14PX rgba(39,239,216,.3)');
  expect(style).toContain('border:1PX solid rgba(126,171,202,.35)');
});

test('profile applies glass to grouped surfaces rather than every setting row', () => {
  const source = read('../pages/profile/index.tsx');
  expect(source).toContain('mp-profile-hero liquid-glass--primary');
  expect(source).toContain('mp-usage-card liquid-glass--secondary');
  expect(source).toContain('mp-profile-section liquid-glass--list');
  expect(source).not.toContain('mp-profile-item liquid-glass');
});

test('profile detail opts summary cards into secondary glass explicitly', () => {
  const source = read('../pages/profile-detail/index.tsx');
  expect(source).toContain('mp-card liquid-glass--secondary');
  expect(source).toContain('data-glass-role="secondary"');
});
