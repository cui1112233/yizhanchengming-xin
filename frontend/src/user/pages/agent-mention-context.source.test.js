import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const source = fs.readFileSync(path.join(path.dirname(fileURLToPath(import.meta.url)), 'AgentPageV2.jsx'), 'utf8');

test('composer at-mentions select stable node or asset ids', () => {
  assert.match(source, /findAgentMentionCandidates\(\{ nodes: canvas\.nodes, assets: workspaceAssets \}, question\)/);
  assert.match(source, /candidate\.mentionSource === 'asset'/);
  assert.match(source, /appendUniqueAssetID\(current, candidate\.referenceId\)/);
  assert.match(source, /candidate\.referenceId/);
});

test('selected node references can be explicitly removed before sending', () => {
  assert.match(source, /<Tag closable key=\{node\.id\}/);
  assert.match(source, /setSelectedCanvasNodes\(current => current\.filter\(id => id !== node\.id\)\)/);
});
