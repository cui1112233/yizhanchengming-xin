import test from 'node:test';
import assert from 'node:assert/strict';
import { createAntTheme } from '../shared/styles/theme.js';

test('dark theme exposes readable text levels and surfaces', () => {
  const { token } = createAntTheme('coral-dark');
  assert.equal(token.colorText, '#F5F0E8');
  assert.equal(token.colorTextSecondary, '#D6CEC2');
  assert.equal(token.colorTextTertiary, '#A79C8C');
  assert.equal(token.colorTextDisabled, 'rgba(199, 208, 221, 0.62)');
  assert.equal(token.colorBgBase, '#12100E');
  assert.equal(token.colorBgContainer, '#1A1714');
  assert.equal(token.colorBgLayout, '#12100E');
  assert.equal(token.colorBgContainerDisabled, '#0b111a');
  assert.equal(token.colorBorder, '#4A4239');
  assert.equal(token.colorBorderSecondary, '#8A7C6A');
});

test('light theme exposes readable text levels and surfaces', () => {
  const { token } = createAntTheme('coral-light');
  assert.equal(token.colorText, '#1C1917');
  assert.equal(token.colorTextSecondary, '#44403C');
  assert.equal(token.colorTextTertiary, '#6B645C');
  assert.equal(token.colorTextDisabled, '#748196');
  assert.equal(token.colorBgBase, '#F6F4F1');
  assert.equal(token.colorBgContainer, '#FFFFFF');
  assert.equal(token.colorBgLayout, '#F6F4F1');
  assert.equal(token.colorBgContainerDisabled, '#ffffff');
  assert.equal(token.colorBorder, '#D8D1C7');
  assert.equal(token.colorBorderSecondary, '#8C8478');
});

test('legacy light/dark selections still resolve through the coral scheme', () => {
  assert.deepEqual(createAntTheme('light').token, createAntTheme('coral-light').token);
  assert.deepEqual(createAntTheme('dark').token, createAntTheme('coral-dark').token);
});
