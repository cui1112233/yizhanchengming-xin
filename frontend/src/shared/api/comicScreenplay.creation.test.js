import assert from 'node:assert/strict';
import test from 'node:test';
import { listComicScreenplayStyles, parseComicScreenplayDraftSource } from './comicScreenplay.js';

test('画风目录向后端请求且不返回生成提示词', async () => {
  const rows = await listComicScreenplayStyles(async path => {
    assert.equal(path, '/api/comic-screenplays/styles');
    return { styles: [{ id: 'ink', name: '水墨', previewUrl: '/api/presets/comic-style/ink/preview' }] };
  });
  assert.equal(rows[0].id, 'ink');
});

test('新建前解析文件无须作品 ID，也不调用创建接口', async () => {
  const file = new File(['原文'], '原文.txt', { type: 'text/plain' });
  let calls = 0;
  assert.equal(await parseComicScreenplayDraftSource(file, async (path, options) => {
    calls += 1;
    assert.equal(path, '/api/comic-screenplays/source/parse');
    assert.equal(options.method, 'POST');
    assert.equal(options.body.get('file').name, '原文.txt');
    return { content: '原文' };
  }), '原文');
  assert.equal(calls, 1);
  await assert.rejects(() => parseComicScreenplayDraftSource(new File(['a'], 'bad.exe'), () => { throw new Error('不能请求'); }), /TXT.*DOCX/);
});
