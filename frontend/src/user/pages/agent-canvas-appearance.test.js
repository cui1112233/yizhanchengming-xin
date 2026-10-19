import test from 'node:test';
import assert from 'node:assert/strict';
import {
  appearanceCssVariables,
  defaultCanvasAppearance,
  edgeStyleForAppearance,
  normalizeCanvasAppearance
} from './agent-canvas-appearance.js';

test('canvas appearance defaults enable the V88 flow presentation', () => {
  assert.deepEqual(defaultCanvasAppearance, {
    edgeColor: 'coral',
    edgeAnimated: true,
    edgeWidth: 'default',
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
    backgroundDots: false,
    alignmentGuides: false
  });
  assert.deepEqual(appearance, {
    edgeColor: 'coral',
    edgeAnimated: false,
    edgeWidth: 'default',
    backgroundDots: false,
    alignmentGuides: false
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
