import assert from 'node:assert/strict';
import fs from 'node:fs';
import test from 'node:test';

const sourcePath = new URL('./AgentNodeComposerReferenceAssets.jsx', import.meta.url);

test('exports an always-visible selected asset strip', () => {
  const source = fs.readFileSync(sourcePath, 'utf8');

  assert.match(source, /export function AgentNodeComposerReferenceStrip/);
  assert.match(source, /agent-node-composer__selected-references/);
  assert.match(source, /aria-label="已引用资产"/);
  assert.match(source, /items\.map/);
});

test('shows media previews or lucide type icons with accessible removal', () => {
  const source = fs.readFileSync(sourcePath, 'utf8');

  assert.match(source, /AuthenticatedAgentImage/);
  assert.match(source, /AuthenticatedAgentVideo/);
  assert.match(source, /Music2/);
  assert.match(source, /FileText/);
  assert.match(source, /aria-label=\{`移除资产/);
  assert.match(source, /onRemove\?\.\(asset, event\)/);
});

test('keeps the compact reference action for adding assets', () => {
  const source = fs.readFileSync(sourcePath, 'utf8');

  assert.match(source, /export function AgentNodeComposerReferenceAssets/);
  assert.match(source, />添加参考素材<\/Button>/);
  assert.match(source, /onAdd\?\.\(event\)/);
});
