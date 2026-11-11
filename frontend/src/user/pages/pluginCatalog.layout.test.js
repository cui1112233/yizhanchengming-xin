import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';

import {
  PLUGIN_CATALOG,
  PLUGIN_TOOL_PANELS,
  filterPlugins,
  pluginLaunchTarget
} from './pluginCatalog.js';

const EXPECTED_PLUGIN_NAMES = [
  '资产参考', 'AI-分段', 'AI-Subtitle', 'AI-Filler', 'AI-双语字幕',
  'AI-CUT插件版', 'AI-BGM', 'AI-Shot', 'AI-Upscale', 'AI-Split',
  '小说转剧', 'AI改文', '分镜大师', '分镜整理', '爆款开头',
  '字幕对轴', '格式化文本', '角色修改', '声音克隆',
  '字幕合并', '文档合并', '封面生成',
  '角色三视图', '音色设计', '草稿合并', '剪映关键帧'
];

// 洛水自己都没有的 4 条（女频性转、长篇拆书、世界观、书名生成）已从目录删除。
const REMOVED_PLUGIN_IDS = ['female-gender-swap', 'long-novel-split', 'world-building', 'book-title'];

test('插件市场按洛水当前可见顺序完整收录26项', () => {
  assert.equal(PLUGIN_CATALOG.length, 26);
  assert.deepEqual(PLUGIN_CATALOG.map(plugin => plugin.name), EXPECTED_PLUGIN_NAMES);
  assert.equal(new Set(PLUGIN_CATALOG.map(plugin => plugin.id)).size, 26);
  assert.ok(PLUGIN_CATALOG.every(plugin => ['workspace', 'route', 'prompt-tool', 'desktop-tool', 'draft-tool', 'local-tool', 'upscale-tool', 'cover-tool', 'turnaround-tool', 'unavailable'].includes(plugin.launch.kind)));
  assert.deepEqual(REMOVED_PLUGIN_IDS.filter(id => PLUGIN_CATALOG.some(plugin => plugin.id === id)), []);
  assert.equal(PLUGIN_CATALOG.find(plugin => plugin.id === 'asset-reference').launch.kind, 'workspace');
  assert.equal(PLUGIN_CATALOG.find(plugin => plugin.id === 'ai-subtitle').launch.kind, 'draft-tool');
  assert.equal(PLUGIN_CATALOG.find(plugin => plugin.id === 'ai-filler').launch.kind, 'draft-tool');
  assert.equal(PLUGIN_CATALOG.find(plugin => plugin.id === 'ai-split').launch.kind, 'draft-tool');
  assert.equal(PLUGIN_CATALOG.find(plugin => plugin.id === 'ai-upscale').launch.kind, 'upscale-tool');
  assert.equal(PLUGIN_CATALOG.find(plugin => plugin.id === 'ai-cut-plugin').launch.kind, 'route');
  assert.equal(PLUGIN_CATALOG.find(plugin => plugin.id === 'cover-generation').launch.kind, 'cover-tool');
});

test('搜索只按洛水插件市场的名称和功能说明过滤', () => {
  assert.deepEqual(filterPlugins(PLUGIN_CATALOG, { query: '剪映' }).map(plugin => plugin.name), [
    'AI-Subtitle', 'AI-双语字幕', 'AI-CUT插件版', 'AI-BGM', 'AI-Split', '草稿合并', '剪映关键帧'
  ]);
  assert.deepEqual(filterPlugins(PLUGIN_CATALOG, { query: '字幕' }).map(plugin => plugin.name), [
    'AI-双语字幕', 'AI-BGM', 'AI-Shot', '字幕对轴', '字幕合并'
  ]);
  assert.deepEqual(filterPlugins(PLUGIN_CATALOG, { query: '不存在' }), []);
});

test('卡片按洛水前端处理器映射进入真实提示词工具、桌面工具或开发中占位', () => {
  assert.deepEqual(pluginLaunchTarget(PLUGIN_CATALOG, 'asset-reference'), { kind: 'workspace', workspace: 'asset-reference' });
  assert.deepEqual(pluginLaunchTarget(PLUGIN_CATALOG, 'ai-segment'), { kind: 'workspace', workspace: 'ai-segment' });
  assert.deepEqual(pluginLaunchTarget(PLUGIN_CATALOG, 'ai-cut-plugin'), { kind: 'route', href: '/ai-cut' });
  assert.deepEqual(pluginLaunchTarget(PLUGIN_CATALOG, 'novel-to-drama'), { kind: 'prompt-tool', toolId: 'novel-to-drama', promptCategory: '中长篇转剧', progressChunkSize: 2000, href: '/plugins/tool/novel-to-drama' });
  assert.deepEqual(pluginLaunchTarget(PLUGIN_CATALOG, 'storyboard-master'), { kind: 'prompt-tool', toolId: 'storyboard-master', promptCategory: '分镜', progressChunkSize: 1500, href: '/plugins/tool/storyboard-master' });
  assert.deepEqual(pluginLaunchTarget(PLUGIN_CATALOG, 'viral-opening'), { kind: 'prompt-tool', toolId: 'viral-opening', promptCategory: '爆款开头', progressChunkSize: 0, href: '/plugins/tool/viral-opening' });
  assert.deepEqual(pluginLaunchTarget(PLUGIN_CATALOG, 'character-edit'), { kind: 'prompt-tool', toolId: 'character-edit', promptCategory: '角色修改', progressChunkSize: 1500, href: '/plugins/tool/character-edit' });
  assert.deepEqual(pluginLaunchTarget(PLUGIN_CATALOG, 'ai-rewrite'), { kind: 'route', href: '/plugins/ai-rewrite' });
  assert.deepEqual(pluginLaunchTarget(PLUGIN_CATALOG, 'voice-clone'), { kind: 'desktop-tool', environment: '声音克隆模型与样本管理工作台' });
  assert.deepEqual(pluginLaunchTarget(PLUGIN_CATALOG, 'ai-subtitle'), { kind: 'draft-tool', tool: 'ai-subtitle' });
  assert.deepEqual(pluginLaunchTarget(PLUGIN_CATALOG, 'ai-upscale'), { kind: 'upscale-tool' });
  assert.deepEqual(pluginLaunchTarget(PLUGIN_CATALOG, 'cover-generation'), { kind: 'cover-tool' });
  assert.equal(pluginLaunchTarget(PLUGIN_CATALOG, 'ai-bgm').kind, 'unavailable');
  assert.equal(pluginLaunchTarget(PLUGIN_CATALOG, 'ai-shot').kind, 'unavailable');
  assert.equal(pluginLaunchTarget(PLUGIN_CATALOG, 'ai-bilingual-subtitle').kind, 'unavailable');
  assert.equal(pluginLaunchTarget(PLUGIN_CATALOG, 'missing-plugin'), null);
});

test('插件市场页面点击整张卡片直达工具且不再提供虚假启停', () => {
  const source = fs.readFileSync(new URL('./PluginCenterPage.jsx', import.meta.url), 'utf8');
  const toolPage = fs.readFileSync(new URL('./PluginToolPage.jsx', import.meta.url), 'utf8');
  const app = fs.readFileSync(new URL('../App.jsx', import.meta.url), 'utf8');
  const layout = fs.readFileSync(new URL('../../shared/layouts/UserLayout.jsx', import.meta.url), 'utf8');
  assert.match(source, /<h1>插件市场<\/h1>/);
  assert.match(source, /className="plugin-market-card"[\s\S]{0,220}onClick=/);
  assert.match(source, /pluginLaunchTarget/);
  assert.match(source, /AssetReferenceWorkbench/);
  assert.match(source, /AISegmentWorkbench/);
  assert.match(source, /target\?\.kind === 'workspace'/);
  assert.match(source, /UpscaleWorkbench/);
  assert.match(source, /target\?\.kind === 'upscale-tool'/);
  // 提示词工具不再弹窗，而是整页面板路由。
  assert.match(source, /target\?\.kind === 'prompt-tool'/);
  assert.match(source, /window\.history\.pushState\(\{\}, '', target\.href\)/);
  assert.doesNotMatch(source, /plugin-prompt-tool/);
  assert.doesNotMatch(source, /placeholder="选择模型"/);
  assert.doesNotMatch(source, /window\.location\.assign/);
  assert.doesNotMatch(source, /\bSwitch\b|localStorage|启用本机目录状态|plugin-center-categories/);
  assert.doesNotMatch(source, /本次要求（可选）|网页可用|需配置|开发中<\/em>/);
  // 面板路由从目录 launch.href 派生，App.jsx 不逐条登记。
  assert.match(app, /import \{ PLUGIN_CATALOG \} from '\.\/pages\/pluginCatalog\.js'/);
  assert.match(app, /const pluginToolRoutes = Object\.fromEntries\(/);
  assert.match(app, /routes\[pathname\] \|\| pluginToolRoutes\[pathname\]/);
  assert.match(toolPage, /PluginPanelShell/);
  assert.match(toolPage, /PLUGIN_TOOL_PANELS/);
  assert.match(layout, /href: '\/plugins', icon: Puzzle, label: '插件'/);
});

test('4 条文字类工具的面板规格逐项对齐洛水实测控件表', () => {
  const novel = PLUGIN_TOOL_PANELS['novel-to-drama'];
  const storyboard = PLUGIN_TOOL_PANELS['storyboard-master'];
  const viral = PLUGIN_TOOL_PANELS['viral-opening'];
  const character = PLUGIN_TOOL_PANELS['character-edit'];

  // 输入框 rows / 高度：小说转剧 6 / 148，其余三条 8 / 193。
  assert.deepEqual([novel.inputRows, novel.inputHeight], [6, 148]);
  for (const panel of [storyboard, viral, character]) assert.deepEqual([panel.inputRows, panel.inputHeight], [8, 193]);

  // 空输入提示原文。
  assert.equal(novel.emptyInputMessage, '请输入或上传小说内容');
  for (const panel of [storyboard, viral, character]) assert.equal(panel.emptyInputMessage, '请输入或上传内容');

  // 输入占位文案。
  assert.equal(novel.inputPlaceholder, '请输入中长篇小说内容，或上传txt/word文档...');
  assert.equal(storyboard.inputPlaceholder, '请输入剧本或小说内容，或上传txt/word文档...');
  assert.equal(viral.inputPlaceholder, '请输入小说原文，或上传txt/word文档...');
  assert.equal(character.inputPlaceholder, '请输入需要修改角色的内容，或上传txt/word文档...');

  // 标题与结果区标题。
  assert.equal(storyboard.heading, 'AI分镜大师');
  assert.deepEqual([novel.resultTitle, storyboard.resultTitle, viral.resultTitle, character.resultTitle], ['生成的剧本', '生成的分镜', '生成结果', '修改结果']);
  assert.deepEqual([novel.resultRows, storyboard.resultRows, viral.resultRows, character.resultRows], [10, 8, 11, 8]);

  // 洛水只有小说转剧有「并发生成」；分镜大师/爆款开头/角色修改不再渲染该开关。
  assert.equal(novel.authorized, true);
  assert.equal(novel.concurrent, true);
  for (const panel of [storyboard, viral, character]) assert.equal(panel.concurrent, undefined);
  for (const panel of [storyboard, viral, character]) assert.equal(panel.authorized, false);

  // 保存文件名和格式化行为来自洛水 bundle：小说转剧可按 TXT/JSON 保存，其余只存 TXT。
  assert.deepEqual([novel.saveFileName, storyboard.saveFileName, viral.saveFileName, character.saveFileName], ['中长篇转剧结果', 'AI分镜结果', '爆款开头结果', '角色修改结果']);
  assert.equal(novel.saveJson, true);
  for (const panel of [storyboard, viral, character]) assert.equal(panel.saveJson, false);
  assert.deepEqual([novel.formatMode, storyboard.formatMode, viral.formatMode, character.formatMode], ['script', 'lines', 'lines', 'lines']);
  assert.equal(novel.showWordCount, true);
  assert.deepEqual([storyboard.showWordCount, viral.showWordCount, character.showWordCount], [false, false, false]);
  assert.deepEqual([novel.iconColor, storyboard.iconColor, viral.iconColor, character.iconColor], ['#14b8a6', '#8b5cf6', '#f59430', '#fb7185']);
});

test('整页面板主按钮点击前不禁用，空输入用提示拦住', () => {
  const source = fs.readFileSync(new URL('./PluginToolPage.jsx', import.meta.url), 'utf8')
    + fs.readFileSync(new URL('./pluginToolLogic.js', import.meta.url), 'utf8')
    + fs.readFileSync(new URL('./pluginGeneration.js', import.meta.url), 'utf8');
  assert.match(source, /plugin-panel-submit/);
  assert.doesNotMatch(source, /plugin-panel-submit[\s\S]{0,200}disabled=/);
  assert.match(source, /const validation = validateGenerationInput\(\{ input, modelId, promptId, emptyMessage: panel\?\.emptyInputMessage \}\)/);
  assert.match(source, /if \(!String\(input \|\| ''\)\.trim\(\)\) return \{ ok: false, message: emptyMessage \|\| '请输入或上传内容' \}/);
  assert.match(source, /showSearch/);
  assert.match(source, /placeholder="选择模型"/);
  assert.match(source, /placeholder="选择提示词"/);
  // 下拉默认值为空：不能再预选第一项。
  assert.doesNotMatch(source, /setModelId\(availableModels\[0\]/);
  assert.doesNotMatch(source, /setPromptId\(promptOptions\[0\]/);
  // 并发生成开关可切换，文案与洛水一致；运行中禁用。
  assert.match(source, /并发生成<\/strong><em>关闭时按分段顺序逐段发送<\/em>/);
  assert.match(source, /<Switch size="small" checked=\{concurrent\} disabled=\{running\}/);
  assert.match(source, /选择模型和提示词，输入内容后点击AI生成/);
});

test('格式化、复制、保存和字数都对齐洛水 bundle 的静态证据', () => {
  const source = fs.readFileSync(new URL('./PluginToolPage.jsx', import.meta.url), 'utf8')
    + fs.readFileSync(new URL('./pluginToolLogic.js', import.meta.url), 'utf8');
  assert.match(source, /格式化完成，共处理 \$\{lines\.length\} 行/);
  assert.match(source, /格式化完成，共\$\{lines\.length\}行/);
  assert.match(source, /剧本正文/);
  assert.match(source, /改编说明/);
  assert.match(source, /已清空/);
  assert.match(source, /已复制到剪贴板/);
  assert.match(source, /已提取并复制，共/);
  assert.match(source, /保存为 TXT/);
  assert.match(source, /保存为 JSON/);
  assert.match(source, /字数：\{output\.length\}/);
  assert.match(source, /panel\.saveFileName/);
});

test('封面生成工作台复用 V88 图片通道并覆盖洛水面板控件', () => {
  const source = fs.readFileSync(new URL('../components/CoverGenerationWorkbench.jsx', import.meta.url), 'utf8');
  const center = fs.readFileSync(new URL('./PluginCenterPage.jsx', import.meta.url), 'utf8');
  assert.match(center, /CoverGenerationWorkbench/);
  assert.match(center, /target\?\.kind === 'cover-tool'/);
  assert.match(source, /createAgentImageGeneration/);
  assert.match(source, /designCoverPrompt/);
  assert.match(source, /'7:10'/);
  assert.match(source, /生成数量支持 1-10 张，默认 1 张，会并发批量生成/);
  assert.match(source, /开始生成/);
  assert.match(source, /改图/);
  assert.match(source, /保存为 5M/);
  assert.match(source, /image_to_image/);
});

test('超分面板用阶段推进、实时计时和动效表达进度，不伪造百分比', () => {
  const source = fs.readFileSync(new URL('../components/UpscaleWorkbench.jsx', import.meta.url), 'utf8');
  // 阶段文案按 上传中 → 已提交 → 处理中 → 完成 推进。
  assert.match(source, /uploading: '上传中'/);
  assert.match(source, /submitted: '已提交'/);
  assert.match(source, /processing: '排队 \/ 处理中'/);
  assert.match(source, /completed: '完成'/);
  // 秒级计时、失败/取消可重试、完成态有结果链接。
  assert.match(source, /已等待 \{formatElapsed\(elapsed\)\}/);
  assert.match(source, /window\.setInterval\(\(\) => setTick\(value => value \+ 1\), 1000\)/);
  assert.match(source, /重新提交/);
  assert.match(source, /打开结果视频/);
  // 洛水左列字段：视频（文件夹+多选）与保存文件夹。
  assert.match(source, /选择文件夹/);
  assert.match(source, /多选视频/);
  assert.match(source, /保存文件夹/);
  // 上游不返回百分比，禁止出现伪造的百分比写法。
  assert.doesNotMatch(source, /percent=\{/);
  assert.doesNotMatch(source, /aria-valuenow=\{(?!100\})\d+\}/);
});

test('超分进度条按阶段流动，不伪造百分比', () => {
  const css = fs.readFileSync(new URL('./plugin-center.css', import.meta.url), 'utf8');
  assert.match(css, /\.upscale-progress\[data-stage="processing"\] span/);
  assert.match(css, /@keyframes upscale-pulse/);
  assert.doesNotMatch(css, /\.upscale-progress\[data-stage="completed"\] span[\s\S]{0,120}right: 0/);
});

test('4 条文字工具按洛水语义分段、显示 x/y 进度并支持 Word 上传', () => {
  const source = fs.readFileSync(new URL('./PluginToolPage.jsx', import.meta.url), 'utf8')
    + fs.readFileSync(new URL('./pluginToolLogic.js', import.meta.url), 'utf8');
  assert.match(source, /splitInputByLength/);
  assert.match(source, /【完整原文】（供参考整体框架，共\$\{full\.length\}字）/);
  assert.match(source, /【重要提示】这是上次对话的历史记录/);
  assert.match(source, /【当前处理分段】这是第\$\{index\}\/\$\{total\}部分/);
  assert.match(source, /AI生成中 \(\$\{progress\.current\}\/\$\{progress\.total\}\)/);
  assert.match(source, /extractDocumentText/);
  assert.match(source, /raw: true/);
  assert.match(source, /attempt < 2/);
});
