import {expect,test,vi} from 'vitest';
import {fetchAccountOptions,scopedFetchRequest,createBooktestManagement,knowledgeFields} from './booktestManagement';
test('manager can only select bound members and dev reads global directory',()=>{
 const rows=[{username:'a',role:'member',boundTo:'m'},{username:'b',role:'member',boundTo:'other'},{username:'n',role:'manager',boundTo:'m'}];
 expect(fetchAccountOptions({username:'m',role:'manager'},rows,[]).map(x=>x.username)).toEqual(['m','a']);
 expect(fetchAccountOptions({username:'a',role:'member'},rows,rows).map(x=>x.username)).toEqual(['a']);
});
test('account header only follows fetch requests, never global tasks or models',async()=>{
 const raw=vi.fn(async(_path:string,_options?:any)=>({})),request=scopedFetchRequest(raw as any,'a');
 await request('/api/novel-fetch-workshop/config');await request('/api/batch-rewrite/config',{header:{Other:'x'}});await request('/api/models?kind=text');
 expect(raw.mock.calls[0][1].header['X-Qiantie-Novel-Fetch-Account']).toBe('a');expect(raw.mock.calls[1][1].header.Other).toBe('x');expect(raw.mock.calls[2][1]?.header).toBeUndefined();
});
test('knowledge preserves collection and unknown fields with revision guard; IDs safely encoded',async()=>{
 const request=vi.fn(async(_path:string,_options?:any)=>({ok:true,item:{id:'one'}}));const client=createBooktestManagement(request);
 await client.saveKnowledge('high_imitation',{id:'one',collection:'references',extra:7,name:'a',content:'正文'},9);
 expect(request.mock.calls[0][1].data).toMatchObject({configRevision:9,item:{collection:'references',extra:7}});
 await client.removeKnowledge('opening_phrases','a/b',10);expect(request.mock.calls[1][0]).toContain('a%2Fb?configRevision=10');
 expect(knowledgeFields('high_imitation','references').map(x=>x[0])).toContain('content');
});
test('queue status reads current batch scope and start/pause are distinct commands',async()=>{
 const request=vi.fn(async(_path:string,_options?:any)=>({ok:true}));const c=createBooktestManagement(request);await c.queue('batch-1','');await c.queue('batch-1','2026-11-24');await c.automation(true);await c.automation(false);
 expect(request.mock.calls.map(x=>x[0])).toEqual(['/api/novel-fetch-workshop/realtime/status?batch_id=batch-1','/api/novel-fetch-workshop/realtime/status?date=2026-11-24','/api/novel-fetch-workshop/automation/start','/api/novel-fetch-workshop/automation/pause']);
});
