import test from 'node:test';
import assert from 'node:assert/strict';
import { listComicAssetLibrary, saveComicAssetLibrarySource, stageComicAssetLibraryEntry } from './comicScreenplayAssetLibrary.js';
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
