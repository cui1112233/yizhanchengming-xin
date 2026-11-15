import test from 'node:test';
import assert from 'node:assert/strict';
import { listComicAssetLibrary, saveComicAssetLibrarySource, stageComicAssetLibraryEntry, uploadComicAssetLibraryFile, updateComicAssetLibraryEntry, deleteComicAssetLibraryEntry } from './comicScreenplayAssetLibrary.js';
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
