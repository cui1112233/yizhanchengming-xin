import test from 'node:test';
import assert from 'node:assert/strict';
import * as api from '../../../shared/api/batchFactoryText.js';

test('预览请求传递取消信号和服务端对比度提示', async () => {
  globalThis.localStorage = {getItem: () => null};
  const controller = new AbortController();
  let signal;
  globalThis.fetch = async (_, options) => {
    signal = options.signal;
    return new Response(new Blob(['preview'], {type:'image/png'}), {headers:{'X-Text-Contrast':'low'}});
  };
  const {blob:image,warning} = await api.previewTextPage('b','k','p',{text:'草稿。'}, {
    signal:controller.signal, withMetadata:true
  });
  assert.equal(image.size,7);
  assert.match(warning,/对比度较低/);
  assert.ok(signal);
});

test('文字页编辑动作携带两级版本和绝对句末位置',async () => {
  const url = new URL('./textMediaActions.js',import.meta.url);
  const {existsSync}=await import('node:fs');
  assert.ok(existsSync(url), '缺少文字页编辑动作');
  const {pageChanges, initialPageDraft, splitChoices, uploadLabel} = await import(url);
  const page={id:'P',revision:2,startOffset:8,endOffset:14,text:'甲😀。乙二。'};
  const doc={revision:4,configSnapshot:{layout:{fontSize:36,padding:0}}};
  const base=initialPageDraft(doc,page);
  assert.equal(base.layout.padding,0);
  assert.deepEqual(pageChanges(base,{...base,text:'甲😀。改了。'},2,4),{expectedRevision:2,expectedDocumentRevision:4,text:'甲😀。改了。'});
  assert.deepEqual(splitChoices(page,[3,8,11,14,20]),[{offset:11,label:'甲😀。'}]);
  assert.deepEqual(splitChoices(page,null),[]);
  assert.equal(uploadLabel({id:'c',status:'uploaded'}),'再次上传');
  assert.equal(uploadLabel({id:'c',status:'ready'}),'上传网络');
});
