import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const pagesDir = path.dirname(fileURLToPath(import.meta.url));
const read = file => fs.readFileSync(path.join(pagesDir, file), 'utf8');
const layout = read('../../shared/layouts/UserLayout.jsx');
const home = read('HomePage.jsx');
const script = read('ScriptPage.jsx');

test('renames the script entry to private customization without moving the route', () => {
  assert.match(layout, /href: '\/script', icon: FilePenLine, label: '私人定制'/);
  assert.match(home, /href: '\/script', className: 'card-write', icon: FilePenLine, title: '私人定制'/);
  assert.match(home, /粘贴内容，生成剧本或漫画制作素材/);
});

test('keeps the existing script layout under a two-workspace shell', () => {
  assert.match(script, /剧本创作/);
  assert.match(script, /漫画生成/);
  assert.match(script, /script-left/);
  assert.match(script, /script-right/);
  assert.match(script, /import '\.\/shuihuo-production\.css'/);
});
