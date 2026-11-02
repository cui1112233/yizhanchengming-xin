import test from 'node:test';
import assert from 'node:assert/strict';
import { buildManualBatchSubmission, hasFetchedManualSources, manualBookIDsFromInput, mergeGroupLines, upsertPlatformGroup, replacePlatformGroup, removePlatformGroup, totalGroupBookCount } from './batchFactoryManualFetch.js';

test('collects each numeric Book ID once from a manual novel list', () => {
  assert.deepEqual(manualBookIDsFromInput('100000000001\t书A\n100000000002  书B\n100000000001\t重复'), ['100000000001', '100000000002']);
});

test('collects short Qimao numeric Book IDs from a manual novel list', () => {
  assert.deepEqual(manualBookIDsFromInput('567168 晚风惊扰旧梦\n705142 她指指方向，我未对路\n697029 旧日微尽赴朝光'), ['567168', '705142', '697029']);
});

test('creation is only ready when every listed book has fetched original text', () => {
  const ids = ['100000000001', '100000000002'];
  assert.equal(hasFetchedManualSources(ids, { '100000000001': '原文' }), false);
  assert.equal(hasFetchedManualSources(ids, { '100000000001': '原文', '100000000002': '另一段原文' }), true);
});

test('manual batch submission keeps schedule metadata with automation explicitly disabled', () => {
  const payload = buildManualBatchSubmission({
    title: ' 白月光回港 ',
    platformId: '15',
    platformName: '知乎付费',
    parseMode: 'smart',
    columnPresetId: 'full_metadata',
    columnOrder: '书籍ID,书名',
    inputText: '2084012035524801698\t白月光回港',
    sourceTextByBookId: { '2084012035524801698': '正文' },
    contentRangeLines: 5,
    contentCaptureCharacters: 4000,
    scheduledAt: '2026-09-20T10:00:00.000Z'
  });

  assert.equal(payload.title, '白月光回港');
  assert.equal(payload.scheduledAt, '2026-09-20T10:00:00.000Z');
  assert.equal(payload.automationEnabled, false);
  assert.equal(payload.autoPublishEnabled, false);
});

test('manual batch submission extracts fetched text instead of persisting object coercion', () => {
  const payload = buildManualBatchSubmission({
    title: '对象正文兼容',
    groups: [{
      platformId: '4',
      platformName: '点众付费',
      inputText: '11010564460\t测试书',
      sourceTextByBookId: {
        '11010564460': { data: { processedText: '第一行\n第二行' } },
        '11010567086': '[object Object]'
      }
    }]
  });

  assert.deepEqual(payload.groups[0].sourceTextByBookId, {
    '11010564460': '第一行\n第二行'
  });
});

test('manual intake treats the leading 6-25 digit number of each row as a book ID', () => {
  assert.deepEqual(
    manualBookIDsFromInput('700000 六位数 ID\n748725 事不过三，过三遭殃\n2086529323515883958 长 ID'),
    ['700000', '748725', '2086529323515883958']
  );
});

test('manual intake ignores leading numbers shorter than 6 digits, aligned with the backend rule', () => {
  assert.deepEqual(manualBookIDsFromInput('73709 五位短号\n567168 六位正常'), ['567168']);
  assert.deepEqual(manualBookIDsFromInput('7 个位数行'), []);
  // mergeGroupLines 里的同款正则也必须忽略不足 6 位的行
  assert.equal(mergeGroupLines('73709 旧短号\n567168 正常书', '73708 新短号'), '567168 正常书');
});

test('mergeGroupLines deduplicates an ID repeated within the old text itself', () => {
  const merged = mergeGroupLines('737092 甲\n737092 甲重复', '');
  assert.equal(merged, '737092 甲重复');
  assert.equal((merged.match(/737092/g) || []).length, 1);
});

test('manual intake does not mistake numbers inside a book title for a book ID', () => {
  assert.deepEqual(manualBookIDsFromInput('事不过三，2026 再相见'), []);
});

test('manual batch submission preserves explicit scheduled full-automation choices', () => {
  const payload = buildManualBatchSubmission({
    title: ' 淮雪 ',
    platformId: '15',
    platformName: '知乎付费',
    parseMode: 'smart',
    columnPresetId: 'full_metadata',
    columnOrder: '书籍ID,书名',
    inputText: '2083601795096502516\t淮雪',
    sourceTextByBookId: { '2083601795096502516': '正文' },
    contentRangeLines: 5,
    contentCaptureCharacters: 4000,
    scheduledAt: '2026-09-22T10:00:00.000Z',
    automationEnabled: true,
    autoPublishEnabled: true
  });

  assert.equal(payload.automationEnabled, true);
  assert.equal(payload.autoPublishEnabled, true);

  const uploadOnly = buildManualBatchSubmission({
    title: '仅上传',
    platformId: '15',
    platformName: '知乎付费',
    parseMode: 'smart',
    columnPresetId: 'full_metadata',
    columnOrder: '书籍ID,书名',
    inputText: '2083601795096502516\t淮雪',
    sourceTextByBookId: { '2083601795096502516': '正文' },
    contentRangeLines: 5,
    contentCaptureCharacters: 4000,
    automationEnabled: false,
    autoPublishEnabled: true
  });

  assert.equal(uploadOnly.automationEnabled, false);
  assert.equal(uploadOnly.autoPublishEnabled, false);
});

test('manual batch submission preserves immediate automation and its frozen book concurrency', () => {
  const payload = buildManualBatchSubmission({
    title: '立即生产',
    platformId: '15',
    inputText: '2083601795096502516\\t淮雪',
    sourceTextByBookId: { '2083601795096502516': '正文' },
    automationEnabled: true,
    scheduledAt: '',
    runMode: 'full_submit',
    automationConcurrency: 4
  });

  assert.equal(payload.automationEnabled, true);
  assert.equal(payload.scheduledAt, '');
  assert.equal(payload.autoPublishEnabled, true);
  assert.equal(payload.automationConcurrency, 4);
});

test('platform group helpers', () => {
  // mergeGroupLines
  assert.equal(mergeGroupLines('737092 甲\n687404 乙', '749269 丙'), '737092 甲\n687404 乙\n749269 丙');
  const merged = mergeGroupLines('737092 旧标题', '737092 新标题\n749269 丙');
  assert.match(merged, /^737092 新标题/m);
  assert.match(merged, /749269 丙/);
  assert.equal((merged.match(/737092/g) || []).length, 1);

  // upsertPlatformGroup
  const g1 = upsertPlatformGroup([], { platformId: '3', platformName: '七猫付费', inputText: '737092 甲' });
  assert.equal(g1.length, 1);
  const g2 = upsertPlatformGroup(g1, { platformId: '3', platformName: '七猫付费', inputText: '687404 乙' });
  assert.equal(g2.length, 1);
  assert.equal(totalGroupBookCount(g2), 2);
  const g3 = upsertPlatformGroup(g2, { platformId: '15', platformName: '知乎付费', inputText: '567168 丙' });
  assert.equal(g3.length, 2);
  assert.equal(totalGroupBookCount(g3), 3);

  // replacePlatformGroup：整体替换；清空则移除
  const r1 = replacePlatformGroup(g3, '3', '737092 甲\n687404 乙\n711720 丁');
  assert.equal(totalGroupBookCount(r1), 4);
  const r2 = replacePlatformGroup(g3, '3', '   ');
  assert.equal(r2.length, 1);
  assert.equal(r2[0].platformId, '15');

  // removePlatformGroup
  assert.equal(removePlatformGroup(g3, '3').length, 1);
  assert.equal(removePlatformGroup(g3, '99').length, 2);
});
