import test from 'node:test';
import assert from 'node:assert/strict';
import {storyboardEpisodeTargets,buildStoryboardInput,createRawStoryboardChapter} from './comic-screenplay-storyboard-input.js';
const chapters=[{id:'a',title:'空章',source_text:''},{id:'b',title:'第二章',source_text:'第二章原文',script_text:''}];
test('原文章节不需要软件剧本，编号不因空章节过滤而变成第一集',()=>{
 const targets=storyboardEpisodeTargets('',chapters,'original');assert.equal(targets.length,1);assert.equal(targets[0].no,2);assert.equal(targets[0].chapterId,'b');assert.equal(targets[0].content,'第二章原文');assert.equal(targets[0].sourceOnly,true);
 assert.deepEqual(storyboardEpisodeTargets('',chapters,'script'),[]);
});
test('原文单章输入始终由后端按真实chapterId读取，保留第二集并不创建假剧本',()=>{
 const workspace={project:{id:'p',model_id:'own',source_text:'全书原文'}};
 const body=buildStoryboardInput(workspace,2,{chapterId:'b',useOriginalContentDirectly:true},'xiaoluo_v9_30s');
 assert.equal(body.episodeNo,2);assert.equal(body.content,undefined);assert.equal(body.episodes,undefined);assert.equal(body.engine,'xiaoluo_v9_30s');
 assert.throws(()=>buildStoryboardInput(workspace,0,{chapterId:'b',useOriginalContentDirectly:true},'e'),/目标/);
});
test('普通剧本按选中集保持完整，空内容/不存在集数拒绝而不静默改第一集',()=>{
 const w={project:{model_id:'own'},document:{content:'# 第1集 A\n第一集完整。\n# 第2集 B\n第二集完整。'}};
 assert.equal(buildStoryboardInput(w,2,{},'e').content,'第二集完整。');assert.throws(()=>buildStoryboardInput(w,3,{},'e'),/不存在/);
 assert.equal(Object.keys(buildStoryboardInput(w,0,{},'e').episodes).length,2);
 assert.throws(()=>buildStoryboardInput({project:{model_id:'own'}},1,{},'e'),/剧本/);
});

test('建立原文章节：空原文与已有章节零写入，回包丢失核对同ID，重复操作不复制',async()=>{
 let rows=[],writes=0;const deps={chapterId:'stable',readWorkspace:async()=>({project:{title:'故事',source_text:'  整篇原文。\n下一段。'}}),readChapters:async()=>({state:{revision:rows.length,chapters:rows}}),saveChapters:async(id,input)=>{writes++;rows=input.chapters;throw Error('回包丢失');}};
 const first=await createRawStoryboardChapter('p',deps);assert.equal(first.state.chapters[0].script_text,'');assert.equal(first.state.chapters[0].source_text,'  整篇原文。\n下一段。');
 await createRawStoryboardChapter('p',deps);assert.equal(writes,1);
 await assert.rejects(createRawStoryboardChapter('p',{...deps,chapterId:'other'}),/已有章节/);assert.equal(writes,1);
 await assert.rejects(createRawStoryboardChapter('p',{...deps,readWorkspace:async()=>({project:{source_text:' '}})}),/没有已保存原文/);assert.equal(writes,1);
});
