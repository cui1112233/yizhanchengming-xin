const minimumRate = 0.25;
const maximumRate = 4;

function finite(value, fallback = 0) {
  const number = Number(value);
  return Number.isFinite(number) ? number : fallback;
}

export const speedCurvePresets = [
  { id: 'uniform', label: '匀速', points: [{ position: 0, rate: 1 }, { position: 1, rate: 1 }] },
  { id: 'hero', label: '英雄时刻', points: [{ position: 0, rate: 1 }, { position: .35, rate: .5 }, { position: .65, rate: .5 }, { position: 1, rate: 1.5 }] },
  { id: 'montage', label: '蒙太奇', points: [{ position: 0, rate: 2 }, { position: .3, rate: 2 }, { position: .5, rate: .5 }, { position: .7, rate: 2 }, { position: 1, rate: 2 }] },
  { id: 'bullet', label: '子弹时间', points: [{ position: 0, rate: 1.5 }, { position: .4, rate: .35 }, { position: .7, rate: .35 }, { position: 1, rate: 1.5 }] },
  { id: 'ramp-up', label: '渐快', points: [{ position: 0, rate: .5 }, { position: 1, rate: 2 }] },
  { id: 'ramp-down', label: '渐慢', points: [{ position: 0, rate: 2 }, { position: 1, rate: .5 }] }
];

export function normalizeSpeedCurve(points) {
  if (!Array.isArray(points) || points.length < 2 || points.length > 12) throw new Error('速度曲线需要 2–12 个控制点');
  const normalized = points.map(point => ({ position: finite(point?.position, NaN), rate: finite(point?.rate, NaN) }))
    .sort((left, right) => left.position - right.position);
  for (const point of normalized) {
    if (!Number.isFinite(point.position) || point.position < 0 || point.position > 1) throw new Error('速度曲线位置必须在 0%–100% 之间');
    if (!Number.isFinite(point.rate) || point.rate < minimumRate || point.rate > maximumRate) throw new Error('速度曲线倍率范围为 0.25x–4x');
  }
  if (normalized[0].position !== 0 || normalized.at(-1).position !== 1) throw new Error('速度曲线必须从 0% 开始并在 100% 结束');
  for (let index = 1; index < normalized.length; index += 1) {
    if (normalized[index].position <= normalized[index - 1].position) throw new Error('速度曲线控制点位置不能重复');
  }
  return normalized;
}

export function insertSpeedCurvePoint(points) {
  const curve = normalizeSpeedCurve(points);
  if (curve.length >= 12) throw new Error('速度曲线最多 12 个控制点');
  let insertAt = 1;
  let widest = -1;
  for (let index = 1; index < curve.length; index += 1) {
    const width = curve[index].position - curve[index - 1].position;
    if (width > widest) {
      widest = width;
      insertAt = index;
    }
  }
  const left = curve[insertAt - 1];
  const right = curve[insertAt];
  const next = [...curve];
  next.splice(insertAt, 0, {
    position: Math.round(((left.position + right.position) / 2) * 1000) / 1000,
    rate: Math.round(((left.rate + right.rate) / 2) * 100) / 100
  });
  return normalizeSpeedCurve(next);
}

export function moveSpeedCurvePoint(points, index, position, rate) {
  const curve = normalizeSpeedCurve(points);
  if (!Number.isInteger(index) || index < 0 || index >= curve.length) throw new Error('速度曲线控制点不存在');
  const next = curve.map(point => ({ ...point }));
  const isFirst = index === 0;
  const isLast = index === next.length - 1;
  const minimumPosition = isFirst ? 0 : next[index - 1].position + .001;
  const maximumPosition = isLast ? 1 : next[index + 1].position - .001;
  next[index] = {
    position: isFirst ? 0 : isLast ? 1 : Math.round(Math.min(maximumPosition, Math.max(minimumPosition, finite(position))) * 1000) / 1000,
    rate: Math.round(Math.min(maximumRate, Math.max(minimumRate, finite(rate, next[index].rate))) * 100) / 100
  };
  return normalizeSpeedCurve(next);
}

export function removeSpeedCurvePoint(points, index) {
  const curve = normalizeSpeedCurve(points);
  if (curve.length <= 2) throw new Error('速度曲线至少保留两个端点');
  if (!Number.isInteger(index) || index < 0 || index >= curve.length) throw new Error('速度曲线控制点不存在');
  if (index === 0 || index === curve.length - 1) throw new Error('速度曲线端点不能删除');
  return normalizeSpeedCurve(curve.filter((_, pointIndex) => pointIndex !== index));
}

export function speedCurveRateAt(points, position) {
  const curve = normalizeSpeedCurve(points);
  const at = Math.min(1, Math.max(0, finite(position)));
  const rightIndex = curve.findIndex(point => point.position >= at);
  if (rightIndex <= 0) return curve[0].rate;
  const left = curve[rightIndex - 1];
  const right = curve[rightIndex];
  const ratio = (at - left.position) / (right.position - left.position);
  return left.rate + (right.rate - left.rate) * ratio;
}

function speedCurveIntegral(points, endPosition = 1) {
  const curve = normalizeSpeedCurve(points);
  const end = Math.min(1, Math.max(0, finite(endPosition)));
  let total = 0;
  for (let index = 1; index < curve.length; index += 1) {
    const left = curve[index - 1];
    const right = curve[index];
    if (end <= left.position) break;
    const segmentEnd = Math.min(end, right.position);
    const span = segmentEnd - left.position;
    if (span <= 0) continue;
    const endRate = left.rate + (right.rate - left.rate) * (span / (right.position - left.position));
    total += span * (left.rate + endRate) / 2;
    if (end <= right.position) break;
  }
  return total;
}

export function speedCurveAverageRate(points) {
  return speedCurveIntegral(points, 1);
}

export function speedCurveDurationMs(sourceSpanMs, points) {
  const source = Math.max(1, finite(sourceSpanMs, 1));
  return Math.max(1, Math.round(source / speedCurveAverageRate(points)));
}

export function speedCurveSourceOffsetMs(outputOffsetMs, outputDurationMs, points) {
  const duration = Math.max(1, finite(outputDurationMs, 1));
  const position = Math.min(1, Math.max(0, finite(outputOffsetMs) / duration));
  return Math.max(0, Math.round(duration * speedCurveIntegral(points, position)));
}

export function sliceSpeedCurve(points, startPosition, endPosition) {
  const curve = normalizeSpeedCurve(points);
  const start = Math.min(1, Math.max(0, finite(startPosition)));
  const end = Math.min(1, Math.max(0, finite(endPosition, 1)));
  if (end <= start) throw new Error('速度曲线裁剪区间无效');
  const selected = [
    { position: start, rate: speedCurveRateAt(curve, start) },
    ...curve.filter(point => point.position > start && point.position < end),
    { position: end, rate: speedCurveRateAt(curve, end) }
  ];
  const span = end - start;
  return selected.map(point => ({ position: (point.position - start) / span, rate: point.rate }));
}
