import test from 'node:test';
import assert from 'node:assert/strict';

test('large multi-book intake sends bounded requests and one text-free manifest', async () => {
  const transport = await import('./batchIntakeTransport.js');
  const a = '甲😀'.repeat(1300000);
  const b = '乙\n'.repeat(2300000);
  const payload = {groups:[{platformId:'3', sourceTextByBookId:{'877777':'甲稿'},rawSourceTextByBookId:{'877777':a}},{platformId:'15',sourceTextByBookId:{'877777':'乙稿'},rawSourceTextByBookId:{'877777':b}}]};
  assert.ok(Buffer.byteLength(JSON.stringify(payload))>16*1024*1024);
  const parts = new Map();
  const result = await transport.stageLargeIntake(payload, async chunk=>{
    assert.ok(Buffer.byteLength(JSON.stringify(chunk)) < 1024*1024);
    const data = parts.get(chunk.platform) || [];
    data[chunk.index] = Buffer.from(chunk.data,'base64');
    parts.set(chunk.platform,data);
    return { ref:'ref-'+chunk.platform };
  });
  assert.ok(JSON.stringify(result).length < 1000);
  assert.equal(result.groups[0].sourceRefsByBookId['877777'], 'ref-3');
  assert.equal(result.groups[1].sourceRefsByBookId['877777'], 'ref-15');
  assert.deepEqual(JSON.parse(Buffer.concat(parts.get('3')).toString()),{sourceText:'甲稿',originalRaw:a});
  assert.deepEqual(JSON.parse(Buffer.concat(parts.get('15')).toString()),{sourceText:'乙稿',originalRaw:b});
  const small = {books:[{bookId:'7',sourceText:'小'}]};
  assert.equal(await transport.stageLargeIntake(small,()=>assert.fail()),small);
});
