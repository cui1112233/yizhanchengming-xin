import test from 'node:test';
import assert from 'node:assert/strict';
import { shortStoryCounts } from './shortStoryInputs.js';
test('short-form counts follow recovered rounding, defaults and caps',()=>{
  assert.deepEqual(shortStoryCounts(12345.6,2.6),{targetWordCount:12346,targetChapters:3});
  for(const value of [null,'',undefined,NaN,Infinity,-1])assert.deepEqual(shortStoryCounts(value,value),{targetWordCount:13000,targetChapters:1});
  assert.deepEqual(shortStoryCounts(999,0),{targetWordCount:13000,targetChapters:1});
  assert.deepEqual(shortStoryCounts(900000,80),{targetWordCount:500000,targetChapters:50});
  assert.deepEqual(shortStoryCounts('1000','1'),{targetWordCount:1000,targetChapters:1});
});
