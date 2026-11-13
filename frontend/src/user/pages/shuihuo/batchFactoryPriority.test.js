import test from 'node:test';
import assert from 'node:assert/strict';
import { batchFactoryPriorityAction } from './batchFactoryPriority.js';

test('only waiting books in a live production queue can be prioritized', () => {
  const queue = {state:'running',books:[{bookId:'a',status:'running'},{bookId:'b',status:'queued'},{bookId:'c',status:'completed'},{bookId:'d',status:'queued',priorityRequested:true}]};
  assert.equal(batchFactoryPriorityAction(queue,'a').enabled,false);
  assert.equal(batchFactoryPriorityAction(queue,'b').enabled,true);
  assert.equal(batchFactoryPriorityAction(queue,'c').enabled,false);
  assert.equal(batchFactoryPriorityAction(queue,'new').enabled,true);
  assert.deepEqual(batchFactoryPriorityAction(queue,'d').label,'已插队');
  assert.equal(batchFactoryPriorityAction(queue,'d').enabled,false);
  assert.equal(batchFactoryPriorityAction({...queue,state:'paused'},'b').enabled,true);
  assert.equal(batchFactoryPriorityAction({...queue,state:'idle'},'b').enabled,false);
});
