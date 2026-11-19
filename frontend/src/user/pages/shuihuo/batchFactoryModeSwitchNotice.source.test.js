import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'BatchFactoryUnifiedSettingsModal.jsx'), 'utf8');

// 模式切换确认弹窗必须明确告知「数据全部保留」：用户担心切换漫画/视频
// 模式会丢数据，弹窗文案是第一道安抚口径。
test('production mode switch notice keeps the data-preserved wording', () => {
  assert.match(source, /漫画页、视频与资产数据全部保留，随时可切回/);
  assert.match(source, /在途\$\{previousMode === 'video' \? '视频' : '漫画'\}任务不会迁移/);
});
