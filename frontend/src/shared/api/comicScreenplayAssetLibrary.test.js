import test from 'node:test';
import assert from 'node:assert/strict';
import * as categoryApi from './comicScreenplayAssetLibrary.js';
test('独立分类接口保持精确名称、版本和取消信号，拒绝非法输入和响应',async()=>{
 const calls=[],state={builtin_categories:['古言'],custom_categories:[' 自定义 '],revision:2},request=async(path,options)=>{calls.push({path,options});return state;},signal=new AbortController().signal;
 assert.equal(typeof categoryApi.listComicAssetLibraryCategories,'function');
 assert.deepEqual(await categoryApi.listComicAssetLibraryCategories({signal},request),state);
 await categoryApi.changeComicAssetLibraryCategory({action:'add',name:' 自定义 ',expectedRevision:0},{signal},request);
 assert.equal(calls[0].path,'/api/comic-screenplays/asset-library/categories');assert.equal(calls[0].options.signal,signal);
 assert.deepEqual(JSON.parse(calls[1].options.body),{action:'add',name:' 自定义 ',expectedRevision:0});
 for(const name of [' ','a\n','a'.repeat(192)])assert.throws(()=>categoryApi.changeComicAssetLibraryCategory({action:'add',name,expectedRevision:0},{},request),/分类/);
 assert.throws(()=>categoryApi.changeComicAssetLibraryCategory({action:'wipe',name:'a',expectedRevision:0},{},request),/操作/);
 assert.throws(()=>categoryApi.changeComicAssetLibraryCategory({action:'add',name:'a',expectedRevision:-1},{},request),/版本/);
 await assert.rejects(categoryApi.listComicAssetLibraryCategories({},async()=>({entries:[]})),/分类.*格式/);
});
import { listComicAssetLibrary, saveComicAssetLibrarySource, stageComicAssetLibraryEntry, uploadComicAssetLibraryFile, updateComicAssetLibraryEntry, deleteComicAssetLibraryEntry, replaceComicAssetLibraryImage, updateComicAssetLibraryAudio } from './comicScreenplayAssetLibrary.js';
test('普通库读取、保存、暂存保留版本和请求编号，库ID不当作项目媒体ID', async () => {
  const calls=[], request=async(path,options)=>{calls.push({path,options});return path.includes('/stage')?{item:{id:'media-copy'}}:options?.method?{entry:{id:'library'}}:{entries:[{id:'library'}]};};
  assert.equal((await listComicAssetLibrary({type:'scene'},request))[0].id,'library');
  const source={requestKey:'stable-source',assetId:'a',expectedAssetRevision:4,expectedSelectedItemId:'image-A',category:'测试'};
  assert.equal((await saveComicAssetLibrarySource('p',source,{},request)).id,'library');
  const stage={requestKey:'stable-stage',targetId:'a',expectedEntryRevision:2};
  assert.equal((await stageComicAssetLibraryEntry('p','library',stage,{},request)).id,'media-copy');
  assert.equal(calls[0].path,'/api/comic-screenplays/asset-library?type=scene');
  assert.equal(calls[1].path,'/api/comic-screenplays/projects/p/asset-library');
  assert.deepEqual(JSON.parse(calls[1].options.body),source);
  assert.equal(calls[2].path,'/api/comic-screenplays/projects/p/asset-library/library/stage');
  assert.deepEqual(JSON.parse(calls[2].options.body),stage);
  assert.throws(()=>stageComicAssetLibraryEntry('p','../bad',stage,{},request),/无效/);
  assert.throws(()=>listComicAssetLibrary({type:'other'},request),/无效/);
});

test('参考音频支持新建随图提交、已有角色仅换音频和图音一起提交，不能写外部音频地址',async()=>{
 const calls=[],request=async(path,options)=>{calls.push({path,options});return {entry:{id:'role'}};};
 const file=new File(['img'],'image.png',{type:'image/png'}),audioFile=new File(['wav'],'参考.wav',{type:'audio/wav'});
 const values={type:'character',name:'角色',requestKey:'reference-audio-key',audioFile};
 await uploadComicAssetLibraryFile({...values,file},{},request);assert.equal(calls[0].options.body.get('audio').name,'参考.wav');
 await updateComicAssetLibraryAudio('role',{...values,expectedRevision:2},{},request);assert.equal(calls[1].path,'/api/comic-screenplays/asset-library/role/audio?expectedRevision=2');assert.equal(calls[1].options.body.has('file'),false);assert.equal(calls[1].options.body.get('requestKey'),values.requestKey);
 await replaceComicAssetLibraryImage('role',{...values,file,expectedRevision:2},{},request);assert(calls[2].options.body.has('file')&&calls[2].options.body.has('audio'));
 assert.throws(()=>updateComicAssetLibraryAudio('role',{...values,expectedRevision:2,audioFile:null,audio:'https://outside.invalid/a.wav'}),/选择.*音频/);
});

test('角色换图只上传编辑草稿和新图，原ID/版本/稳定请求编号不能丢',async()=>{
 const calls=[],request=async(path,options)=>{calls.push({path,options});return {entry:{id:'same-entry',revision:4}};};
 const file=new File(['replacement'],'new.png',{type:'image/png'}),signal=new AbortController().signal;
 await replaceComicAssetLibraryImage('same-entry',{type:'character',name:'保留名称',description:'新描述',category:'古言',prompt:'保留提示词',voice:'保留声音',file,expectedRevision:3,requestKey:'same-replacement-key'},{signal},request);
 assert.equal(calls[0].path,'/api/comic-screenplays/asset-library/same-entry/image?expectedRevision=3');assert.equal(calls[0].options.method,'PUT');assert.equal(calls[0].options.signal,signal);
 const form=calls[0].options.body;assert.equal(form.get('requestKey'),'same-replacement-key');assert.equal(form.get('file').name,'new.png');assert.equal(form.get('prompt'),'保留提示词');
 assert.throws(()=>replaceComicAssetLibraryImage('same-entry',{file,expectedRevision:0}),/版本/);assert.equal(calls.length,1);
});

test('普通库管理上传仅单真实file，编辑完整版本元数据，删除携带版本',async()=>{
  const calls=[],request=async(path,options)=>{calls.push({path,options});return options.method==='DELETE'?{deleted:true}:{entry:{id:'entry'}};};
  const file=new File(['png'],'scene.png',{type:'image/png'}),signal=new AbortController().signal;
  await uploadComicAssetLibraryFile({file,type:'scene',requestKey:'stable-upload',description:'说明',category:'古代',prompt:'画面',voice:'声音'},{signal},request);
  const form=calls[0].options.body;assert.ok(form instanceof FormData);assert.equal(form.getAll('file').length,1);assert.equal(form.get('name'),'scene.png');assert.equal(form.get('requestKey'),'stable-upload');assert.equal(form.has('source_path'),false);assert.equal(form.has('url'),false);assert.equal(calls[0].options.signal,signal);
  await updateComicAssetLibraryEntry('entry',{expectedRevision:3,name:'新名',description:'说明',category:'古代',prompt:'画面',voice:'声音'},{},request);
  assert.deepEqual(JSON.parse(calls[1].options.body),{expectedRevision:3,name:'新名',description:'说明',category:'古代',prompt:'画面',voice:'声音'});assert.equal(calls[1].options.method,'PATCH');
  await deleteComicAssetLibraryEntry('entry',3,{},request);assert.equal(calls[2].path,'/api/comic-screenplays/asset-library/entry?expectedRevision=3');assert.equal(calls[2].options.method,'DELETE');
  assert.throws(()=>uploadComicAssetLibraryFile({file,type:'character',requestKey:'k'}),/名称/);
  assert.throws(()=>deleteComicAssetLibraryEntry('../other',3),/无效/);
});
