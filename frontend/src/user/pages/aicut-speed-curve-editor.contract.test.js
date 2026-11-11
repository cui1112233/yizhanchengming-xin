import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

const editor = readFileSync(new URL('./aicut/SpeedCurveEditor.jsx', import.meta.url), 'utf8');
const inspector = readFileSync(new URL('./aicut/InspectorPanel.jsx', import.meta.url), 'utf8');

test('自由变速曲线提供拖点、新增、删除和精确输入且提交真实命令', () => {
  for (const evidence of ['onPointerDown', 'onPointerMove', '新增控制点', '删除控制点', '位置（%）', '倍率（x）']) assert.match(editor, new RegExp(evidence.replace(/[（）]/g, value => `\\${value}`)));
  assert.match(editor, /insertSpeedCurvePoint/);
  assert.match(editor, /moveSpeedCurvePoint/);
  assert.match(editor, /removeSpeedCurvePoint/);
  assert.match(editor, /onChange\(next\)/);
  assert.match(inspector, /<SpeedCurveEditor/);
  assert.match(inspector, /onChange=\{onSetSpeedCurve\}/);
});
