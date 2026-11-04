import assert from 'node:assert/strict';
import test from 'node:test';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'AgentPageV2.jsx'), 'utf8');
const css = fs.readFileSync(path.join(here, '..', '..', 'shared', 'styles', 'agent-workspace.css'), 'utf8');
const canvas = fs.readFileSync(path.join(here, '..', 'components', 'AgentCanvas.jsx'), 'utf8');

test('Agent 工作区接入 V88 主题与连接线外的青绿主色', () => {
  assert.match(source, /import \{ Button, ConfigProvider,[^}]*\} from 'antd'/);
  assert.match(source, /createV88AntTheme\('dark'\)/);
  assert.match(source, /<ConfigProvider theme=\{antTheme\}>/);
  assert.match(css, /\.agent-page\.utility-workbench[\s\S]*?--agent-accent:\s*var\(--v88-color-primary\)/);
  assert.match(css, /--agent-accent-hover:\s*var\(--v88-color-primary-hover\)/);
});

test('Agent 工作区保留珊瑚连线强调色，不改画布结构', () => {
  assert.match(css, /--agent-edge-coral:\s*#f07167/);
  assert.match(css, /\.agent-page \.agent-canvas-workspace\s*\{[^}]*grid-template-rows:\s*44px\s+minmax\(0,\s*1fr\)/s);
});

test('Agent 工作区过渡使用 V88 令牌并支持减少动态', () => {
  assert.match(css, /\.agent-page \.agent-task-row[\s\S]*?transition:[^\n]*var\(--v88-motion-base\)/);
  assert.match(css, /@media \(prefers-reduced-motion: reduce\)[\s\S]*?\.agent-page \.agent-task-row[\s\S]*?transition:\s*none/);
});

test('Agent 工作区原有入口一个不少', () => {
  ['添加附件', '选择技能', '选择专家', '发送消息', '开始创作'].forEach(label => {
    assert.ok(source.includes(label), `缺少入口：${label}`);
  });
  ['画布外观', '连线颜色', '连线动效', '连线粗细', '背景网点', '辅助线对齐'].forEach(label => {
    assert.ok(canvas.includes(label), `缺少入口：${label}`);
  });
});
