const LIVE = new Set(['running','queued','scheduled','paused','needs_attention']);
const WEEK = 7 * 86400000;
const safeRead = (storage,key,fallback) => {try{return JSON.parse(storage?.getItem(key) || 'null') ?? fallback;}catch{return fallback;}};
const safeWrite = (storage,key,value) => {try{storage?.setItem(key,JSON.stringify(value));}catch{/* In-memory observation remains available. */}};
const readList = (storage,key) => {const value=safeRead(storage,key,[]);return Array.isArray(value)?value:[];};
const identityKey = identity => identity?.owner && identity?.session ? `${identity.owner}\0${identity.session}` : '';

export function createBatchCompletionObserver({initial=[],save=()=>{},now=Date.now}={}){
  const armed=new Map(initial.filter(item=>item?.key && now()-item.at<WEEK).map(item=>[item.key,item.at]));
  return {observe(jobs){
    const notices=[];let changed=false;
    for(const job of jobs || []){
      if(!job?.id || !job.batchId)continue;
      const key=`${job.batchId}:${job.id}`;
      if(LIVE.has(job.state)){if(!armed.has(key)){armed.set(key,now());changed=true;}}
      else if(job.state==='completed' && job.success===true && armed.has(key)){
        armed.delete(key);changed=true;notices.push(job);
      } else if(['failed','stopped','cancelled'].includes(job.state) && armed.delete(key)){changed=true;}
    }
    for(const [key,at]of armed){if(now()-at>WEEK){armed.delete(key);changed=true;}}
    if(changed)save([...armed].slice(-200).map(([key,at])=>({key,at})));
    return notices;
  }};
}

// Independent of React routes: navigation and fullscreen editors do not unmount
// this monitor. All requests and delayed callbacks belong to one login session.
export function createBatchCompletionMonitor({getIdentity,fetchJobs,fetchPreferences,storage,sessionStorage,locks,notify,play,now=Date.now}){
  let scope='',owner='',observer=null,prefs=null,busy=false,disposed=false,unauthorizedScope='',preferenceRevision=0;
  const pending=new Map(),seen=new Map();
  const valid=expected=>!disposed && expected===identityKey(getIdentity());
  function setPreferences(value){preferenceRevision++;prefs={soundEnabled:value?.soundEnabled!==false,soundVolume:Number.isFinite(value?.soundVolume)?Math.min(100,Math.max(0,value.soundVolume)):60};}
  const savePending=()=>safeWrite(sessionStorage,`qiantie:batch-completion-pending:${owner}`,[...pending.values()].slice(-200).map(job=>({job,at:now()})));
  async function deliver(job,expected){
    const key=job.completionKey || job.id;
    const run=async()=>{
      if(!valid(expected))return false;
      const ledgerKey=`qiantie:batch-completion:${owner}`;
      const ledger=()=>readList(storage,ledgerKey).filter(item=>item?.key && Number.isFinite(item.at) && now()-item.at<WEEK);
      const stored=ledger().find(item=>item.key===key),local=seen.get(key);
      const record={key,at:stored?.at || local?.at || now(),noticed:Boolean(stored?.noticed || local?.noticed),sounded:Boolean(stored?.sounded || local?.sounded)};
      const save=()=>{seen.set(key,{...record});safeWrite(storage,ledgerKey,[...ledger().filter(item=>item.key!==key),record].slice(-200));};
      if(record.sounded)return true;
      if(!record.noticed){record.noticed=true;save();notify(job);}
      if(!prefs.soundEnabled || prefs.soundVolume===0){record.sounded=true;save();return true;}
      let sounded=false;
      try{sounded=await play({...prefs},{isCurrent:()=>valid(expected),getPreferences:()=>prefs})===true;}catch{/* An unlocked tab may retry the pending sound. */}
      if(!valid(expected))return false;
      record.sounded=sounded || !prefs.soundEnabled || prefs.soundVolume===0;
      save();return record.sounded;
    };
    if(locks?.request)return locks.request(`batch-completion:${owner}:${key}`,run);
    return run();
  }
  return {
    setPreferences,
    hasPending(){return pending.size>0;},
    dispose(){disposed=true;pending.clear();},
    async poll(){
      if(disposed || busy)return;
      const identity=getIdentity(),expected=identityKey(identity);
      if(!expected){scope='';prefs=null;pending.clear();return;}
      if(expected===unauthorizedScope)return;
      if(scope!==expected){
        scope=expected;owner=identity.owner;prefs=null;pending.clear();seen.clear();
        const key=`qiantie:batch-completion-armed:${owner}`;
        observer=createBatchCompletionObserver({initial:readList(sessionStorage,key),save:value=>safeWrite(sessionStorage,key,value),now});
        for(const item of readList(sessionStorage,`qiantie:batch-completion-pending:${owner}`)){if(item?.job?.completionKey && now()-item.at<WEEK)pending.set(item.job.completionKey,item.job);}
      }
      busy=true;
      try{
        const response=await fetchJobs();if(!valid(expected))return;
        for(const job of observer.observe(Array.isArray(response)?response:response?.jobs || []))pending.set(job.completionKey || job.id,job);
        savePending();
        if(!prefs){const revision=preferenceRevision,config=await fetchPreferences();if(!valid(expected))return;if(revision===preferenceRevision)setPreferences(config);}
        for(const [key,job]of pending){if(!valid(expected))return;if(await deliver(job,expected)){pending.delete(key);savePending();}}
      }catch(error){if(error?.status===401)unauthorizedScope=expected;}
      finally{busy=false;}
    }
  };
}
