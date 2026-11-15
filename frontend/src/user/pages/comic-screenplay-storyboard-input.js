import {assertUniqueEpisodeNumbers} from './comic-screenplay-episode-headings.js';
import {comicScriptEpisodes} from './comic-screenplay-stages.js';
export function storyboardEpisodeTargets(document,chapters=[],source='script'){
 if(source!=='original')return comicScriptEpisodes(document);
 return chapters.map((c,i)=>({no:i+1,title:c.title||'第'+(i+1)+'章',chapterId:c.id,content:String(c.source_text||''),sourceOnly:true})).filter(ep=>ep.content.trim());
}
export function buildStoryboardInput(workspace,episodeNo,options={},engine){
 const project=workspace?.project||{};
 if(!String(project.model_id||'').trim())throw Error('请先在页头选择并保存文本模型');
 const payload={engine,style:project.art_style||'',synopsis:project.synopsis||'',model:project.model_id};
 const no=Number(episodeNo);
 if(options.useOriginalContentDirectly&&options.chapterId){
  if(!Number.isInteger(no)||no<1)throw Error('原文分镜目标集无效');
  // The server reads the saved chapter. Never send a fabricated script or whole-book fallback.
  return {...payload,episodeNo:no};
 }
 const episodes=comicScriptEpisodes(workspace?.document?.content);
 if(!episodes.length)throw Error('请先保存剧本，或切换到章节原文分镜');
 assertUniqueEpisodeNumbers(episodes);
 const inputContent=ep=>ep.prefix?`${ep.prefix}\n\n${ep.content}`:ep.content;
 if(no===0)return {...payload,episodes:Object.fromEntries(episodes.map(ep=>[ep.no,inputContent(ep)]))};
 const selected=episodes.find(ep=>ep.no===no);if(!selected)throw Error('所选剧本集不存在，请刷新后重新选择');
 if(!selected.content.trim())throw Error('所选集剧本内容为空');return {...payload,episodeNo:no,content:inputContent(selected)};
}
export async function createRawStoryboardChapter(projectId,{chapterId,readChapters,readWorkspace,saveChapters}){
 if(!chapterId)throw Error('缺少原文章节编号');
 const [workspace,data]=await Promise.all([readWorkspace(projectId),readChapters(projectId)]);
 const source=String(workspace?.project?.source_text||'');if(!source.trim())throw Error('没有已保存原文，请先输入并保存原文');
 if(!Array.isArray(data?.state?.chapters))throw Error('章节读取失败，尚未保存');
 const existing=data.state.chapters.find(c=>c.id===chapterId);
 if(existing){if(existing.source_text===source)return data;throw Error('原文已变化，请在章节工具核对已保存章节');}
 if(data.state.chapters.length)throw Error('已有章节，请使用章节工具管理，不会覆盖');
 const chapter={id:chapterId,title:workspace.project.title||'原文章节',source_text:source,script_text:''};
 try{return await saveChapters(projectId,{expected_revision:data.state.revision,chapters:[chapter]});}
 catch(error){try{const latest=await readChapters(projectId);if(latest.state?.chapters?.some(c=>c.id===chapterId&&c.source_text===source))return latest;}catch{}throw error;}
}
