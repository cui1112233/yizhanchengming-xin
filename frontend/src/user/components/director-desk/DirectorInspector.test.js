import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import { createRequire } from 'node:module';
import { pathToFileURL } from 'node:url';
import React from 'react';
import { renderToStaticMarkup } from 'react-dom/server';
import { transformSync } from 'esbuild';

const require = createRequire(import.meta.url);
const source = fs.readFileSync(new URL('./DirectorInspector.jsx', import.meta.url), 'utf8');
const compiled = transformSync(source, { loader: 'jsx', jsx: 'automatic', format: 'esm' }).code
  .replace('"react/jsx-runtime"', JSON.stringify(pathToFileURL(require.resolve('react/jsx-runtime')).href))
  .replace('"react"', JSON.stringify(pathToFileURL(require.resolve('react')).href))
  .replace('"../../pages/agent-director-desk.js"', JSON.stringify(new URL('../../pages/agent-director-desk.js', import.meta.url).href))
  .replace('"./director-inspector-state.js"', JSON.stringify(new URL('./director-inspector-state.js', import.meta.url).href));
const { default: Inspector } = await import(`data:text/javascript;base64,${Buffer.from(compiled).toString('base64')}`);

test('inspector renders actor asset name and axis controls with accessible labels', () => {
  const markup = renderToStaticMarkup(React.createElement(Inspector, { project: { actors: [{ id: 'a', assetId: 'image-a' }] }, images: [{ id: 'image-a', name: '角色图片' }], selection: { kind: 'actor', id: 'a' } }));
  for (const label of ['角色位置 X', '角色旋转 X', '角色旋转 Y', '角色旋转 Z', '角色缩放 Z', '角色统一缩放']) assert.ok(markup.includes(`aria-label="${label}"`));
  assert.match(markup, /角色图片/);
  assert.doesNotMatch(markup, /image-a/);
});

test('inspector falls back to asset id when actor image name is unavailable', () => {
  const markup = renderToStaticMarkup(React.createElement(Inspector, { project: { actors: [{ id: 'a', assetId: 'image-a' }] }, images: [], selection: { kind: 'actor', id: 'a' } }));
  assert.match(markup, /image-a/);
});

test('inspector renders scene information and camera controls for their selections', () => {
  const project = { scene: { assetId: 'bg', name: '场景图' }, aspectRatio: '16:9' };
  const scene = renderToStaticMarkup(React.createElement(Inspector, { project, selection: { kind: 'scene', id: 'scene' } }));
  assert.match(scene, /场景图/); assert.match(scene, /16:9/); assert.match(scene, /更换场景图片/);
  const camera = renderToStaticMarkup(React.createElement(Inspector, { project, selection: { kind: 'camera', id: 'camera-main' } }));
  assert.match(camera, /机位目标 Z/); assert.match(camera, /机位滚转/); assert.match(camera, /机位 FOV/); assert.match(camera, /设为活动机位/);
  assert.doesNotMatch(camera, /机位旋转 [XYZ]/);
});

test('camera semantic controls read the selected camera values', () => {
  const project = { cameras: [{ id: 'camera-main', shotSize: 'wide', focalLength: 35, azimuth: 20, elevation: -5 }] };
  const markup = renderToStaticMarkup(React.createElement(Inspector, { project, selection: { kind: 'camera', id: 'camera-main' } }));
  assert.match(markup, /<option value="wide" selected="">大全景<\/option>/);
  assert.match(markup, /aria-label="焦段 mm"[^>]*value="35"/);
});
