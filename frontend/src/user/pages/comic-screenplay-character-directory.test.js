import test from 'node:test';
import assert from 'node:assert/strict';
import { buildCharacterDirectoryImport } from './comic-screenplay-character-directory.js';
function file(path,type='image/png',size=3){const value=new File(['x'.repeat(size)],path.split('/').at(-1),{type});Object.defineProperty(value,'webkitRelativePath',{value:path});return value;}
test('CM目录计划保留浏览器顺序，中文根分类编号，嵌套统一归根并公开跳过项',()=>{
  let count=0;const files=[file('古言/乙.WEBP','image/webp'),file('古言/说明.txt','text/plain'),file('古言/子目录/甲.png',''),file('古言/动图.gif','image/gif'),file('古言/示例.svg','image/svg+xml')];
  const result=buildCharacterDirectoryImport(files,()=>`key-${++count}`);
  assert.equal(result.category,'古言');assert.deepEqual(result.items.map(item=>[item.name,item.category,item.requestKey,item.relativePath]),[['古言1','古言','key-1','古言/乙.WEBP'],['古言2','古言','key-2','古言/子目录/甲.png'],['古言3','古言','key-3','古言/动图.gif']]);
  assert.equal(result.items[1].file.type,'image/png');assert.equal(result.items[1].file.size,3);assert.equal(result.skipped.length,2);assert.equal(result.skipped[0].relativePath,'古言/说明.txt');assert.equal(count,3);assert.equal(buildCharacterDirectoryImport([]),null);
});
test('目录混根/缺相对路径/路径跳转或伪装文件名拒绝，不创建上传编号',()=>{
  let count=0;const key=()=>{count++;return 'key';};
  assert.throws(()=>buildCharacterDirectoryImport([file('古言/a.png'),file('现言/b.png')],key),/同一个根文件夹/);
  for(const path of ['../a.png','古言/../a.png','/古言/a.png','古言//a.png','C:/古言/a.png','古言\\a.png'])assert.throws(()=>buildCharacterDirectoryImport([file(path)],key),/路径/);
  const mismatch=file('古言/a.png');Object.defineProperty(mismatch,'name',{value:'别名.png'});assert.throws(()=>buildCharacterDirectoryImport([mismatch],key),/文件名/);
  assert.equal(count,0);
});
