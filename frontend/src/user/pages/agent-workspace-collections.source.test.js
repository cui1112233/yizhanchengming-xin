import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const source = fs.readFileSync(path.join(path.dirname(fileURLToPath(import.meta.url)), 'AgentPageV2.jsx'), 'utf8');

test('asset and task workspaces project existing page state instead of loading duplicate stores', () => {
  assert.match(source, /viewState\.activeView === 'assets'/);
  assert.match(source, /workspaceAssets\.map/);
  assert.match(source, /<AgentTaskWorkspace executions=\{executions\} generations=\{generations\} \/>/);
  assert.match(source, /\[\.\.\.executions, \.\.\.generations\]/);
  assert.doesNotMatch(source, /getAgentAssetWorkspace|getAgentTaskWorkspace/);
});

test('workspace root class exposes the active four-view state', () => {
  assert.match(source, /is-\$\{viewState\.activeView\}/);
});
