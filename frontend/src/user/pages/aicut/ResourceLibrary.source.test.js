import assert from 'node:assert/strict';
import fs from 'node:fs';
import test from 'node:test';

const source = fs.readFileSync(new URL('./ResourceLibrary.jsx', import.meta.url), 'utf8');

test('资源库提供六个中文主分类和直接搜索筛选', () => {
  for (const label of ['媒体', '文字', '转场', '特效', '滤镜', '模板']) {
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

test('媒体卡支持上传、识别字幕、预览和加入时间线', () => {
  assert.match(source, /beforeUpload={onUpload}/);
  assert.match(source, /onOpenASR/);
  assert.match(source, /onAddAsset/);
  assert.match(source, /application\/x-aicut-asset/);
});

test('资源库不再通过大弹窗作为资源的唯一入口', () => {
  assert.doesNotMatch(source, /Modal|setToolWorkbenchOpen|功能说明/);
});
