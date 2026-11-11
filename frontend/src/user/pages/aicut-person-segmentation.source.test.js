import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';

const generator = fs.readFileSync(new URL('./aicutPersonSegmentation.js', import.meta.url), 'utf8');
const page = fs.readFileSync(new URL('./AICutPage.jsx', import.meta.url), 'utf8');
const inspector = fs.readFileSync(new URL('./aicut/InspectorPanel.jsx', import.meta.url), 'utf8');

test('人物抠像使用本地 MediaPipe 模型和 WASM 并生成可上传蒙版视频', () => {
  assert.match(generator, /@mediapipe\/tasks-vision/);
  assert.match(generator, /\/aicut\/mediapipe\/wasm/);
  assert.match(generator, /\/aicut\/mediapipe\/models\/selfie_segmenter\.tflite/);
  assert.match(generator, /ImageSegmenter/);
  assert.match(generator, /MediaRecorder/);
  assert.match(generator, /captureStream/);
  assert.match(generator, /confidenceMasks/);
  assert.match(generator, /signal\.aborted|throwIfAborted/);
  assert.match(page, /uploadAICutAsset/);
  assert.match(page, /clip\.visual_processing_set/);
  assert.doesNotMatch(page, /if \(controller\.signal\.aborted\) return;/, '取消生成不能静默返回，必须进入统一取消提示');
});

test('人物抠像界面明确生成、取消、过期和导出规则', () => {
  assert.match(inspector, /AI 人物抠像/);
  assert.match(inspector, /生成人物蒙版/);
  assert.match(inspector, /取消生成/);
  assert.match(inspector, /蒙版已过期/);
  assert.match(inspector, /预览、保存和云端导出/);
});
