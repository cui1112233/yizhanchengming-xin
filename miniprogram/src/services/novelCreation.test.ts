import { expect, test, vi } from 'vitest';
import { createNovelClient } from './novelCreation';
import { buildNovelCreation, splitNovelChapters, continuationMessage } from './novelCreationInputs';
import { createNovelJournal } from './novelRecovery';
import { moduleRoute } from '../config/modules';
test('novels use an independent entry without breaking existing fetch task links',()=>{
  expect(moduleRoute('novel')).toBe('/comic/pages/novel/index');
  expect(moduleRoute('fetch')).toBe('/comic/pages/module-detail/index?key=fetch');
});
test('original discussions use novel studio endpoints, not fetch or generic chat',async()=>{
  const request=vi.fn().mockResolvedValue({session:{id:'s1'}});const client=createNovelClient(request);
  await client.createStudio('short_original','key1');await client.sendStudio('s1',{content:'一个故事',modelId:'m1',idempotencyKey:'msg1'});
  expect(request.mock.calls.map(([path])=>path)).toEqual(['/api/novel-creation/studio-sessions','/api/novel-creation/studio-sessions/s1/messages']);
  expect(request.mock.calls[0][1].data).toEqual({mode:'short_original',idempotencyKey:'key1'});
});
test('long and short imitation keep different count rules and endpoints',()=>{
  const base={modelId:'m1',referenceText:'文'.repeat(100),targetWordCount:5000,targetChapters:100};
  expect(buildNovelCreation('long_imitation',base,'key').path).toBe('/long-imitation');
  expect(()=>buildNovelCreation('short_imitation',base,'key')).toThrow('章节');
  expect(buildNovelCreation('short_imitation',{...base,targetWordCount:13000,targetChapters:1},'key').data).toMatchObject({mode:'short_imitation',targetWordCount:13000,targetChapters:1});
});
test('spinoffs require their parent and continuation imports real chapters',()=>{
  expect(()=>buildNovelCreation('fanfiction',{modelId:'m1',title:'番外'},'key')).toThrow('正传');
  const request=buildNovelCreation('continuation',{title:'导入作品',referenceText:'第1章 出发\n正文一\n第2章 重逢\n正文二'},'key');
  expect(request.path).toBe('/chapter-imports');expect((request.data as Record<string,any>).chapters).toHaveLength(2);
  expect(continuationMessage([{number:2,content:'最新正文'},{number:1,content:'旧正文'}],'继续', 'm1','key')).toMatchObject({modeInputs:{previousText:'最新正文'}});
  expect(()=>continuationMessage([],'继续','m1','key')).toThrow('正式章节');
});
test('chapter parser never loses an unheaded source paragraph',()=>{
  expect(splitNovelChapters('序言正文\n第1章 起点\n内容').map(item=>item.content).join('\n')).toContain('序言正文');
});
test('chapter save sends the expected version and does not retry conflicts',async()=>{
  const request=vi.fn().mockRejectedValue(Object.assign(new Error('conflict'),{statusCode:409}));
  await expect(createNovelClient(request).saveChapter('p1',{id:'c1',version:7,title:'标题',content:'新正文'})).rejects.toThrow('conflict');
  expect(request).toHaveBeenCalledExactlyOnceWith('/api/novel-creation/projects/p1/chapters/c1',{method:'PUT',data:{expectedVersion:7,title:'标题',content:'新正文'}});
});
test('recovery records isolate accounts and failed persistence must stop before submission',()=>{
  const map=new Map();const storage={get:(key:string)=>map.get(key),set:(key:string,value:any)=>map.set(key,value),remove:(key:string)=>map.delete(key)};
  const a=createNovelJournal(storage,'origin:alice'),b=createNovelJournal(storage,'origin:bob');a.save('create',{idempotencyKey:'same-key',content:'原参数'});
  expect(b.read('create')).toBe(null);expect(a.read('create')).toMatchObject({idempotencyKey:'same-key'});
  expect(()=>createNovelJournal({...storage,set:()=>{throw Error('full');}},'alice').save('create',{})).toThrow('恢复记录');
});
