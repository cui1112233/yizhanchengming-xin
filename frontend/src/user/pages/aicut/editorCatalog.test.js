import assert from 'node:assert/strict';
import test from 'node:test';
import {
  catalogById,
  catalogForTab,
  editorCatalog,
  normalizeCatalogParameters,
} from './editorCatalog.js';

test('资源目录提供约定数量的四类中文资源', () => {
  assert.equal(catalogForTab('filter').length, 16);
  assert.equal(catalogForTab('effect').length, 12);
  assert.equal(catalogForTab('transition').length, 16);
  assert.equal(catalogForTab('text').length, 12);
  assert.equal(editorCatalog.length, 56);
});

test('每项资源都有唯一 ID、中文名称和浏览器与导出映射', () => {
  const ids = editorCatalog.map((item) => item.id);
  assert.equal(new Set(ids).size, ids.length);
  for (const item of editorCatalog) {
    assert.match(item.labelZh, /[\u3400-\u9fff]/, `${item.id} 缺少中文名称`);
    assert.ok(item.description.length > 0, `${item.id} 缺少说明`);
    assert.ok(item.supportedKinds.length > 0, `${item.id} 缺少适用对象`);
    assert.ok(item.previewAdapter, `${item.id} 缺少浏览器预览映射`);
    assert.ok(item.renderAdapter, `${item.id} 缺少服务端导出映射`);
  }
});

test('参数声明包含有效边界且查找未知资源返回空值', () => {
  assert.equal(catalogById('not-exists'), null);
  assert.equal(catalogById('filter-warm-sun')?.labelZh, '暖阳');
  for (const item of editorCatalog) {
    for (const parameter of item.parameters) {
      assert.ok(parameter.key);
      assert.match(parameter.labelZh, /[\u3400-\u9fff]/);
      if (parameter.type === 'number') {
        assert.ok(Number.isFinite(parameter.min));
        assert.ok(Number.isFinite(parameter.max));
        assert.ok(parameter.min <= parameter.defaultValue && parameter.defaultValue <= parameter.max);
        assert.ok(parameter.step > 0);
      }
    }
  }
});

test('参数规范化会截断数值、使用默认值并丢弃未知字段', () => {
  assert.deepEqual(
    normalizeCatalogParameters('effect-blur', { intensity: 9, radius: -20, ignored: 88 }),
    { intensity: 1, radius: 0 },
  );
  assert.deepEqual(
    normalizeCatalogParameters('transition-cross-dissolve', { durationMs: 'oops' }),
    { durationMs: 500 },
  );
  assert.deepEqual(normalizeCatalogParameters('not-exists', { intensity: 1 }), {});
});
