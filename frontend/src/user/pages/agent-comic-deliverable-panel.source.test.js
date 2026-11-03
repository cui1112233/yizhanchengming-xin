import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const dirname = path.dirname(fileURLToPath(import.meta.url));
const read = relativePath => fs.readFileSync(path.resolve(dirname, relativePath), 'utf8');

test('漫画交付物面板只读取画布投影节点', () => {
  const panel = read('../components/AgentComicDeliverablePanel.jsx');
  assert.match(panel, /buildComicDeliverables\(nodes\)/);
  assert.match(panel, /AuthenticatedAgentImage/);
  assert.doesNotMatch(panel, /shared\/api/);
  assert.doesNotMatch(panel, /onPublish|onSubmit|onSave|fetch\(/);
});

test('画布挂载漫画交付物只读面板', () => {
  const canvas = read('../components/AgentCanvas.jsx');
  assert.match(canvas, /import \{ AgentComicDeliverablePanel \} from '\.\/AgentComicDeliverablePanel'/);
  assert.match(canvas, /<AgentComicDeliverablePanel nodes=\{currentCanvas\.nodes\} \/>/);
});

test('面板样式独立存放并复用工作区主题变量', () => {
  const css = read('../components/agent-comic-deliverable-panel.css');
  assert.match(css, /\.agent-page \.agent-comic-panel \{/);
  assert.match(css, /var\(--agent-panel\)/);
});
