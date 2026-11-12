import assert from 'node:assert/strict';
import fs from 'node:fs';
import test from 'node:test';

const source = fs.readFileSync(new URL('./ResourceLibrary.jsx', import.meta.url), 'utf8')
  + fs.readFileSync(new URL('./resourceLibraryModel.js', import.meta.url), 'utf8');
const css = fs.readFileSync(new URL('./editor.css', import.meta.url), 'utf8');

test('资源库提供八个中文主分类并把文字和字幕分开', () => {
  for (const label of ['媒体', '音频', '文字', '字幕', '转场', '特效', '滤镜', '模板']) {
    assert.match(source, new RegExp(`['"]${label}['"]`));
  }
  for (const label of ['搜索资源', '全部', '收藏', '最近使用']) {
    assert.match(source, new RegExp(label));
  }
});

test('资源卡支持预览、双击应用、收藏和拖入时间线', () => {
  assert.match(source, /onClick=.*onPreviewResource/);
  assert.match(source, /onDoubleClick=.*onApplyResource/);
  assert.match(source, /draggable/);
  assert.match(source, /application\/x-aicut-resource/);
  assert.match(source, /onToggleFavorite/);
});

test('资源预览区会显示真实参数并把用户参数交给应用命令', () => {
  assert.match(source, /InputNumber/);
  assert.match(source, /normalizeCatalogParameters/);
  assert.match(source, /selectedResource\.parameters\.map/);
  assert.match(source, /onApplyResource\(selectedResource, normalizedParameters\)/);
  assert.match(source, /应用到时间线/);
});

test('媒体卡支持上传、识别字幕、预览和加入时间线', () => {
  assert.match(source, /beforeUpload={onUpload}/);
  assert.match(source, /onOpenASR/);
  assert.match(source, /onAddAsset/);
  assert.match(source, /application\/x-aicut-asset/);
});

test('视频素材卡显示并执行代理媒体操作', () => {
  assert.match(source, /onGenerateProxy/);
  assert.match(source, /proxyBusyId/);
  assert.match(source, /代理已启用/);
  assert.match(source, /生成代理/);
  assert.match(source, /重新生成代理/);
});

test('图片素材卡提供真实背景移除入口和任务取消状态', () => {
  assert.match(source, /onRemoveBackground/);
  assert.match(source, /backgroundBusyId/);
  assert.match(source, /背景移除/);
  assert.match(source, /取消去背/);
});

test('音频分类显示项目音频并提供真实波形节拍分析入口', () => {
  assert.match(source, /category === 'audio'/);
  assert.match(source, /asset\.kind === 'audio'/);
  assert.match(source, /onAnalyzeAudio/);
  assert.match(source, />分析节拍</);
  assert.match(source, /onOpenTTS/);
  assert.match(source, />文字转语音</);
});

test('音频分类提供声音克隆入口而不是展示卡片', () => {
  assert.match(source, /onOpenVoiceClone/);
  assert.match(source, />声音克隆</);
});

test('资源库不再通过大弹窗作为资源的唯一入口', () => {
  assert.doesNotMatch(source, /Modal|setToolWorkbenchOpen|功能说明/);
});

test('资源库提供用户可见的收起与展开按钮', () => {
  assert.match(source, /aria-label={collapsed \? '展开素材区' : '收起素材区'}/);
  assert.match(source, /onToggleCollapsed/);
});

test('文字资源页提供可操作的自定义文字编辑入口', () => {
  assert.match(source, /onOpenTextEditor/);
  assert.match(source, /新建自定义文字/);
  assert.match(source, /上传字体/);
  assert.match(source, /\.ttf,\.otf/);
  assert.match(source, /onUseFont/);
});

test('资源主分类采用桌面剪辑器顶部横向页签而不是竖向图标栏', () => {
  assert.match(css, /\.aicut-resource-library\s*\{[^}]*grid-column:\s*1\s*\/\s*3/);
  assert.match(css, /\.aicut-resource-tabs\s*\{[^}]*flex-direction:\s*row/);
  assert.match(css, /\.aicut-resource-panel\s*\{[^}]*grid-row:\s*2/);
});
