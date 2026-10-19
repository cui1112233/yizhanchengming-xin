export const canvasEdgeColorOptions = Object.freeze([
  Object.freeze({ value: 'coral', label: '珊瑚' }),
  Object.freeze({ value: 'cyan', label: '青蓝' }),
  Object.freeze({ value: 'violet', label: '紫罗兰' }),
  Object.freeze({ value: 'gold', label: '鎏金' })
]);

export const canvasEdgeWidthOptions = Object.freeze([
  Object.freeze({ value: 'thin', label: '细' }),
  Object.freeze({ value: 'default', label: '默认' }),
  Object.freeze({ value: 'thick', label: '粗' })
]);

export const defaultCanvasAppearance = Object.freeze({
  edgeColor: 'coral',
  edgeAnimated: true,
  edgeWidth: 'default',
  backgroundDots: true,
  alignmentGuides: true
});

const edgeColors = new Set(canvasEdgeColorOptions.map(option => option.value));
const edgeWidths = new Set(canvasEdgeWidthOptions.map(option => option.value));
const edgeWidthPixels = Object.freeze({ thin: 1, default: 2, thick: 3 });

export function normalizeCanvasAppearance(value) {
  const input = value && typeof value === 'object' ? value : {};
  const edgeColor = edgeColors.has(input.edgeColor) ? input.edgeColor : defaultCanvasAppearance.edgeColor;
  const edgeWidth = edgeWidths.has(input.edgeWidth) ? input.edgeWidth : defaultCanvasAppearance.edgeWidth;
  return {
    edgeColor,
    edgeAnimated: typeof input.edgeAnimated === 'boolean' ? input.edgeAnimated : defaultCanvasAppearance.edgeAnimated,
    edgeWidth,
    backgroundDots: typeof input.backgroundDots === 'boolean' ? input.backgroundDots : defaultCanvasAppearance.backgroundDots,
    alignmentGuides: typeof input.alignmentGuides === 'boolean' ? input.alignmentGuides : defaultCanvasAppearance.alignmentGuides
  };
}

export function edgeStyleForAppearance(value) {
  const appearance = normalizeCanvasAppearance(value);
  return {
    stroke: `var(--agent-edge-${appearance.edgeColor})`,
    strokeWidth: edgeWidthPixels[appearance.edgeWidth]
  };
}

export function appearanceCssVariables(value) {
  const appearance = normalizeCanvasAppearance(value);
  return {
    '--agent-edge-color': `var(--agent-edge-${appearance.edgeColor})`,
    '--agent-edge-width': `${edgeWidthPixels[appearance.edgeWidth]}px`,
    '--agent-dot-opacity': appearance.backgroundDots ? '0.5' : '0'
  };
}
