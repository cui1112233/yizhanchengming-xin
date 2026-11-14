import OriginalIcon from './ComicScreenplayIcon.jsx';
import { Alert, Button, Empty, Input, Modal, Select, Spin, Upload } from 'antd';
import { Check, FileText, ImageOff, Palette, Search, UploadCloud, X } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { createComicCreativeStyle, getComicStylePreview } from '../../shared/api/comicScreenplay.js';
import { listAvailableModels } from '../../shared/api/modelCatalog.js';
import { creativePromptFields, creativePromptOptions } from './comic-screenplay-creative.js';

export const comicDegreeOptions = [
  { value: 'master', label: '大师模式 · 3D', description: '人物、场景、道具 · 3D 设定', detail: '包含颜值定位与出现集数' },
  { value: 'master_real', label: '大师模式 · 写实', description: '人物、场景、道具 · 写实设定', detail: '包含颜值定位与出现集数' },
  { value: 'more_diff', label: '差异化', description: '强化人物的外形差异', detail: '场景与道具使用标准解析' },
  { value: 'standard', label: '标准', description: '提取基础人物设定', detail: '场景与道具使用标准解析' },
];

export function ComicDegreeCards({ value, onChange, compact = false }) {
  return <div className={`comic-degree-cards${compact ? ' is-compact' : ''}`} role="group" aria-label="资产解析程度">
    {comicDegreeOptions.map(option => <button type="button" key={option.value} aria-pressed={value === option.value} className={value === option.value ? 'is-active' : ''} onClick={() => onChange(option.value)}>
      <span className="comic-degree-title"><b>{option.label}</b><span className="comic-choice-check">{value === option.value ? <Check size={13} /> : null}</span></span>
      <span>{option.description}</span><small>{option.detail}</small>
    </button>)}
  </div>;
}

function StylePreview({ src, name }) {
  const [image, setImage] = useState('');
  const [failed, setFailed] = useState(false);
  useEffect(() => {
    let alive = true;
    let objectURL = '';
    setImage(''); setFailed(false);
    if (!src) return undefined;
    // 私人素材通过带登录凭据的同源接口读取，不把凭据放进图片地址。
    getComicStylePreview(src).then(blob => {
      if (!alive) return;
      objectURL = URL.createObjectURL(blob); setImage(objectURL);
    }).catch(() => { if (alive) setFailed(true); });
    return () => { alive = false; if (objectURL) URL.revokeObjectURL(objectURL); };
  }, [src]);
  return image && !failed ? <img src={image} alt={`${name}画风预览`} onError={() => setFailed(true)} />
    : <span className="comic-style-placeholder"><ImageOff size={22} /><small>{failed ? '预览加载失败' : src ? '加载预览…' : '暂无预览图'}</small></span>;
}

export function ComicStylePicker({ value, onChange, styles = [], disabled = false, onLibraryChange, compact = false }) {
  const [open, setOpen] = useState(false);
  const [draft, setDraft] = useState(null);
  const [query, setQuery] = useState('');
  const [addOpen, setAddOpen] = useState(false), [adding, setAdding] = useState(false);
  const [newName, setNewName] = useState(''), [newPrompt, setNewPrompt] = useState(''), [addError, setAddError] = useState('');
  const addPending = useRef(false);
  async function addStyle() {
    if (addPending.current) return;
    addPending.current = true; setAdding(true); setAddError('');
    try {
      const style = await createComicCreativeStyle({ name: newName, prompt: newPrompt });
      setDraft({ style_id: style.id, style_name: style.name }); setAddOpen(false);
      void onLibraryChange?.();
    } catch (error) { setAddError(error.message || '画风保存失败'); }
    finally { addPending.current = false; setAdding(false); }
  }
  const filtered = styles.filter(item => item.name.toLowerCase().includes(query.trim().toLowerCase()));
  function choose() { setDraft({ ...value }); setQuery(''); setOpen(true); }
  return <>
    <div className="comic-style-current"><button disabled={disabled} type="button" className="style-picker-trigger" onClick={choose}><OriginalIcon name="ColorPaletteOutline" size={compact ? 13 : 17} /><span>{value.style_name || '无画风（可选）'}</span>{compact?<OriginalIcon name="ChevronDownOutline" size={12}/>:null}<span className="comic-style-change">选择</span></button>
      {value.style_id ? <Button disabled={disabled} type="text" size="small" aria-label="清除画风选择" icon={<OriginalIcon name="CloseOutline" size={14} />} onClick={() => onChange({ style_name: '', style_id: null })} /> : null}</div>
    <Modal className="screenplay-style-modal" title="选择我的画风" open={open} onCancel={() => setOpen(false)} footer={<><Button onClick={() => setOpen(false)}>取消</Button><Button type="primary" onClick={() => { onChange({ style_id: draft?.style_id || null, style_name: draft?.style_name || '' }); setOpen(false); }}>确认选择</Button></>} width={780}>
      <p className="comic-style-caption">画风来自你的私人画风库；不选择也可以保存作品。</p><Button onClick={() => { setNewName(''); setNewPrompt(''); setAddError(''); setAddOpen(true); }}>新增私人画风</Button>
      <Input prefix={<OriginalIcon name="SearchOutline" size={14} />} value={query} onChange={event => setQuery(event.target.value)} placeholder="搜索画风名称" allowClear />
      <div className="comic-style-selection">待确认：{draft?.style_name || '无画风'}<Button size="small" onClick={() => setDraft({ style_id: null, style_name: '' })}>不使用画风</Button></div>
      {draft?.style_id && !styles.some(item => item.id === draft.style_id) ? <Alert type="warning" showIcon message="已保存画风当前不在库中；原选择仍保留，可重新选择或清空。" /> : null}
      <div className="comic-style-library">{filtered.map(style => <button type="button" key={style.id} aria-pressed={draft?.style_id === style.id} className={draft?.style_id === style.id ? 'is-active' : ''} onClick={() => setDraft({ style_name: style.name, style_id: style.id })}>
        <StylePreview src={style.images?.[0]} name={style.name} /><div><b>{style.name}</b><span className="comic-choice-check">{draft?.style_id === style.id ? <Check size={13} /> : null}</span></div>
      </button>)}</div>{!filtered.length ? <Empty description={query ? '没有匹配的画风' : '私人画风库暂无画风，可先不选择'} /> : null}
    </Modal>
    <Modal title="新增私人画风" open={addOpen} onCancel={() => { if (!adding) setAddOpen(false); }} maskClosable={!adding} closable={!adding} confirmLoading={adding} onOk={addStyle} okText="保存到我的画风库" cancelButtonProps={{ disabled: adding }}>
      {addError ? <Alert type="error" showIcon message={addError} /> : null}
      <div className="comic-creative-fields"><label><span>画风名称</span><Input disabled={adding} maxLength={80} value={newName} onChange={event => setNewName(event.target.value)} placeholder="例如：暖色水彩" /></label><label><span>画风文字</span><Input.TextArea disabled={adding} value={newPrompt} onChange={event => setNewPrompt(event.target.value)} rows={5} placeholder="填写你希望使用的画风特征；此文字只保存到你的私人画风库。" /></label></div>
    </Modal>
  </>;
}

export function ComicCreativeFields({ value, onChange, library, loading, error, onRetry, onLibraryChange, disabled = false }) {
  const [imageModels, setImageModels] = useState([]), [modelError, setModelError] = useState('');
  const [modelAttempt, setModelAttempt] = useState(0);
  useEffect(() => {
    let alive = true;
    setModelError('');
    listAvailableModels('image').then(rows => { if (alive) setImageModels(rows); }).catch(error => { if (alive) setModelError(error.message || '图片模型读取失败'); });
    return () => { alive = false; };
  }, [modelAttempt]);
  function modelOptions(id) {
    const options = imageModels.map(model => ({ value: model.id, label: model.displayName || model.name || model.id }));
    if (id && !options.some(option => option.value === id)) options.unshift({ value: id, label: `已保存模型（${id}）` });
    return options;
  }
  return <div className="comic-creative-fields">
    {error ? <Alert type="error" showIcon message={error} action={<Button size="small" onClick={onRetry}>重新读取</Button>} /> : null}
    <Spin spinning={loading}>
      {creativePromptFields.map(({ field, purpose, label }) => <label key={field}><span>{label}</span><Select aria-label={label} disabled={disabled || loading || !!error} value={value[field] || undefined} placeholder={`请选择${label}`} options={creativePromptOptions(library.prompts, purpose, value[field])} onChange={id => onChange({ [field]: id })} /></label>)}
      <label><span>作品画风</span><ComicStylePicker disabled={disabled || loading || !!error} value={value} styles={library.styles} onChange={onChange} onLibraryChange={onLibraryChange} /></label>
      <label><span>图片核心要求</span><Input.TextArea disabled={disabled} rows={2} value={value.image_core_instruction??''} placeholder="人物与场景图片共同遵守的补充要求" onChange={event => onChange({ image_core_instruction: event.target.value })} /></label>
      <label><span>图片生成模型</span><Select disabled={disabled} allowClear value={value.image_model_id || undefined} options={modelOptions(value.image_model_id)} placeholder="使用账号默认模型（可选）" onChange={id => onChange({ image_model_id: id || '' })} /></label>
      <label><span>图片编辑模型</span><Select disabled={disabled} allowClear value={value.image_edit_model_id || undefined} options={modelOptions(value.image_edit_model_id)} placeholder="使用账号默认模型（可选）" onChange={id => onChange({ image_edit_model_id: id || '' })} /></label>{modelError ? <Alert type="warning" showIcon message={modelError} action={<Button size="small" onClick={() => setModelAttempt(current => current + 1)}>重试模型列表</Button>} /> : null}
    </Spin>
  </div>;
}

export function ComicSourceInput({ mode, onModeChange, content, onContentChange, filename, parsing, onFile, onRemove, disabled = false }) {
  return <section className="comic-source-input" aria-label="原文内容">
    <div className="comic-source-heading"><b>原文内容 <i>*</i></b><span>小说正文 / 已有剧本 / 故事大纲</span></div>
    <div className="comic-source-tabs" role="group" aria-label="内容输入方式">
      <button type="button" aria-pressed={mode === 'text'} className={mode === 'text' ? 'is-active' : ''} onClick={() => onModeChange('text')}><FileText size={15} />直接输入</button>
      <button type="button" aria-pressed={mode === 'file'} className={mode === 'file' ? 'is-active' : ''} onClick={() => onModeChange('file')}><UploadCloud size={15} />上传文件</button>
    </div>
    {mode === 'file' ? <Upload.Dragger accept=".txt,.docx" disabled={parsing || disabled} beforeUpload={file => { void onFile(file); return false; }} showUploadList={false}><UploadCloud size={23} /><p>{parsing ? '正在读取文件…' : filename || '点击或拖入 TXT / DOCX 文件'}</p><small>导入后可在下方预览并修改内容</small></Upload.Dragger> : null}
    {filename ? <div className="comic-source-file"><FileText size={14} /><span>{filename}</span><Button type="text" size="small" onClick={onRemove}>移除文件</Button></div> : null}
    <Input.TextArea aria-label="原文内容" value={content} onChange={event => onContentChange(event.target.value)} rows={6} placeholder="在这里输入或粘贴内容，也可以切换到上传文件。" />
    <div className="comic-source-count"><span>原文按你的编辑内容保存</span><span>{[...content].length.toLocaleString()} / 600,000 字</span></div>
  </section>;
}
