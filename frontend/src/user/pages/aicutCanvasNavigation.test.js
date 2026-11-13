import test from 'node:test';
import assert from 'node:assert/strict';
import { creativeCanvasHref, cutEditorHref, creativeTimelineHint, openBoundCreativeCanvas, persistLatestCanvasNavigation } from './aicutCanvasNavigation.js';

test('one canonical canvas URL preserves the explicit return timeline', () => {
  assert.equal(creativeCanvasHref('canvas-1','version-2'),'/agent/canvas?taskId=canvas-1&view=workflow&cutTimeline=version-2');
  assert.equal(cutEditorHref('cut-1','version-2'),'/ai-cut?project=cut-1&timeline=version-2');
  assert.equal(creativeTimelineHint('?taskId=canvas-1&cutTimeline=version-2','canvas-1'),'version-2');
  assert.equal(creativeTimelineHint('?taskId=other&cutTimeline=version-2','canvas-1'),'default');
  assert.throws(()=>creativeCanvasHref('https://evil'),/无效/);
});

test('save and existing binding precede navigation, and scope changes cancel it', async () => {
  const calls=[];let scope='cut-1:version-2';
  const dependencies={scope, getScope:()=>scope, save:async()=>calls.push('save'), ensure:async id=>{calls.push(id);return{taskId:'existing-canvas'}}, navigate:href=>calls.push(href)};
  await openBoundCreativeCanvas({projectId:'cut-1',timelineId:'version-2'},dependencies);
  assert.deepEqual(calls,['save','cut-1','save','/agent/canvas?taskId=existing-canvas&view=workflow&cutTimeline=version-2']);
  calls.length=0;
  await assert.rejects(openBoundCreativeCanvas({projectId:'cut-1'}, {...dependencies,save:async()=>{throw Error('save failed')}}),/save failed/);
  assert.equal(calls.length,0);
  await assert.rejects(openBoundCreativeCanvas({projectId:'cut-1'}, {...dependencies,ensure:async()=>{scope='other';return{taskId:'existing-canvas'}}}),/已切换/);
  assert.deepEqual(calls,['save']);
});

test('edits made while binding loads are saved before leaving the editor', async () => {
  let changed = false, saved = false, navigated = false;
  await openBoundCreativeCanvas({ projectId:'cut-1' }, { scope:'one',getScope:()=> 'one',save:async()=>{if(changed)saved=true},ensure:async()=>{changed=true;return{taskId:'existing-canvas'}},navigate:()=>{assert.equal(saved,true);navigated=true} });
  assert.equal(navigated,true);
});

test('final history wait cannot hide newer edits before navigation', async () => {
  let edits=1,persisted=0,calls=0;
  await persistLatestCanvasNavigation({scope:'one',getScope:()=> 'one',isDirty:()=>edits>persisted,save:async()=>{persisted=edits;calls++;if(calls===1)edits++;return{revision:calls}}});
  assert.equal(calls,2); assert.equal(persisted,edits);
  await assert.rejects(persistLatestCanvasNavigation({scope:'one',getScope:()=> 'one',isDirty:()=>true,save:async()=>null}),/保存失败/);
});
