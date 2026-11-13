import assert from 'node:assert/strict';
import test from 'node:test';
import { comicAssetMediaImage, positioningImageFingerprint, mergePositioningRecovery, positioningAttemptState, positioningServerRequestKey, positioningRefreshDecision, comicStages, comicStageLocation, comicScriptEpisodes, comicEpisodesDocument, reorderComicEpisodes, comicNextState } from './comic-screenplay-stages.js';
import { bindComicPositioningImage, editComicPositioningImage, getComicPositioning, saveComicPositioning, generateComicPositioningImage, analyzeComicPositioning } from '../../shared/api/comicScreenplayWorkflow.js';

test('主入口严格三步，旧偏好定位到对应阶段真实工具而不是多出主入口',()=>{
 assert.deepEqual(comicStages.map(item=>item.label),['剧本','资产','分镜']);
 assert.deepEqual(comicStageLocation('chapters'),{stage:'source',tool:'chapters'});
 assert.deepEqual(comicStageLocation('media'),{stage:'storyboard',tool:'media'});
 assert.deepEqual(comicStageLocation('assistant'),{stage:'source',agent:true});
 assert.equal(comicStageLocation('unknown').stage,'source');
});

test('实际剧本只按集切分，编辑与重排保留完整正文和前置文字，不复制原文',()=>{
 const episodes=comicScriptEpisodes('前置概要\n\n# 第1集 相遇\n甲对白。\n动作一。\n\n# 第2集 离开\n乙对白。');
 assert.equal(episodes.length,2);assert.equal(episodes[0].title,'相遇');assert.equal(episodes[1].content,'乙对白。');
 const reordered=reorderComicEpisodes(episodes,0,1);
 assert.equal(reordered[0].title,'离开');assert.equal(reordered[0].prefix,'前置概要');
 const document=comicEpisodesDocument(reordered);
 assert.ok(document.startsWith('前置概要\n\n# 第1集 离开'));
 assert.ok(document.includes('# 第2集 相遇\n\n甲对白。\n动作一。'));
 assert.equal(comicScriptEpisodes('独立的已有剧本。')[0].content,'独立的已有剧本。');
});

test('下一步始终允许查看阶段，但只有真实已存剧本才显示整理完毕，分镜不加假下一步',()=>{
 const originalOnly=comicNextState('source',{project:{source_text:'这只是原文'}});
 assert.equal(originalOnly.next,'assets');assert.equal(originalOnly.ready,false);assert.doesNotMatch(originalOnly.title,/完毕/);
 assert.equal(comicNextState('source',{document:{content:'# 第1集\n剧本对白。'}}).ready,true);
 assert.equal(comicNextState('source',{document:{content:'# 第1集\n剧本对白。'}},true).ready,false);
 assert.equal(comicNextState('source',{document:{content:'# 第1集'}}).ready,false);
 assert.equal(comicNextState('storyboard',{}),null);
});

test('站位真实接口保存独立修订号与顺序，图片请求携幂等键，分析可取消',async()=>{
 const calls=[],request=async(path,options)=>{calls.push({path,options});return {state:{revision:3,items:[]}};};
 await getComicPositioning('p',request);
 await saveComicPositioning('p',{expected_revision:2,items:[{id:'b'},{id:'a'}],prefix_text:'手填要求',image_model:'C',show_character_names:false},request);
 await generateComicPositioningImage('p','a',{expected_revision:3,idempotencyKey:'reuse',allow_missing_refs:true},request);
 const controller=new AbortController();await analyzeComicPositioning('p',{expected_revision:3,model:'text',analyze_mode:'complex'},controller.signal,request);
 assert.deepEqual(calls.map(item=>item.path),['/api/comic-screenplays/projects/p/positioning','/api/comic-screenplays/projects/p/positioning','/api/comic-screenplays/projects/p/positioning/a/image','/api/comic-screenplays/projects/p/positioning/analyze']);
 assert.deepEqual(JSON.parse(calls[1].options.body).items.map(item=>item.id),['b','a']);
 assert.equal(JSON.parse(calls[2].options.body).idempotencyKey,'reuse');assert.equal(calls[3].options.signal,controller.signal);
});


test('站位上传绑定和标注改图走真实独立接口，不发送任意图片URL',async()=>{
 const calls=[],request=async(path,options)=>{calls.push({path,options});return {state:{revision:5,items:[]}};};
 await bindComicPositioningImage('p','row',{expected_revision:3,media_item_id:'uploaded'},request);
 const controller=new AbortController();await editComicPositioningImage('p','row',{expected_revision:4,idempotencyKey:'retry',marked_image_item_id:'marked',edit_prompt:'调整人物站位',signal:controller.signal},request);
 assert.equal(calls[0].path,'/api/comic-screenplays/projects/p/positioning/row/image/upload');
 assert.equal(calls[1].path,'/api/comic-screenplays/projects/p/positioning/row/image/edit');
 assert.equal(calls[1].options.signal,controller.signal);assert.equal(JSON.parse(calls[1].options.body).signal,undefined);
 assert.equal(JSON.parse(calls[1].options.body).marked_image_item_id,'marked');
});

test('站位失败恢复只更新服务端状态与版本，保留输入；只有已确认结果决定新尝试',()=>{
 const draft={revision:1,items:[{id:'a',position_detail:'保留手填要求',scene_name:'教室',status:'',image_path:'old'}]};
 const remote={revision:3,items:[{id:'a',position_detail:'服务器旧要求',status:'failed',request_key:'one',error:'模型失败',image_path:'saved'}]};
 const recovered=mergePositioningRecovery(draft,remote);
 assert.equal(recovered.revision,1);assert.equal(recovered.items[0].position_detail,'保留手填要求');assert.equal(recovered.items[0].status,'failed');
 assert.equal(positioningAttemptState(remote.items[0],'one'),'failed');assert.equal(positioningAttemptState({...remote.items[0],status:'succeeded'},'one'),'succeeded');assert.equal(positioningAttemptState({...remote.items[0],status:'running'},'one'),'running');assert.equal(positioningAttemptState(remote.items[0],'different'),'unknown');
 const before=positioningImageFingerprint(draft.items[0],draft),after=positioningImageFingerprint({...draft.items[0],status:'failed',request_key:'one',image_path:'signed-new-url',error:'failure'},draft);
 assert.equal(before,after);assert.notEqual(before,positioningImageFingerprint({...draft.items[0],position_detail:'修改了要求'},draft));
});

test('资产卡与站位参考遵循选中图片优先，否则最新成功图片，不显示最旧图',()=>{
 const media={items:[{id:'old',kind:'image',status:'succeeded',targetId:'hero',url:'old',createdAt:'2026-01-01'},{id:'new',kind:'image',status:'succeeded',targetId:'hero',url:'new',createdAt:'2026-01-02'},{id:'failed',kind:'image',status:'failed',targetId:'hero',url:'bad',createdAt:'2026-01-03'}],settings:{selected:{}}};
 assert.equal(comicAssetMediaImage(media,'hero').id,'new');media.settings.selected.hero='old';assert.equal(comicAssetMediaImage(media,'hero').id,'old');assert.equal(comicAssetMediaImage(media,undefined),null);
});


test('站位结果核对使用实际后端generate/edit哈希键，而非直接比较客户端UUID',async()=>{
 const generate=await positioningServerRequestKey('generate','a','one');
 const edit=await positioningServerRequestKey('edit','a','one');
 assert.equal(generate,'positioning-d70780bd16d7f8953c156c0210e962f2ca6f0ad05a59ced4e571d1cb9a2850b2');assert.equal(edit,'positioning-2a446568957c7b581e5465a5b3c8c6dd129c9994155f8bac3168f948bd264a97');
 assert.equal(positioningAttemptState({request_key:generate,status:'failed'},generate),'failed');
 assert.equal(positioningAttemptState({request_key:generate,status:'succeeded'},'one'),'unknown');
 assert.notEqual(generate,edit);
});


test('刷新站位无草稿时完整采用远端新增删除，有草稿时不升级可覆盖的修订号',()=>{
 const local={revision:1,items:[{id:'one',position_detail:'本地待保存'}]};
 const remoteAdded={revision:2,items:[{id:'one',position_detail:'服务器原值'},{id:'two',position_detail:'他人新增'}]};
 const clean=positioningRefreshDecision(local,remoteAdded,false,1);
 assert.deepEqual(clean.draft.items.map(item=>item.id),['one','two']);assert.equal(clean.revision,2);assert.equal(clean.conflict,false);
 const dirty=positioningRefreshDecision(local,remoteAdded,true,1);
 assert.equal(dirty.revision,1);assert.equal(dirty.draft.revision,1);assert.equal(dirty.conflict,true);assert.equal(dirty.draft.items[0].position_detail,'本地待保存');
 const deleted={revision:3,items:[]};
 assert.deepEqual(positioningRefreshDecision(local,deleted,false,1).draft.items,[]);
 assert.equal(positioningRefreshDecision(local,deleted,true,1).revision,1);
 assert.equal(positioningRefreshDecision(local,deleted,true,1).conflict,true);
});
