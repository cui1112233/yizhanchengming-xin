import {expect,test} from 'vitest';
import {moduleRoute,MODULES} from '../config/modules';
import {taskSourceForModule} from '../pages/workspace/workbenchState';
import {fetchTaskOverallStatus,normalizeFetchTask,fetchWorkspaceState} from './fetchWorkspace';
import {fetchDetailBody} from './fetchDetailBody';
test('book testing opens fetch processing while factory tasks keep the existing batch route',()=>{
 expect(MODULES.find(item=>item.label==='批量测书')?.key).toBe('booktest');
 expect(moduleRoute('booktest')).toBe('/comic/pages/module-detail/index?key=booktest');
 expect(moduleRoute('batch')).toBe('/pages/batch/index');
 expect(taskSourceForModule('booktest')).toBe('fetch');
 expect(taskSourceForModule('batch')).toBe('batch');
});
test('body reading preserves complete raw text and server counts independently of the processed window',()=>{
 expect(fetchDetailBody({original:'处理正文',original_raw:'完整原文',processed_chars:4000,original_raw_chars:27831})).toEqual({processed:'处理正文',raw:'完整原文',processedCount:4000,rawCount:27831});
 expect(fetchDetailBody({})).toEqual({processed:'',raw:'',processedCount:null,rawCount:null});
});
test('input ready and running retain the real classification and execution meaning',()=>{
 expect(fetchTaskOverallStatus(normalizeFetchTask({book_id:'1',status:'input_ready'}))).toBe('分类信息已就绪');
 expect(fetchTaskOverallStatus(normalizeFetchTask({book_id:'1',status:'running',original_status:'done'}))).toBe('正在执行中…');
});
test('disabled models cannot become a default processing model',()=>{
 expect(fetchWorkspaceState({app_config:{text_model_id:'disabled'}},[{id:'disabled',enabled:false},{id:'real',name:'可用'}]).textModelId).toBe('real');
});
test('escaped platform labels are readable while ids are preserved',()=>{
 expect(fetchWorkspaceState({platforms:[{id:1,name:'u9ed1u5ca9u4ed8u8d39'},{id:2,name:'番茄付费'}]},[]).platforms).toEqual([{id:'1',name:'黑岩付费'},{id:'2',name:'番茄付费'}]);
});
