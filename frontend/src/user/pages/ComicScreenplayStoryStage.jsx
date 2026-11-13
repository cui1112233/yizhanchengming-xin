import { Alert, Button, Empty, Input, message, Popconfirm, Space, Upload } from 'antd';
import { ChevronDown, GripVertical, Pencil, Plus, UploadCloud } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { comicEpisodesDocument, comicScriptEpisodes, reorderComicEpisodes } from './comic-screenplay-stages.js';

export default function ComicScreenplayStoryStage({ project, document, draft, onDraft, onSave, onTools, readOnly, working }) {
  const episodes = useMemo(() => comicScriptEpisodes(draft), [draft]);
  const [collapsed, setCollapsed] = useState({}), [edit, setEdit] = useState(null), [dragFrom, setDragFrom] = useState(-1);
  useEffect(() => { setEdit(null); setCollapsed({}); }, [document?.id]);
  const dirty = draft !== (document?.content || '');
  async function commit(index, value) {
    const rows = episodes.map((item, no) => no === index ? { ...item, ...value, raw: false } : item);
    const content = comicEpisodesDocument(rows); onDraft(content);
    if (await onSave(content)) setEdit(null);
  }
  async function importScript(file) {
    try { const content = await file.text(); if (!content.trim()) throw Error('剧本文件没有正文'); onDraft(content); await onSave(content); }
    catch (error) { message.error(error.message || '剧本导入失败'); }
    return false;
  }
  return <div className="comic-story-stage">
    <div className="comic-stage-title-row"><h1>{project.title || '未命名项目'}</h1><Space wrap><Button onClick={() => onTools('source')}>原文与剧本工具</Button><Button onClick={() => onTools('chapters')}>章节规划</Button><Button onClick={() => onTools('history')}>版本历史</Button></Space></div>
    <p className="comic-story-synopsis">{project.synopsis || '还没有故事概要。可在模型配置中补充，或使用作品 Agent 讨论故事。'}</p>
    <div className="comic-story-count">共 {episodes.length} 集</div>
    {!document?.content ? <Alert type="info" showIcon message="尚未保存剧本" description="原文已保存不等于剧本已生成。可打开原文与剧本工具明确生成剧本，或直接导入已有剧本；切换阶段不会自动调用模型。" /> : null}
    {dirty ? <Alert type="warning" showIcon message="本地修改尚未保存为剧本版本" action={<Button disabled={readOnly || !!working} onClick={() => onSave(draft)}>保存剧本修改</Button>} /> : null}
    <section className="comic-episodes-card"><header><div><ChevronDown size={17} /><h2>分集剧本</h2><Button size="small" type="text" onClick={() => setCollapsed(Object.fromEntries(episodes.map((_, i) => [i, true])))}>收起全部</Button><Button size="small" type="text" onClick={() => setCollapsed({})}>展开全部</Button></div><Space wrap><Button disabled={readOnly || !!working || !!edit} icon={<Plus size={14} />} onClick={() => { const rows = [...episodes, { no: episodes.length+1, title: `第${episodes.length+1}集`, content: '', raw: false }]; onDraft(comicEpisodesDocument(rows)); setEdit({ index: rows.length-1, title: rows.at(-1).title, content: '', originalDraft: draft }); }}>新增一集</Button><Upload disabled={readOnly || !!working} accept=".txt,.md" showUploadList={false} beforeUpload={importScript}><Button disabled={readOnly || !!working} icon={<UploadCloud size={14} />}>导入剧本</Button></Upload></Space></header>
      {episodes.map((episode, index) => <article key={index} className={`comic-episode-item${dragFrom === index ? ' is-dragging' : ''}`} onDragOver={event => { if (!readOnly) event.preventDefault(); }} onDrop={event => { event.preventDefault(); if (dragFrom < 0 || edit) return; onDraft(comicEpisodesDocument(reorderComicEpisodes(episodes, dragFrom, index))); setDragFrom(-1); }}>
        <div className="comic-episode-heading"><button type="button" className="comic-episode-drag" disabled={readOnly || !!edit || !!working} draggable={!readOnly && !edit && !working} onDragStart={event => { event.dataTransfer.effectAllowed = 'move'; event.dataTransfer.setData('text/plain', String(index)); setDragFrom(index); }} onDragEnd={() => setDragFrom(-1)} title="拖动调整顺序"><GripVertical size={16} /></button><button type="button" className="comic-episode-toggle" aria-label={`展开或收起第${index+1}集`} aria-expanded={!collapsed[index]} onClick={() => setCollapsed(current => ({ ...current, [index]: !current[index] }))}><ChevronDown size={14} style={{ transform: collapsed[index] ? 'rotate(-90deg)' : undefined }} /></button><span className="comic-episode-number">{index+1}.</span>{edit?.index === index ? <Input aria-label="集名" value={edit.title} onChange={event => setEdit(current => ({ ...current, title: event.target.value }))} /> : <b className="comic-episode-name" onDoubleClick={() => { if (!readOnly) setEdit({ index, title: episode.title, content: episode.content }); }}>{episode.title}</b>}
        {edit?.index === index ? <Space><Button size="small" loading={working === 'document-save'} disabled={readOnly || !!working} onClick={() => commit(index, edit)}>保存</Button><Button size="small" disabled={!!working} onClick={() => { if(edit.originalDraft !== undefined) onDraft(edit.originalDraft); setEdit(null); }}>取消</Button></Space> : <button type="button" className="comic-episode-edit" title="编辑本集（可改集名与正文）" disabled={readOnly || !!working || !!edit} onClick={() => { setEdit({ index, title: episode.title, content: episode.content }); setCollapsed(current => ({ ...current, [index]: false })); }}><Pencil size={15} /></button>}</div>
        {!collapsed[index] ? edit?.index === index ? <Input.TextArea aria-label="本集正文" rows={16} value={edit.content} onChange={event => setEdit(current => ({ ...current, content: event.target.value }))} /> : <div className="comic-episode-body">{episode.content || '本集还没有正文，点击铅笔编辑后保存。'}</div> : null}
      </article>)}
      {!episodes.length ? <Empty description="还没有剧本分集，可导入已有剧本或明确生成剧本"><Button type="primary" onClick={() => onTools('source')}>生成剧本 / 输入已有剧本</Button></Empty> : null}
    </section>
  </div>;
}
