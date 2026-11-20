import test from 'node:test';
import assert from 'node:assert/strict';
import {runProgress} from './runProgress.js';

test('首个模型失败或暂停的100终止事件不能伪装生成完成',()=>{
 const events=[{type:'run.stage',progress:5},{type:'run.paused',progress:100}];
 assert.deepEqual(runProgress({status:'paused',errorCode:'novel_upstream_result_unknown'},events),{percent:5,status:'exception'});
 assert.deepEqual(runProgress({status:'failed'},[{type:'run.failed',progress:100}]),{percent:0,status:'exception'});
});
test('确认与取消沿用已完成步骤进度，仅成功显示100',()=>{
 assert.deepEqual(runProgress({status:'awaiting_confirmation'},[{type:'run.stage',progress:60},{type:'run.awaiting_confirmation',progress:100}]),{percent:60,status:'normal'});
 assert.deepEqual(runProgress({status:'cancelled'},[{type:'run.stage',progress:20},{type:'run.cancelled',progress:100}]),{percent:20,status:'exception'});
 assert.deepEqual(runProgress({status:'succeeded'},[]),{percent:100,status:'success'});
 assert.deepEqual(runProgress({status:'writing'},[{type:'run.stage',progress:100}]),{percent:99,status:'active'});
});
