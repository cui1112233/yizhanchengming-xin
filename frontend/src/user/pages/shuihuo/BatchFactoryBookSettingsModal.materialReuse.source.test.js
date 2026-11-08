import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const source = readFileSync(join(dirname(fileURLToPath(import.meta.url)), 'BatchFactoryBookSettingsModal.jsx'), 'utf8');

// 23 版需求 1：原「素材使用」实际写的是 biaohong_reuse（121 的“按行号标红→复用”），
// 必须改名「标红复用」；真正控制 121 aiHeadVideoMode 的 materialReuse 要单独一项。
test('单书配置把素材使用拆成「标红复用」和「AI头部素材使用」', () => {
  assert.match(source, /<b>标红复用<\/b>/);
  assert.match(source, /<b>AI头部素材使用<\/b>/);
  assert.doesNotMatch(source, /<b>素材使用<\/b>/);
});

test('materialReuse 只绑定 AI 头部素材使用，标红复用走 biaohongReuse', () => {
  assert.match(source, /publish\.materialReuse/);
  assert.match(source, /publish\.biaohongReuse/);
});
