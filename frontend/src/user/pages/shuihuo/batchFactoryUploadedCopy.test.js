import test from 'node:test';
import assert from 'node:assert/strict';
import { batchFactoryUploadedBooksForCopy, batchFactoryNovelCopyText } from './batchFactoryRuntimeLogic.js';

test('只复制当前确认上传成功的文本书，排除待回执、失败、仅生成和别的书的回执', () => {
  const books=['ok','pending','failed','generated','wrong'].map(id=>({id,title:id,bookId:id,platform:'zhihu'}));
  const receipt={kind:'text',status:'published',carouselId:9,imageIds:[11,12],pageCount:2};
  const stageSummaries={
    ok:{textPublishSource:'manual',textPublishRecord:{...receipt,bookId:'ok'},runs:[]},
    pending:{textPublishSource:'manual',textPublishRecord:{...receipt,bookId:'pending',status:'pending'},runs:[]},
    failed:{textPublishSource:'manual',textPublishRecord:{...receipt,bookId:'failed',status:'failed'},textLastPublishedRecord:{...receipt,bookId:'failed'},runs:[]},
    generated:{runs:[{stage:'text_collection',status:'succeeded'}]},
    wrong:{textPublishSource:'manual',textPublishRecord:{...receipt,bookId:'someone-else'},runs:[]}
  };
  const copied=batchFactoryUploadedBooksForCopy(books,{productionMode:'text',stageSummaries});
  assert.deepEqual(copied.map(book=>book.id),['ok']);
  assert.equal(batchFactoryNovelCopyText(copied,{zhihu:'知乎'}),'ok\tok\t知乎');
});

test('视频和漫画使用各自上传状态，保留列表顺序且不修改原书目', () => {
  const books=[{id:'v',title:'视频',sourceMetadata:{websiteSubmitStatus:'uploaded'}},{id:'c',title:'漫画'},{id:'waiting',title:'待上传'}];
  const copied=batchFactoryUploadedBooksForCopy(books,{stageSummaries:{c:{comicPublishedCollectionCount:1}}},book=>book.id==='c'?'comic':'video');
  assert.deepEqual(copied.map(book=>book.id),['v','c']);
  assert.equal(books.length,3);
  assert.deepEqual(batchFactoryUploadedBooksForCopy([],{productionMode:'text'}),[]);
});
