import {expect,test,vi} from 'vitest';
import {createBooktestJobs,booktestInput,fetchVersionTexts} from './booktestJobs';
test('the actual start contract includes selected versions, rewrite methods, selected model and the processing limit',async()=>{
 const request=vi.fn().mockResolvedValue({id:'go-workshop-1',status:'queued'});
 const input=booktestInput({input:'123456',platformId:'2',maxTxt:4000,selectedVersions:['ai2'],aiSlotMethods:{ai2:'instruction'},sensitiveAiEnabled:true},'m1');
 await createBooktestJobs(request).start(input);
 expect(request).toHaveBeenCalledExactlyOnceWith('/api/novel-fetch-workshop/process/start',{method:'POST',data:{inputText:'123456',platform_id:'2',maxTxt:4000,selected_versions:['ai2'],rewrite_methods:{ai2:'instruction'},sensitive_ai_enabled:true,text_model_id:'m1'},timeout:180000});
});
test('job reads are read-only and stopping is limited to the displayed job',async()=>{
 const request=vi.fn().mockResolvedValue({id:'go-workshop-1'}),client=createBooktestJobs(request);await client.read('go-workshop-1');await client.latest();await client.stop('go-workshop-1');
 expect(request.mock.calls).toEqual([['/api/novel-fetch-workshop/process/jobs/go-workshop-1'],['/api/novel-fetch-workshop/process/jobs/latest?limit=10'],['/api/novel-fetch-workshop/process/stop',{method:'POST',data:{job_id:'go-workshop-1',pause_auto_queue:false}}]]);
});
test('real body-store version maps and legacy arrays both expose saved text instead of a fabricated empty result',()=>{
 expect(fetchVersionTexts({versions:{ai1:{state:'ready',text:'一稿'},ai2:'二稿',original_raw:'原文'}})).toEqual([{key:'ai1',label:'AI1',content:'一稿'},{key:'ai2',label:'AI2',content:'二稿'}]);
 expect(fetchVersionTexts({ai_texts:[{version:'ai3',text:'三稿'}]})).toEqual([{key:'ai3',label:'ai3',content:'三稿'}]);
});
