import { useEffect, useRef, useState } from 'react';
import OriginalIcon from './ComicScreenplayIcon.jsx';
import AssetImage from './ComicScreenplayImage.jsx';
import './comic-screenplay-asset-library-picker.css';

const labels = { character: '角色', scene: '场景', prop: '道具' };
const imageUrl = entry => entry.image?.url || entry.thumbnail || (typeof entry.image === 'string' ? entry.image : '') || entry.thumbnail_url || entry.file_path || '';

// This dialog only returns a choice. Importing or saving belongs to its caller.
export default function ComicScreenplayAssetLibraryPicker({ type = 'character', variant = 'binding', embedded = false, disabled = false, loadEntries, onSelect, onClose }) {
  const [entries, setEntries] = useState([]), [loading, setLoading] = useState(true), [error, setError] = useState('');
  const [query, setQuery] = useState(''), [category, setCategory] = useState(''), [selected, setSelected] = useState(null), [attempt, setAttempt] = useState(0), [closed, setClosed] = useState(false);
  const finished = useRef(false), generation = useRef(0);
  const libraryType = variant === 'legacy-add' ? 'character' : type;
  const binding = variant === 'binding', hasCategories = !binding && libraryType === 'character';
  useEffect(() => {
    let active = true;
    const request = ++generation.current;
    setLoading(true); setError(''); setSelected(null);
    Promise.resolve().then(() => loadEntries({ type: libraryType })).then(rows => {
      if (!active || finished.current || request !== generation.current) return;
      if (!Array.isArray(rows)) throw new Error('素材库返回格式不正确，请重新读取');
      setEntries(rows);
    }).catch(e => {
      if (active && !finished.current && request === generation.current) setError(e?.message || '素材库读取失败');
    }).finally(() => {
      if (active && !finished.current && request === generation.current) setLoading(false);
    });
    return () => { active = false; };
  }, [libraryType, loadEntries, attempt]);
  useEffect(() => {
    if (embedded) return;
    const keydown = event => { if (event.key === 'Escape') { event.stopImmediatePropagation(); close(); } };
    document.addEventListener('keydown', keydown);
    return () => document.removeEventListener('keydown', keydown);
  });
  function close() {
    if (finished.current) return;
    finished.current = true; setClosed(true); onClose?.();
  }
  function choose(entry) {
    if (finished.current || disabled || loading || error) return;
    if (embedded) { setSelected(entry); onSelect?.(entry); return; }
    finished.current = true; setClosed(true);
    try { onSelect?.(entry); } finally { onClose?.(); }
  }
  const categories = [...new Set(entries.map(entry => entry.category).filter(Boolean))].sort();
  const search = query.trim().toLowerCase();
  const filtered = entries.filter(entry => (!hasCategories || !category || entry.category === category) && (!search || [entry.name, entry.description].some(value => String(value || '').toLowerCase().includes(search))));
  const visible = binding ? filtered : filtered.slice(0, 60);
  const title = binding ? `选择${labels[type] || '资产'}` : variant === 'legacy-add' || type === 'character' ? '从角色库选择' : `更换${labels[type]}图片`;
  if (closed) return null;
  return <div className={embedded ? 'comic-library-embedded' : 'comic-library-overlay'} onClick={event => { if (!embedded && event.target === event.currentTarget) close(); }}>
    <section role={embedded ? 'region' : 'dialog'} aria-modal={embedded ? undefined : 'true'} aria-label={title} className={`comic-library-picker is-${variant} type-${libraryType}`} style={embedded ? {width:'100%',maxWidth:'100%',maxHeight:'none',boxShadow:'none',borderRadius:0,background:'transparent'} : undefined}>
      {!embedded ? <header><strong>{title}</strong><button type="button" aria-label="关闭普通库" onClick={close}><OriginalIcon name="CloseOutline" size={18} /></button></header> : <small>库来源：角色库</small>}
      <div className="comic-library-content" style={embedded?{padding:'8px 0 0'}:undefined}>
        <div className="comic-library-filters"><label className="comic-library-search" style={{flexDirection:'row'}}><OriginalIcon name="SearchOutline" size={14} /><input disabled={disabled} aria-label="搜索普通库" placeholder="搜索..." value={query} onChange={event => setQuery(event.target.value)} /></label>{hasCategories ? <select disabled={disabled} aria-label="普通库分类" value={category} onChange={event => setCategory(event.target.value)}><option value="">全部</option>{categories.map(value => <option key={value} value={value}>{value}</option>)}</select> : null}</div>
        {loading ? <p role="status">加载中...</p> : error ? <div role="alert"><p>{error}</p><button type="button" disabled={disabled} onClick={() => setAttempt(value => value + 1)}>重新读取</button></div> : visible.length ? <div className="comic-library-grid">{visible.map((entry, index) => <button type="button" disabled={disabled} key={entry.id || index} className={`comic-library-card${selected === entry ? ' is-selected' : ''}`} aria-pressed={selected === entry} onClick={() => binding || embedded ? choose(entry) : setSelected(entry)} onKeyDown={event => { if (event.key === 'Enter' || event.key === ' ') { event.preventDefault(); if (binding || embedded) choose(entry); else setSelected(entry); } }}><div className="comic-library-thumb"><AssetImage src={imageUrl(entry)} name={entry.name || '未命名'} placeholderIcon={libraryType === 'character' ? 'PeopleOutline' : 'ImageOutline'} placeholderSize={16} /></div><span>{entry.name || '未命名'}</span></button>)}</div> : <p className="comic-library-empty">{binding ? '暂无可绑定资产' : libraryType === 'character' ? '角色库为空' : `${labels[type]}资产库为空`}</p>}
      </div>
      {!embedded ? <footer><button type="button" onClick={close}>取消</button>{!binding ? <button type="button" className="comic-library-confirm" disabled={disabled || !selected || loading || !!error} onClick={() => choose(selected)}>{variant === 'replace' ? '确认更换' : '保存'}</button> : null}</footer> : null}
    </section>
  </div>;
}
