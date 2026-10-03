import test from 'node:test';
import assert from 'node:assert/strict';

import {
  BATCH_FACTORY_PLATFORMS,
  addBookstoreGroup,
  buildIntakePayload,
  groupLabel,
  parseBookIds,
  submitIntakeAndStartJob,
} from './batchFactoryIntake.js';

test('keeps the verified v88 bookstore platform mapping', () => {
  const byName = Object.fromEntries(BATCH_FACTORY_PLATFORMS.map(item => [item.name, item.id]));
  assert.equal(byName['点众付费'], '4');
  assert.equal(byName['黑岩付费'], '1');
  assert.equal(byName['知乎付费'], '15');
});

test('parses pasted book ids, removes blanks and deduplicates in input order', () => {
  assert.deepEqual(
    parseBookIds('7673480334440139800\nabc-1, abc-1  xyz-2；xyz-2'),
    ['7673480334440139800', 'abc-1', 'xyz-2'],
  );
});

test('adds a bookstore group with shared settings and produces the requested label', () => {
  const groups = addBookstoreGroup([], {
    platformId: '15',
    platformName: '知乎付费',
    rawBookIds: 'a\nb\nc',
    manualGender: 'female',
    style: '现言',
    maxTxt: 2000,
  });

  assert.equal(groups.length, 1);
  assert.equal(groupLabel(groups[0]), '知乎付费3本');
  assert.deepEqual(groups[0].books, [
    { book_id: 'a', manual_gender: 'female', style: '现言' },
    { book_id: 'b', manual_gender: 'female', style: '现言' },
    { book_id: 'c', manual_gender: 'female', style: '现言' },
  ]);
  assert.equal(groups[0].max_txt, 2000);
});

test('rejects duplicate book ids inside the same platform across groups', () => {
  const first = addBookstoreGroup([], {
    platformId: '4', platformName: '点众付费', rawBookIds: 'book-1', maxTxt: 4000,
  });
  assert.throws(() => addBookstoreGroup(first, {
    platformId: '4', platformName: '点众付费', rawBookIds: 'book-1\nbook-2', maxTxt: 4000,
  }), /book-1/);
});

test('builds the Go intake payload without a client-controlled owner', () => {
  const groups = addBookstoreGroup([], {
    platformId: '1', platformName: '黑岩付费', rawBookIds: 'b1\nb2', maxTxt: 5000,
  });
  const payload = buildIntakePayload({ title: '今日批量', groups });
  assert.deepEqual(payload, { title: '今日批量', groups });
  assert.equal(Object.hasOwn(payload, 'owner'), false);
});

test('immediate execution creates the intake first and then starts a job with intake_id', async () => {
  const calls = [];
  const fetchImpl = async (url, options) => {
    calls.push({ url, options, body: JSON.parse(options.body) });
    if (url === '/api/batch-factory/intakes') {
      return { ok: true, json: async () => ({ intake_id: 'intake-9', group_count: 1, book_count: 2 }) };
    }
    if (url === '/api/batch-factory/jobs') {
      return { ok: true, json: async () => ({ job_id: 'job-9', status: 'queued' }) };
    }
    throw new Error(`unexpected url ${url}`);
  };

  const groups = addBookstoreGroup([], {
    platformId: '15', platformName: '知乎付费', rawBookIds: 'b1\nb2', maxTxt: 4000,
  });
  const result = await submitIntakeAndStartJob({ fetchImpl, title: '立即执行', groups });

  assert.equal(calls.length, 2);
  assert.equal(calls[0].url, '/api/batch-factory/intakes');
  assert.equal(calls[1].url, '/api/batch-factory/jobs');
  assert.deepEqual(calls[1].body, { intake_id: 'intake-9' });
  assert.deepEqual(result, {
    intake: { intake_id: 'intake-9', group_count: 1, book_count: 2 },
    job: { job_id: 'job-9', status: 'queued' },
  });
});
