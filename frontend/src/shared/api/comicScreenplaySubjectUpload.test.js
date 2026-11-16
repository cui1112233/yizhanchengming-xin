import test from 'node:test';
import assert from 'node:assert/strict';
import * as api from './comicScreenplayAssetLibrary.js';
test('真人上传仅调用本站接口、同键核对及四个应用入口携带引用',async()=>{
 const calls=[],request=async(path,options={})=>{calls.push({path,...options});return path.endsWith('capability')?{provider:'volcano',label:'上传火山引擎素材',available:false,reason:'请配置'}:path.includes('subject-uploads')?{operation:{requestKey:'same-key',status:'succeeded',result:{assetId:'subject'}}}:{entry:{id:'a'}};};
 assert.equal(typeof api.getComicSubjectUploadCapability,'function','缺少真人上传能力接口');
 assert.equal((await api.getComicSubjectUploadCapability({},request)).available,false);
 const file=new File(['image'],'x.png',{type:'image/png'}),input={requestKey:'same-key',name:'甲',description:'说明',groupId:'g',file};
 await api.createComicSubjectUpload(input,{},request);assert.equal(calls.at(-1).body.get('requestKey'),'same-key');assert.equal(calls.at(-1).body.get('file'),file);await api.createComicSubjectUpload({...input,groupId:''},{},request);assert.equal(calls.at(-1).body.get('groupId'),'');
 await api.getComicSubjectUpload('same-key',{},request);assert.ok(calls.at(-1).path.endsWith('/subject-uploads/same-key'));
 await api.updateComicAssetLibraryEntry('a',{name:'甲',expectedRevision:1,subjectUploadKey:''},{},request);assert.equal(JSON.parse(calls.at(-1).body).subjectUploadKey,'');
 assert.throws(()=>api.createComicSubjectUpload({...input,name:' '},{},request),/名称/);
 for(const fn of [api.uploadComicAssetLibraryFile,api.replaceComicAssetLibraryImage,api.updateComicAssetLibraryAudio,api.updateComicAssetLibraryEntry]){const values={...input,type:'character',expectedRevision:1,audioFile:new File(['audio'],'a.mp3'),subjectUploadKey:'same-key'};await(fn===api.uploadComicAssetLibraryFile?fn(values,{},request):fn('a',values,{},request));const body=calls.at(-1).body;assert.equal(body instanceof FormData?body.get('subjectUploadKey'):JSON.parse(body).subjectUploadKey,'same-key');}
});
