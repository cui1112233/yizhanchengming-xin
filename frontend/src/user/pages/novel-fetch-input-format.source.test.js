import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

const source = readFileSync(new URL('./NovelFetchPage.jsx', import.meta.url), 'utf8');
const css = readFileSync(new URL('./novel-fetch.css', import.meta.url), 'utf8');

test('处理页在执行与定时前使用同一份输入格式和解析结果', () => {
  assert.match(source, /normalizeNovelFetchInput/);
  assert.match(source, /const \[inputFormat, setInputFormat\] = useState\('smart'\)/);
  assert.match(source, /input_format:\s*inputFormat/);
  assert.match(source, /column_preset:\s*inputFormatLabel/);
  assert.match(source, /输入格式/);
  assert.match(source, /localInputPreview\.validRows/);
  assert.match(source, /自定义列顺序/);
  assert.match(source, /serializeNovelFetchRows\(parsed\.rows\)/);
  assert.match(source, /columnMapping:\s*inputColumnMapping/);
});

test('输入格式与解析预览区域跟随全局主题变量', () => {
  for (const selector of [
    '.novel-fetch-input-summary',
    '.novel-fetch-input-preview'
  ]) {
    const block = css.match(new RegExp(`${selector.replaceAll('.', '\\.') }\\s*\\{([^}]*)\\}`, 's'))?.[1] || '';
    assert.match(block, /background:\s*var\(--(?:legacy-panel|legacy-card|legacy-input)\)/, `${selector} must use a themed surface`);
    assert.match(block, /border:\s*1px solid var\(--legacy-border/, `${selector} must use a themed border`);
  }
});
