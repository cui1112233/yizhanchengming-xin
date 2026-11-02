import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';

const source = fs.readFileSync(path.resolve('src/user/components/AgentCanvas.jsx'), 'utf8');

test('图片和视频生成完成后自动采用第一份结果', () => {
  assert.match(source, /mergedGeneration\.task\?\.state === 'awaiting_confirmation'/);
  assert.match(source, /void confirmImage\(props\.id, first\.id\)/);
  assert.match(source, /videoJob\?\.state === 'awaiting_confirmation'/);
  assert.match(source, /void confirmVideo\(props\.id\)/);
});

test('自动采用前使用稳定键避免重复提交', () => {
  assert.match(source, /const autoApplyRef = useRef\(''\)/);
  assert.match(source, /autoApplyRef\.current !== key/);
});
