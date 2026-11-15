import test from 'node:test';
import assert from 'node:assert/strict';
import * as archive from './comicScreenplayCharacterArchive.js';
test('整库接口使用会话封装、原键、相对文件名与全量版本，不接受坏输入或假成功',async()=>{
 const calls=[],request=async(path,options)=>{calls.push({path,options});return path.endsWith('export')?new Blob(['v2']):path.endsWith('import')?{imported_count:1,entries:[{id:'a'}]}:{cleared:true,cleared_count:2};};
 const file=new File(['archive'],'characters.lschar');Object.defineProperty(file,'webkitRelativePath',{value:'folder/characters.lschar'});
 await archive.importCharacterArchive({files:[file],requestKey:'same'},{},request);
 await archive.importCharacterArchive({files:[file],requestKey:'same'},{},request);
 assert.equal(calls[0].options.body.get('files').name,'folder/characters.lschar');assert.equal(calls[1].options.body.get('requestKey'),'same');
 await archive.clearCharacterArchive({requestKey:'clear',expected:[{id:'a',revision:1},{id:'b',revision:2}]},{},request);assert.equal(JSON.parse(calls[2].options.body).expected.length,2);
 await archive.exportCharacterArchive({},request);assert.equal(calls[3].options.responseType,'blob');assert.equal(calls[3].options.suppressGlobalError,true);
 for(const files of [[],[new File(['bad'],'other.json')],[new File(['bad'],'characters.lschar',{type:'text/plain'})]]){if(files[0]?.name==='characters.lschar')continue;assert.throws(()=>archive.importCharacterArchive({files,requestKey:'x'},{},request),/characters/);}
 await assert.rejects(archive.importCharacterArchive({files:[file],requestKey:'same'},{},async()=>({imported_count:1})),/未确认/);
 await assert.rejects(archive.clearCharacterArchive({requestKey:'x',expected:[]},{},async()=>({deleted:true})),/未确认/);
 await assert.rejects(archive.exportCharacterArchive({},async()=>new Blob([])),/空/);
});
