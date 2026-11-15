import test from 'node:test';
import assert from 'node:assert/strict';
import { taskDetailNeedsContent } from './novel-fetch-detail-content.js';
test('summary and diagnostic tabs do not fetch novel bodies; body tabs load only once', () => {
 for (const tab of ['summary','sensitive','rules','submit']) assert.equal(taskDetailNeedsContent({content_loaded:false},tab),false);
 for (const tab of ['original','ai','logs']) assert.equal(taskDetailNeedsContent({content_loaded:false},tab),true);
 assert.equal(taskDetailNeedsContent({original:'已加载'},'original'),false);
});
