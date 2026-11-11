import test from 'node:test';
import assert from 'node:assert/strict';

import {
  insertSpeedCurvePoint,
  moveSpeedCurvePoint,
  normalizeSpeedCurve,
  removeSpeedCurvePoint,
  speedCurveAverageRate,
  speedCurveDurationMs,
  speedCurveRateAt,
  speedCurveSourceOffsetMs,
  sliceSpeedCurve
} from './aicutSpeedCurve.js';

const curve = [
  { position: 0, rate: 0.5 },
  { position: 0.5, rate: 2 },
  { position: 1, rate: 0.5 }
];

test('速度曲线规范化位置、顺序和倍率边界', () => {
  assert.deepEqual(normalizeSpeedCurve([...curve].reverse()), curve);
  assert.throws(() => normalizeSpeedCurve([{ position: 0, rate: 1 }, { position: 0.5, rate: 5 }, { position: 1, rate: 1 }]), /0.25x–4x/);
  assert.throws(() => normalizeSpeedCurve([{ position: 0.2, rate: 1 }, { position: 1, rate: 1 }]), /必须从 0% 开始并在 100% 结束/);
});

test('速度曲线用线性积分计算平均倍率、时长和素材位置', () => {
  assert.equal(speedCurveAverageRate(curve), 1.25);
  assert.equal(speedCurveDurationMs(2000, curve), 1600);
  assert.equal(speedCurveRateAt(curve, 0.25), 1.25);
  assert.equal(speedCurveSourceOffsetMs(800, 1600, curve), 1000);
});

test('分割速度曲线时在边界补点并重新映射到完整片段', () => {
  assert.deepEqual(sliceSpeedCurve(curve, 0, 0.5), [
    { position: 0, rate: 0.5 },
    { position: 1, rate: 2 }
  ]);
  assert.deepEqual(sliceSpeedCurve(curve, 0.5, 1), [
    { position: 0, rate: 2 },
    { position: 1, rate: 0.5 }
  ]);
});

test('自由变速曲线可在最大空档新增控制点并保持顺序', () => {
  assert.deepEqual(insertSpeedCurvePoint([{ position: 0, rate: 1 }, { position: 1, rate: 3 }]), [
    { position: 0, rate: 1 },
    { position: .5, rate: 2 },
    { position: 1, rate: 3 }
  ]);
});

test('自由变速曲线拖点限制倍率和相邻位置且端点不横移', () => {
  const curve = [{ position: 0, rate: 1 }, { position: .5, rate: 2 }, { position: 1, rate: 1 }];
  assert.deepEqual(moveSpeedCurvePoint(curve, 1, 1.2, 9), [
    { position: 0, rate: 1 },
    { position: .999, rate: 4 },
    { position: 1, rate: 1 }
  ]);
  assert.deepEqual(moveSpeedCurvePoint(curve, 0, .4, .1)[0], { position: 0, rate: .25 });
});

test('自由变速曲线只允许删除中间点并至少保留两个端点', () => {
  const curve = [{ position: 0, rate: 1 }, { position: .5, rate: 2 }, { position: 1, rate: 1 }];
  assert.deepEqual(removeSpeedCurvePoint(curve, 1), [{ position: 0, rate: 1 }, { position: 1, rate: 1 }]);
  assert.throws(() => removeSpeedCurvePoint(curve, 0), /端点不能删除/);
  assert.throws(() => removeSpeedCurvePoint([{ position: 0, rate: 1 }, { position: 1, rate: 1 }], 1), /至少保留两个/);
});
