import type {CmRow} from './cmCanvas';
import type {CmIntent} from './cmOperationRecovery';
type Request=(path:string,options?:any)=>Promise<any>;
const id=(value:any)=>{const s=String(value??'').trim();if(!s)throw Error('缺少项目或操作编号');return encodeURIComponent(s);};
const path=(project:string)=>`/api/agent/tasks/${id(project)}`;
export function createCmAdvanced(request:Request){return{
 approve:(project:string,plan:string,revision:number)=>request(`${path(project)}/director/plans/${id(plan)}/approve`,{method:'POST',data:{revision}}),
 shots:async(project:string,node:string)=>{const r=await request(`${path(project)}/script-nodes/${id(node)}/shots`);if(!Array.isArray(r?.shots))throw Error('镜头列表读取失败');return r.shots as CmRow[];},
 split:async(project:string,node:CmRow,modelId:string)=>{const r=await request(`${path(project)}/script-splits`,{method:'POST',data:{nodeId:node.id,nodeVersion:node.version,modelId},timeout:180000});if(!r?.generation?.id)throw Error('拆镜提交结果未知，请核对任务');return r;},
 saveShot:async(project:string,shot:CmRow)=>{const data:CmRow={scriptNodeId:shot.scriptNodeId,revision:shot.revision};for(const k of ['sequence','scene','characters','dialogue','narration','imagePrompt','videoPrompt','ttsText'])if(shot[k]!==undefined)data[k]=shot[k];const r=await request(`${path(project)}/shots/${id(shot.id)}`,{method:'PATCH',data});if(!r?.shot?.id)throw Error('镜头保存结果未确认');return r.shot as CmRow;},
 batch:(project:string,node:string,shots:string[],model:CmRow,requestId:string)=>{if(!shots.length||!['image','video','audio'].includes(model.kind))throw Error('请选择镜头与对应媒体模型');return request(`${path(project)}/shots/batch-nodes`,{method:'POST',data:{scriptNodeId:node,shotIds:[...new Set(shots)],kind:model.kind,provider:model.provider||'',modelId:model.id,requestId},timeout:180000});},
 merge:async(project:string,node:string,seconds:number)=>{if(![10,15,30].includes(seconds))throw Error('请选择10、15或30秒');const r=await request(`${path(project)}/shots/merge`,{method:'POST',data:{scriptNodeId:node,targetDurationSeconds:seconds}});if(!Array.isArray(r?.shots))throw Error('镜头合并结果未确认');return r.shots as CmRow[];},
 skill:async(skill:string)=>{const r=await request(`/api/agent/skills/mine/${id(skill)}`);if(!r?.skill?.id)throw Error('个人技能读取失败');return r.skill as CmRow;},
 saveSkill:async(skill:CmRow,creating=false)=>{const data:CmRow={...(creating?{id:skill.id}:{}),name:String(skill.name||'').trim(),description:skill.description||'',category:skill.category||'',inputTemplate:skill.inputTemplate||'',body:skill.body||''};if(!data.name||!String(data.body).trim())throw Error('请填写技能名称和内容');const r=await request(`/api/agent/skills/mine${creating?'':`/${id(skill.id)}`}`,{method:creating?'POST':'PUT',data});if(!r?.skill?.id)throw Error('技能保存结果未确认，请刷新技能列表核对');return r.skill as CmRow;},
 removeSkill:(skill:string)=>request(`/api/agent/skills/mine/${id(skill)}`,{method:'DELETE'}),
};}
export type CmOperation=(label:string,work:()=>Promise<void>,confirmation?:string,key?:string,intent?:CmIntent)=>Promise<boolean>;
