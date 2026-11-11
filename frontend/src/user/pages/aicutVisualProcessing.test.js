import assert from 'node:assert/strict';
import test from 'node:test';
import { detectChromaScreenColor, normalizeVisualProcessing, sampleBezierMaskPoints, sampleChromaColorFromPixels, visualProcessingPreviewPlan } from './aicutVisualProcessing.js';
import * as visualProcessing from './aicutVisualProcessing.js';

test('高级画面参数提供安全默认值并限制边界', () => {
  assert.deepEqual(normalizeVisualProcessing(), {
    color: { exposure: 0, temperature: 0, tint: 0, highlights: 0, shadows: 0, gamma: 1, curve: [0, .25, .5, .75, 1], curvePoints: [{ x: 0, y: 0 }, { x: .25, y: .25 }, { x: .5, y: .5 }, { x: .75, y: .75 }, { x: 1, y: 1 }] },
    chromaKey: { enabled: false, removeBlack: false, color: '#00ff00', similarity: .2, blend: .08 },
    mask: { type: 'none', x: .5, y: .5, width: 1, height: 1, feather: 0, invert: false, points: [], bezierPoints: [] },
    geometry: { cropLeft: 0, cropTop: 0, cropRight: 0, cropBottom: 0, flipHorizontal: false, flipVertical: false },
    details: { sharpen: 0, vignette: 0, grain: 0 },
    frameInterpolation: 'off',
    stabilization: { enabled: false, smoothness: 15, zoom: 0, tripod: false },
    objectRemoval: { enabled: false, x: .5, y: .5, width: .2, height: .2 },
    personSegmentation: { enabled: false, matteAssetId: '', threshold: .5, feather: .1, signature: '' }
  });
  const value = normalizeVisualProcessing({
    color: { exposure: 9, temperature: -2, tint: 2, highlights: -4, shadows: 4, gamma: 9, curve: [-1, .2, 2, .8, 4] },
    chromaKey: { enabled: true, removeBlack: true, color: 'bad', similarity: 2, blend: -1 },
    mask: { type: 'ellipse', x: 2, y: -1, width: 0, height: 4, feather: 2, invert: true },
    geometry: { cropLeft: -1, cropTop: 2, cropRight: .2, cropBottom: .3, flipHorizontal: true, flipVertical: true },
    details: { sharpen: 4, vignette: -.5, grain: .65 }
  });
  assert.deepEqual(value.color, { exposure: 2, temperature: -1, tint: 1, highlights: -1, shadows: 1, gamma: 2, curve: [0, .2, 1, .8, 1], curvePoints: [{ x: 0, y: 0 }, { x: .25, y: .2 }, { x: .5, y: 1 }, { x: .75, y: .8 }, { x: 1, y: 1 }] });
  assert.deepEqual(value.chromaKey, { enabled: true, removeBlack: true, color: '#00ff00', similarity: 1, blend: 0 });
  assert.deepEqual(value.mask, { type: 'ellipse', x: 1, y: 0, width: .05, height: 1, feather: .5, invert: true, points: [], bezierPoints: [] });
  assert.deepEqual(value.geometry, { cropLeft: 0, cropTop: .9, cropRight: .2, cropBottom: .3, flipHorizontal: true, flipVertical: true });
  assert.deepEqual(value.details, { sharpen: 1, vignette: 0, grain: .65 });
  assert.equal(normalizeVisualProcessing({ frameInterpolation: 'optical-flow' }).frameInterpolation, 'optical-flow');
  assert.equal(normalizeVisualProcessing({ frameInterpolation: 'unknown' }).frameInterpolation, 'off');
  assert.equal(visualProcessingPreviewPlan({ frameInterpolation: 'blend' }).frameInterpolation, 'blend');
  assert.deepEqual(normalizeVisualProcessing({ stabilization: { enabled: true, smoothness: 60, zoom: 12, tripod: true } }).stabilization, { enabled: true, smoothness: 60, zoom: 12, tripod: true });
  assert.deepEqual(normalizeVisualProcessing({ stabilization: { enabled: true, smoothness: 999, zoom: -4 } }).stabilization, { enabled: true, smoothness: 100, zoom: 0, tripod: false });
});

test('区域物体消除提供安全默认框并把越界输入限制在画面内', () => {
  const defaults = normalizeVisualProcessing().objectRemoval;
  assert.deepEqual(defaults, { enabled: false, x: .5, y: .5, width: .2, height: .2 });
  assert.deepEqual(normalizeVisualProcessing({ objectRemoval: { enabled: false, x: 0, y: 0, width: 0, height: 0 } }).objectRemoval, defaults, '旧工程中的全零结构回退到居中安全框');
  assert.deepEqual(normalizeVisualProcessing({ objectRemoval: { enabled: true, x: 2, y: -1, width: 0, height: 4 } }).objectRemoval, {
    enabled: true,
    x: 1,
    y: 0,
    width: .05,
    height: 1
  });
  assert.deepEqual(visualProcessingPreviewPlan({ objectRemoval: { enabled: true, x: .4, y: .3, width: .25, height: .15 } }).objectRemoval, {
    enabled: true,
    x: .4,
    y: .3,
    width: .25,
    height: .15
  });
});

test('吸管按画面坐标读取像素并返回稳定十六进制颜色', () => {
  const pixels = new Uint8ClampedArray([
    1, 2, 3, 255, 17, 34, 51, 255,
    170, 187, 204, 255, 250, 240, 230, 255
  ]);
  assert.equal(sampleChromaColorFromPixels(pixels, 2, 2, 1, 0), '#112233');
  assert.equal(sampleChromaColorFromPixels(pixels, 2, 2, 9, 9), null);
  assert.equal(sampleChromaColorFromPixels(null, 2, 2, 0, 0), null);
});

test('自动识别幕布优先选择画面边缘占比最高的高饱和颜色', () => {
  const width = 8;
  const height = 8;
  const pixels = new Uint8ClampedArray(width * height * 4);
  for (let y = 0; y < height; y += 1) {
    for (let x = 0; x < width; x += 1) {
      const index = (y * width + x) * 4;
      const border = x < 2 || x >= width - 2 || y < 2 || y >= height - 2;
      const color = border ? [7, 246, 18] : [180, 90, 40];
      pixels.set([...color, 255], index);
    }
  }
  assert.equal(detectChromaScreenColor(pixels, width, height), '#00f010');
  assert.equal(detectChromaScreenColor(new Uint8ClampedArray(0), 0, 0), null);
});

test('浏览器预览计划输出可直接用于滤镜、色键和蒙版的参数', () => {
  const plan = visualProcessingPreviewPlan({
    color: { exposure: 1, temperature: .4, tint: -.2, highlights: .3, shadows: -.1, gamma: 1.2, curve: [0, .18, .55, .88, 1] },
    chromaKey: { enabled: true, color: '#112233', similarity: .3, blend: .1 },
    mask: { type: 'rectangle', x: .5, y: .5, width: .8, height: .6, feather: .12 },
    geometry: { cropLeft: .1, cropTop: .2, cropRight: .1, cropBottom: 0, flipHorizontal: true },
    details: { sharpen: .4, vignette: .55, grain: .2 }
  });
  assert.deepEqual(plan.keyRGB.map(value => Math.round(value * 255)), [17, 34, 51]);
  assert.equal(plan.mask.type, 'rectangle');
  assert.deepEqual(plan.curve, [0, .18, .55, .88, 1]);
  assert.equal(plan.mask.feather, .12);
  assert.equal(plan.geometry.cropTop, .2);
  assert.equal(plan.geometry.flipHorizontal, true);
  assert.equal(plan.exposureMultiplier, 2);
  assert.deepEqual(plan.details, { sharpen: .4, vignette: .55, grain: .2 });
});

test('自由调色曲线保留任意横纵坐标并兼容旧五点工程', () => {
  const points = [{ x: 0, y: .05 }, { x: .18, y: .3 }, { x: .63, y: .52 }, { x: 1, y: .95 }];
  const normalized = normalizeVisualProcessing({ color: { curvePoints: points } });
  assert.deepEqual(normalized.color.curvePoints, points);
  assert.equal(normalized.color.curve.length, 5, '继续生成旧版五点采样，保证旧预览代码和存量工程兼容');
  assert.deepEqual(visualProcessingPreviewPlan({ color: { curvePoints: points } }).curvePoints, points);
});

test('手绘多边形蒙版保留3到12个二维控制点并为新蒙版提供安全默认形状', () => {
  const points = [{ x: .12, y: .2 }, { x: .82, y: .18 }, { x: .68, y: .86 }, { x: .25, y: .74 }];
  const normalized = normalizeVisualProcessing({ mask: { type: 'polygon', points, feather: .04, invert: true } });
  assert.deepEqual(normalized.mask.points, points);
  assert.equal(normalized.mask.type, 'polygon');
  assert.equal(normalized.mask.feather, .04);
  assert.equal(normalized.mask.invert, true);
  const fallback = normalizeVisualProcessing({ mask: { type: 'polygon', points: [{ x: 2, y: 2 }] } });
  assert.equal(fallback.mask.points.length, 4);
  assert.ok(fallback.mask.points.every(point => point.x >= 0 && point.x <= 1 && point.y >= 0 && point.y <= 1));
});

test('贝塞尔曲线蒙版保留锚点与双控制柄并稳定采样为闭合轮廓', () => {
  const bezierPoints = [
    { x: .2, y: .2, inX: .1, inY: .35, outX: .35, outY: .1 },
    { x: .8, y: .2, inX: .65, inY: .1, outX: .9, outY: .35 },
    { x: .8, y: .8, inX: .9, inY: .65, outX: .65, outY: .9 },
    { x: .2, y: .8, inX: .35, inY: .9, outX: .1, outY: .65 }
  ];
  const normalized = normalizeVisualProcessing({ mask: { type: 'bezier', bezierPoints } });
  assert.deepEqual(normalized.mask.bezierPoints, bezierPoints);
  assert.equal(normalized.mask.points.length, 24);
  assert.deepEqual(normalized.mask.points[0], { x: .2, y: .2 });
  assert.deepEqual(sampleBezierMaskPoints(bezierPoints), normalized.mask.points);
  const fallback = normalizeVisualProcessing({ mask: { type: 'bezier', bezierPoints: [{ x: 2, y: 2 }] } });
  assert.equal(fallback.mask.bezierPoints.length, 4);
});

test('低耗防抖预览能从灰度采样中识别整幅画面的水平平移', () => {
  assert.equal(typeof visualProcessing.estimateFrameTranslation, 'function');
  const width = 12;
  const height = 8;
  const previous = new Uint8Array(width * height);
  for (let y = 2; y < 6; y += 1) {
    for (let x = 2; x < 8; x += 1) previous[y * width + x] = 40 + x * 9 + y * 5;
  }
  const current = new Uint8Array(width * height);
  for (let y = 0; y < height; y += 1) {
    for (let x = 0; x < width - 2; x += 1) current[y * width + x + 2] = previous[y * width + x];
  }
  const motion = visualProcessing.estimateFrameTranslation(previous, current, width, height, 3);
  assert.deepEqual({ x: motion.x, y: motion.y }, { x: 2, y: 0 });
  assert.ok(motion.confidence > 0);
});

test('防抖平滑器把累计相机位移转换成采样校正并可重置', () => {
  assert.equal(typeof visualProcessing.createPreviewStabilizer, 'function');
  const stabilizer = visualProcessing.createPreviewStabilizer({ smoothness: 20, tripod: true });
  assert.deepEqual(stabilizer.update({ x: 2, y: -1 }), { x: 2, y: -1 });
  assert.deepEqual(stabilizer.update({ x: 1, y: 1 }), { x: 3, y: 0 });
  stabilizer.reset();
  assert.deepEqual(stabilizer.update({ x: 0, y: 0 }), { x: 0, y: 0 });
});

test('区域跟踪只比较目标框并识别目标的水平位移', () => {
  assert.equal(typeof visualProcessing.estimateRegionTranslation, 'function');
  const width = 16;
  const height = 10;
  const previous = new Uint8Array(width * height);
  for (let y = 3; y < 7; y += 1) {
    for (let x = 4; x < 9; x += 1) previous[y * width + x] = 30 + x * 13 + y * 7;
  }
  const current = new Uint8Array(width * height);
  for (let y = 0; y < height; y += 1) {
    for (let x = 0; x < width - 3; x += 1) current[y * width + x + 3] = previous[y * width + x];
  }
  const motion = visualProcessing.estimateRegionTranslation(previous, current, width, height, {
    x: 4 / width, y: 3 / height, width: 5 / width, height: 4 / height
  }, 4);
  assert.deepEqual({ x: motion.x, y: motion.y }, { x: 3, y: 0 });
  assert.ok(motion.confidence > 0);
});

test('区域跟踪序列累计位移并输出可写入时间线的归一化位置', () => {
  assert.equal(typeof visualProcessing.trackRegionSequence, 'function');
  const width = 12;
  const height = 8;
  const makeFrame = (shiftX) => {
    const frame = new Uint8Array(width * height);
    for (let y = 2; y < 6; y += 1) {
      for (let x = 2; x < 6; x += 1) frame[y * width + x + shiftX] = 50 + x * 17 + y * 9;
    }
    return frame;
  };
  const tracked = visualProcessing.trackRegionSequence([
    { timeMs: 1000, pixels: makeFrame(0) },
    { timeMs: 1500, pixels: makeFrame(1) },
    { timeMs: 2000, pixels: makeFrame(2) }
  ], width, height, { x: 2 / width, y: 2 / height, width: 4 / width, height: 4 / height }, { maxShift: 2 });
  assert.deepEqual(tracked.map(item => item.offsetMs), [0, 500, 1000]);
  assert.deepEqual(tracked.map(item => Math.round(item.x * width)), [0, 1, 2]);
  assert.deepEqual(tracked.map(item => Math.round(item.y * height)), [0, 0, 0]);
});
