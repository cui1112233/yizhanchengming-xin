import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

test('声像滑块只在松开时提交，有数值和居中按钮', () => {
  const source = readFileSync(new URL('./InspectorPanel.jsx', import.meta.url), 'utf8');
  assert.match(source, /aria-label="左右声像" min=\{-1\} max=\{1\} step=\{\.01\}/);
  assert.match(source, /onChangeComplete=\{\(pan\) => onSetAudioProcessing\(\{ \.\.\.audioProcessing, pan \}\)\}/);
  assert.match(source, /onClick=\{\(\) => onSetAudioProcessing\(\{ \.\.\.audioProcessing, pan: 0 \}\)\}>居中/);
  assert.match(source, /audioProcessing.pan\.toFixed\(2\)/);
});
