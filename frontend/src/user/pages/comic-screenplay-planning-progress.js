export function applyPlanningProgress(previous,result,chapterId,requestKey){
 if(result?.chapter_id!==chapterId||result?.request_key!==requestKey)throw Error('规划进度与当前请求不一致');
 if(!['running','succeeded','partial','failed','cancelled'].includes(result.status))throw Error('规划状态无法确认');
 const p=result.progress;if(!p)return previous;
 if(p.type!=='plan'||p.chapter_id!==chapterId||p.request_key!==requestKey||!Number.isInteger(p.sequence)||p.sequence<1||!Number.isInteger(p.round)||p.round<1||p.round>8||p.total!==8||!Number.isInteger(p.chars)||p.chars<0||!Array.isArray(p.missing))throw Error('规划进度返回不完整');
 if(previous&&p.sequence<previous.sequence)return previous;
 return {...p,status:result.status};
}
