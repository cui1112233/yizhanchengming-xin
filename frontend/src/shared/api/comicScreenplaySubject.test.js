import test from 'node:test';
import assert from 'node:assert/strict';
import {updateComicAssetLibraryEntry,uploadComicAssetLibraryFile,replaceComicAssetLibraryImage,updateComicAssetLibraryAudio} from './comicScreenplayAssetLibrary.js';
test('主体存档原文、明确清空、旧调用省略、16KiB及所有写入入口',async()=>{
 const calls=[],request=async(path,options)=>{calls.push(options.body);return {entry:{id:'a'}}},text='  {"主体":"雪🌨"}\n  ';
 for(const sora2Info of [text,'']) {await updateComicAssetLibraryEntry('a',{name:'甲',expectedRevision:1,sora2Info},{},request);assert.equal(JSON.parse(calls.at(-1)).sora2Info,sora2Info);}
 await updateComicAssetLibraryEntry('a',{name:'甲',expectedRevision:1},{},request);assert.equal(Object.hasOwn(JSON.parse(calls.at(-1)),'sora2Info'),false);
 for(const fn of [uploadComicAssetLibraryFile,replaceComicAssetLibraryImage,updateComicAssetLibraryAudio]){const input={name:'甲',type:'character',requestKey:'key',expectedRevision:1,sora2Info:text,file:new File(['x'],'x.png',{type:'image/png'}),audioFile:new File(['a'],'a.mp3')};await(fn===uploadComicAssetLibraryFile?fn(input,{},request):fn('a',input,{},request));assert.equal(calls.at(-1).get('sora2Info'),text);}
 assert.throws(()=>updateComicAssetLibraryEntry('a',{name:'甲',expectedRevision:1,sora2Info:'雪'.repeat(5462)},{},request),/16KiB/);
 assert.throws(()=>uploadComicAssetLibraryFile({type:'scene',name:'景',requestKey:'x',file:new File(['x'],'x.png',{type:'image/png'}),sora2Info:'x'},{},request),/角色/);
});
