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
  assert.match(comic, /className="script-workbench private-comic-workbench private-studio-workbench private-studio-resizable"/);
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
  assert.match(script, /已读取 \$\{scriptReadCharCount\} 字/);
  assert.match(script, /自动保存/);
  assert.match(script, /label: hasScriptText \|\| hasScriptOutput \? '内容已读取' : '待读取'/);
  assert.match(script, /label: '确认人物与场景'/);
  assert.match(script, /label: '生成创作结果'/);
  assert.match(steps, /private-studio-step-dot/);
  assert.match(chips, /addLabel = '\+ 添加'/);
});

test('阶段 2：左栏主按钮保留真实语义并位于工具行右端', () => {
  assert.match(script, /className="private-studio-extract"[\s\S]{0,80}htmlType="submit"/);
  assert.match(script, /提取人物与场景；生成剧本请用右侧「开始生成」/);
  assert.match(comic, /className="private-studio-extract"[\s\S]{0,600}onClick={onExtractAssets}/);
  assert.match(comic, /重新提取资产/);
  assert.doesNotMatch(comic, /className="private-studio-primary"/);
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
  assert.match(comic, /private-studio-tool-primary[\s\S]{0,600}>\{primaryAction\.label\}<\/Button>/);
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
  assert.match(note, /先提取资产，再生成漫画提示词，确认后生成漫画图片。/);
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

test('修正：0 字时步骤 1 不得为完成态，也不得显示“已读取”', () => {
  assert.match(script, /const hasScriptText = Boolean\(novelText\.trim\(\)\);/);
  assert.match(script, /const hasScriptOutput = Boolean\(output\);/);
  assert.match(script, /const hasScriptAssets = scriptEntityCount > 0;/);
  // 空正文 → 待读取（pending），不是高亮/完成
  assert.match(script, /label: hasScriptText \|\| hasScriptOutput \? '内容已读取' : '待读取', state: hasScriptText \|\| hasScriptOutput \? 'done' : 'pending'/);
  assert.doesNotMatch(script, /label: '内容已读取', state: novelText\.trim\(\) \? 'done' : 'active'/);
  // 三个态在 PrivateStudioSteps 里仍有区分：done 打勾，active/pending 显示序号
  assert.match(steps, /step\.state === 'done' \? '✓' : index \+ 1/);
  assert.match(css, /\.private-studio-step\.is-done \.private-studio-step-dot \{/);
  assert.match(css, /\.private-studio-step\.is-active \.private-studio-step-dot \{/);
  assert.match(css, /\.private-studio-step-dot \{/);
});

test('修正：0 字时字数行显示“尚未读取”，右栏空态文案按真实条件切换', () => {
  assert.match(script, /hasScriptText \? `已读取 \$\{scriptReadCharCount\} 字` : '尚未读取'/);
  assert.match(script, /hasScriptText \? \(hasScriptAssets \? '素材已就绪，点击右侧「开始生成」' : '提取素材后开始生成'\) : '先粘贴正文，再提取素材'/);
});

test('修正：提取在左栏固定区工具行右端，不在底部滚动区', () => {
  const fixedAt = script.indexOf('private-studio-left-fixed');
  const bodyAt = script.indexOf('private-studio-left-scroll-body');
  const extractAt = script.indexOf('private-studio-extract');
  assert.ok(fixedAt > -1 && bodyAt > -1 && extractAt > -1);
  assert.ok(extractAt > fixedAt && extractAt < bodyAt, '提取按钮必须位于 fixed 区且早于滚动区');
  assert.doesNotMatch(script, /className="private-studio-primary" type="primary" htmlType="submit"/);
  assert.match(css, /\.private-studio-left-scroll-body \{[\s\S]{0,120}overflow-y: auto;/);
});

test('修正：0 字时提取 disabled 并给出原因，loading 来自真实请求', () => {
  assert.match(script, /disabled=\{generating \|\| extractionUnavailable \|\| !hasScriptText\}/);
  assert.match(script, /'先粘贴或输入内容'/);
  assert.match(script, /loading=\{extracting\}/);
});

test('修正：人物/场景卡压成单行，空态说明移入 title，不内联展开', () => {
  assert.doesNotMatch(script, /<div className="entity-card">可手动添加/);
  assert.match(script, /const hint = `可手动添加\$\{title\}，或提取后显示\$\{title\}信息`;/);
  assert.match(script, /aria-label=\{hint\}/);
  assert.match(script, /entity-section-info/);
  assert.match(script, /entity-section-summary/);
  assert.match(script, />查看全部<\/Button>/);
  assert.doesNotMatch(script, /className="entity-list"[^}]{0,40}\{items\.length === 0 \?/);
});

test('修正：草稿保存接真实三态，失败可读且不再有静态「草稿已保存」', () => {
  assert.doesNotMatch(script, /<i \/>草稿已保存/);
  assert.match(script, /status === 'saving'[\s\S]{0,40}保存中…/);
  assert.match(script, /已保存 · \$\{formatDraftSavedTime\(draftSaveState\.savedAt\)/);
  assert.match(script, /草稿保存失败，内容仍在页面上，请勿关闭/);
  assert.match(script, /console\.error\('\[私人定制\] 草稿保存失败：', result\.reason\)/);
  assert.match(script, /draftSaveState\.status === 'failed' \?/);
  const storage = read('./scriptDraftStorage.js');
  assert.match(storage, /return \{ ok: false, reason:/);
  assert.match(storage, /QuotaExceededError/);
  assert.match(storage, /return \{ ok: true, savedAt \}/);
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

test('修正：顶栏不再重复页面标题，只在有真实全局状态时出现', () => {
  const layout = read('../../shared/layouts/UserLayout.jsx');
  assert.doesNotMatch(layout, /legacy-page-title/);
  assert.match(layout, /!isAgentCanvasRoute && globalStatus\.text/);
  assert.match(layout, /legacy-global-status/);
});

test('漫画页签：与剧本页签共用同一套左栏固定/滚动结构', () => {
  const fixedAt = comic.indexOf('private-studio-left-fixed');
  const bodyAt = comic.indexOf('private-studio-left-scroll-body');
  const extractAt = comic.indexOf('className="private-studio-extract"');
  assert.ok(fixedAt > -1 && bodyAt > -1 && extractAt > -1);
  assert.ok(extractAt > fixedAt && extractAt < bodyAt, '漫画提取按钮必须在 fixed 区工具行内');
  assert.doesNotMatch(comic, /className="private-studio-primary"/);
  assert.doesNotMatch(comic, /className="private-studio-primary-wrap"/);
  assert.match(comic, /autoSize=\{\{ minRows: 4, maxRows: 10 \}\}/);
  assert.match(comic, /private-studio-count-row[\s\S]{0,120}尚未读取/);
  assert.match(comic, /private-comic-asset-row/);
  assert.match(comic, /查看全部<\/Button>/);
  assert.match(css, /\.private-studio-left-scroll-body \{[\s\S]{0,120}overflow-y: auto;/);
  assert.match(css, /\.private-comic-asset-row \{/);
});

test('漫画页签：草稿保存状态与剧本同一套真实三态', () => {
  assert.doesNotMatch(comic, /<i \/>草稿已保存/);
  assert.doesNotMatch(comic, />已保存<\/span>/);
  assert.match(comic, /draftSaveState\.status === 'saving'[\s\S]{0,40}保存中…/);
  assert.match(comic, /已保存 · \$\{formatDraftSavedTime\(draftSaveState\.savedAt\)/);
  assert.match(comic, /草稿保存失败，正文仍在页面上，请勿关闭/);
  assert.match(script, /setPrivateComicSaveState\(\{ status: 'failed', savedAt: null, reason \}\)/);
});

test('约束设置：剧本与漫画共用同一个弹窗组件与字段表', () => {
  const constraints = read('private-studio/PrivateStudioConstraintsModal.jsx');
  assert.match(constraints, /export const PRIVATE_STUDIO_CONSTRAINT_FIELDS/);
  assert.match(constraints, /key: 'prefix'/);
  assert.match(constraints, /key: 'comic_quality'/);
  assert.match(constraints, /export function PrivateStudioConstraintsModal/);
  assert.match(script, /<PrivateStudioConstraintsModal/);
  assert.match(comic, /<PrivateStudioConstraintsModal/);
  assert.match(comic, /PRIVATE_STUDIO_CONSTRAINT_FIELDS\.comic\.map/);
  assert.doesNotMatch(comic, /<Drawer/);
  assert.match(css, /\.private-studio-constraints-modal \.private-studio-constraint-section \{/);
});

test('打包下载：漫画复用批量工厂 download-zip 接口，入口在工具行与更多操作', () => {
  assert.match(comic, /下载提示词与图片（ZIP）/);
  assert.match(comic, /if \(key === 'download'\) onDownloadAll\(\);/);
  assert.match(comic, /onDownloadAll=\{onDownloadAll\}/);
  assert.match(script, /downloadComicPromptZip\(batchId, bookId\)/);
});

test('左右拖动：漫画页签补上分栏拖拽手柄，与剧本同一套交互', () => {
  assert.match(comic, /const \[leftPanelWidth, setLeftPanelWidth\] = useState\(null\);/);
  assert.match(comic, /className="script-resize-handle"[\s\S]{0,200}aria-orientation="vertical"[\s\S]{0,80}onPointerDown=\{handleResizeStart\}/);
  assert.match(comic, /gridTemplateColumns: `\$\{leftPanelWidth\}px 8px minmax\(460px, 1fr\)`/);
  assert.match(css, /\.script-workbench-form > \.private-studio-workbench\.private-comic-workbench \{[\s\S]{0,60}326px 8px minmax\(0, 1fr\)/);
});
