import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const source = fs.readFileSync(path.join(path.dirname(fileURLToPath(import.meta.url)), 'AgentWorkspaceHeader.jsx'), 'utf8');

test('workspace header exposes the four user-facing project views', () => {
  for (const [id, label] of [['workflow', '画布'], ['storyboard', '故事板'], ['assets', '素材'], ['tasks', '任务']]) {
    assert.match(source, new RegExp(`aria-label="${label}"[\\s\\S]{0,180}activeView === '${id}'`));
    assert.match(source, new RegExp(`onViewChange\\('${id}'\\)`));
  }
});
