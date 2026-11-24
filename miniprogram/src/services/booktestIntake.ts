import Taro from '@tarojs/taro';
import type {FetchRequest} from './booktestManagement';
import type {NovelJournal} from './novelRecovery';
const utf8=(s:string)=>Uint8Array.from((encodeURIComponent(s).match(/%[A-F0-9]{2}|./g)||[]).map(c=>c[0]==='%'?parseInt(c.slice(1),16):c.charCodeAt(0)));
export async function prepareFetchIntake(request:FetchRequest,ids:string[],version:string){
 if(!ids.length||!['sensitive','original','ai1','ai2','ai3','ai4','ai5'].includes(version))throw Error('请选择当前批次和正文版本');
 const books:any[]=[];let total=0;
 for(const bookId of [...new Set(ids)]){const detail:any=await request(`/api/novel-fetch-workshop/tasks/${encodeURIComponent(bookId)}`),meta=detail?.meta||{};
 const originalRaw=[detail.originalRaw,detail.rawData].find(v=>typeof v==='string'&&v.trim()&&v!=='[object Object]')||'';
 const sourceText=String(version==='original'?originalRaw:version==='sensitive'?detail.processedText||detail.original||'':detail.versions?.[version]||'').trim();
 if(!sourceText)throw Error(`《${meta.bookName||bookId}》没有 ${version} 正文，请先生成或选择其他版本`);
 const title=String(meta.bookName||detail.bookName||bookId);total+=utf8(sourceText).length+utf8(originalRaw).length;if(total>48*1024*1024)throw Error('本批正文超过48MiB，请缩小当前批次');
 books.push({id:bookId,bookId,sourceTaskId:bookId,title,platform:String(meta.platformId||detail.platformId||''),sourceText,originalRaw,sourceMetadata:{sourceMode:'novel_fetch',sourceContentVersion:version,sourceTaskId:bookId,bookName:title,gender:String(detail.gender||meta.gender||''),genderSource:String(detail.genderSource||meta.genderSource||''),style:String(detail.style||meta.style||''),styleSource:String(detail.styleSource||meta.styleSource||''),processingRuleVersion:String(meta.processingRuleVersion||'novel-fetch-rules-v1'),originalRawChars:detail.originalRawChars||originalRaw.length,processedChars:detail.processedChars||sourceText.length,processingSummary:{stages:meta.rulesTrace||[]}}});}
 return {books,metadata:{sourceMode:'novel_fetch',sourceContentVersion:version,selectedTaskIds:ids}};
}
export async function submitFetchIntake(request:FetchRequest,payload:any,journal:NovelJournal){
 if(journal.read('intake'))throw Error('上次入厂结果待核对，请先到批量工厂查看，勿重复创建');
 const books:any[]=[];
 for(const book of payload.books){const {sourceText,originalRaw,...rest}=book,bytes=utf8(JSON.stringify({sourceText,originalRaw}));if(bytes.length>24*1024*1024||Math.max(utf8(sourceText).length,utf8(originalRaw).length)>16777215)throw Error('单书正文超过暂存容量，请减少截取范围');
 let ref='';const uploadId=Array.from({length:24},()=>Math.floor(Math.random()*256).toString(16).padStart(2,'0')).join('');
 for(let offset=0;offset<bytes.length;offset+=524288){const chunk=bytes.slice(offset,offset+524288);const result:any=await request('/api/batch-factory/v12/intakes/source-chunks',{method:'POST',data:{platform:book.platform,bookId:book.bookId,bytes:bytes.length,index:offset/524288,ref,uploadId,maxFieldBytes:Math.max(utf8(sourceText).length,utf8(originalRaw).length),data:Taro.arrayBufferToBase64(chunk.buffer)}});if(!result?.ref||(ref&&ref!==result.ref))throw Error('正文暂存回执无效，未创建批次');ref=result.ref;}
 books.push({...rest,sourceRef:ref});}
 journal.save('intake',{phase:'intake',ids:payload.metadata.selectedTaskIds});
 const result:any=await request('/api/batch-factory/v12/intakes/novel-fetch',{method:'POST',data:{...payload,books}});const intakeId=result.intake?.id||result.intakeId||result.id;if(!intakeId)throw Error('未取得入厂编号，请核对批量工厂，勿重复提交');
 journal.save('intake',{phase:'batch',intakeId});const created:any=await request(`/api/batch-factory/v12/intakes/${encodeURIComponent(intakeId)}/batches`,{method:'POST',data:{title:'测书入厂',automationEnabled:false}});const batchId=created?.batch?.id;if(!batchId)throw Error('批次创建结果尚未确认，请到批量工厂核对');journal.clear('intake');return {batchId};
}

export async function recoverFetchIntake(request:FetchRequest,journal:NovelJournal){
 const pending=journal.read('intake');if(!pending?.intakeId)throw Error('本机未收到入厂编号，请核对批量工厂后再决定是否放弃此记录');
 const intake:any=await request(`/api/batch-factory/v12/intakes/${encodeURIComponent(pending.intakeId)}`);
 if(intake?.intake?.id!==pending.intakeId)throw Error('入厂记录编号不匹配');
 const locate=async()=>{const r:any=await request('/api/batch-factory/v12/batches');if(!Array.isArray(r.batches))throw Error('批次列表读取失败');return r.batches.find((b:any)=>b.sourceIntakeId===pending.intakeId);};
 const existing=await locate();if(existing?.id){journal.clear('intake');return {batchId:existing.id};}
 if(intake.intake.consumedAt)throw Error('入厂记录已使用，但未定位到关联批次，请保留编号核对');
 try{const r:any=await request(`/api/batch-factory/v12/intakes/${encodeURIComponent(pending.intakeId)}/batches`,{method:'POST',data:{title:'测书入厂',automationEnabled:false}});if(!r?.batch?.id)throw Error('创建结果未确认');journal.clear('intake');return{batchId:r.batch.id};}
 catch(c){const found=await locate();if(found?.id){journal.clear('intake');return{batchId:found.id};}throw c;}
}
