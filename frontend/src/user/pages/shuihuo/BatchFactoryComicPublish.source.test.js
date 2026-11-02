import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'BatchFactoryNovelList.jsx'), 'utf8');

test('publishes a comic collection only after confirming the N8 fields', () => {
	assert.match(source, /getN8ComicPublishStatus/);
	assert.match(source, /publishComicCollection/);
	assert.match(source, /发布到漫画后台/);
	assert.match(source, /音频分类 ID/);
	assert.match(source, /作品 ID/);
	assert.match(source, /2 到 35 页/);
	assert.match(source, /workIdsText/);
	assert.match(source, /comicPublishSettings/);
});

test('reuses the server-managed producer account selector when confirming publication', () => {
	assert.match(source, /producerAccounts/);
	assert.match(source, /producerAccountOptions/);
	assert.match(source, /showSearch/);
	assert.match(source, /账号目录尚未配置/);
	assert.doesNotMatch(source, /<Input value=\{publishForm\.producerAdminName\}/);
});
