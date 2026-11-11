import assert from 'node:assert/strict';
import fs from 'node:fs';
import test from 'node:test';

const source = fs.readFileSync(new URL('./InspectorPanel.jsx', import.meta.url), 'utf8');
const visualProcessingSource = fs.readFileSync(new URL('../aicutVisualProcessing.js', import.meta.url), 'utf8');

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

test('选中文字后可以继续修改字体排版并提交轨道设置', () => {
  assert.match(source, /onPatchTrack/);
  for (const label of ['字体', '字号', '文字颜色', '描边', '字距', '行高', '对齐']) assert.match(source, new RegExp(label));
});

test('选中音频后可以设置统一音频预设、降噪、三段均衡器、声道混音和响度匹配', () => {
  assert.match(source, /onSetAudioProcessing/);
  assert.match(source, /options=\{audioEffectOptions\}/);
  for (const label of ['音频效果', '降噪', '低频', '中频', '高频', '声道混音', '立体声', '混合为单声道', '仅左声道', '仅右声道', '左右交换', '响度自动匹配', '目标响度', 'LUFS']) {
    assert.match(source, new RegExp(label));
  }
});

test('选中音频后可以精确设置两端淡入淡出', () => {
  for (const label of ['淡入时长', '淡出时长', '音频淡入淡出']) assert.match(source, new RegExp(label));
  assert.match(source, /onSetAudioFades/);
  assert.match(source, /updateAudioFade\('fadeInMs', fadeInMs\)/);
  assert.match(source, /updateAudioFade\('fadeOutMs', fadeOutMs\)/);
});

test('选中画面后可以设置高级调色、自由曲线、色键和羽化蒙版', () => {
  assert.match(source, /onSetVisualProcessing/);
  for (const label of ['高级调色', '曝光', '色温', '色调', '高光', '阴影', '伽马', '自由调色曲线', '恢复直线', '色键抠像', '自动识别幕布', '吸管取色', 'Remove Black', '批量应用', '相似度', '柔化', '蒙版', '矩形', '椭圆', '羽化', '反转蒙版', '裁剪与翻转', '左裁剪', '上裁剪', '右裁剪', '下裁剪', '水平翻转', '垂直翻转']) {
    assert.match(source, new RegExp(label));
  }
  for (const label of ['新增控制点', '删除选中点', '双击曲线区域新增']) assert.match(source, new RegExp(label));
  for (const label of ['手绘多边形', '新增蒙版点', '删除蒙版点', '重置多边形', '双击画布新增']) assert.match(source, new RegExp(label));
  for (const label of ['贝塞尔曲线', '新增锚点', '删除锚点', '重置曲线', '入控制柄', '出控制柄']) assert.match(source, new RegExp(label));
});

test('选中画面后可以设置真实GPU细节效果', () => {
  for (const label of ['画面细节', '锐化', '暗角', '颗粒', 'WebGL实时预览', 'FFmpeg成片']) assert.match(source, new RegExp(label));
  assert.match(source, /patchVisual\('details'/);
  assert.match(source, /aria-label=\{`\$\{label\}百分比`\}/);
  assert.match(source, /Number\(value \?\? 0\) \/ 100/);
});

test('选中视频后可以设置真实导出补帧而不是展示按钮', () => {
  assert.match(source, /视频补帧/);
  assert.match(source, /optical-flow/);
  assert.match(source, /frameInterpolation/);
  assert.match(source, /最终帧率跟随导出设置/);
});

test('选中视频后提供两遍分析的真实防抖参数', () => {
  assert.match(source, /视频防抖/);
  assert.match(source, /平滑度/);
  assert.match(source, /边缘缩放/);
  assert.match(source, /三脚架模式/);
  assert.match(source, /云端导出再做两遍分析/);
});

test('属性栏提供可取消的画面区域跟踪并让用户设置目标框', () => {
  assert.match(source, /画面区域跟踪/);
  assert.match(source, /目标中心X/);
  assert.match(source, /目标中心Y/);
  assert.match(source, /onRunRegionTracking/);
  assert.match(source, /onCancelRegionTracking/);
  assert.match(source, /取消跟踪/);
  assert.match(source, /最多抽取 60 张低清帧/);
});

test('播放器开启防抖后进入低耗运动估计预览而不是继续显示原素材', () => {
  const page = fs.readFileSync(new URL('../AICutPage.jsx', import.meta.url), 'utf8');
  assert.match(page, /processing\.stabilization\.enabled/);
  assert.match(visualProcessingSource, /estimateFrameTranslation/);
  assert.match(visualProcessingSource, /u_stabilization_offset/);
  assert.match(source, /编辑器使用低耗平移防抖预览/);
});

test('选中视频后可以真实设置倒放和当前帧定格', () => {
  assert.match(source, /onSetReverse/);
  assert.match(source, /onSetFreezeFrame/);
  for (const label of ['片段播放', '倒放', '冻结当前帧', '取消定格']) assert.match(source, new RegExp(label));
});

test('选中画面后可以设置常用混合模式', () => {
  assert.match(source, /onSetBlendMode/);
  for (const label of ['画面合成', '正常', '正片叠底', '滤色', '叠加', '柔光', '差值', '添加']) {
    assert.match(source, new RegExp(label));
  }
});

test('关键帧区提供缓动预设和可拖动贝塞尔曲线编辑器', () => {
  for (const label of ['关键帧缓动', '线性', '缓入', '缓出', '缓入缓出', '定格', '自定义', '控制柄1', '控制柄2']) assert.match(source, new RegExp(label));
  assert.match(source, /onSetKeyframeEasing/);
  assert.match(source, /aicut-keyframe-curve/);
});

test('关键帧区提供可撤销的一次性批量选择、偏移、统一缓动和删除', () => {
  for (const label of ['批量选择', '全选', '取消选择', '统一缓动', '时间偏移', '应用偏移', '删除所选']) assert.match(source, new RegExp(label));
  assert.match(source, /onBatchEditKeyframes/);
  assert.match(source, /selectedIds/);
});

test('关键帧区可以复制所选并把剪贴板粘贴到当前播放头', () => {
  for (const label of ['复制所选', '粘贴到播放头', '关键帧剪贴板']) assert.match(source, new RegExp(label));
  assert.match(source, /onCopyKeyframes/);
  assert.match(source, /onPasteKeyframes/);
});
