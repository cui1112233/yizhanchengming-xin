import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'SettingsPage.jsx'), 'utf8');

test('global settings page owns the production retention control', () => {
  assert.match(source, /name="productionRetentionDays"/);
	assert.match(source, /productionRetentionDays: \[3, 7, 14, 30\]/);
	assert.match(source, /productionRetentionDays: 7/);
	assert.match(source, /上传 TOS 并校验成功后/);
	assert.match(source, /磁盘达到 75%/);
	assert.match(source, /未上传或校验失败的文件绝不删除/);
});
