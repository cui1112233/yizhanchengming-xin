import test from 'node:test';
import assert from 'node:assert/strict';
import {
  applyCanvasEdgeAppearance,
  appearanceCssVariables,
  defaultCanvasAppearance,
  edgeStyleForAppearance,
  normalizeCanvasAppearance
} from './agent-canvas-appearance.js';

test('canvas appearance defaults enable the V88 flow presentation', () => {
  assert.deepEqual(defaultCanvasAppearance, {
    edgeColor: 'neutral',
    edgeAnimated: false,
    edgeWidth: 'thin',
    edgesVisible: true,
    backgroundDots: true,
    alignmentGuides: true
  });
  assert.deepEqual(normalizeCanvasAppearance(), defaultCanvasAppearance);
});

test('canvas appearance retains explicit false settings and rejects unsupported presets', () => {
  const appearance = normalizeCanvasAppearance({
    edgeColor: 'unsupported',
    edgeAnimated: false,
    edgeWidth: 'huge',
    edgesVisible: false,
    backgroundDots: false,
    alignmentGuides: false
  });
  assert.deepEqual(appearance, {
    edgeColor: 'neutral',
    edgeAnimated: false,
    edgeWidth: 'thin',
    edgesVisible: false,
    backgroundDots: false,
    alignmentGuides: false
  });
});

test('legacy colored edge preferences remain available after neutral becomes the default', () => {
  assert.deepEqual(edgeStyleForAppearance({ edgeColor: 'coral', edgeWidth: 'default', edgeAnimated: true }), {
    stroke: 'var(--agent-edge-coral)',
    strokeWidth: 2
  });
});

test('canvas appearance exposes safe edge styles and background variables', () => {
  assert.deepEqual(edgeStyleForAppearance({ edgeColor: 'cyan', edgeWidth: 'thick' }), {
    stroke: 'var(--agent-edge-cyan)',
    strokeWidth: 3
  });
  assert.deepEqual(appearanceCssVariables({ edgeColor: 'gold', edgeWidth: 'thin', backgroundDots: false }), {
    '--agent-edge-color': 'var(--agent-edge-gold)',
    '--agent-edge-width': '1px',
    '--agent-dot-opacity': '0'
  });
});

test('edge visual preferences do not override the workspace connection visibility', () => {
  const edge = { id: 'edge-image-video', source: 'image-1', target: 'video-1', hidden: false };
  assert.equal(applyCanvasEdgeAppearance(edge, { edgesVisible: false }).hidden, false);
});
