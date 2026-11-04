import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'BatchFactoryNovelList.jsx'), 'utf8');

test('keeps generated videos from an earlier round visible as a selectable previous version', () => {
  // 历史轮次（orphaned）的成功媒体必须挂回同一分镜并标注「上一版」，不能随轮换消失。
  assert.match(source, /const mediaVersions = bookVideoVersions\(book, productionStatus\)/);
  assert.match(source, /version\.isHistory \? ' · 上一版' : ''/);
  assert.match(source, /main\.isHistory \? ' · 上一版' : ''/);
  // 历史记录不属于当前片段库，不提供“移除候选”操作。
  assert.match(source, /!isPrimary && !isUpload && !version\.isHistory/);
  // 上传来源只允许当前 active 记录下的任务，避免 video/task 错配。
  assert.match(source, /state\.main\?\.mediaUrl && !state\.main\.isHistory/);
});
