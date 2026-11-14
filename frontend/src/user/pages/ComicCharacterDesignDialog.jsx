import { Modal } from 'antd';
import { useEffect, useState } from 'react';
import OriginalIcon from './ComicScreenplayIcon.jsx';
import recovered from './comic-character-design-catalog.json';
import { characterDesignInitialSelection, characterDesignOptionValue, characterDesignSelectionText, toggleCharacterDesignOption } from './comic-character-design.js';
import './comic-character-design.css';

export default function ComicCharacterDesignDialog({ open, initialSelection, onSave, onCancel, catalog = recovered.catalog }) {
  const [category, setCategory] = useState('basic');
  const [selection, setSelection] = useState(() => characterDesignInitialSelection(catalog, initialSelection));
  useEffect(() => {
    if (open) { setCategory('basic'); setSelection(characterDesignInitialSelection(catalog, initialSelection)); }
  }, [open]);
  const icons={basic:'PersonOutline',makeup:'BrushOutline',features:'PersonOutline',body:'PersonOutline',hair:'BrushOutline',style:'SparklesOutline'};
  const fields = catalog.find(item => item.key === category)?.fields || [];
  const reset = () => setSelection(characterDesignInitialSelection(catalog));
  function save() {
    onSave({ text: characterDesignSelectionText(selection, catalog), selection: characterDesignInitialSelection(catalog, selection) });
  }
  return <Modal open={open} centered width={920} style={{ maxWidth: '92vw' }} footer={null} closable={false} maskClosable={false} onCancel={onCancel} className="comic-character-design-dialog" destroyOnHidden>
    <header className="comic-character-design-header"><h3>角色设计</h3><button type="button" className="comic-character-design-icon-button" aria-label="关闭角色设计" title="关闭" onClick={onCancel}><OriginalIcon name="CloseOutline" size={18} /></button></header>
    <nav className="comic-character-design-tabs" role="tablist" aria-label="角色设计分类">{catalog.map(item => <button key={item.key} type="button" role="tab" aria-selected={category === item.key} className={category === item.key ? 'is-active' : ''} onClick={() => setCategory(item.key)}><OriginalIcon name={icons[item.key]} size={15}/>{item.label}</button>)}<button type="button" className="comic-character-design-icon-button comic-character-design-reset" aria-label="重置角色设计" title="重置" onClick={reset}><OriginalIcon name="RefreshOutline" size={16} /></button></nav>
    <div className="comic-character-design-body">{fields.map(field => <section className="comic-character-design-section" key={field.key}><div className="comic-character-design-label">{field.label}</div><div className={`comic-character-design-options${field.type === 'color' ? ' is-color' : ''}`} role="group" aria-label={field.label}>{field.options.map(option => { const value = characterDesignOptionValue(option), label = typeof option === 'string' ? option : option.label || option.value; const selected = field.multiple ? Array.isArray(selection[field.key]) && selection[field.key].includes(value) : selection[field.key] === value; return <button key={value} type="button" title={label} aria-pressed={selected} className={`comic-character-design-chip${selected ? ' is-active' : ''}${field.type === 'color' ? ' is-color-chip' : ''}`} onClick={() => setSelection(previous => toggleCharacterDesignOption(previous, field, option))}>{field.type === 'color' ? <span className="comic-character-design-swatch" style={{ background: option.value }} aria-hidden="true" /> : null}<span>{label}</span></button>; })}</div></section>)}</div>
    <footer className="comic-character-design-footer"><button type="button" onClick={onCancel}>取消</button><button type="button" className="is-primary" onClick={save}>保存</button></footer>
  </Modal>;
}
