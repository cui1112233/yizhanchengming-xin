import { expect, test } from "vitest";
import { waterImageModels, createProductionClients, latestRequestGuard } from "./productionClients";

test('water projects become available even while the model request remains pending', async()=>{
  let complete!:(value:any)=>void;
  const models=new Promise(resolve=>{complete=resolve;});
  const client=createProductionClients(async path=>path.endsWith('/models')?models:{projects:[{id:7}]});
  let projects:any[]=[];
  const loading=client.waterCatalog(next=>{projects=next;},()=>{});
  await new Promise(resolve=>setTimeout(resolve,0));
  expect(projects).toEqual([{id:7}]);
  complete({models:[]}); await loading;
});

test("Agent uses task history, canvas and executions and sends chat prompt without Studio model payload", async () => {
  const calls: unknown[] = [];
  const client = createProductionClients(async (path, options) => { calls.push([path, options]); return path.endsWith('/canvas') ? { canvas: { nodes: [] } } : path.endsWith('/executions') ? { executions: [] } : { task: { id: 'a/b', messages: [] } }; });
  const result = await client.agentDetail('a/b');
  expect(result.task.id).toBe('a/b');
  await client.ask('a/b', '  hello  ');
  expect(calls).toEqual([
    ['/api/agent/tasks/a%2Fb', undefined], ['/api/agent/tasks/a%2Fb/canvas', undefined], ['/api/agent/tasks/a%2Fb/executions', undefined],
    ['/api/agent/chat', { method: 'POST', data: { taskId: 'a/b', prompt: 'hello', context: { page: 'Agent 工作区' }, skillIds: [] }, timeout: 180000 }],
  ]);
});

test("Shuihuo reads own detail and tasks, validates generation selections", async () => {
  const calls: unknown[] = [];
  const client = createProductionClients(async (path, options) => { calls.push([path, options]); return path.endsWith('/tasks') ? { tasks: [] } : { project: { id: 7 }, segments: [], assets: [], media: [] }; });
  expect((await client.waterDetail('7')).project.id).toBe(7);
  await expect(client.waterTask('7', '', '9', 'video')).rejects.toThrow();
  await client.waterTask('7', '3', '9', 'video');
  expect(calls[2]).toEqual(['/api/shuihuo-production/projects/7/tasks', { method: 'POST', data: { segmentId: '3', modelId: '9', kind: 'video' }, timeout: 180000 }]);
});

test("malformed success cannot become empty data and stale requests cannot publish", async () => {
  const client = createProductionClients(async () => ({}));
  await expect(client.agentDetail('x')).rejects.toThrow();
  await expect(client.waterDetail('x')).rejects.toThrow();
  const guard = latestRequestGuard(); const old = guard.begin(); const current = guard.begin();
  expect(guard.current(old)).toBe(false); expect(guard.current(current)).toBe(true);
  guard.invalidate(); expect(guard.current(current)).toBe(false);
});

test('new Shuihuo project saves pasted source through source endpoint before returning', async () => {
  const calls: unknown[] = [];
  const client = createProductionClients(async (path, options) => { calls.push([path, options]); return { id: 11 }; });
  await client.createWater(' Story ', ' Original text ');
  expect(calls).toEqual([
    ['/api/shuihuo-production/projects', { method: 'POST', data: { name: 'Story' } }],
    ['/api/shuihuo-production/projects/11/source', { method: 'PUT', data: { sourceText: 'Original text' } }],
  ]);
});

test('partial project creation exposes the existing project instead of making a second one', async () => {
  let created:any;const calls:string[]=[];
  const client=createProductionClients(async(path)=>{calls.push(path);if(path.endsWith('/projects'))return{id:'p1',name:'故事'};throw Error('source timeout');});
  await expect(client.createWater('故事','原文',project=>{created=project;})).rejects.toMatchObject({createdProject:{id:'p1'}});
  expect(created.id).toBe('p1');expect(calls).toEqual(['/api/shuihuo-production/projects','/api/shuihuo-production/projects/p1/source']);
});

test('smart segmentation and task actions use their own real endpoints',async()=>{
  const calls:any[]=[];const client=createProductionClients(async(path,options)=>{calls.push([path,options]);return{candidates:[{text:'分镜'}]};});
  await client.waterSmartCandidates('p1');await client.waterTaskAction('task-1','retry');await client.waterTaskAction('task-1','cancel');
  expect(calls).toEqual([
    ['/api/shuihuo-production/projects/p1/segmentation/smart',{method:'POST',data:{},timeout:180000}],
    ['/api/shuihuo-production/tasks/task-1/retry',{method:'POST',data:{}}],
    ['/api/shuihuo-production/tasks/task-1/cancel',{method:'PUT',data:{}}],
  ]);
});

test('source edit and paragraph candidates require explicit confirmation before persisted segments', async () => {
  const calls: unknown[] = [];
  const client = createProductionClients(async (path, options) => { calls.push([path, options]); return { candidates: [{ text: 'Scene', speaker: '旁白' }] }; });
  await client.saveWaterSource('7', '  Source  ');
  const candidates = await client.waterCandidates('7');
  expect(candidates).toEqual([{ text: 'Scene', speaker: '旁白' }]);
  expect(calls).toHaveLength(2);
  await client.confirmWaterCandidates('7', candidates);
  expect(calls).toEqual([
    ['/api/shuihuo-production/projects/7/source', { method: 'PUT', data: { sourceText: 'Source' } }],
    ['/api/shuihuo-production/projects/7/segmentation/paragraphs', { method: 'POST', data: {} }],
    ['/api/shuihuo-production/projects/7/segmentation/confirm', { method: 'POST', data: { candidates: [{ text: 'Scene', speaker: '旁白' }] } }],
  ]);
});

test('segment prompt edits and asset edits keep established manual mutation contracts', async () => {
  const calls: unknown[] = [];
  const client = createProductionClients(async (path, options) => { calls.push([path, options]); return {}; });
  await client.saveWaterSegment({ id: 3, subtitleText: 'Caption', imagePrompt: 'Image', videoPrompt: 'Video', speaker: '旁白', confirmed: true });
  await client.saveWaterAsset('7', { id: 4, name: ' Hero ', category: 'character', description: 'Coat', prompt: 'Portrait' });
  await client.saveWaterAsset('7', { name: ' Square ', category: 'scene', description: 'Evening' });
  expect(calls[0]).toEqual(['/api/shuihuo-production/segments/3', { method: 'PUT', data: { id: 3, subtitleText: 'Caption', imagePrompt: 'Image', videoPrompt: 'Video', speaker: '旁白', confirmed: true } }]);
  expect(calls[1]).toEqual(['/api/shuihuo-production/assets/4', { method: 'PUT', data: { id: 4, name: 'Hero', category: 'character', description: 'Coat', prompt: 'Portrait', source: 'manual', manuallyEdited: true } }]);
  expect(calls[2]).toEqual(['/api/shuihuo-production/projects/7/assets', { method: 'POST', data: { name: 'Square', category: 'scene', description: 'Evening', source: 'manual', manuallyEdited: true } }]);
});

test('image catalog rejects missing provider ids and ambiguous channels',()=>{
 expect(waterImageModels([{id:'missing',kind:'image'},{id:'a',kind:'image',modelId:'same'},{id:'b',kind:'image',modelId:'same'},{id:'c',kind:'image',modelId:'unique'},{id:'d',kind:'image',modelId:'unique',enabled:false}]).map(item=>item.id)).toEqual(['c']);
});
