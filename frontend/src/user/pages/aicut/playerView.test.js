import test from 'node:test';
import assert from 'node:assert/strict';
import { previewSource, previewDimensions } from './playerView.js';
test('原画使用原片，流畅仅在有代理时使用代理', () => {
 const asset={kind:'video',url:'/original.mp4',proxyUrl:'/proxy.mp4'};
 assert.equal(previewSource(asset),'/original.mp4');
 assert.equal(previewSource(asset,'proxy'),'/proxy.mp4');
 assert.equal(previewSource({...asset,proxyUrl:''},'proxy'),'/original.mp4');
});
test('预览缩放保持画面比例且不改工程内容', () => {
 const fit=previewDimensions('9:16',500,500);
 assert.ok(fit.width<=468 && fit.height<=476);
 assert.equal(fit.width/fit.height,9/16);
 assert.deepEqual(previewDimensions('16:9',500,500,50),{width:960,height:540,scale:.5});
});
