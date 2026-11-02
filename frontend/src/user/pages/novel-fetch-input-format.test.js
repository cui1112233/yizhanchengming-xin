import test from 'node:test';
import assert from 'node:assert/strict';

import { normalizeNovelFetchInput, serializeNovelFetchRows } from './novel-fetch-input-format.js';

test('固定列预设保留 ID、书名、男女频和理由', () => {
  assert.deepEqual(
    normalizeNovelFetchInput('2094511573566019532\t鼓励式老公\t女频\t闺蜜背刺', {
      format: 'paid_id_title_gender_reason'
    }).rows[0],
    {
      bookId: '2094511573566019532',
      bookName: '鼓励式老公',
      gender: '女频',
      reason: '闺蜜背刺',
      error: ''
    }
  );
});

test('输入归一化保留错误行、重复 ID 和空 ID', () => {
  const result = normalizeNovelFetchInput([
    '2094511573566019532\t鼓励式老公\t女频\t闺蜜背刺',
    '2094511573566019532\t重复',
    '\t缺失 ID'
  ].join('\n'), { format: 'paid_id_title_gender_reason' });

  assert.equal(result.validRows, 1);
  assert.equal(result.duplicateIds, 1);
  assert.equal(result.emptyIdRows, 1);
  assert.equal(result.rows[1].error, '重复书籍 ID');
  assert.equal(result.rows[2].error, '缺少书籍 ID');
});

test('智能识别兼容原有的空格分隔 ID 和书名', () => {
  const result = normalizeNovelFetchInput('2094511573566019532    鼓励式老公，把行动都给了闺蜜');

  assert.deepEqual(result.rows[0], {
    bookId: '2094511573566019532',
    bookName: '鼓励式老公，把行动都给了闺蜜',
    gender: '',
    reason: '',
    error: ''
  });
});

test('智能识别跳过字段表头并按别名映射列', () => {
  const result = normalizeNovelFetchInput([
    '书籍ID\t书名\t性别\t推荐理由',
    '2094511573566019532\t鼓励式老公\t女频\t闺蜜背刺'
  ].join('\n'));

  assert.equal(result.rows.length, 1);
  assert.deepEqual(result.rows[0], {
    bookId: '2094511573566019532',
    bookName: '鼓励式老公',
    gender: '女频',
    reason: '闺蜜背刺',
    error: ''
  });
});

test('智能识别可处理两行分组表头', () => {
  const result = normalizeNovelFetchInput([
    '书籍信息\t书籍信息\t分类信息\t补充信息',
    '书籍ID\t作品名称\t男女频\t入选理由',
    '2094511573566019532\t鼓励式老公\t女频\t闺蜜背刺'
  ].join('\n'));

  assert.equal(result.rows.length, 1);
  assert.equal(result.rows[0].bookId, '2094511573566019532');
  assert.equal(result.rows[0].bookName, '鼓励式老公');
  assert.equal(result.rows[0].gender, '女频');
  assert.equal(result.rows[0].reason, '闺蜜背刺');
});

test('固定输入格式允许调用方调整列位置', () => {
  const result = normalizeNovelFetchInput('女频\t闺蜜背刺\t鼓励式老公\t2094511573566019532', {
    format: 'paid_id_title_gender_reason',
    columnMapping: { gender: 0, reason: 1, bookName: 2, bookId: 3 }
  });

  assert.equal(result.validRows, 1);
  assert.deepEqual(result.rows[0], {
    bookId: '2094511573566019532',
    bookName: '鼓励式老公',
    gender: '女频',
    reason: '闺蜜背刺',
    error: ''
  });
});

test('把解析行整理成后端统一可读的四列格式', () => {
  assert.equal(serializeNovelFetchRows([{
    bookId: '2094511573566019532', bookName: '鼓励式老公', gender: '女频', reason: '闺蜜背刺'
  }]), '2094511573566019532\t鼓励式老公\t女频\t闺蜜背刺');
});
