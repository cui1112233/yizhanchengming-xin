import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import vm from 'node:vm';

const page = fs.readFileSync(new URL('./NovelFetchPage.jsx', import.meta.url), 'utf8');
function handler(name, context) {
  let start = page.indexOf(`  async function ${name}(`);
  if (start < 0) start = page.indexOf(`  function ${name}(`);
  assert.ok(start >= 0, name);
  const end = page.indexOf('\n  }', start) + 4;
  return vm.runInNewContext(`(${page.slice(start, end)})`, context);
}

test('batch deletion never includes other batches and empty batch never means all', async () => {
  const deleted = [];
  const context = {
    tasks: [{ bookId: 'a' }, { book_id: 'b' }],
    message: { info() {}, success() {} },
    Modal: { confirm: async options => { assert.match(options.content, /1/); await options.onOk(); } },
    deleteWorkshopTasks: async ids => deleted.push(...ids),
    setSelectedTaskKeys() {}, loadTasks: async () => {}, loadCurrentBatch: async () => {}
  };
  const remove = handler('deleteTasksByFilter', context);
  await remove(() => true, 'empty', ['a']);
  await remove(() => true, 'empty', []);
  assert.deepEqual(deleted, ['a']);
});

test('current batch submit never silently submits visible tasks when no current batch', async () => {
  const requests = [];
  await handler('submitVisibleTasks', {
    currentBatchIds: [], visibleTasks: [{ bookId: 'other' }],
    message: { info() {} }, handleWebSubmit: async fn => fn(),
    startWebSubmit: async body => requests.push(body), setWebSubmitResults() {}
  })();
  assert.equal(requests.length, 0);
});

test('submit selected directly submits only the checked task ids', async () => {
  const requests = [];
  const results = [];
  let warned = '';
  let confirmOptions = null;
  await handler('submitSelectedTasks', {
    selectedTaskKeys: ['a', 'b'],
    message: { warning: text => { warned = text; } },
    Modal: { confirm: options => { confirmOptions = options; } },
    handleWebSubmit: async fn => fn(),
    startWebSubmit: async body => {
      requests.push(body);
      return { results: [{ bookId: 'a', version: 'ai1', status: 'confirmed' }] };
    },
    setWebSubmitResults: value => results.push(value),
    setWebSubmitHistory() {}, setSubmitView() {}, changeTab() {},
    loadTasks: async () => {}, loadCurrentBatch: async () => {}
  })();
  assert.ok(confirmOptions);
  await confirmOptions.onOk();
  assert.equal(warned, '');
  assert.equal(JSON.stringify(requests), JSON.stringify([{ taskIds: ['a', 'b'] }]));
  assert.equal(results.length, 1);
});

test('submit selected refuses to submit when nothing is checked', async () => {
  const requests = [];
  let warned = '';
  await handler('submitSelectedTasks', {
    selectedTaskKeys: [],
    message: { warning: text => { warned = text; } },
    Modal: { confirm: async () => { throw new Error('should not confirm'); } },
    handleWebSubmit: async fn => fn(),
    startWebSubmit: async body => requests.push(body),
    setWebSubmitResults() {}, setWebSubmitHistory() {}, setSubmitView() {}, changeTab() {},
    loadTasks: async () => {}, loadCurrentBatch: async () => {}
  })();
  assert.match(warned, /请先选择/);
  assert.equal(requests.length, 0);
});

test('submit all submits every visible task id, not the current batch', async () => {
  const requests = [];
  let confirmOptions = null;
  await handler('submitAllTasks', {
    visibleTaskIds: ['a', 'b', 'c'],
    message: { info() {} },
    Modal: { confirm: options => { confirmOptions = options; } },
    handleWebSubmit: async fn => fn(),
    startWebSubmit: async body => { requests.push(body); return { results: [] }; },
    setWebSubmitResults() {}, setWebSubmitHistory() {}, setSubmitView() {}, changeTab() {},
    loadTasks: async () => {}, loadCurrentBatch: async () => {}
  })();
  assert.ok(confirmOptions);
  await confirmOptions.onOk();
  assert.equal(JSON.stringify(requests), JSON.stringify([{ taskIds: ['a', 'b', 'c'] }]));
});

test('submit page confirmation no longer requires or sends a preview revision', async () => {
  const requests = [];
  let warned = '';
  let confirmOptions = null;
  await handler('confirmWebSubmit', {
    selectedTaskKeys: ['a'],
    message: { warning: text => { warned = text; } },
    Modal: { confirm: options => { confirmOptions = options; } },
    handleWebSubmit: async fn => fn(),
    startWebSubmit: async body => { requests.push(body); return { results: [] }; },
    setWebSubmitResults() {}, setWebSubmitHistory() {}, setSubmitView() {}
  })();
  assert.equal(warned, '');
  assert.ok(confirmOptions);
  await confirmOptions.onOk();
  assert.equal(JSON.stringify(requests), JSON.stringify([{ taskIds: ['a'] }]));
});

test('retry reports partial failure rather than unconditional success', async () => {
  const feedback = [];
  await handler('retryFailedTasks', {
    currentBatchIds: ['a'], currentBatchRows: [{ bookId: 'a' }], visibleTasks: [],
    novelFetchTaskStatus: () => ({ status: 'failed' }),
    retryWorkshopTasks: async () => ({ requested: 1, failed: 1, retried: 0 }),
    novelFetchBatchResultSummary: () => '失败 1 项',
    setTaskActionBusy() {}, loadTasks: async () => {}, loadCurrentBatch: async () => {},
    message: { info() {}, success: text => feedback.push(['success', text]), warning: text => feedback.push(['warning', text]), error: text => feedback.push(['error', text]) }
  })();
  assert.notEqual(feedback[0]?.[0], 'success');
  assert.match(feedback[0]?.[1], /失败 1 项/);
});

test('configuration save includes version profile edits in the same request', async () => {
  let sent;
  await handler('saveConfig', {
    configLoading: false, setConfigLoading() {},
    configForm: { getFieldsValue: () => ({ configRevision: 7, rewrite: {}, platforms: [], styles: [], web_submit: { retry_times: 2 } }), setFieldsValue() {} },
    selectedUploadProfile: 'new', profileBindings: { ai1: 'p1' },
    withNovelFetchVersionProfiles: value => value,
    saveWorkshopConfig: async body => { sent = body; return { settings: body }; },
    setConfigEffective() {}, setConfigDirty() {},
    message: { success() {}, error: error => { throw new Error(error); } }
  })();
  assert.equal(sent.web_submit.selected_profile, 'new');
  assert.equal(sent.web_submit.profile_bindings.ai1, 'p1');
  assert.equal(sent.web_submit.retry_times, 2);
  assert.equal(sent.configRevision, 7);
});

test('switching rule types restores unfinished drafts without requesting stale data', async () => {
  const edits = [];
  await handler('loadRuleConfig', {
    ruleReadSequence: { current: 0 },
    ruleDrafts: { current: new Map([['symbol_rules', { text: '{"draft":true}', revision: 8 }]]) },
    setRuleConfigText: text => edits.push(text), setRuleConfigRevision: revision => edits.push(revision),
    getWorkshopConfig: () => { throw new Error('cached draft must not be reloaded'); }
  })('symbol_rules');
  assert.deepEqual(edits, ['{"draft":true}', 8]);
});

test('late rule response cannot overwrite the rule type selected later', async () => {
  const pending = [];
  const texts = [];
  const load = handler('loadRuleConfig', {
    ruleReadSequence: { current: 0 }, ruleDrafts: { current: new Map() },
    getWorkshopConfig: () => new Promise(resolve => pending.push(resolve)),
    setRuleConfigText: text => texts.push(text), setRuleConfigRevision() {},
    message: { error: error => { throw new Error(error); } }
  });
  const first = load('symbol_rules', false);
  const second = load('chapter_exact_rules', false);
  pending[1]({ appConfig: { knowledge: { chapter_rules: { chapter_exact_rules: ['latest'] } } }, settings: { configRevision: 2 } });
  await second;
  pending[0]({ appConfig: { knowledge: { symbol_rules: { rules: ['old'] } } }, settings: { configRevision: 1 } });
  await first;
  assert.equal(texts.length, 1);
  assert.match(texts[0], /latest/);
});

test('duplicate submit/save/AI controls are removed and rules have distinct targets', () => {
  assert.doesNotMatch(page, /label: '重新提交当前批次'/);
  assert.doesNotMatch(page, />保存版本配置</);
  assert.doesNotMatch(page, />AI 规则建议</);
  for (const kind of ['sensitive', 'symbol_rules', 'pair_fill_rules', 'chapter_exact_rules', 'chapter_inline_rules']) {
    assert.ok(page.includes(`openRuleCollection('${kind}')`), kind);
  }
});
