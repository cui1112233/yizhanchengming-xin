export function filterVoiceSwitchLibrary(rows,query='') {
  const text=query.trim().toLowerCase();
  return rows.filter(item=>!text||[item.name,...(Array.isArray(item.tags)?item.tags:[])].join(' ').toLowerCase().includes(text)).slice(0,40);
}
export function projectVoiceLibrary(items,projectId,projectTitle='') {
  return items.filter(item=>item.kind==='audio'&&item.voiceId&&['minimax','qwen-tts'].includes(item.voiceEngine)&&['succeeded','ready','completed'].includes(item.status)).map(item=>({project_id:projectId,project_title:projectTitle,media_item_id:item.id,name:item.name||item.voiceId,voice_id:item.voiceId,engine:item.voiceEngine,target_model:item.voiceTargetModel,url:item.url,tags:[voiceEngineLabel(item.voiceEngine)]}));
}
export function voiceEngineLabel(engine) {return ({minimax:'MiniMax','qwen-tts':'Qwen',custom:'参考音频'})[engine]||engine||'参考音频';}
export function voiceBindingPreview(media,targetId) {
  const binding=media?.settings?.voiceBindings?.[targetId];if(!binding)return null;
  const item=media.items?.find(item=>item.id===binding.media_item_id&&item.kind==='audio');
  const reference=binding.reference_only===true||binding.engine==='custom'||!binding.voice_id;
  return {binding,item,reference,label:reference?'参考音频':voiceEngineLabel(binding.engine),voiceId:reference?'':binding.voice_id};
}
export function validateVoiceReferenceFile(file) {
  if(!file||!Number(file.size))return '音频文件为空，请重新选择';
  if(!/\.(wav|mp3)$/i.test(file.name||''))return '只支持WAV或MP3音频';
  if(file.size>32*1024*1024)return '参考音频不能超过32MB';
  return '';
}
