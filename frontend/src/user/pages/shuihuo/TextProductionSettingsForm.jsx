import ColorPicker from './TextColorPicker.jsx';
import { Alert, Button, Input, InputNumber, Select, Switch, Tabs } from 'antd';
import { lazy, Suspense, useEffect, useState } from 'react';
import { describeModelCatalogFailure, getAvailableModelCatalog, modelSelectOptions } from '../../../shared/api/modelCatalog';
import { listPersonalTextBackgrounds, uploadPersonalTextBackground } from '../../../shared/api/batchFactoryText';
import { TEXT_PRODUCTION_SECTIONS, textProductionInitialSection } from './textProductionConfig';
import TextBackgroundGallery from './TextBackgroundGallery.jsx';
import './publish-settings.css';
import {TEXT_LAYOUT_TEMPLATES,publishOverrideRestores,textColorScheme,lightTextChanges} from './textLayoutTemplates.js';
import TextContrastNotice from './TextContrastNotice.jsx';
import TextHighlightRuleSelect from './TextHighlightRuleSelect.jsx';

const SharedImagePublishSettings = lazy(() => import('./BatchFactoryImagePublishSettingsForm').then(module => ({default:module.BatchFactoryComicPublishSettingsForm})));
const SettingsPreview = lazy(()=>import('./TextSettingsPreview.jsx'));

function BackgroundPicker({value,items,onChange,onUploaded}) {
  const upload=async file=>{
      const item=await uploadPersonalTextBackground(file);
      onUploaded(item);
      return item;
  };
  return <div className="batch-factory-background-controls">
    <ColorPicker label="背景颜色" value={value?.color || '#111318'} onChange={color=>onChange({kind:'color',color})}/>
    <TextBackgroundGallery value={value} items={items} onChange={onChange} onUpload={upload}/>
  </div>;
}

// All three entry points render this form. value is the resolved display
// model while changes contain only the path the person actually touched.
export default function TextProductionSettingsForm({value, explicitKeys=[], onChange, onRestore, scope='batch', open=true, activeSection='models',batchId,bookId}) {
  const [tab,setTab]=useState(()=>textProductionInitialSection(activeSection));
  const [models,setModels]=useState([]);
  const [backgrounds,setBackgrounds]=useState([]);
  const [error,setError]=useState('');
  const [modelRetry,setModelRetry]=useState(0);
  useEffect(()=>{
    if (!open) return;
    let alive=true;
    setError('');
    const load=tab==='models' ? getAvailableModelCatalog({kinds:['text']}).then(catalog=>{
      if (!alive) return;
      setModels(modelSelectOptions(catalog.models,'text'));
      setError(describeModelCatalogFailure(catalog));
    })
      : tab==='layout' ? listPersonalTextBackgrounds().then(x=>{if(alive)setBackgrounds(x.items || []);})
      : Promise.resolve();
    load.catch(e=>{if(alive)setError(e.message || '目录读取失败，请稍后重试');});
    return ()=>{alive=false;};
  },[tab,open,modelRetry]);
  const choices=field=>field.type==='model' || field.type==='models' ? models : field.options;
  const control=field=>{
    const v=value[field.key];
    const set=next=>onChange({[field.key]:next});
    const attrs={'aria-label':field.label};
    if(field.type==='highlightRules')return <TextHighlightRuleSelect value={v} legacyKeywords={value.highlightReuseEnabled} onChange={next=>onChange({textHighlightRules:next,...(next.split(',').includes('lines')?{textHighlightLines:value.textHighlightLines || '1,3'}:{})})}/>;
    if(field.type==='background')return <BackgroundPicker value={v} items={backgrounds} onChange={set} onUploaded={item=>setBackgrounds(old=>[item,...old.filter(x=>x.id!==item.id)])}/>;
    if(field.type==='switch')return <Switch {...attrs} checked={v} onChange={set}/>;
    if(field.type==='color')return <div><ColorPicker label={field.label} value={v} onChange={set}/>{field.key==='textColor' && <TextContrastNotice color={v} background={value.textBackground} onChange={onChange}/>}</div>;
    if(field.type==='number' || field.type==='percent'){
      const percent=field.type==='percent';
      return <InputNumber {...attrs} min={field.min} max={field.max} step={field.step || 1} addonAfter={percent ? '%' : undefined}
        disabled={field.key==='textPageTarget' && value.textPaginationMode!=='target'}
        value={percent ? Math.round(v*100) : v} onChange={n=>{if(n!=null)set(percent ? n/100 : n);}}/>;
    }
    if(field.type==='input')return <Input {...attrs} value={v} placeholder={field.key==='textHighlightLines'?'如 1,3（定位整句）':undefined} disabled={field.key==='textHighlightLines' && !value.textHighlightRules?.split(',').includes('lines')} onChange={e=>set(e.target.value)}/>;
    const many=field.type==='models';
    const disabled=many && value.sensitiveMode!=='ai';
    return <Select {...attrs} showSearch optionFilterProp="label" style={{minWidth:150}} mode={many?'multiple':undefined}
      allowClear={field.type!=='choice'} disabled={disabled} options={choices(field)}
      placeholder={many?'默认使用文本模型':'请选择'} value={v==='' ? undefined : v} onChange={n=>set(n ?? (many?[]:''))}/>;
  };
  const fieldGroups=section=>{
    if(section.key==='layout') return [
      {title:'正文排版',fields:section.fields.filter(field=>field.key!=='textBackground' && !field.key.startsWith('textCover.'))},
      {title:'背景与标题',fields:section.fields.filter(field=>field.key==='textBackground' || field.key.startsWith('textCover.'))}
    ];
    if(section.key==='publish') return [
      {title:'自动发布',fields:section.fields.filter(field=>field.key.startsWith('text'))}
    ];
    return [{title:'',fields:section.fields}];
  };
  const renderField=field=>{
      const own=explicitKeys.includes(field.key);
      return <div className={`batch-factory-engine-field${field.type==='switch' ? ' is-switch-row' : ''}`} key={field.key}>
        <span><b>{field.label}</b>{scope==='book' && own && <small> · 本书覆盖</small>}</span>
        {control(field)}
        {scope==='book' && own && <Button className="batch-factory-field-restore" size="small" type="link" onClick={()=>onRestore(field.key)}>恢复继承</Button>}
      </div>;
  };
  const sharedPublishValue = Object.fromEntries(Object.entries(value).filter(([key])=>key.startsWith('comicPublishSettings.')).map(([key,v])=>[key.slice('comicPublishSettings.'.length),v]));
  const panel=section=><>{section.key==='publish' && <Suspense fallback={null}><SharedImagePublishSettings productionMode="text" singleBook={scope==='book'} value={{comicPublishSettings:sharedPublishValue}} onChange={next=>{
    const changes=Object.fromEntries(Object.entries(next.comicPublishSettings || {}).filter(([key,v])=>JSON.stringify(v)!==JSON.stringify(sharedPublishValue[key])).map(([key,v])=>['comicPublishSettings.'+key,v]));
    if(Object.keys(changes).length)onChange(changes);
  }}/></Suspense>}{section.key==='publish' && scope==='book' && <div className="text-publish-restores">{publishOverrideRestores(explicitKeys).map(group=><Button key={group.label} size="small" onClick={()=>onChange(Object.fromEntries(group.keys.map(key=>[key,undefined])))}>{group.label} · 恢复继承</Button>)}</div>}{section.key==='layout' && <div className="text-layout-presets"><b>排版预设</b>{TEXT_LAYOUT_TEMPLATES.map(preset=><Button key={preset.id} onClick={()=>onChange(structuredClone(preset.values))}>{preset.label}</Button>)}<Button onClick={()=>onChange(textColorScheme('dark'))}>深色文字</Button><Button onClick={()=>onChange(lightTextChanges(value.textHighlightRules || (value.highlightReuseEnabled?'keywords':'')))}>浅色文字</Button></div>}<div className={`batch-factory-engine-card batch-factory-settings-card is-${section.key}`}><div className="batch-factory-engine-card-body batch-factory-engine-grid">
    {fieldGroups(section).map(group=><section className="batch-factory-engine-group" key={group.title || section.key}>
      {group.title ? <div className="batch-factory-engine-group-title">{group.title}</div> : null}
      {group.fields.filter(field=>!field.hidden).map(renderField)}
    </section>)}
  </div></div></>;
  return <>{error && <Alert type="warning" showIcon message={error} action={tab==='models' ? <Button size="small" onClick={()=>setModelRetry(value=>value+1)}>重新读取</Button> : undefined}/>}
    {tab==='layout' && batchId && bookId && <Suspense fallback={null}><SettingsPreview key={batchId+bookId} batchId={batchId} bookId={bookId} value={value}/></Suspense>}
    <Tabs activeKey={tab} onChange={setTab} items={TEXT_PRODUCTION_SECTIONS.map(s=>({key:s.key,label:s.label,children:panel(s)}))}/>
  </>;
}
