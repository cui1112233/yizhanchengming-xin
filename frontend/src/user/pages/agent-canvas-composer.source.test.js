import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'AgentPageV2.jsx'), 'utf8');

test('画布生成框保留当前真实能力入口', () => {
  for (const label of ['添加文件', '模式', '专家', '技能', '连接器']) {
    assert.match(source, new RegExp(`>${label}<`));
  }
  assert.match(source, /problemSearchMode \? '退出找问题' : '找问题'/);
  assert.match(source, /openComposerTool\('file'\)/);
  assert.match(source, /openComposerTool\('mode'\)/);
  assert.match(source, /openComposerTool\('expert'\)/);
  assert.match(source, /openComposerTool\('skill'\)/);
});

test('生成框浮层由单一 composerPanel 状态控制', () => {
  assert.match(source, /const \[composerPanel, setComposerPanel\] = useState\(null\)/);
  assert.match(source, /composerPanel === 'mode'/);
  assert.match(source, /composerPanel === 'expert'/);
  assert.match(source, /composerPanel === 'skill'/);
  assert.match(source, /setComposerPanel\(current => current === tool \? null : tool\)/);
});

test('附件和素材引用继续走真实校验与移除逻辑', () => {
  assert.match(source, /classifyComposerAttachment\(file\)/);
  assert.match(source, /仅支持 TXT、MD、CSV、JSON、图片、视频或音频文件/);
  assert.match(source, /移除素材 \$\{asset\.name\}/);
  assert.match(source, /onDrop=\{handleConversationAssetDrop\}/);
});

test('发送状态与故事板专家入口规则不回归', () => {
  assert.match(source, /loading=\{asking\}/);
  assert.match(source, /disabled=\{!question\.trim\(\) \|\| taskDetailLoading\}/);
  assert.match(source, /viewState\.activeView !== 'storyboard'/);
  assert.match(source, /problemSearchMode \? <SearchCheck/);
});
