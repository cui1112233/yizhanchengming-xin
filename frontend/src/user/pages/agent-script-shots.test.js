import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';
import {
  buildShotBatchPayload,
  invalidateScriptSplitForSource,
  normalizeScriptShots,
  normalizeScriptShotsForConfig,
  runAfterDurableCanvasSave,
  scriptShotsByNode,
  scriptSourceMarker,
  scriptSplitReady,
  selectedShotIDs
} from './agent-script-shots.js';
import { createCanvasSaveQueue } from './agent-canvas-state.js';

const composerSource = readFileSync(new URL('../components/AgentScriptNodeComposer.jsx', import.meta.url), 'utf8');
const nodeComposerSource = readFileSync(new URL('../components/AgentNodeComposer.jsx', import.meta.url), 'utf8');
const canvasSource = readFileSync(new URL('../components/AgentCanvas.jsx', import.meta.url), 'utf8');
const apiSource = readFileSync(new URL('../../shared/api/agent.js', import.meta.url), 'utf8');

test('normalization keeps structured shots in sequence order', () => {
  const result = normalizeScriptShots({
    generation: { state: 'completed' },
    split: {
      shots: [
        { id: 'shot-2', sequence: 2, scene: '室内', dialogue: '第二句', imagePrompt: '近景' },
        { id: 'shot-1', sequence: 1, scene: '门外', narration: '开场', ttsText: '旁白' }
      ]
    }
  });

  assert.equal(result.state, 'completed');
  assert.deepEqual(result.shots.map(shot => shot.id), ['shot-1', 'shot-2']);
  assert.equal(result.shots[0].ttsText, '旁白');
  assert.equal(result.shots[1].imagePrompt, '近景');
});

test('selected shot IDs follow shot order and ignore stale selection entries', () => {
  const shots = [{ id: 'shot-1', sequence: 1 }, { id: 'shot-2', sequence: 2 }];
  assert.deepEqual(selectedShotIDs(shots, { 'shot-2': true, missing: true, 'shot-1': false }), ['shot-2']);
  assert.deepEqual(selectedShotIDs(shots, new Set(['shot-2', 'shot-1'])), ['shot-1', 'shot-2']);
});

test('shared storyboard shots are grouped by their original script node without copying them', () => {
  const shots = [
    { id: 'shot-2', scriptNodeId: 'script-b', sequence: 1, scene: 'B' },
    { id: 'shot-1', scriptNodeId: 'script-a', sequence: 2, scene: 'A2' },
    { id: 'shot-0', scriptNodeId: 'script-a', sequence: 1, scene: 'A1' }
  ];

  const grouped = scriptShotsByNode(shots);

  assert.deepEqual(Object.keys(grouped), ['script-b', 'script-a']);
  assert.deepEqual(grouped['script-a'].map(shot => shot.id), ['shot-0', 'shot-1']);
  assert.strictEqual(grouped['script-a'][0], shots[2]);
});

test('batch payload deduplicates selected shots and keeps source markers', () => {
  assert.deepEqual(buildShotBatchPayload({
    scriptNodeId: ' script-1 ',
    selectedIDs: ['shot-2', 'shot-2', '', 'shot-1'],
    kind: 'video',
    modelId: ' video-1 '
  }), {
    scriptNodeId: 'script-1',
    shotIds: ['shot-2', 'shot-1'],
    kind: 'video',
    modelId: 'video-1'
  });
});

test('batch payload keeps the selected provider for local executor models', () => {
  assert.deepEqual(buildShotBatchPayload({
    scriptNodeId: 'script-1',
    selectedIDs: ['shot-1'],
    kind: 'image',
    provider: 'doubao_local_executor',
    modelId: 'Seedream 4.5'
  }), {
    scriptNodeId: 'script-1',
    shotIds: ['shot-1'],
    kind: 'image',
    provider: 'doubao_local_executor',
    modelId: 'Seedream 4.5'
  });
});

test('both batch image entry points expose channel-aware model selection', () => {
  const workspaceSource = readFileSync(new URL('../components/AgentStoryboardWorkspace.jsx', import.meta.url), 'utf8');
  const pickerSource = readFileSync(new URL('../components/AgentBatchModelPicker.jsx', import.meta.url), 'utf8');
  assert.match(composerSource, /<AgentBatchModelPicker[\s\S]*?kind=\{activeBatchKind\.kind\}/);
  assert.match(workspaceSource, /<AgentBatchModelPicker[\s\S]*?kind=\{shotBatchKind\}/);
  assert.match(pickerSource, /providerOptionsForModels/);
  assert.match(pickerSource, /modelOptionsForProvider/);
});

test('failed split status preserves the provider error instead of looking complete', () => {
  const result = normalizeScriptShots({
    generation: { state: 'failed', errorMessage: '上游失败' },
    split: { shots: [{ id: 'shot-1', sequence: 1 }] }
  });

  assert.equal(result.state, 'failed');
  assert.equal(result.errorMessage, '上游失败');
});

test('durable save failure prevents the split or batch request from starting', async () => {
  let submitted = 0;
  const queue = createCanvasSaveQueue({
    revision: 3,
    delay: 5000,
    save: async () => { throw new Error('offline'); },
    reload: async () => ({ canvas: { version: 1, revision: 3, nodes: [], edges: [] } })
  });
  queue.enqueue({ version: 1, revision: 3, nodes: [{ id: 'script-1' }], edges: [] });

  await assert.rejects(
    runAfterDurableCanvasSave(() => queue.flushRequired(), async () => { submitted += 1; }),
    /offline/
  );
  assert.equal(submitted, 0);
  queue.dispose();
});

test('unresolved save conflict prevents the split or batch request from starting', async () => {
  let saveCalls = 0;
  let submitted = 0;
  const queue = createCanvasSaveQueue({
    revision: 1,
    delay: 5000,
    save: async () => {
      saveCalls += 1;
      throw Object.assign(new Error('revision conflict'), { status: 409 });
    },
    reload: async () => ({ canvas: { version: 1, revision: 7, nodes: [{ id: 'script-1' }], edges: [] } })
  });
  queue.enqueue({ version: 1, revision: 1, nodes: [{ id: 'script-1' }], edges: [] });

  await assert.rejects(
    runAfterDurableCanvasSave(() => queue.flushRequired(), async () => { submitted += 1; }),
    /revision conflict/
  );
  assert.equal(saveCalls, 2);
  assert.equal(submitted, 0);
  queue.dispose();
});

test('editing source text invalidates the completed split and hides old shots immediately', () => {
  const original = {
    sourceText: '旧正文',
    splitModelId: 'text-1',
    unrelatedSetting: 'keep-me',
    splitState: {
      state: 'completed',
      generationId: 'generation-1',
      sourceMarker: scriptSourceMarker('旧正文')
    }
  };
  assert.equal(scriptSplitReady(original), true);

  const changed = invalidateScriptSplitForSource(original, '新正文');
  const visible = normalizeScriptShotsForConfig(changed, {
    state: 'completed',
    shots: [{ id: 'shot-1', sequence: 1 }]
  });

  assert.equal(original.splitState.state, 'completed');
  assert.equal(changed.sourceText, '新正文');
  assert.equal(changed.splitModelId, 'text-1');
  assert.equal(changed.unrelatedSetting, 'keep-me');
  assert.deepEqual(changed.splitState, { state: 'stale', errorMessage: '', generationId: '', sourceMarker: '' });
  assert.equal(scriptSplitReady(changed), false);
  assert.equal(visible.state, 'stale');
  assert.deepEqual(visible.shots, []);
  assert.deepEqual(selectedShotIDs(visible.shots, new Set(['shot-1'])), []);
});

test('editing source text keeps stale ahead of an old running split response', () => {
  const changed = invalidateScriptSplitForSource({
    sourceText: '旧正文',
    splitState: {
      state: 'completed',
      sourceMarker: scriptSourceMarker('旧正文')
    }
  }, '新正文');

  const visible = normalizeScriptShotsForConfig(changed, {
    generation: { state: 'running' },
    split: { shots: [{ id: 'shot-1', sequence: 1 }] }
  });

  assert.equal(visible.state, 'stale');
  assert.deepEqual(visible.shots, []);
});

test('editing source text suppresses an old failed split response and error', () => {
  const changed = invalidateScriptSplitForSource({
    sourceText: '旧正文',
    splitState: {
      state: 'completed',
      sourceMarker: scriptSourceMarker('旧正文')
    }
  }, '新正文');

  const visible = normalizeScriptShotsForConfig(changed, {
    generation: { state: 'failed', errorMessage: '旧正文拆分失败' },
    split: { shots: [{ id: 'shot-1', sequence: 1 }] }
  });

  assert.equal(visible.state, 'stale');
  assert.equal(visible.errorMessage, '');
  assert.deepEqual(visible.shots, []);
});

test('a stale script configuration also hides shared storyboard shots returned by a later refresh', () => {
  const changed = invalidateScriptSplitForSource({
    sourceText: '旧正文',
    splitState: {
      state: 'completed',
      sourceMarker: scriptSourceMarker('旧正文')
    }
  }, '新正文');

  const visible = normalizeScriptShotsForConfig(changed, {
    shots: [{ id: 'shot-1', scriptNodeId: 'script-1', sequence: 1 }]
  });

  assert.equal(visible.state, 'stale');
  assert.deepEqual(visible.shots, []);
});

test('script node UI uses Task 2 split, list, and batch APIs', () => {
  assert.match(apiSource, /createAgentScriptSplit/);
  assert.match(apiSource, /listAgentScriptShots/);
  assert.match(apiSource, /createAgentShotBatchNodes/);
  assert.match(composerSource, /sourceText/);
  assert.match(composerSource, /splitModelId/);
  assert.match(composerSource, /selectedShotIDs/);
  assert.match(nodeComposerSource, /AgentScriptNodeComposer/);
  assert.match(canvasSource, /nodeType === 'script'/);
  assert.match(canvasSource, /handleContextCreate\('script'\)/);
});

test('script node guides users through script, shots, and one selected generation type', () => {
  assert.match(composerSource, /剧本/);
  assert.match(composerSource, /镜头/);
  assert.match(composerSource, /生成内容/);
  assert.match(composerSource, /Segmented/);
  assert.match(composerSource, /selectedIDs\.length/);
  assert.match(composerSource, /kind=\{activeBatchKind\.kind\}/);
});

test('script node keeps shot selection and batch execution inside the workflow', () => {
  assert.doesNotMatch(composerSource, /onOpenStoryboard/);
  assert.doesNotMatch(composerSource, /打开故事板/);
  assert.doesNotMatch(composerSource, /在故事板打开镜头/);
  assert.match(composerSource, /镜头 \$\{String\(shot\.sequence\)/);
  assert.match(composerSource, /buildShotBatchPayload/);
  assert.match(composerSource, /onBatchShots\?\.\(node\.id, payload\)/);
  assert.doesNotMatch(nodeComposerSource, /onOpenStoryboard=\{onOpenStoryboard\}/);
  assert.doesNotMatch(canvasSource, /onOpenStoryboard/);
});
