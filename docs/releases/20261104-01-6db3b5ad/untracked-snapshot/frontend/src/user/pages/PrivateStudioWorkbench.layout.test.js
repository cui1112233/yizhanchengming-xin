import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const pagesDir = path.dirname(fileURLToPath(import.meta.url));
const read = file => fs.readFileSync(path.join(pagesDir, file), 'utf8');

const script = read('ScriptPage.jsx');
const comic = read('PrivateComicWorkspace.jsx');
const css = read('private-studio/private-studio-workbench.css');
const headerActions = read('private-studio/PrivateStudioHeaderActions.jsx');
const steps = read('private-studio/PrivateStudioSteps.jsx');
const chips = read('private-studio/PrivateStudioChips.jsx');
const note = read('private-studio/PrivateStudioNote.jsx');

test('阶段 1：两个模式共用三栏工作台外壳', () => {
  assert.match(script, /className="script-workbench utility-workbench private-studio-workbench/);
  assert.match(comic, /className="script-workbench private-comic-workbench private-studio-workbench"/);
  assert.match(script, /private-studio-side/);
  assert.match(script, /private-studio-main/);
  assert.match(comic, /private-studio-side/);
  assert.match(comic, /private-studio-main/);
  assert.match(script, /private-studio-content/);
  assert.match(comic, /private-studio-content/);
});

test('阶段 1：顶部保留历史版本与更多设置入口，文案逐字一致', () => {
  assert.match(headerActions, />历史版本<\/Button>/);
  assert.match(headerActions, />更多设置…<\/Button>/);
  assert.match(script, /<PrivateStudioHeaderActions/);
  assert.match(comic, /<PrivateStudioHeaderActions/);
});

test('阶段 1：左栏文案与三段进度逐字取自设计规格', () => {
  assert.match(script, /支持 TXT \/ 长文本/);
  assert.match(script, /已读取 \{scriptReadCharCount\} 字/);
  assert.match(script, /自动保存/);
  assert.match(script, /label: '内容已读取'/);
  assert.match(script, /label: '确认人物与场景'/);
  assert.match(script, /label: '生成创作结果'/);
  assert.match(steps, /private-studio-step-dot/);
  assert.match(chips, /addLabel = '\+ 添加'/);
});

test('阶段 2：左栏主按钮保留真实语义并落到底部', () => {
  assert.match(script, /className="private-studio-primary" type="primary" htmlType="submit"/);
  assert.match(script, /提取人物与场景；生成剧本请用右侧「生成」/);
  assert.match(comic, /className="private-studio-primary"/);
  assert.match(comic, /onClick={onExtractAssets}/);
  assert.match(comic, /重新提取资产/);
});

test('阶段 2：漫画保持现有四段链路，不套用剧本三段', () => {
  assert.match(comic, /key: 'assets'/);
  assert.match(comic, /key: 'assetImages'/);
  assert.match(comic, /key: 'prompts'/);
  assert.match(comic, /key: 'images'/);
});

test('阶段 1：下一步提示卡按模式逐字显示', () => {
  assert.match(note, /确认人物与场景后，点击“继续：生成分镜”。/);
  assert.match(note, /先粘贴正文，再从左侧提取资产。/);
  assert.match(script, /<PrivateStudioNote mode="script" \/>/);
  assert.match(comic, /<PrivateStudioNote mode="comic" \/>/);
});

test('阶段 1：新样式独立成文件且不污染全局样式', () => {
  assert.match(css, /\.private-studio-workbench \{/);
  assert.match(css, /grid-template-columns: 326px 8px minmax\(0, 1fr\)/);
  assert.match(css, /grid-template-columns: minmax\(0, 1fr\) 260px/);
  assert.match(css, /@media \(max-width: 900px\)/);
  assert.doesNotMatch(css, /(^|\n)\s*:root\s*\{/);
  assert.doesNotMatch(css, /(^|\n)(html|body|#root)\s*\{/);
  assert.match(script, /import '\.\/private-studio\/private-studio-workbench\.css';/);
});
