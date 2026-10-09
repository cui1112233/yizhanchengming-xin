import test from 'node:test';
import assert from 'node:assert/strict';
import { createAntTheme } from '../shared/styles/theme.js';

test('dark theme exposes readable text levels and surfaces', () => {
  const { token } = createAntTheme('dark');
  assert.equal(token.colorText, '#f5f7fb');
  assert.equal(token.colorTextSecondary, '#c7d0dd');
  assert.equal(token.colorTextTertiary, '#a7b4c5');
  assert.equal(token.colorTextDisabled, 'rgba(199, 208, 221, 0.62)');
  assert.equal(token.colorBgBase, '#070b11');
  assert.equal(token.colorBgContainer, '#0e1621');
  assert.equal(token.colorBgLayout, '#070b11');
  assert.equal(token.colorBgContainerDisabled, '#0b111a');
  assert.equal(token.colorBorder, 'rgba(151, 168, 190, 0.22)');
  assert.equal(token.colorBorderSecondary, 'rgba(176, 195, 219, 0.36)');
});

test('light theme exposes readable text levels and surfaces', () => {
  const { token } = createAntTheme('light');
  assert.equal(token.colorText, '#172033');
  assert.equal(token.colorTextSecondary, '#3f4e63');
  assert.equal(token.colorTextTertiary, '#596a80');
  assert.equal(token.colorTextDisabled, '#748196');
  assert.equal(token.colorBgBase, '#f4f6fa');
  assert.equal(token.colorBgContainer, '#ffffff');
  assert.equal(token.colorBgLayout, '#f4f6fa');
  assert.equal(token.colorBgContainerDisabled, '#ffffff');
  assert.equal(token.colorBorder, '#c8d2df');
  assert.equal(token.colorBorderSecondary, '#9daabd');
});
