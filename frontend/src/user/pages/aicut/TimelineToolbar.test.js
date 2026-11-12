import assert from 'node:assert/strict';
import test from 'node:test';
import React from 'react';
import { renderToStaticMarkup } from 'react-dom/server';
import { createServer } from 'vite';
import { fileURLToPath } from 'node:url';
import { buildTimelinePrimaryActions } from './timelineTools.js';

test('toolbar renders bound primary buttons with disabled reasons and hides an empty more menu', async () => {
  const server = await createServer({ root: fileURLToPath(new URL('../../../../', import.meta.url)), configFile: false, esbuild: { jsx: 'automatic' }, optimizeDeps: { noDiscovery: true, include: [] }, server: { middlewareMode: true }, appType: 'custom' });
  try {
    const { default: TimelineToolbar } = await server.ssrLoadModule('/src/user/pages/aicut/TimelineToolbar.jsx');
    const primaryActions = buildTimelinePrimaryActions({}, { delete: () => {}, addMarker: () => {} });
    const markup = renderToStaticMarkup(React.createElement(TimelineToolbar, { primaryActions }));
    assert.match(markup, /aria-label="删除"/);
    assert.match(markup, /title="删除：请先选择片段"/);
    assert.match(markup, /disabled=""/);
    assert.match(markup, /aria-label="添加标记"/);
    assert.doesNotMatch(markup, /aria-label="更多操作"/);
    const withMenu = renderToStaticMarkup(React.createElement(TimelineToolbar, { advancedItems: [{ key: 'nextEdit', label: '下一剪辑点', onClick: () => {} }] }));
    assert.match(withMenu, /aria-label="更多操作"/);
  } finally {
    await server.close();
  }
});
