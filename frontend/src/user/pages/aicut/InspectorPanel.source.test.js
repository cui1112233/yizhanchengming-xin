import assert from 'node:assert/strict';
import fs from 'node:fs';
import test from 'node:test';

const source = fs.readFileSync(new URL('./InspectorPanel.jsx', import.meta.url), 'utf8');

test('属性栏为不同选中对象提供中文分组', () => {
  for (const label of ['未选择对象', '画面', '文字与字幕', '音频', '转场', '特效', '位置与大小', '时间', '不透明度']) {
    assert.match(source, new RegExp(label));
  }
});

test('属性修改通过明确回调提交而不是只改展示状态', () => {
  assert.match(source, /onPatchClip/);
  assert.match(source, /onPatchEffect/);
  assert.match(source, /onPatchTransition/);
  assert.match(source, /onBlur/);
  assert.match(source, /onPointerUp/);
});

test('未知资源只读显示并提供窄屏关闭入口', () => {
  assert.match(source, /未知资源/);
  assert.match(source, /关闭属性栏/);
});

test('桌面属性栏也可以收起', () => {
  assert.match(source, /收起属性栏/);
  assert.match(source, /onCollapse/);
});
