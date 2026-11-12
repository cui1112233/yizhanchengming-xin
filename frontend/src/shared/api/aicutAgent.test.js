import test from 'node:test';
import assert from 'node:assert/strict';

let api = {};
try {
  api = await import('./aicutAgent.js');
} catch {}

test('AI-CUT Agent创建运行携带模式模型项目版本和幂等编号', async () => {
  assert.equal(typeof api.createAICutAgentRun, 'function');
  let received;
  const run = await api.createAICutAgentRun('cut-1', {
    instruction: '删除第二段', mode: 'agent', modelId: 'text-main', activeSkillId: 'rough-cut', expectedRevision: 3, clientRequestId: 'client-1'
  }, async (path, options) => {
    received = { path, options };
    return { run: { id: 'run-1', state: 'queued' } };
  });
  assert.equal(received.path, '/api/ai-cut/projects/cut-1/agent-runs');
  assert.deepEqual(JSON.parse(received.options.body), {
    instruction: '删除第二段', mode: 'agent', modelId: 'text-main', activeSkillId: 'rough-cut', expectedRevision: 3, clientRequestId: 'client-1'
  });
  assert.equal(run.id, 'run-1');
});

test('AI-CUT Agent增量事件、消息和取消使用持久接口', async () => {
  assert.equal(typeof api.listAICutAgentEvents, 'function');
  const calls = [];
  const request = async (path, options = {}) => {
    calls.push({ path, options });
    if (path.endsWith('/cancel')) return { run: { id: 'run-1', state: 'cancel_requested' } };
    if (path.includes('/events')) return { events: [{ sequence: 2, type: 'tool_done', revision: 4 }] };
    return { messages: [{ id: 'msg-1', role: 'user', content: '删除第二段' }] };
  };
  assert.equal((await api.listAICutAgentEvents('cut-1', 'run-1', 1, request))[0].sequence, 2);
  assert.equal((await api.listAICutAgentMessages('cut-1', 100, request))[0].id, 'msg-1');
  assert.equal((await api.cancelAICutAgentRun('cut-1', 'run-1', request)).state, 'cancel_requested');
  assert.equal(calls[0].path, '/api/ai-cut/projects/cut-1/agent-runs/run-1/events?after=1');
  assert.equal(calls[2].options.method, 'POST');
});

test('AI-CUT Agent传送当前选择和播放位置作为独立数据', async () => {
 let body;
 await api.createAICutAgentRun('cut-1', {instruction:'剪选中的这段',uiContext:{selectedClipKeys:['video-1::clip-b'],playheadMs:1234}},async (_,options)=>{body=JSON.parse(options.body);return {run:{id:'run-ui'}}});
 assert.deepEqual(body.uiContext,{selectedClips:[{trackId:'video-1',clipId:'clip-b'}],playheadMs:1234});
 assert.equal(body.instruction,'剪选中的这段');
});

test('AI-CUT Agent拒绝格式错误和越界的页面数据', async () => {
 for (const uiContext of [
  {selectedClipKeys:['video-1::clip-b::extra'],playheadMs:10},
  {selectedClipKeys:[],playheadMs:-1},
  {selectedClipKeys:[],playheadMs:Infinity},
  {selectedClipKeys:Array(101).fill('video-1::clip-b'),playheadMs:10}
 ]) {
  let called=false;
  await assert.rejects(api.createAICutAgentRun('cut-1',{instruction:'剪这段',uiContext},async()=>{called=true;return {run:{}}}));
  assert.equal(called,false);
 }
});
