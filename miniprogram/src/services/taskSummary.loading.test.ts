import {expect,test,vi} from 'vitest';
import {loadTaskSummary} from './taskSummary';

test('workbench loads only the selected functional source',async()=>{
 const request=vi.fn(async(path:string)=>{if(path==='/api/batch-rewrite/tasks')return {tasks:[{bookId:'book',title:'current',status:'running'}]};throw Error('unrelated source');});
 const result=await loadTaskSummary(request,{sources:['fetch']});
 expect(request.mock.calls.map(([path])=>path)).toEqual(['/api/batch-rewrite/tasks']);
 expect(result.sources.map(s=>s.source)).toEqual(['fetch']);
 expect(result.entries).toHaveLength(1);
});

test('leaving a page prevents queued summary requests reaching the server',async()=>{
 let visible=true;const releases:Array<()=>void>=[];
 const request=vi.fn((path:string)=>new Promise<unknown>(resolve=>releases.push(()=>resolve({entries:[],tasks:[],batches:[],projects:[]}))));
 const result=loadTaskSummary(request,{shouldContinue:()=>visible});
 const outcome=result.catch(error=>error);
 await Promise.resolve();await Promise.resolve();
 expect(request).toHaveBeenCalledTimes(4);
 visible=false;releases.forEach(release=>release());
 expect(await outcome).toBeInstanceOf(Error);
 expect(request).toHaveBeenCalledTimes(4);
});

test('an explicitly empty source list makes no unrelated calls',async()=>{
 const request=vi.fn();expect(await loadTaskSummary(request,{sources:[]})).toEqual({entries:[],sources:[]});expect(request).not.toHaveBeenCalled();
});

test('a ready source is displayed before a slow optional batch source',async()=>{
 let finish!: (value:unknown)=>void;
 const slow=new Promise(resolve=>{finish=resolve;});const progress=vi.fn();
 const pending=loadTaskSummary(async path=>path==='/api/history'?{entries:[{id:'ready',title:'可用',status:'completed'}]}:slow,{sources:['script','batch'],onProgress:progress});
 for(let i=0;i<12;i++)await Promise.resolve();
 expect(progress).toHaveBeenCalled();expect(progress.mock.calls[0][0].entries[0].id).toBe('ready');
 expect(progress.mock.calls[0][0].sources).toHaveLength(1);
 finish({batches:[]});expect((await pending).sources).toHaveLength(2);
});
