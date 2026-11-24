import {expect,test,vi} from 'vitest';
vi.mock('@tarojs/taro',()=>({default:{arrayBufferToBase64:(b:ArrayBuffer)=>Buffer.from(b).toString('base64')}}));
import {prepareFetchIntake,submitFetchIntake,recoverFetchIntake} from './booktestIntake';
import {createNovelJournal} from './novelRecovery';
const journal=()=>{const m=new Map();return createNovelJournal({get:k=>m.get(k),set:(k,v)=>m.set(k,v),remove:k=>m.delete(k)},'self');};
test('selected AI body must exist and original/classification sources survive intake',async()=>{
 const request=vi.fn(async()=>({originalRaw:'完整原文',processedText:'处理正文',versions:{ai2:'选定AI版本'},meta:{bookName:'甲',gender:'男频',style:'玄幻',genderSource:'manual'}}));
 const p=await prepareFetchIntake(request as any,['b'],'ai2');expect(p.books[0]).toMatchObject({sourceText:'选定AI版本',originalRaw:'完整原文',sourceMetadata:{sourceContentVersion:'ai2',genderSource:'manual'}});
 await expect(prepareFetchIntake(request as any,['b'],'ai3')).rejects.toThrow('没有 ai3');
});
test('staged content then intake then disabled production batch; no generation commands',async()=>{
 const j=journal(),calls:any[]=[];const request=vi.fn(async(p:string,o?:any)=>{calls.push([p,o]);if(p.endsWith('source-chunks'))return {ref:'ref1'};if(p.endsWith('novel-fetch'))return {intake:{id:'i1'}};return {batch:{id:'b1'}};});
 await expect(submitFetchIntake(request as any,{books:[{bookId:'b',platform:'p',sourceText:'正文',originalRaw:'原文'}],metadata:{selectedTaskIds:['b']}},j)).resolves.toEqual({batchId:'b1'});
 expect(calls[2]).toEqual(['/api/batch-factory/v12/intakes/i1/batches',{method:'POST',data:{title:'测书入厂',automationEnabled:false}}]);expect(j.read('intake')).toBeNull();
});
test('unknown intake outcome survives reentry and blocks duplicate writes',async()=>{
 const j=journal(),r=vi.fn(async(p:string)=>{if(p.endsWith('source-chunks'))return {ref:'ref'};throw Error('timeout');});const p={books:[{bookId:'b',sourceText:'正文',originalRaw:'原文'}],metadata:{selectedTaskIds:['b']}};
 await expect(submitFetchIntake(r as any,p,j)).rejects.toThrow('timeout');const n=r.mock.calls.length;await expect(submitFetchIntake(r as any,p,j)).rejects.toThrow('待核对');expect(r.mock.calls).toHaveLength(n);
});

test('upload identity follows backend 24-byte hex contract',async()=>{const j=journal(),r=vi.fn(async(p:string,o?:any)=>{if(p.endsWith('source-chunks')){expect(o.data.uploadId).toMatch(/^[a-f0-9]{48}$/);return {ref:'ref'};}if(p.endsWith('novel-fetch'))return {intake:{id:'i'}};return {batch:{id:'b'}};});await submitFetchIntake(r as any,{books:[{bookId:'b',sourceText:'文',originalRaw:'原'}],metadata:{}},j);});
test('second step recovery uses original unconsumed intake and never duplicates intake',async()=>{const j=journal();j.save('intake',{phase:'batch',intakeId:'i'});const r=vi.fn(async(p:string,o?:any)=>p.endsWith('/intakes/i')?{intake:{id:'i'}}:p.endsWith('/batches')&&o?.method==='POST'?{batch:{id:'b'}}:{batches:[]});await expect(recoverFetchIntake(r as any,j)).resolves.toEqual({batchId:'b'});expect(r.mock.calls.some(([p])=>p.endsWith('novel-fetch'))).toBe(false);});
test('lost batch receipt locates consumed source intake instead of creating again',async()=>{const j=journal();j.save('intake',{phase:'batch',intakeId:'i'});const r=vi.fn(async(p:string,o?:any)=>p.endsWith('/intakes/i')?{intake:{id:'i',consumedAt:'now'}}:{batches:[{id:'b',sourceIntakeId:'i'}]});await expect(recoverFetchIntake(r as any,j)).resolves.toEqual({batchId:'b'});expect(r.mock.calls.every(([,o])=>!o?.method)).toBe(true);});
