import { Alert, Button, Input, InputNumber, Select, Space, Switch, Tabs, Upload, message } from 'antd';
import { useEffect, useState } from 'react';
import { getAvailableModelCatalog, modelSelectOptions } from '../../../shared/api/modelCatalog';
import { getN8ComicPublishStatus } from '../../../shared/api/batchFactoryV12';
import { syncWebSubmitConfigs } from '../../../shared/api/novelFetch';
import { listPersonalTextBackgrounds, uploadPersonalTextBackground } from '../../../shared/api/batchFactoryText';
import { producerAccountOptions } from './batchFactoryComicPublishAccounts';
import { TEXT_PRODUCTION_SECTIONS, textProfileChanges } from './textProductionConfig';

function ColorPicker({ label, value, onChange }) {
  const [draft, setDraft] = useState(value);
  useEffect(() => setDraft(value), [value]);
  function edit(color) {
    setDraft(color);
    if (/^#[0-9a-f]{6}$/i.test(color)) onChange(color);
  }
  return <Space>
    <input type="color" aria-label={label} value={value} onChange={event => edit(event.target.value)} />
    <Input aria-label={label+'色值'} value={draft} maxLength={7} style={{width: 100}}
      onChange={event => edit(event.target.value)} onBlur={() => setDraft(value)} />
  </Space>;
}

function BackgroundPicker({value,items,onChange,onUploaded}) {
  const [busy,setBusy]=useState(false);
  const image = Boolean(value?.id);
  const upload=async file=>{
    setBusy(true);
    try {
      const item=await uploadPersonalTextBackground(file);
      onUploaded(item);
      onChange({kind:'library',id:item.id});
    } catch(error) {message.error(error.message || '背景上传失败');}
    finally {setBusy(false);}
  };
  return <Space wrap>
    <ColorPicker label="背景颜色" value={value?.color || '#111318'} onChange={color=>onChange({kind:'color',color})}/>
    {image && <Button onClick={()=>onChange({kind:'color',color:'#111318'})}>使用纯色背景</Button>}
    <Select aria-label="个人背景库" showSearch optionFilterProp="label" style={{minWidth:150}} placeholder="选择背景图" value={image ? value.id : undefined}
      options={items.map(x=>({value:x.id,label:x.name || x.id}))} onChange={id=>onChange({kind:'library',id})}/>
    <Upload accept="image/png,image/jpeg,image/webp" showUploadList={false} beforeUpload={file=>{void upload(file);return false;}}><Button loading={busy}>上传背景</Button></Upload>
  </Space>;
}

// All three entry points render this form. value is the resolved display
// model while changes contain only the path the person actually touched.
export default function TextProductionSettingsForm({value, explicitKeys=[], onChange, onRestore, scope='batch', open=true}) {
  const [tab,setTab]=useState('models');
  const [directory,setDirectory]=useState({});
  const [models,setModels]=useState([]);
  const [backgrounds,setBackgrounds]=useState([]);
  const [syncedProfiles,setSyncedProfiles]=useState(null);
  const [error,setError]=useState('');
  const [syncing,setSyncing]=useState(false);
  const profileId = 'publishSettings.websiteProfileId';
  const profiles = syncedProfiles ?? value['publishSettings.websiteProfiles'] ?? [];
  useEffect(() => {
    if (!open) setSyncedProfiles(null);
  }, [open]);
  useEffect(()=>{
    if (!open) return;
    let alive=true;
    setError('');
    const load=tab==='models' ? getAvailableModelCatalog().then(catalog=>{
      if (!alive) return;
      setModels(modelSelectOptions(catalog.models,'text'));
      if (catalog.failedKinds?.includes('text')) setError('文本模型目录读取失败，请稍后重试');
    })
      : tab==='layout' ? listPersonalTextBackgrounds().then(x=>{if(alive)setBackgrounds(x.items || []);})
      : getN8ComicPublishStatus({silent:true,suppressGlobalError:true}).then(x=>{if(alive)setDirectory(x || {});});
    load.catch(e=>{if(alive)setError(e.message || '目录读取失败，请稍后重试');});
    return ()=>{alive=false;};
  },[tab,open]);
  const syncProfiles=async()=>{
    setSyncing(true);
    try {
      const result = await syncWebSubmitConfigs({persist:false});
      if (!Array.isArray(result.groups)) throw new Error('网站配置档目录格式错误');
      const changes = textProfileChanges(result.groups,value[profileId]);
      setSyncedProfiles(changes['publishSettings.websiteProfiles']);
      if (scope !== 'book' || explicitKeys.includes(profileId)) {
        onChange(changes);
        if (value[profileId] && !changes[profileId]) message.warning('原网站配置档已失效，请重新选择');
      }
    }
    catch(e){message.error(e.message || '网站配置档同步失败');}
    finally{setSyncing(false);}
  };
  const choices=field=>{
    switch(field.type){
      case 'model': case 'models':return models;
      case 'producer':return producerAccountOptions(directory.producerAccounts,value[field.key],value['comicPublishSettings.organizationId'],directory.producerAccountOrganizations);
      case 'organization':return (directory.organizations || []).map(o=>({value:String(o.id),label:o.name || String(o.id)}));
      case 'profile':return profiles.map(p=>({value:String(p.id),label:p.name || String(p.id)}));
      default:return field.options;
    }
  };
  const control=field=>{
    const v=value[field.key];
    const set=next=>onChange(field.type==='profile' ? textProfileChanges(profiles,next) : {[field.key]:next});
    const attrs={'aria-label':field.label};
    if(field.type==='background')return <BackgroundPicker value={v} items={backgrounds} onChange={set} onUploaded={item=>setBackgrounds(old=>[item,...old.filter(x=>x.id!==item.id)])}/>;
    if(field.type==='switch')return <Switch {...attrs} checked={v} onChange={set}/>;
    if(field.type==='color')return <ColorPicker label={field.label} value={v} onChange={set}/>;
    if(field.type==='number' || field.type==='percent'){
      const percent=field.type==='percent';
      return <InputNumber {...attrs} min={field.min} max={field.max} step={field.step || 1} addonAfter={percent ? '%' : undefined}
        value={percent ? Math.round(v*100) : v} onChange={n=>{if(n!=null)set(percent ? n/100 : n);}}/>;
    }
    if(field.type==='input')return <Input {...attrs} value={v} onChange={e=>set(e.target.value)}/>;
    const many=field.type==='models';
    const disabled=many && value.sensitiveMode!=='ai';
    return <Select {...attrs} showSearch optionFilterProp="label" style={{minWidth:150}} mode={many?'multiple':undefined}
      allowClear={field.type!=='choice'} disabled={disabled} options={choices(field)}
      placeholder={many?'默认使用文本模型':'请选择'} value={v==='' ? undefined : v} onChange={n=>set(n ?? (many?[]:''))}/>;
  };
  const panel=section=><div className="batch-factory-engine-card"><div className="batch-factory-engine-card-body batch-factory-engine-grid">
    {section.fields.map(field=>{
      const own=explicitKeys.includes(field.key);
      return <div className="batch-factory-engine-field" key={field.key}>
        <span><b>{field.label}</b>{scope==='book' && own && <small> · 本书覆盖</small>}</span>
        {control(field)}
        {scope==='book' && own && <Button size="small" type="link" onClick={()=>onRestore(field.key)}>恢复继承</Button>}
        {field.key==='sensitiveMode' && <small>{value.sensitiveMode==='ai'?field.aiHint:value.sensitiveMode==='off'?'直接使用生产稿。':'系统优化在本地执行。'}</small>}
        {field.key==='highlightReuseEnabled' && <small>复用已有标红词，本地排版。</small>}
        {field.key==='textFontId' && value[field.key]==='ma-shan-zheng' && <small>部分字回退到 Noto Sans SC。</small>}
        {field.type==='profile' && <Button size="small" loading={syncing} onClick={syncProfiles}>同步网站配置档</Button>}
      </div>;
    })}
  </div></div>;
  return <>{error && <Alert type="warning" showIcon message={error}/>}
    <Tabs activeKey={tab} onChange={setTab} items={TEXT_PRODUCTION_SECTIONS.map(s=>({key:s.key,label:s.label,children:panel(s)}))}/>
  </>;
}
