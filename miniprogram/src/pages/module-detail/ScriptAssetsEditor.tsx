import { Button, Input, Picker, ScrollView, Text, Textarea, View } from '@tarojs/components';
import { useRef, useState } from 'react';
import AppIcon from '../../components/AppIcon';
import ScriptEntityMedia from './ScriptEntityMedia';
import { newScriptEntity, scriptEntityName, type EntityType, type ScriptAssets } from '../../services/scriptAssets';

const labels: Record<string, string> = { name: '名称', 角色名称: '名称', 场景名称: '名称', 名称: '名称', identity: '身份', appearance: '外形（剧本）', 外形: '外形（剧本）', personality: '性格', description: '描述', time: '时段', atmosphere: '氛围', relation: '关系', voice: '声音', style: '风格' };
export default function ScriptAssetsEditor({ value, onChange, model, novel, preset, models = [], onModelChange }: { value: ScriptAssets; onChange: (value: ScriptAssets) => void; model: string; novel: string; preset: string; models?: Array<{id:string;name:string}>; onModelChange?: (id:string)=>void }) {
  const [type, setType] = useState<EntityType>('characters'), [selected, setSelected] = useState(0);
  const current = useRef(value); current.current = value;
  const entity = value[type][selected];
  const fields = entity ? entity.data && typeof entity.data === 'object' && !Array.isArray(entity.data) ? entity.data as Record<string, any> : entity as Record<string, any> : {};
  const nested = Boolean(entity?.data && typeof entity.data === 'object' && !Array.isArray(entity.data));
  const update = (next: any) => { const source = current.current; onChange({ ...source, [type]: source[type].map((item, index) => index === selected ? next : item) }); };
  const changeField = (key:string,text:string) => { const sourceEntity=current.current[type][selected]; if(sourceEntity)update(nested?{...sourceEntity,data:{...(sourceEntity.data as object),[key]:text}}:{...sourceEntity,[key]:text}); };
  return <View className="studio-asset-editor">
    <View className="studio-asset-tabs">{(['characters', 'scenes'] as EntityType[]).map(item => <Button key={item} className={type === item ? 'is-active' : ''} onClick={() => { setType(item); setSelected(0); }}>{item === 'characters' ? '人物' : '场景'} · {value[item].length}</Button>)}</View>
    <ScrollView scrollX className="studio-asset-selector"><View>{value[type].map((item, index) => <Button key={item.id || index} className={index === selected ? 'is-active' : ''} onClick={() => setSelected(index)}>{scriptEntityName(item)}</Button>)}</View></ScrollView>
    <Button className="studio-add-entity" onClick={() => { const id = `script-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 9)}`; const source = current.current; setSelected(source[type].length); onChange({ ...source, [type]: [...source[type], newScriptEntity(type, id)] }); }}><AppIcon name="plus" size={16} />添加{type === 'characters' ? '人物' : '场景'}</Button>
    {entity ? <View className="studio-single-entity"><Text className="studio-caption">一次编辑一个设定；点击底部“应用修改”后用于本次创作。</Text>
      {Object.entries(fields).filter(([key, item]) => key !== 'id' && typeof item === 'string' && !/url|token|key|secret/i.test(key)).map(([key, text]) => <View key={key}><Text className="mp-field-label">{labels[key] || key}</Text>{['name','名称','角色名称','场景名称','身份','identity','时段','time'].includes(key)?<Input id={['name','名称','角色名称','场景名称'].includes(key)?'script-entity-name':undefined} className="mp-input studio-entity-short" maxlength={-1} value={text} onInput={event=>changeField(key,event.detail.value)}/>:<Textarea className="mp-textarea" maxlength={-1} value={text} onInput={event=>changeField(key,event.detail.value)}/>}</View>)}
      {onModelChange?<View className="studio-reference-model"><Text className="mp-field-label">参考图模型</Text><Picker disabled={!models.length} range={models.map(item=>item.name)} value={Math.max(0,models.findIndex(item=>item.id===model))} onChange={event=>{const selectedModel=models[Number(event.detail.value)];if(selectedModel)onModelChange(selectedModel.id);}}><View className="studio-setting">{models.find(item=>item.id===model)?.name || (models.length?'请选择图片模型':'暂无可用图片模型')}</View></Picker></View>:null}
      <ScriptEntityMedia key={entity.id || `${type}-${selected}`} type={type} entity={entity} model={model} novel={novel} preset={preset} onChange={next => { const sourceEntity = current.current[type][selected]; if (sourceEntity) update({ ...sourceEntity, id: next.id || sourceEntity.id, imageUrls: next.imageUrls, mainImageUrl: next.mainImageUrl }); }} />
      <Button className="studio-remove-entity" onClick={() => { const source = current.current; onChange({ ...source, [type]: source[type].filter((_, index) => index !== selected) }); setSelected(Math.max(0, selected - 1)); }}>移除此{type === 'characters' ? '人物' : '场景'}</Button>
    </View> : <View className="studio-empty-assets"><AppIcon name={type === 'characters' ? 'user-round' : 'camera'} size={30} /><Text>还没有{type === 'characters' ? '人物' : '场景'}设定</Text><Text className="studio-caption">可手动添加，或返回正文后提取。</Text></View>}
  </View>;
}
