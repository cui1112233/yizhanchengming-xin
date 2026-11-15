import test from 'node:test';
import assert from 'node:assert/strict';
import {storyboardPromptRows,storyboardPromptSections,storyboardPromptDate,normalizeStoryDensity,importStoryboardPrefixes,validateStoryboardPrefix,matchStoryboardChapter,storyboardPlanningAttempt,storyboardGenerationCanRetry,storyboardRestorableGeneration} from './comic-screenplay-storyboard-settings.js';

test('分镜词按用途与名称描述搜索、按时间排序，不搜索或发送正文',()=>{
 const prompts=[{id:'2',purpose:'storyboard',name:'V7',updated_at:'2026-01-01'},{id:'3',purpose:'storyboard',name:'V9',description:'30秒',updated_at:'2026-01-01'},{id:'1',purpose:'script',name:'剧本',content:'30秒'},{id:'4',purpose:'storyboard',name:'新的',updated_at:'2026-02-01'}];
 assert.deepEqual(storyboardPromptRows(prompts,'').map(item=>item.id),['4','3','2']);assert.deepEqual(storyboardPromptRows(prompts,' 30秒 ').map(item=>item.id),['3']);
 assert.equal(storyboardPromptDate({updated_at:'not-a-real-date-value'}),'not-a-real-date-');
 assert.deepEqual(storyboardPromptRows([{id:'x',purpose:'storyboard',name:'标题',description:'说明'}],'标题 说明'),[]);
});

test('收藏与最近仅使用实际目录ID，最近排除收藏并限制8条',()=>{
 const prompts=Array.from({length:12},(_,i)=>({id:String(i),name:String(i),purpose:'storyboard'}));
 const result=storyboardPromptSections(prompts,['0','missing'],['0','1','1','2','3','4','5','6','7','8','9']);
 assert.deepEqual(result.favorites.map(item=>item.id),['0']);assert.deepEqual(result.recent.map(item=>item.id),['1','2','3','4','5','6','7','8']);
});

test('原版搜索同时筛选收藏和最近，在搜索过滤后取最近8项',()=>{
 const prompts=Array.from({length:12},(_,i)=>({id:String(i),name:i===0?'已收藏 Alpha':`最近 ${i}`,description:i>=9?'Alpha':'其他',purpose:'storyboard'}));
 const result=storyboardPromptSections(prompts,['0','1'],['0','1','2','3','4','5','6','7','8','9','10','11'],' alpha ');
 assert.deepEqual(result.favorites.map(item=>item.id),['0']);
 assert.deepEqual(result.recent.map(item=>item.id),['9','10','11']);
 assert.deepEqual(storyboardPromptSections(prompts,['0'],['9'],'没有匹配'),{favorites:[],recent:[]});
});

test('同时间按原编号倒序，不被CM内部ID名称影响；原日期去空白并规范空格',()=>{
 const prompts=[{id:'luoshui_z_137',source_id:137,purpose:'storyboard',updated_at:'2026-11-09 11:39:38'},
  {id:'luoshui_a_310',source_id:310,purpose:'storyboard',updated_at:'2026-11-09T11:39:38'},
  {id:'luoshui_b_202',source_id:202,purpose:'storyboard',updated_at:'2026-11-09 11:39:38'}];
 assert.deepEqual(storyboardPromptRows(prompts).map(item=>item.source_id),[310,202,137]);
 assert.equal(storyboardPromptDate({updated_at:'  2026-11-09 11:39:38  '}),'11-09 11:39');
 assert.equal(storyboardPromptDate({updated_at:'   '}),'');
});

test('规划密度严格四值，detail不是更多小节，布尔不冒充数字',()=>{
 assert.equal(normalizeStoryDensity('more'),'more');assert.equal(normalizeStoryDensity('detail'),'standard');assert.equal(normalizeStoryDensity(true),'standard');assert.equal(normalizeStoryDensity(3),'sd25');assert.equal(normalizeStoryDensity('sd2.5'),'sd25');
});

test('前缀导入只接数组并去重不覆盖，名称正文必填、5000限制，额外字段不进入保存',()=>{
 const old=[{id:'same',name:'原名',content:'原内容'}];
 const result=importStoryboardPrefixes(JSON.stringify([{id:'same',name:'新名',content:'覆盖尝试'},{name:' 新增 ',content:' 内容 ',token:'不保存'},{id:'empty',name:' ',content:'无效'}]),old,()=> 'generated');
 assert.deepEqual(result.items,[...old,{id:'generated',name:'新增',content:'内容'}]);assert.equal(result.added,1);assert.equal(result.duplicates,1);assert.equal(result.invalid,1);
 assert.throws(()=>importStoryboardPrefixes('{}',old),/数组/);assert.throws(()=>validateStoryboardPrefix({name:'a',content:'x'.repeat(5001)}),/5000/);assert.throws(()=>validateStoryboardPrefix({name:' ',content:'x'}),/名称/);
});


test('目标章节只按唯一完整正文匹配，不拿标题数字或数组首项冒充关联；规划状态按真实uuid核对',()=>{
 const chapters=[{id:'wrong',title:'第1集',script_text:'另一段正文'},{id:'right',title:'第9章',script_text:'完整正文',planning_request_key:'uuid',planning_status:'succeeded'}];
 assert.equal(matchStoryboardChapter(chapters,{no:1,content:'完整正文'}),'right');assert.equal(matchStoryboardChapter([...chapters,{id:'duplicate',script_text:'完整正文'}],{content:'完整正文'}),'');assert.equal(storyboardPlanningAttempt(chapters[1],'uuid'),'succeeded');assert.equal(storyboardPlanningAttempt(chapters[1],'other'),'unknown');
});


test('真正生成的attempt独立恢复，running/unknown不准重试，成功后不拿旧失败冒充当前批次',()=>{
 const failed={request_key:'old',created_at:'2026-01-01',episode_numbers:[1],status:'failed',can_retry:true},succeeded={request_key:'new',created_at:'2026-01-02',episode_numbers:[1],status:'succeeded'};
 assert.equal(storyboardGenerationCanRetry(failed),true);for(const status of ['running','unknown','succeeded'])assert.equal(storyboardGenerationCanRetry({...failed,status}),false);
 assert.equal(storyboardRestorableGeneration([failed,succeeded],1),null);assert.equal(storyboardRestorableGeneration([failed],1).request_key,'old');assert.equal(storyboardRestorableGeneration([{...failed,status:'running'},succeeded],1).status,'running');
});
