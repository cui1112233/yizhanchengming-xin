import test from 'node:test';
import assert from 'node:assert/strict';

import {
  novelFetchAiVersions,
  novelFetchSensitiveComparisons,
  novelFetchSubmitDiagnostic,
  novelFetchTaskSummaryItems
} from './novel-fetch-task-detail.js';

test('task detail reads persisted per-version rewrite evidence without inventing history', () => {
  const data = { versions: {ai1: '正文一', ai2: '正文二'}, meta: { rewriteModel: 'model-used', rewriteEvidenceByVersion: {
    ai1: {strategy_name: '方案一', template_name: '模板一'},
    ai2: {strategy_name: '方案二', template_name: '模板二', opening_phrase_name: '开头二'}
  }}};
  const versions = novelFetchAiVersions(data);
  assert.equal(versions[0].knowledge.template_name, '模板一');
  assert.equal(versions[1].knowledge.strategy_name, '方案二');
  const summary = Object.fromEntries(novelFetchTaskSummaryItems(data).map(row => [row.label, row.value]));
  assert.equal(summary['AI 当前版本'], 'AI2');
  assert.equal(summary['改文方案'], '方案二');
  assert.equal(summary['改文模板'], '模板二');
  assert.equal(summary['开头词'], '开头二');
  assert.equal(summary['改文模型'], 'model-used');
  data.meta.aiCurrentVersion = 'ai1';
  data.meta.rewriteEvidenceByVersion.ai1.model = 'model-one';
  const selected = Object.fromEntries(novelFetchTaskSummaryItems(data).map(row => [row.label, row.value]));
  assert.equal(selected['改文方案'], '方案一');
  assert.equal(selected['改文模型'], 'model-one');
  assert.equal(Object.fromEntries(novelFetchTaskSummaryItems({versions: {ai1:'旧正文'}}).map(row=>[row.label,row.value]))['改文模型'], '—');
});

test('submission diagnostics distinguish an empty result, real receipt and no attempt', () => {
  const empty = novelFetchSubmitDiagnostic({ result: {}, logs: [], reason: '文本模型不可用' });
  assert.equal(empty.result, null);
  assert.equal(empty.reason, '文本模型不可用');
  const receipt = novelFetchSubmitDiagnostic({result: {status:'submitted',version:'ai1',remote_receipt:{verified:true,submission_id:'287175'}},logs:[{event:'site_submit',status:'submitted',version:'ai1',time:'2026-11-13T04:44:11Z'}]});
  assert.equal(receipt.rows[0].statusLabel, '已回读确认');
  assert.equal(receipt.receiptId, '287175');
  assert.equal(receipt.verified, true);
});

test('fallback completion marks old AI failure as recovered and retains its evidence', () => {
  const data = novelFetchSensitiveComparisons({sensitive_hits:{hits:[{hit_index:1,keyword:'测试'}]},sensitive_fixed:{status:'done',items:[{hit_index:1,status:'done',source:'fallback',error:'旧模型错误',after:'处理后'}]}});
  assert.equal(data.items[0].source, '规则兜底替换');
  assert.equal(data.items[0].recovered, true);
  assert.equal(data.items[0].error, '旧模型错误');
});

test('task summary exposes stable Chinese labels instead of internal meta keys', () => {
  const items = novelFetchTaskSummaryItems({
    bookId: 'book-7',
    originalRawChars: 27831,
    processedChars: 4000,
    meta: {
      bookName: '测试小说',
      platformName: '番茄免费',
      style: '都市',
      gender: '男频',
      genderReason: '主角视角与题材判断',
      originalStatus: 'input_ready',
      sensitiveStatus: 'done',
      sensitiveHitCount: 16,
      sensitiveFixedCount: 16,
      aiStatus: 'ai_done',
      aiGeneratedCount: 2,
      batchId: 'batch-9'
    }
  });

  assert.deepEqual(items.map(item => item.label), [
    '书籍 ID', '书名', '平台', '风格', '男女频', '判断依据', '分类模型', '分类失败原因',
    '原文状态', '原文失败原因', '原文错误码', '处理字数', '敏感词状态', '敏感词模型',
    '命中 / 已处理', 'AI 状态', 'AI 失败原因', 'AI 当前版本', 'AI 文案版本', 'AI 尝试次数',
    'AI 最近失败', '网站提交失败原因', '改文模型', '改文方案', '改文模板', '开头词', '高仿参考',
    '创建时间', '更新时间', '当前批次'
  ]);
  assert.equal(items.find(item => item.label === '原文状态').value, '分类信息已就绪');
  assert.equal(items.find(item => item.label === '处理字数').value, '4000 / 27831');
  assert.equal(items.find(item => item.label === '命中 / 已处理').value, '16 / 16');
  assert.equal(items.find(item => item.label === 'AI 文案版本').value, '2 个');
  assert.equal(items.some(item => item.label === 'aiError'), false);
});

test('task summary reads fetched title, nested sensitive counts and current batch metadata', () => {
  const items = novelFetchTaskSummaryItems({
    bookId: '2094511573566019532',
    bookInfo: { workTitle: '鼓励式老公，却把行动都给了闺蜜' },
    meta: {
      currentBatchId: 'batch-20261018',
      sensitiveHits: { hit_count: 7 },
      sensitiveFixed: { fixed_count: 7 }
    }
  });

  assert.equal(items.find(item => item.label === '书名').value, '鼓励式老公，却把行动都给了闺蜜');
  assert.equal(items.find(item => item.label === '命中 / 已处理').value, '7 / 7');
  assert.equal(items.find(item => item.label === '当前批次').value, 'batch-20261018');
});

test('task summary keeps REA diagnostic provenance fields', () => {
  const items = novelFetchTaskSummaryItems({
    meta: {
      classifier_model: '分类模型',
      original_error: '原文失败',
      ai_last_attempt_count: 2,
      ai_last_attempt_error: 'AI失败',
      rewrite_model: '改文模型',
      rewrite_knowledge: { strategy_name: '高仿文章', opening_phrase_name: '强冲突开头' }
    }
  });
  const values = new Map(items.map(item => [item.label, item.value]));
  assert.equal(values.get('分类模型'), '分类模型');
  assert.equal(values.get('原文失败原因'), '原文失败');
  assert.equal(values.get('AI 尝试次数'), '2');
  assert.equal(values.get('AI 最近失败'), 'AI失败');
  assert.equal(values.get('改文模型'), '改文模型');
  assert.equal(values.get('改文方案'), '高仿文章');
  assert.equal(values.get('开头词'), '强冲突开头');
});

test('AI versions are read from versions and support object bodies and legacy fallback', () => {
  assert.deepEqual(novelFetchAiVersions({
    versions: {
      original: '原文不应进入 AI 标签',
      ai2: { content: '第二版正文' },
      ai1: '第一版正文',
      draft: '草稿不展示'
    }
  }), [
    { key: 'ai1', label: 'AI 版本 1', content: '第一版正文' },
    { key: 'ai2', label: 'AI 版本 2', content: '第二版正文' }
  ]);

  assert.deepEqual(novelFetchAiVersions({ ai_text: '旧版正文' }), [
    { key: 'ai1', label: 'AI 版本 1', content: '旧版正文' }
  ]);
});

test('sensitive comparisons join hits and fixes by hit index', () => {
  const result = novelFetchSensitiveComparisons({
    meta: { book_id: 'book-7', book_name: '测试小说' },
    sensitive_hits: {
      mode: 'group', status: 'done', hit_count: 1, updated_at: '2026-10-18 10:00',
      hits: [{ hit_index: 3, keyword: '佛', start: 8, end: 9, snippet: '仿佛有一团棉花' }]
    },
    sensitive_fixed: {
      status: 'done', fixed_count: 1, updated_at: '2026-10-18 10:01',
      items: [{ hit_index: 3, source: 'ai', model: '模型A', attempt: 2, before: '仿佛有一团棉花', after: '好像有一团棉花' }]
    }
  });

  assert.equal(result.summary.mode, 'AI 整组修改');
  assert.equal(result.summary.hitCount, 1);
  assert.equal(result.summary.fixedCount, 1);
  assert.deepEqual(result.items[0], {
    index: 3,
    keyword: '佛',
    status: '已完成',
    position: '8 - 9',
    source: 'AI 处理',
    model: '模型A',
    attempts: 2,
    error: '—',
    recovered: false,
    before: '仿佛有一团棉花',
    after: '好像有一团棉花'
  });
});

test('sensitive comparisons return a truthful empty state when no hit detail exists', () => {
  const result = novelFetchSensitiveComparisons({ meta: { sensitive_hit_count: 4, sensitive_fixed_count: 0 } });
  assert.equal(result.summary.hitCount, 4);
  assert.equal(result.summary.fixedCount, 0);
  assert.deepEqual(result.items, []);
});
