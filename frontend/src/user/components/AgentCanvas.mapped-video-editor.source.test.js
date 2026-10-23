import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const canvas = fs.readFileSync(path.join(here, 'AgentCanvas.jsx'), 'utf8');

test('mapped video cards keep an editor entry even when their media preview fails', () => {
  assert.match(canvas, /className="agent-canvas-node-card-header[\s\S]*?onClick=\{openEditor\}/);
  assert.match(canvas, /aria-label=\{`编辑\$\{canvasNodeTypeLabel\(nodeType\)\}节点`\}/);
});
