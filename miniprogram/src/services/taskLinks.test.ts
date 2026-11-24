import {expect,test} from 'vitest';
import {taskDestination} from './taskSummary';
test('task links use owned project identifiers instead of child job identifiers',()=>{
 expect(taskDestination({source:'batch',id:'child',projectId:'batch',bookId:'book'} as any)).toBe('/pages/batch/index?batchId=batch&bookId=book');
 expect(taskDestination({source:'agent',id:'generation',projectId:'cm'} as any)).toBe('/comic/pages/module-detail/index?key=agent&taskId=cm&tab=tasks');
 expect(taskDestination({source:'script',id:'history'} as any)).toBe('/comic/pages/module-detail/index?key=script&historyId=history');
 expect(taskDestination({source:'novel',id:'run',projectId:'novel',sessionId:'session'} as any)).toBe('/comic/pages/novel/index?projectId=novel&sessionId=session&runId=run');
 expect(taskDestination({source:'aicut',id:'export',projectId:'cut',timelineId:'alt'} as any)).toBe('/comic/pages/aicut/index?projectId=cut&timelineId=alt&tab=exports');
 expect(taskDestination({source:'novel',id:'schedule',projectId:'novel',sessionId:'session',taskType:'writing-plan',planKind:'schedules'} as any)).toBe('/comic/pages/novel/index?projectId=novel&sessionId=session&tab=automation&planId=schedule&planKind=schedules');
});
