import test from 'node:test';
import assert from 'node:assert/strict';
import { shouldRecoverCanvasViewport } from './agent-canvas-viewport.js';

const canvasSize = { width: 1200, height: 800 };
const nodeSizes = {
  visible: { width: 320, height: 180 },
  partial: { width: 320, height: 180 },
  far: { width: 320, height: 180 }
};

test('keeps a saved viewport when at least one workflow node is visible', () => {
  const nodes = [
    { id: 'visible', position: { x: 120, y: 80 } },
    { id: 'far', position: { x: 4200, y: 3200 } }
  ];

  assert.equal(shouldRecoverCanvasViewport(nodes, { x: 0, y: 0, zoom: 1 }, canvasSize, nodeSizes), false);
});

test('keeps a saved viewport when a node still intersects the usable canvas edge', () => {
  const nodes = [{ id: 'partial', position: { x: 1140, y: 740 } }];

  assert.equal(shouldRecoverCanvasViewport(nodes, { x: 0, y: 0, zoom: 1 }, canvasSize, nodeSizes), false);
});

test('recovers a saved viewport when every workflow node is off screen', () => {
  const nodes = [
    { id: 'visible', position: { x: 1600, y: 1200 } },
    { id: 'far', position: { x: 3600, y: -900 } }
  ];

  assert.equal(shouldRecoverCanvasViewport(nodes, { x: -40, y: 20, zoom: 0.8 }, canvasSize, nodeSizes), true);
});

test('does not recover empty canvases or canvases without a valid saved viewport', () => {
  assert.equal(shouldRecoverCanvasViewport([], { x: 0, y: 0, zoom: 1 }, canvasSize, nodeSizes), false);
  assert.equal(shouldRecoverCanvasViewport([{ id: 'visible', position: { x: 0, y: 0 } }], null, canvasSize, nodeSizes), false);
  assert.equal(shouldRecoverCanvasViewport([{ id: 'visible', position: { x: 0, y: 0 } }], { x: 0, y: 0, zoom: 0 }, canvasSize, nodeSizes), false);
});

