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
  assert.match(script, /提取人物与场景；生成剧本请用右侧「开始生成」/);
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

test('阶段 3：剧本中栏元信息接真实镜头数与计算时长，不写死', () => {
  assert.doesNotMatch(script, /预计 2 分钟/);
  assert.match(script, /const scriptHeaderShotCount = shotCards\.length;/);
  assert.match(script, /shotCards\.length \* \(selectedDuration === '15s' \? 15 : 10\)/);
  assert.match(script, /scriptCardDurations\.every\(value => value > 0\)/);
  assert.doesNotMatch(comic, /预计 2 分钟/);
});

test('阶段 3：工具行按稿子视觉重排且保留现有入口', () => {
  assert.match(script, /className="script-toolbar private-studio-toolbar"/);
  assert.match(comic, /className="script-toolbar private-studio-toolbar"/);
  assert.match(script, /private-studio-tool-label">输出<\/span>/);
  assert.match(comic, /private-studio-tool-label">文本模型<\/span>/);
  assert.match(comic, /private-studio-tool-label">图片模型<\/span>/);
  assert.match(script, />约束条件<\/Button>/);
  assert.match(comic, />约束条件<\/Button>/);
  assert.match(script, /private-studio-tool-primary[\s\S]{0,400}>开始生成<\/Button>/);
  assert.match(comic, /private-studio-tool-primary[\s\S]{0,400}>开始生成<\/Button>/);
  // P1：漫画「配置」与表头「更多设置…」两者都保留
  assert.match(comic, />配置<\/Button>/);
  assert.match(comic, /<PrivateStudioHeaderActions/);
});

test('阶段 3：漫画徽标与进度复用后端 comicReadiness，不另算口径', () => {
  assert.match(script, /const \[privateComicReadiness, setPrivateComicReadiness\] = useState\(null\)/);
  assert.match(script, /setPrivateComicReadiness\(pageResult\?\.readiness \|\| null\)/);
  assert.match(script, /readiness={privateComicReadiness}/);
  assert.match(comic, /Number\(readinessValue\.promptReadyCount \|\| 0\)/);
  assert.match(comic, /Number\(readinessValue\.imageReadyCount \|\| 0\)/);
  assert.match(comic, /const readinessComplete = hasReadiness && pagesTotal > 0 && imageReadyCount >= pagesTotal/);
  assert.match(comic, /const pageGenerating = pageList\.some\(page => text\(page\?\.generationState\) === 'generating'\)/);
});

test('阶段 3：漫画页状态标签只读后端派生字段', () => {
  assert.match(comic, /text\(page\?\.promptStatus\)/);
  assert.match(comic, /text\(page\?\.imageStatus\)/);
  assert.match(comic, /text\(page\?\.generationState\)/);
  assert.match(comic, /className="private-studio-shot-status"/);
  assert.match(css, /\.private-studio-shot-status \{/);
});

test('阶段 3：中栏卡头与工具行样式仍在独立作用域', () => {
  assert.match(css, /\.private-studio-canvas-head \{/);
  assert.match(css, /\.private-studio-toolbar \{/);
  assert.match(css, /\.private-studio-tool-pill \{/);
  assert.match(script, /className="private-studio-canvas-head"[\s\S]{0,80}分镜草稿/);
});

test('阶段 1：下一步提示卡按模式逐字显示', () => {
  assert.match(note, /确认人物与场景后，点击右侧“开始生成”。/);
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

test('修正：工作台宽度不再锁死 1180px，随可用区域自适应', () => {
  assert.doesNotMatch(css.split('@media')[0], /max-width:\s*1180px/);
  assert.match(css, /\.private-studio-workbench \{[\s\S]{0,180}max-width: none;/);
  assert.match(css, /\.script-workbench-form > \.private-studio-workbench \{[\s\S]{0,140}width: 100%;[\s\S]{0,60}max-width: none;/);
  assert.match(css, /\.script-workbench-form \{[\s\S]{0,120}padding: 14px 18px 18px;/);
  assert.match(css, /@media \(max-width: 1180px\)/);
  assert.match(css, /@media \(max-width: 1100px\)/);
});

test('修正：素材胶囊可点且键盘可达，漫画占位保持不可点', () => {
  assert.match(script, /onClick: \(\) => openEntityEditor\('characters'/);
  assert.match(script, /onClick: \(\) => openEntityEditor\('scenes'/);
  assert.match(chips, /items\.map/);
  assert.match(chips, /<button\s+key=\{item\.key\}/);
  assert.match(chips, /<span key=\{item\.key\}/);
  assert.match(css, /button\.private-studio-chip:focus-visible/);
  assert.match(css, /span\.private-studio-chip \{/);
});

test('修正：/script 顶部不再重复「私人定制」标题，状态区与退出导航保留', () => {
  const layout = read('../../shared/layouts/UserLayout.jsx');
  assert.match(layout, /const isScriptRoute = pathname === '\/script';/);
  assert.match(layout, /!isScriptRoute \|\| globalStatus\.text/);
  assert.match(layout, /\{isScriptRoute \? null : <span className="legacy-page-title">/);
  assert.match(layout, /legacy-global-status/);
});
