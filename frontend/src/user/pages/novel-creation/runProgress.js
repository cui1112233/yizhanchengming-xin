const succeeded=new Set(['succeeded','succeeded_with_warnings']);
const stopped=new Set(['failed','cancelled','paused','awaiting_confirmation']);
export function runProgress(run,events=[]){
 if(succeeded.has(run.status))return {percent:100,status:'success'};
 const usable=events.filter(event=>!/^run\.(failed|cancelled|paused|awaiting_confirmation)$/.test(event.type||''));
 const percent=Math.min(99,Math.max(0,...usable.map(event=>Number(event.progress)||0)));
 const exceptional=['failed','cancelled'].includes(run.status)||run.errorCode==='novel_upstream_result_unknown';
 return {percent,status:exceptional?'exception':stopped.has(run.status)?'normal':'active'};
}
