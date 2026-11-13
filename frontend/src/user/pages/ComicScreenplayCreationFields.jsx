import { Alert, Button, Empty, Input, Modal, Spin, Upload } from 'antd';
import { Check, FileText, ImageOff, Palette, Search, UploadCloud, X } from 'lucide-react';
import { useEffect, useState } from 'react';
import { listComicScreenplayStyles } from '../../shared/api/comicScreenplay.js';
import { getComicReferencePreviewBlob } from '../../shared/api/batchFactory.js';

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
    // 受保护的预设图使用现有带登录凭据的读取方式，不把凭据放进图片地址。
    getComicReferencePreviewBlob(src).then(blob => {
      if (!alive) return;
      objectURL = URL.createObjectURL(blob); setImage(objectURL);
    }).catch(() => { if (alive) setFailed(true); });
    return () => { alive = false; if (objectURL) URL.revokeObjectURL(objectURL); };
  }, [src]);
  return image && !failed ? <img src={image} alt={`${name}画风预览`} onError={() => setFailed(true)} />
    : <span className="comic-style-placeholder"><ImageOff size={22} /><small>{failed ? '预览加载失败' : src ? '加载预览…' : '暂无预览图'}</small></span>;
}

export function ComicStylePicker({ value, onChange }) {
  const [open, setOpen] = useState(false);
  const [styles, setStyles] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [query, setQuery] = useState('');
  const [attempt, setAttempt] = useState(0);
  useEffect(() => {
    if (!open) return undefined;
    let alive = true;
    setLoading(true); setError('');
    listComicScreenplayStyles().then(rows => { if (alive) setStyles(rows); })
      .catch(err => { if (alive) { setStyles([]); setError(err.message || '画风库读取失败'); } })
      .finally(() => { if (alive) setLoading(false); });
    return () => { alive = false; };
  }, [open, attempt]);
  const selected = styles.find(item => item.id === value.selectedStyle);
  const filtered = styles.filter(item => item.name.toLowerCase().includes(query.trim().toLowerCase()));
  return <>
    <div className="comic-style-current"><button type="button" className="style-picker-trigger" onClick={() => setOpen(true)}><Palette size={17} /><span>{value.artStyle || '点击选择画风'}</span><span className="comic-style-change">{value.artStyle ? '更换' : '选择'}</span></button>
      {value.artStyle ? <Button type="text" size="small" aria-label="清除画风选择" title="清除画风选择" icon={<X size={15} />} onClick={() => onChange({ artStyle: '', selectedStyle: '' })} /> : null}</div>
    <Modal className="screenplay-style-modal" title="选择画风" open={open} onCancel={() => setOpen(false)} footer={<Button onClick={() => setOpen(false)}>完成</Button>} width={780}>
      <p className="comic-style-caption">选择作品的视觉风格，应用于资产设定与分镜描述。</p>
      <Input prefix={<Search size={15} />} value={query} onChange={event => setQuery(event.target.value)} placeholder="搜索画风名称" allowClear />
      {value.artStyle ? <div className="comic-style-selection"><Check size={14} />当前选择：{value.artStyle}</div> : null}
      {!loading && !error && value.selectedStyle && !selected ? <Alert type="warning" showIcon message="原画风不在当前可用库中，请重新选择后保存。" /> : null}
      {error ? <Alert type="error" showIcon message={error} action={<Button size="small" onClick={() => setAttempt(current => current + 1)}>重试</Button>} /> : null}
      <Spin spinning={loading}><div className="comic-style-library">{filtered.map(style => <button type="button" key={style.id} aria-pressed={value.selectedStyle === style.id} className={value.selectedStyle === style.id ? 'is-active' : ''} onClick={() => { onChange({ artStyle: style.name, selectedStyle: style.id }); setOpen(false); }}>
        <StylePreview src={style.previewUrl} name={style.name} /><div><b>{style.name}</b><span className="comic-choice-check">{value.selectedStyle === style.id ? <Check size={13} /> : null}</span></div>
      </button>)}</div>{!loading && !error && !filtered.length ? <Empty description={query ? '没有匹配的画风' : '暂无可用画风，请先在预设库添加画风'} /> : null}</Spin>
    </Modal>
  </>;
}

export function ComicSourceInput({ mode, onModeChange, content, onContentChange, filename, parsing, onFile, onRemove }) {
  return <section className="comic-source-input" aria-label="原文内容">
    <div className="comic-source-heading"><b>原文内容 <i>*</i></b><span>小说正文 / 已有剧本 / 故事大纲</span></div>
    <div className="comic-source-tabs" role="group" aria-label="内容输入方式">
      <button type="button" aria-pressed={mode === 'text'} className={mode === 'text' ? 'is-active' : ''} onClick={() => onModeChange('text')}><FileText size={15} />直接输入</button>
      <button type="button" aria-pressed={mode === 'file'} className={mode === 'file' ? 'is-active' : ''} onClick={() => onModeChange('file')}><UploadCloud size={15} />上传文件</button>
    </div>
    {mode === 'file' ? <Upload.Dragger accept=".txt,.docx" disabled={parsing} beforeUpload={file => { void onFile(file); return false; }} showUploadList={false}><UploadCloud size={23} /><p>{parsing ? '正在读取文件…' : filename || '点击或拖入 TXT / DOCX 文件'}</p><small>导入后可在下方预览并修改内容</small></Upload.Dragger> : null}
    {filename ? <div className="comic-source-file"><FileText size={14} /><span>{filename}</span><Button type="text" size="small" onClick={onRemove}>移除文件</Button></div> : null}
    <Input.TextArea aria-label="原文内容" value={content} onChange={event => onContentChange(event.target.value)} rows={6} placeholder="在这里输入或粘贴内容，也可以切换到上传文件。" />
    <div className="comic-source-count"><span>原文按你的编辑内容保存</span><span>{[...content].length.toLocaleString()} / 600,000 字</span></div>
  </section>;
}
