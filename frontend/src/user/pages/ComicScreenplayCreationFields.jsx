import OriginalIcon from './ComicScreenplayIcon.jsx';
import { Alert, Button, Input, Modal, Select, Spin, Upload } from 'antd';
import { Check, FileText, ImageOff, Palette, Search, UploadCloud, X } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { createComicCreativeStyle, getComicStylePreview } from '../../shared/api/comicScreenplay.js';
import { listAvailableModels } from '../../shared/api/modelCatalog.js';
import { creativePromptFields, creativePromptOptions } from './comic-screenplay-creative.js';

export const comicDegreeOptions = [
  { value: 'master', label: '大模型-3D', description: '人物、场景、道具 · 3D 设定', detail: '包含颜值定位与出现集数' },
  { value: 'master_real', label: '大模型-写实', description: '人物、场景、道具 · 写实设定', detail: '包含颜值定位与出现集数' },
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

// 洛水「新建剧本」资产解析程度是横排 4 段切换 + 一行说明，不是 V88 的大卡片。
export function ComicDegreeSegments({ value, onChange, disabled = false }) {
  const active = comicDegreeOptions.find(option => option.value === value) || comicDegreeOptions[0];
  return <div className="comic-degree-segments-field">
    <div className="degree-segments" role="group" aria-label="资产解析程度">
      {comicDegreeOptions.map(option => <button type="button" key={option.value} disabled={disabled} aria-pressed={value === option.value} className={value === option.value ? 'is-active' : ''} onClick={() => onChange(option.value)}>{option.label}</button>)}
    </div>
    <small className="comic-degree-segments-hint">{active.description}；{active.detail}</small>
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

function ComicStyleEmptyArt() {
  return <svg className="comic-style-empty-art" viewBox="0 0 220 150" aria-hidden="true" focusable="false">
    <ellipse cx="108" cy="138" rx="55" ry="7" fill="#495261" opacity=".42" />
    <path d="M77 69h72a7 7 0 0 1 7 7v45a7 7 0 0 1-7 7H77a7 7 0 0 1-7-7V76a7 7 0 0 1 7-7z" fill="#6d7584" />
    <rect x="79" y="77" width="68" height="44" rx="3" fill="#929bab" opacity=".72" />
    <path d="M95 128h36l7 9H88z" fill="#6d7584" />
    <rect x="80" y="134" width="66" height="7" rx="3.5" fill="#7d8695" />
    <path d="M133 38a15 15 0 0 1 15 15c0 5.1-2.6 9.6-6.5 12.2l1.2 9.3-8.4-5.2a15.2 15.2 0 0 1-3.8.5 15 15 0 1 1 2.5-29.8z" fill="#8993a3" />
    <circle cx="134" cy="52" r="3" fill="#c7cdd7" opacity=".82" />
  </svg>;
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
  function confirmStyle() {
    onChange({ style_id: draft?.style_id || null, style_name: draft?.style_name || '' });
    setOpen(false);
  }
  return <>
    <div className="comic-style-current"><button disabled={disabled} type="button" className="style-picker-trigger" onClick={choose}><OriginalIcon name="ColorPaletteOutline" size={compact ? 13 : 17} /><span>{value.style_name || '点击选择画风'}</span>{compact?<OriginalIcon name="ChevronDownOutline" size={12}/>:null}<span className="comic-style-change">选择</span></button>
      {value.style_id ? <Button disabled={disabled} type="text" size="small" aria-label="清除画风选择" icon={<OriginalIcon name="CloseOutline" size={14} />} onClick={() => onChange({ style_name: '', style_id: null })} /> : null}</div>
    <Modal
      className="screenplay-style-modal comic-style-parity-modal"
      title="选择我的画风"
      closeIcon={<X size={18} />}
      open={open}
      onCancel={() => setOpen(false)}
      footer={<div className="comic-style-modal-footer"><Button className="comic-style-cancel" onClick={() => setOpen(false)}>取消</Button><Button className="comic-style-confirm" type="primary" onClick={confirmStyle}>确认选择</Button></div>}
      width={780}
    >
      <p className="comic-style-caption"><Palette size={13} />画风来自你的私人画风库；不选择也可以保存作品。</p>
      <Button className="comic-style-add-btn" onClick={() => { setNewName(''); setNewPrompt(''); setAddError(''); setAddOpen(true); }}>新增私人画风</Button>
      <Input className="comic-style-search" prefix={<Search size={14} />} value={query} onChange={event => setQuery(event.target.value)} placeholder="搜索画风名称" allowClear />
      <div className="comic-style-selection"><span>待确认：{draft?.style_name || '无画风'}</span><Button className="comic-style-none-btn" size="small" onClick={() => setDraft({ style_id: null, style_name: '' })}>不使用画风</Button></div>
      {draft?.style_id && !styles.some(item => item.id === draft.style_id) ? <Alert type="warning" showIcon message="已保存画风当前不在库中；原选择仍保留，可重新选择或清空。" /> : null}
      {filtered.length ? <div className="comic-style-library">{filtered.map(style => <button type="button" key={style.id} aria-pressed={draft?.style_id === style.id} className={draft?.style_id === style.id ? 'is-active' : ''} onClick={() => setDraft({ style_name: style.name, style_id: style.id })}>
        <StylePreview src={style.images?.[0]} name={style.name} /><div><b>{style.name}</b><span className="comic-choice-check">{draft?.style_id === style.id ? <Check size={13} /> : null}</span></div>
      </button>)}</div> : <div className="comic-style-empty"><ComicStyleEmptyArt /><p>{query ? '没有匹配的画风' : '私人画风库暂无画风，可先不选择'}</p></div>}
    </Modal>
    <Modal title="新增私人画风" open={addOpen} onCancel={() => { if (!adding) setAddOpen(false); }} maskClosable={!adding} closable={!adding} confirmLoading={adding} onOk={addStyle} okText="保存到我的画风库" cancelButtonProps={{ disabled: adding }}>
      {addError ? <Alert type="error" showIcon message={addError} /> : null}
      <div className="comic-creative-fields"><label><span>画风名称</span><Input disabled={adding} maxLength={80} value={newName} onChange={event => setNewName(event.target.value)} placeholder="例如：暖色水彩" /></label><label><span>画风文字</span><Input.TextArea disabled={adding} value={newPrompt} onChange={event => setNewPrompt(event.target.value)} rows={5} placeholder="填写你希望使用的画风特征；此文字只保存到你的私人画风库。" /></label></div>
    </Modal>
  </>;
}

export function ComicCreativeFields({ value, onChange, library, loading, error, onRetry, onLibraryChange, disabled = false, hideStyle = false }) {
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
      {!hideStyle ? <label><span>选择风格</span><ComicStylePicker disabled={disabled || loading || !!error} value={value} styles={library.styles} onChange={onChange} onLibraryChange={onLibraryChange} /></label> : null}
      <label><span>图片核心要求</span><Input.TextArea disabled={disabled} rows={2} value={value.image_core_instruction??''} placeholder="人物与场景图片共同遵守的补充要求" onChange={event => onChange({ image_core_instruction: event.target.value })} /></label>
      <label><span>图片生成模型</span><Select disabled={disabled} allowClear value={value.image_model_id || undefined} options={modelOptions(value.image_model_id)} placeholder="使用账号默认模型（可选）" onChange={id => onChange({ image_model_id: id || '' })} /></label>
      <label><span>图片编辑模型</span><Select disabled={disabled} allowClear value={value.image_edit_model_id || undefined} options={modelOptions(value.image_edit_model_id)} placeholder="使用账号默认模型（可选）" onChange={id => onChange({ image_edit_model_id: id || '' })} /></label>{modelError ? <Alert type="warning" showIcon message={modelError} action={<Button size="small" onClick={() => setModelAttempt(current => current + 1)}>重试模型列表</Button>} /> : null}
    </Spin>
  </div>;
}

export function ComicSourceInput({ mode, onModeChange, content, onContentChange, filename, parsing, onFile, onRemove, disabled = false }) {
  return <section className="comic-source-input" aria-label="原文内容">
    <div className="comic-source-heading"><b>上传剧本格式 <i>*</i></b><span>小说正文 / 已有剧本 / 故事大纲</span></div>
    <div className="comic-source-tabs" role="group" aria-label="内容输入方式">
      <button type="button" aria-pressed={mode === 'text'} className={mode === 'text' ? 'is-active' : ''} onClick={() => onModeChange('text')}><FileText size={15} />直接输入</button>
      <button type="button" aria-pressed={mode === 'file'} className={mode === 'file' ? 'is-active' : ''} onClick={() => onModeChange('file')}><UploadCloud size={15} />上传文件</button>
    </div>
    {mode === 'file' ? <Upload.Dragger accept=".txt,.docx" disabled={parsing || disabled} beforeUpload={file => { void onFile(file); return false; }} showUploadList={false}><UploadCloud size={23} /><p>{parsing ? '正在读取文件…' : filename || '点击或拖动文件 (txt / docx)'}</p><small>导入后可在下方预览并修改内容</small></Upload.Dragger> : null}
    {filename ? <div className="comic-source-file"><FileText size={14} /><span>{filename}</span><Button type="text" size="small" onClick={onRemove}>移除文件</Button></div> : null}
    <Input.TextArea aria-label="原文内容" value={content} onChange={event => onContentChange(event.target.value)} rows={6} placeholder="在这里输入或粘贴内容，也可以切换到上传文件。" />
    <div className="comic-source-count"><span>原文按你的编辑内容保存</span><span>{[...content].length.toLocaleString()} / 600,000 字</span></div>
  </section>;
}
