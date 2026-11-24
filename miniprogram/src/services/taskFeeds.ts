type Row=Record<string,any>;
type Request=(path:string)=>Promise<unknown>;
type Feed={rows:Row[];warning:string};
const rows=(value:any,key:string):Row[]=>{if(!Array.isArray(value?.[key]))throw Error('服务端任务列表不完整');return value[key];};
const id=(value:any)=>{if(value===undefined||value===null||!String(value).trim())throw Error('作品编号缺失');return encodeURIComponent(String(value));};
const message=(e:any)=>e?.message||'读取失败';
export function limitTaskRequests(request:Request,limit=4):Request{
 let active=0;const queue:Array<()=>void>=[];
 return async(path)=>{if(active>=limit)await new Promise<void>(resolve=>queue.push(resolve));else active++;try{return await request(path);}finally{const next=queue.shift();if(next)next();else active--;}};
}
async function collect(projects:Row[],work:(project:Row)=>Promise<Feed>):Promise<Feed>{const results=await Promise.all(projects.map(async p=>{try{return await work(p);}catch(e){return{rows:[],warning:`${p.title||p.name||p.id}：${message(e)}`};}}));return{rows:results.flatMap(r=>r.rows),warning:results.map(r=>r.warning).filter(Boolean).join('；')};}
export async function loadCmTaskFeed(request:Request,projects:Row[]):Promise<Feed>{return collect(projects,async p=>{
 const base=`/api/agent/tasks/${id(p.id)}`;
 const data=await Promise.allSettled([request(`${base}/generations`).then(r=>rows(r,'tasks')),request(`${base}/video-jobs`).then(r=>rows(r,'jobs'))]);
 const tasks=data[0].status==='fulfilled'?data[0].value:[],jobs=data[1].status==='fulfilled'?data[1].value:[];
 const records:Row[]=[...tasks.filter(t=>!jobs.some(j=>j.id===t.id)),...jobs.map(j=>({...j,kind:'video'}))];
 return{rows:records.map(t=>({...t,projectId:String(p.id),title:t.title||`${p.title||'CM项目'} · ${({image:'图片生成',video:'视频生成',audio:'配音生成',text:'文本生成'} as Record<string,string>)[t.kind]||'创作任务'}`,stage:t.mode,taskType:t.kind})),warning:data.flatMap(v=>v.status==='rejected'?[`${p.title||p.id}：${message(v.reason)}`]:[]).join('；')};
 });}
export async function loadCutTaskFeed(request:Request,projects:Row[]):Promise<Feed>{return collect(projects,async p=>({rows:rows(await request(`/api/ai-cut/projects/${id(p.id)}/export-jobs`),'jobs').map(j=>({...j,projectId:String(p.id),timelineId:j.timelineId||j.options?.timelineId||'default',title:`${p.title||'AI-CUT工程'} · 成片导出`,kind:'export',taskType:'ai-cut',...(typeof j.progress==='number'&&Number.isFinite(j.progress)&&j.progress>=0&&j.progress<=100?{completed:j.progress,total:100}:{})})),warning:''}));}
export async function loadWaterTaskFeed(request:Request,projects:Row[]):Promise<Feed>{return collect(projects,async p=>({rows:rows(await request(`/api/shuihuo-production/projects/${id(p.id)}/tasks`),'tasks').map(t=>({...t,projectId:String(p.id),title:t.title||t.name||`${p.name||p.id} · ${t.kind||'任务'}`})),warning:''}));}
export async function loadNovelTaskFeed(request:Request,projects:Row[]):Promise<Feed>{return collect(projects.filter(p=>p.status!=='purging'),async p=>{
 const base=`/api/novel-creation/projects/${id(p.id)}`;
 const data=await Promise.allSettled(['sessions','schedules','batches'].map(k=>request(`${base}/${k}`).then(r=>rows(r,k))));
 const sessions=data[0].status==='fulfilled'?data[0].value:[],schedules=data[1].status==='fulfilled'?data[1].value:[],batches=data[2].status==='fulfilled'?data[2].value:[];
 const runIds=[...new Set([...sessions.map(s=>s.currentRunId),...[...schedules,...batches].flatMap(v=>[v.currentRunId,...(v.rounds||[]).map((r:Row)=>r.runId)])].filter(Boolean))];
 const runs=await Promise.allSettled(runIds.map(async runId=>{const r=await request(`/api/novel-creation/runs/${id(runId)}`) as Row;if(r?.run?.id!==runId||String(r.run.projectId)!==String(p.id))throw Error('小说任务作品归属不匹配');return r.run as Row;}));
 const records:Row[]=runs.flatMap(v=>v.status==='fulfilled'?[{...v.value,projectId:String(p.id),title:`${p.title||'小说作品'} · 章节写作`,kind:'novel',taskType:'writing'}]:[]);
 const warnings=[...data,...runs].flatMap(v=>v.status==='rejected'?[`${p.title||p.id}：${message(v.reason)}`]:[]);
 // An upcoming schedule is a real plan, explicitly labelled separately from its running round.
 for(const plan of [...schedules,...batches])if(!plan.currentRunId&&!plan.rounds?.length&&['waiting','not_started','running','paused','failed'].includes(plan.status))records.push({...plan,projectId:String(p.id),title:`${p.title||'小说作品'} · ${schedules.includes(plan)?'定时写作计划':'连续写作计划'}`,kind:'novel',taskType:'writing-plan',planKind:schedules.includes(plan)?'schedules':'batches'});
 return{rows:records,warning:warnings.join('；')};
 });}
