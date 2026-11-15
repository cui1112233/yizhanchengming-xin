import { Alert, Button, Empty, Input, Select, Spin, Switch } from 'antd';
import { useEffect, useState } from 'react';
import { listAvailableModels } from '../../shared/api/modelCatalog.js';
import { creativePromptOptions } from './comic-screenplay-creative.js';
import {getComicMedia} from '../../shared/api/comicScreenplayWorkflow.js';
import {availableComicImageModels,comicImageModelFamily,comicImageModelGroups,comicImageModelChange} from './comic-screenplay-model-options.js';

const fallbackOptions={aspect_ratios:['9:16','16:9','1:1','3:4','4:3','2:3','3:2','21:9','4:5','5:4'],image_resolutions:['1K','2K','4K'],asset_prompt_types:['general','western']};
export default function ComicScreenplayModelConfig({projectId,value,onChange,library,options,disabled,loading,error,onRetry}) {
  const [models,setModels]=useState([]),[modelError,setModelError]=useState(''),[attempt,setAttempt]=useState(0),[modelsLoading,setModelsLoading]=useState(true),[genChannel,setGenChannel]=useState('all'),[editChannel,setEditChannel]=useState('all');
  useEffect(()=>{let alive=true;setModelsLoading(true);setModelError('');listAvailableModels('image').then(rows=>{if(alive)setModels(rows);}).catch(error=>{if(alive)setModelError(error.message||'图片模型读取失败');}).finally(()=>{if(alive)setModelsLoading(false);});return()=>{alive=false;};},[attempt]);
  const [voiceModels,setVoiceModels]=useState([]),[voiceError,setVoiceError]=useState(''),[voiceLoading,setVoiceLoading]=useState(!!projectId);
  useEffect(()=>{let alive=true;setVoiceModels([]);setVoiceError('');if(!projectId){setVoiceLoading(false);return()=>{alive=false;};}setVoiceLoading(true);getComicMedia(projectId).then(data=>{if(alive){if(!Array.isArray(data?.voiceDesignModels))throw Error('音色设计服务目录暂不可用');setVoiceModels(data.voiceDesignModels);}}).catch(error=>{if(alive)setVoiceError(error.message||'音色设计服务读取失败');}).finally(()=>{if(alive)setVoiceLoading(false);});return()=>{alive=false;};},[projectId,attempt]);
  const enums={...fallbackOptions,...options};
  const generationModels=availableComicImageModels(models),editModels=availableComicImageModels(models,'edit');
  const generationGroups=comicImageModelGroups(generationModels),editGroups=comicImageModelGroups(editModels);
  useEffect(()=>{if(!generationGroups.some(group=>group.key===genChannel))setGenChannel('all');if(!editGroups.some(group=>group.key===editChannel))setEditChannel('all');},[models]);
  const voiceRows=voiceModels.map(model=>({value:model.id,label:model.label}));
  const voiceSelectedAvailable=voiceRows.some(row=>row.value===value.voice_design_model);
  const savedVoice=value.voice_design_model;
  if(savedVoice&&!voiceSelectedAvailable)voiceRows.unshift({value:savedVoice,label:`${savedVoice==='minimax'?'MiniMax':savedVoice==='qwen-tts'?'Qwen':savedVoice}（未配置）`,disabled:true});
  function chips(kind,selected) {const groups=kind==='gen'?generationGroups:editGroups;return <div className="comic-config-channel-chips" role="group" aria-label={`${kind==='gen'?'生图':'改图'}模型分类`}>{groups.map(group=><button type="button" key={group.key} disabled={disabled||modelsLoading} aria-pressed={selected===group.key} className={selected===group.key?'is-active':''} onClick={()=>{const key=selected===group.key||group.key==='all'?'all':group.key;if(kind==='gen'){setGenChannel(key);setEditChannel(editGroups.some(group=>group.key===key)?key:'all');}else setEditChannel(key);}}>{group.label} · {group.count}</button>)}</div>;}
  function modelSelect(field,channel,label) {
    const eligible=field==='image_model_id'?generationModels:editModels;
    const filtered=eligible.filter(model=>channel==='all'||comicImageModelFamily(model).key===channel);const rows=filtered.map(model=>({value:model.id,label:model.displayName||model.name||model.modelId||model.id}));
    if(value[field]&&!eligible.some(model=>model.id===value[field]))rows.unshift({value:value[field],label:`已保存模型（${value[field]}，当前不可用或不支持此用途）`,disabled:true});
    return <><Select aria-label={label} size="small" showSearch optionFilterProp="label" disabled={disabled||modelsLoading||!!modelError} loading={modelsLoading} value={value[field]||undefined} placeholder={modelsLoading?'正在读取模型':eligible.length?`选择${label}`:`尚未配置可用的${label}`} options={rows} labelRender={item=>eligible.find(model=>model.id===item.value)?.displayName||eligible.find(model=>model.id===item.value)?.name||item.label} notFoundContent={<Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="当前没有匹配的已配置模型"/>} onChange={id=>onChange(comicImageModelChange(field,id,models))}/>{!modelsLoading&&!modelError&&!filtered.length?<small className="comic-config-empty">当前没有匹配的{label}，请到设置中的模型管理配置。</small>:null}</>;
  }
  function templateOptions(purpose,field,idField) {const choices=creativePromptOptions(library.prompts,purpose,value[idField]).map(item=>({...item,value:`typedprompt:${item.value}`}));if(purpose==='scene_design')choices.unshift({value:'none',label:'无（不使用模板）'});const current=value[field]||`typedprompt:${value[idField]||''}`;if(current&&!choices.some(item=>item.value===current))choices.unshift({value:current,label:`已保存模式（${current}）`});return choices;}
  function chooseTemplate(field,idField,mode){onChange({[field]:mode,...(mode.startsWith('typedprompt:')?{[idField]:mode.slice(12)}:{})});}
  function templateDescription(purpose,mode,id) {
    const name=library.prompts.find(item=>item.id===id)?.name;
    if(mode?.startsWith('typedprompt:')||mode?.startsWith('prompt:')||!mode)return name?`${purpose==='character'?'人设设计':'场景设计'}：使用提示词库「${name}」作为${purpose==='character'?'角色':'场景'}生图模板`:'已保存模板当前不在目录中，请重新读取或选择。';
    if(purpose==='scene')return '无：直接使用场景描述，不叠加场景出图模板。';
    return ({three_view:'三视图：角色生图和角色改图都走三视图设定提示词',three_view_realistic_sketch:'三视图真人写实-线稿过真人：左侧为线稿风格面部特写，右侧为真人写实三视图',all_around:'全方位：角色生图和角色改图都走三排多角度展示提示词',single:'单人照：角色生图和角色改图都走单人全身图提示词'})[mode]||'已保存的兼容模式；可重新选择库中的角色设定模板。';
  }
  const locked=disabled||loading||!!error;
  return <div className="comic-config-fields">{error?<Alert type="error" message={error} action={<Button size="small" onClick={onRetry}>重新读取</Button>}/>:null}<Spin spinning={loading}><div className="comic-model-config-columns">
    <section className="comic-config-column"><header><h3>剧本与资产策略</h3><p>控制剧本比例和资产解析类型</p></header>
      <div className="comic-config-control-group"><span>剧本比例</span><div className="comic-config-ratios" role="group" aria-label="剧本比例">{enums.aspect_ratios.map(ratio=>{const[w,h]=ratio.split(':').map(Number),scale=18/Math.max(w,h);return <button key={ratio} type="button" title={({'9:16':'9:16（竖屏）','16:9':'16:9（横屏）','1:1':'1:1（方形）'})[ratio]||ratio} disabled={locked} aria-pressed={value.aspect_ratio===ratio} className={value.aspect_ratio===ratio?'is-active':''} onClick={()=>onChange({aspect_ratio:ratio})}><i style={{width:Math.max(7,Math.round(w*scale)),height:Math.max(7,Math.round(h*scale))}}/><span>{ratio}</span></button>;})}</div></div>
      <div className="comic-config-control-group"><span>出图分辨率</span><div className="comic-config-segments" role="group" aria-label="出图分辨率">{enums.image_resolutions.map(resolution=><button key={resolution} type="button" disabled={locked} aria-pressed={value.image_resolution===resolution} className={value.image_resolution===resolution?'is-active':''} onClick={()=>onChange({image_resolution:resolution})}>{resolution}</button>)}</div></div>
      <div className="comic-config-control-group"><span>资产类型</span><div className="comic-config-segments" role="group" aria-label="资产类型">{enums.asset_prompt_types.map(type=><button key={type} type="button" disabled={locked} aria-pressed={value.asset_prompt_type===type} className={value.asset_prompt_type===type?'is-active':''} onClick={()=>onChange({asset_prompt_type:type})}>{type==='western'?'海外':'通用'}</button>)}</div></div>
      <p>{value.asset_prompt_type==='western'?'海外：角色/场景/道具解析走海外版提示词。':'通用：角色/场景/道具解析走通用提示词。'} 剧本比例与出图分辨率保存后，下次打开沿用。</p>
    </section>
    <section className="comic-config-column"><header><h3>图像与音色模型</h3><p>与设置中的模型管理共用目录，仅显示本账号可用模型</p></header>
      <div className="comic-config-control-group"><span>生图模型</span>{chips('gen',genChannel)}{modelSelect('image_model_id',genChannel,'生图模型')}</div>
      <div className="comic-config-control-group"><span>改图模型</span>{chips('edit',editChannel)}{modelSelect('image_edit_model_id',editChannel,'改图模型')}</div>
      {modelError?<Alert type="warning" message={modelError} action={<Button size="small" onClick={()=>setAttempt(n=>n+1)}>重试模型列表</Button>}/>:null}
      <label><span>生图核心指令</span><Input.TextArea aria-label="生图核心指令" size="small" disabled={locked} autoSize={{minRows:2,maxRows:6}} value={value.image_core_instruction??''} onChange={event=>onChange({image_core_instruction:event.target.value})} placeholder="可自定义；清空即不叠加"/><small>仅叠加到人物设定图；场景、道具、站位和分镜使用各自提示词，避免混入角色三视图。可改可删。</small></label>
      <label><span>资产处理</span><Input.TextArea aria-label="资产处理" size="small" disabled={locked} autoSize={{minRows:2,maxRows:6}} value={value.asset_process_prompt??''} onChange={event=>onChange({asset_process_prompt:event.target.value})} placeholder="填写资产处理要求"/><small>仅用于「资产处理」图生图，不会自动叠加到普通生图。保存后沿用。</small></label>
      <label><span>音色设计模型</span><Select aria-label="音色设计模型" size="small" disabled={locked||voiceLoading||!!voiceError||!voiceModels.length} loading={voiceLoading} value={voiceSelectedAvailable?value.voice_design_model:undefined} placeholder={voiceLoading?'正在读取音色服务':voiceError?'音色服务读取失败':voiceModels.length?'选择已配置音色设计服务':'尚未配置音色设计服务'} options={voiceRows} onChange={id=>onChange({voice_design_model:id})}/><small>{voiceError||(!voiceLoading&&!voiceModels.length?'当前仅显示已配置的音色设计服务。普通配音 TTS 不能代替音色设计。':'使用本账号已配置的音色设计服务，不自动更换已保存选择。')}</small></label>
    </section>
    <section className="comic-config-column"><header><h3>出图模板与显示</h3><p>选择角色/场景出图模板和界面显示</p></header>
      <label><span>角色出图</span><Select aria-label="角色出图" size="small" disabled={locked} showSearch optionFilterProp="label" value={value.character_image_mode||`typedprompt:${value.character_design_prompt_id||''}`} options={templateOptions('character_design','character_image_mode','character_design_prompt_id')} onChange={mode=>chooseTemplate('character_image_mode','character_design_prompt_id',mode)}/><small>{templateDescription('character',value.character_image_mode,value.character_design_prompt_id)}</small></label>
      <label><span>场景出图</span><Select aria-label="场景出图" size="small" disabled={locked} showSearch optionFilterProp="label" value={value.scene_image_mode||`typedprompt:${value.scene_design_prompt_id||''}`} options={templateOptions('scene_design','scene_image_mode','scene_design_prompt_id')} onChange={mode=>chooseTemplate('scene_image_mode','scene_design_prompt_id',mode)}/><small>{templateDescription('scene',value.scene_image_mode,value.scene_design_prompt_id)}</small></label>
      <label className="comic-config-display"><Switch aria-label="显示原文" disabled={locked} checked={value.show_original_text!==false} onChange={checked=>onChange({show_original_text:checked})}/><span>{value.show_original_text!==false?'显示原文':'隐藏原文'}</span></label>
    </section>
  </div></Spin></div>;
}
