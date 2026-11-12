import test from 'node:test';
import assert from 'node:assert/strict';
import { operationReviewDetails, selectedOperationIds } from './aicutEditPlanReview.js';
test('review exposes exact source and timeline times, text, identifiers and all parameters',()=>{
 const details=operationReviewDetails({id:'p',type:'place_asset',assetId:'actual',trackId:'v',startMs:1234,sourceStartMs:321,durationMs:4567,text:'真实对白',parameters:{intensity:0.4},volume:0});
 assert.match(details,/1234/);assert.match(details,/321/);assert.match(details,/4567/);assert.match(details,/真实对白/);assert.match(details,/actual/);assert.match(details,/intensity/);assert.match(details,/0.4/);assert.match(details,/"volume": 0/);
});
test('selected operations follow draft order and never add unknown identifiers',()=>assert.deepEqual(selectedOperationIds({operations:[{id:'a'},{id:'b'}]},['b','unknown']),['b']));
