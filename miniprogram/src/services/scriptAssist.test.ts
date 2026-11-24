import {expect,test,vi} from 'vitest';
import {createScriptAssist,mergeEntityEnrichment,replaceScriptShot} from './scriptAssist';
test('online entity enrichment uses server prompt type and preserves existing user fields',async()=>{
 const request=vi.fn(async(_p:string,_o?:any)=>({enrichment:{fields:{名称:'被覆盖',外形:'黑发'},evidence:['原文依据'],uncertainties:['年龄未知']}}));const c=createScriptAssist(request);
 const r=await c.enrich('characters',{id:'a',data:{名称:'小明',外形:''}},'小说原文','p','m',{characters:[],scenes:[]});
 expect(request.mock.calls[0][1].data).toMatchObject({promptType:'entity_enrich',entityType:'character',textModelId:'m',entity:{名称:'小明'}});
 expect(mergeEntityEnrichment({id:'a',data:{名称:'小明',外形:''}},r)).toMatchObject({id:'a',data:{名称:'小明',外形:'黑发'}});
});
test('editing repeated shot text updates exact indexed range and preserves the shared prefix',()=>{
 const source='公共设定\n### 分镜1\n相同\n### 分镜2\n相同';expect(replaceScriptShot(source,1,'### 分镜2\n新内容')).toBe('公共设定\n### 分镜1\n相同\n### 分镜2\n新内容');
 expect(()=>replaceScriptShot(source,5,'x')).toThrow('分镜');
});
