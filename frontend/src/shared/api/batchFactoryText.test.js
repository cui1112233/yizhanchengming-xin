import test from 'node:test';
import assert from 'node:assert/strict';
import * as api from './batchFactoryV12.js';

test('text requests encode book and page identifiers and send optimistic revisions', async () => {
 globalThis.localStorage={ getItem:()=>null };
 const seen=[];
 globalThis.fetch=async(url,opts)=>{seen.push([url,opts]);return new Response('{}',{status:200,headers:{'Content-Type':'application/json'}})};
 await api.patchTextPage('batch /','book/1','page+2',{expectedRevision:3});
 assert.equal(seen[0][0],'/api/batch-factory/v12/batches/batch%20%2F/books/book%2F1/text-pages/page%2B2');
 assert.equal(seen[0][1].method,'PATCH');assert.deepEqual(JSON.parse(seen[0][1].body),{expectedRevision:3});
 await api.previewTextCover('b','k',{});
 assert.equal(seen[1][0],'/api/batch-factory/v12/batches/b/books/k/text-cover/preview');
 await api.uploadTextPageArtifact('b','k','p',new Blob(['test'],{type:'image/png'}),9);
 assert.equal(seen[2][0],'/api/batch-factory/v12/batches/b/books/k/text-pages/p/upload?expectedRevision=9');
 assert.equal(seen[2][1].body.size,4);
});
