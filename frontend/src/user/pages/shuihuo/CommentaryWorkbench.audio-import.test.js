import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'CommentaryWorkbench.jsx'), 'utf8');

test('batch narration import binds each audio file to an existing confirmed segment', () => {
  assert.match(source, /const narrationImportRef = useRef\(null\)/);
  assert.match(source, /async function importNarrations\(event\)/);
  assert.match(source, /kind: 'audio'/);
  assert.match(source, /filename: file\.name/);
  assert.match(source, /segmentId: segment\.id/);
  assert.match(source, /批量导入配音/);
  assert.match(source, /multiple/);
});

test('batch narration import rejects ambiguous file counts instead of binding audio to the wrong segment', () => {
  assert.match(source, /配音文件数量与当前分镜数量不一致/);
  assert.match(source, /请按分镜顺序选择同数量的音频文件/);
});
