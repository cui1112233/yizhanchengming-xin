import {Button,Input,Picker,ScrollView,Text,Textarea,View} from '@tarojs/components';
import {useState} from 'react';
import {waterImageModels,type Row} from '../../services/productionClients';
export function WaterField({id,label,value,onChange,multiline=false,disabled=false}:{id:string;label:string;value:any;onChange:(value:string)=>void;multiline?:boolean;disabled?:boolean}){
 return <View className="water-field"><Text className="water-label">{label}</Text>{multiline?<Textarea id={id} className="water-input water-textarea" maxlength={-1} value={String(value||'')} onInput={event=>onChange(event.detail.value)} disabled={disabled}/>:<Input id={id} className="water-input" maxlength={-1} value={String(value||'')} onInput={event=>onChange(event.detail.value)} disabled={disabled}/>}</View>;
}
export function WaterSelect({id,label,items,value,onChange,disabled=false}:{id:string;label:string;items:Row[];value:string;onChange:(value:string)=>void;disabled?:boolean}){
 return <Picker id={id} disabled={disabled||!items.length} range={items.map(item=>item.name||item.displayName||item.modelId||item.id)} value={Math.max(0,items.findIndex(item=>String(item.id)===value))} onChange={event=>{const item=items[Number(event.detail.value)];if(item)onChange(String(item.id));}}><View className="water-select"><Text className="water-label">{label}</Text><Text>{items.find(item=>String(item.id)===value)?.name||items.find(item=>String(item.id)===value)?.displayName||(items.length?'请选择':'暂无可用选项')}</Text></View></Picker>;
}
export function WaterSheet({title,busy,onClose,children}:{title:string;busy:boolean;onClose:()=>void;children:any}){
 return <View className="water-overlay"><View className="water-mask" onClick={()=>{if(!busy)onClose();}}/><View className="water-sheet"><View className="water-heading"><Text className="water-title">{title}</Text><Button className="water-link" disabled={busy} onClick={onClose}>取消</Button></View><ScrollView scrollY className="water-sheet-body">{children}</ScrollView></View></View>;
}
export default function WaterEditor({edit,onChange,onSave,onClose,busy,error}:{edit:Row;onChange:(value:Row)=>void;onSave:()=>void;onClose:()=>void;busy:boolean;error:string}){
 const kind=edit.editorKind;
 const fields=kind==='asset'?[['name','资产名称'],['description','外观与描述'],['prompt','素材提示词']]:kind==='candidate'?[['text','分镜正文'],['speaker','说话人']]:[['subtitleText','分镜正文'],['speaker','说话人'],['imagePrompt','图片提示词'],['videoPrompt','视频提示词']];
 return <WaterSheet title={kind==='asset'?'编辑资产':kind==='candidate'?'编辑候选分镜':'编辑分镜'} busy={busy} onClose={onClose}>{kind==='asset'?<WaterSelect id="water-asset-category" label="资产类型" items={[{id:'character',name:'人物'},{id:'scene',name:'场景'},{id:'prop',name:'道具'}]} value={edit.category} onChange={category=>onChange({...edit,category})} disabled={busy||Boolean(edit.id)}/>:null}{fields.map(([key,label])=><WaterField key={key} id={`water-editor-${key}`} label={label} value={edit[key]} multiline={key!=='name'&&key!=='speaker'} onChange={value=>onChange({...edit,[key]:value})} disabled={busy}/>)}{error?<Text className="water-error">{error}</Text>:null}<Button className="water-primary water-wide" disabled={busy} onClick={onSave}>{kind==='candidate'?'保留候选修改':kind==='asset'?'保存资产':'保存分镜'}</Button></WaterSheet>;
}
export function WaterGenerate({segment,models,onSave,onClose,busy,error}:{segment:Row;models:Row[];onSave:(modelId:string,kind:string)=>void;onClose:()=>void;busy:boolean;error:string}){
 const kind='image';const [model,setModel]=useState('');
 const options=waterImageModels(models);
 return <WaterSheet title="生成分镜图片" busy={busy} onClose={onClose}><Text className="water-note">仅为当前分镜生成图片。视频需在作品工作区按整部作品执行；配音在分镜卡片单独提交，使用服务端配音配置。</Text><Text className="water-prose">{segment.subtitleText||segment.text}</Text><WaterSelect id="water-generation-model" label="生成模型" items={options.map(item=>({...item,id:String(item.id)}))} value={model} onChange={setModel} disabled={busy}/>{!options.length?<Text className="water-note">暂无可明确匹配的图片模型；缺少模型编号或同编号多通道时，请先在网页端核对配置。</Text>:null}{model?<Text className="water-note">模型编号：{model}</Text>:null}{error?<Text className="water-error">{error}</Text>:null}<Button className="water-primary water-wide" disabled={busy||!options.some(item=>String(item.id)===model)} onClick={()=>onSave(model,kind)}>开始生成</Button></WaterSheet>;
}
