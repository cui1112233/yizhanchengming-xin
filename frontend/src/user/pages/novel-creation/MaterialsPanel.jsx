import { Button, Checkbox, Input, Modal, Select, message } from 'antd';
import { useEffect, useRef, useState } from 'react';
import { novelProjectRequest, novelRequest } from '../../../shared/api/novelCreation.js';
import useWorkspaceScope from './useWorkspaceScope.js';

const LABELS = { uploading: '原件上传中', queued: '等待解析', parsing: '正在解析', ready: '可以引用', failed: '解析失败', upload_failed: '上传失败，请重新上传' };
export default function MaterialsPanel({ projectId, selectedIds, onSelect, onReferencesChanged, onSelection, styleMode = false, styleId, onStyle, onMaintenance }) {
  const scope = useWorkspaceScope(projectId); const [docs, setDocs] = useState([]); const [styles, setStyles] = useState([]);
  const [availability, setAvailability] = useState(null); const [title, setTitle] = useState(''); const [content, setContent] = useState('');
  const [picks, setPicks] = useState({}); const [edit, setEdit] = useState(null); const [description, setDescription] = useState('');
  const [busy, setBusy] = useState(false); const busyRef = useRef(false);
  const [ownedStyleId, setOwnedStyleId] = useState('');
  const [connection, setConnection] = useState('');
  const request = (path, input, options = {}) => novelProjectRequest(projectId, path, input, { signal: scope.signal, ...options });
  const applyDocs = r => { setDocs(r.documents || []); onReferencesChanged?.((r.documents || []).filter(d => d.status === 'ready' && d.selection).map(d => d.selection)); };
  async function load() { await scope.call(request('/documents'), applyDocs); await scope.call(request('/styles'), r => setStyles(r.styles || [])); }
  useEffect(() => {
    let timer, failures = 0;
    async function poll() {
      try { await load(); if (scope.isCurrent()) { failures = 0; setConnection(''); } } catch { if (scope.isCurrent()) { failures++; setConnection('资料连接暂时中断，正在重新连接；已选资料保留。'); } }
      if (scope.isCurrent()) timer = setTimeout(poll, Math.min(4000 * 2 ** failures, 30000));
    }
    scope.call(novelRequest('/documents/availability', undefined, { signal: scope.signal }), setAvailability).catch(e => message.error(e.message));
    poll(); return () => clearTimeout(timer);
  }, [projectId]);
  async function action(fn) {
    if (busyRef.current || !scope.isCurrent()) return; busyRef.current = true; setBusy(true);
    try { await fn(); } catch (e) { if (scope.isCurrent()) message.error(e.message); }
    finally { busyRef.current = false; if (scope.isCurrent()) setBusy(false); }
  }
  async function upload(files) {
    if (!files.length) return;
    const maxFiles = Math.min(8, availability?.maxBatchFiles ?? 8), maxBytes = Math.min(20 * 1024 * 1024, availability?.maxFileBytes ?? 20 * 1024 * 1024);
    if (files.length > maxFiles || files.some(f => f.size > maxBytes || !/\.(txt|docx|pdf)$/i.test(f.name))) { message.warning('每批最多8个TXT、DOCX或PDF文件，单个不超过20MiB'); return; }
    await action(async () => {
      const data = new FormData(); files.forEach(f => data.append('files', f));
      await scope.call(request('/documents/upload', data), r => { setDocs(old => [...(r.documents || []), ...old]); if (r.failures?.length) message.warning(r.failures.map(f => f.error || f.message || f.filename).join('；')); });
    });
  }
  async function selectChapters(doc) {
    const ids = picks[doc.id] ?? doc.selection?.selectedChapterIds ?? [];
    const size = ids.reduce((n, id) => n + (doc.chapters.find(c => c.id === id)?.characters ?? 0), 0);
    if (!ids.length || size > 12000) { message.warning('每份参考请选择正文，总计不超过12000字'); return; }
    await action(() => scope.call(request(`/documents/${doc.id}/selection`, { expectedVersion: doc.version, chapterIds: ids }), r => {
      setDocs(old => old.map(d => d.id === doc.id ? { ...d, selection: r.selection } : d));
      onReferencesChanged?.([r.selection]); onSelection?.(r.selection); onSelect([...new Set([...selectedIds, r.selection.id])]); message.success('已保存参考范围');
    }));
  }
  return <div className="novel-panel">
    {connection && <p role="status">{connection}</p>}
    <p>{styleMode ? '纯统计不会调用模型；AI分析需要单独点击。可以选择当前账号已保存的文风。' : '选择章节后保存参考范围；每份参考最多12000字，可按勾选顺序排列。只有解析成功且有正文的资料可以引用。'}</p>
    {styleMode && <><label>选用文风<Select aria-label="选用文风" allowClear value={styleId || undefined} onChange={id => onStyle(id || '')} options={styles.map(s => ({ value: s.profile.id, label: s.profile.title }))} /></label><Input aria-label="已有文风编号" value={ownedStyleId} onChange={e => setOwnedStyleId(e.target.value)} placeholder="粘贴当前账号在其他作品保存的文风编号" /><Button disabled={!ownedStyleId.trim()} onClick={() => action(() => scope.call(novelRequest(`/style-profiles/${encodeURIComponent(ownedStyleId.trim())}`, undefined, { signal: scope.signal }), r => { setStyles(old => [...old.filter(s => s.profile.id !== r.styleProfile.id), { profile: r.styleProfile, external: true }]); onStyle(r.styleProfile.id); }))}>选用已有文风</Button><Input aria-label="文风描述" value={description} onChange={e => setDescription(e.target.value)} placeholder="文风描述（可选）" /></>}
    <Input aria-label="素材标题" value={title} onChange={e => setTitle(e.target.value)} placeholder="素材标题" />
    <Input.TextArea aria-label="素材内容" rows={5} value={content} onChange={e => setContent(e.target.value)} placeholder="粘贴文本，按章节拆分" />
    <div className="novel-actions"><Button loading={busy} disabled={!content.trim()} onClick={() => action(() => scope.call(request('/documents/paste', { title, content }), r => { setDocs(old => [r.document, ...old]); setContent(''); message.success('文本已解析'); }))}>解析粘贴文本</Button>
      <label className="novel-upload">上传文件<input aria-label="上传素材文件" type="file" multiple accept=".txt,.docx,.pdf" disabled={!availability?.uploadAvailable || busy} onChange={e => { const files = [...e.target.files]; e.target.value = ''; upload(files); }} /></label></div>
    <small>{availability?.uploadAvailable ? '支持TXT、DOCX及可提取文字的PDF。每批8个，单个20MiB；扫描件不支持。' : availability?.reason || '正在检查文件上传能力'}</small>
    {docs.map(doc => {
      const ready = doc.status === 'ready' && doc.chapters?.length > 0;
      const chosen = picks[doc.id] ?? doc.selection?.selectedChapterIds ?? [];
      const total = chosen.reduce((n, id) => n + (doc.chapters?.find(c => c.id === id)?.characters ?? 0), 0);
      return <section className="novel-item" key={doc.id}><b>{doc.filename}</b><span>{LABELS[doc.status] || '处理中'} · 版本 {doc.version}</span>{doc.error && <p role="alert">{doc.error}</p>}
        {doc.emptyPages?.length > 0 && <small>第{doc.emptyPages.join('、')}页未提取到文字</small>}
        {ready && <><div className="novel-document-chapters">{doc.chapters.map(c => <Checkbox key={c.id} checked={chosen.includes(c.id)} onChange={e => setPicks(old => ({ ...old, [doc.id]: e.target.checked ? [...chosen, c.id] : chosen.filter(id => id !== c.id) }))}>{c.title}（{c.characters}字）</Checkbox>)}</div>
          <p>按选择顺序：{chosen.map(id => doc.chapters.find(c => c.id === id)?.title).join(' → ') || '未选择'} · 共 {total} 字 / 12000字</p>
          <div className="novel-actions"><Button disabled={!chosen.length || total > 12000} onClick={() => selectChapters(doc)}>保存参考范围</Button><Button onClick={() => setEdit({ ...doc, chapters: doc.chapters.map(c => ({ ...c })) })}>编辑与预览分段</Button></div>
          {doc.selection && <Checkbox checked={selectedIds.includes(doc.selection.id)} onChange={e => { if (e.target.checked) onSelection?.(doc.selection); onSelect(e.target.checked ? [...new Set([...selectedIds, doc.selection.id])] : selectedIds.filter(id => id !== doc.selection.id)); }}>引用 {doc.selection.characters} 字（来源版本{doc.selection.sourceVersion}）</Checkbox>}
          {styleMode && <Button disabled={!doc.selection} onClick={() => action(() => scope.call(request('/styles', { documentId: doc.id, title: doc.filename, description }), r => { setStyles(old => [r.style, ...old]); onStyle(r.style.profile.id); message.success('文风统计已保存，未调用模型'); }))}>保存文风统计</Button>}
        </>}
        <div className="novel-actions">{doc.canRetry && <Button onClick={() => action(() => scope.call(request(`/documents/${doc.id}/retry`, {}), r => setDocs(old => old.map(d => d.id === doc.id ? r.document : d))))}>重试解析原件</Button>}
          {doc.status === 'uploading' && !doc.canRetry && <small>上传尚在进行，结束后可管理；中断后可刷新查看重试入口。</small>}
          <Button danger disabled={doc.canDelete === false || (!Object.hasOwn(doc, 'canDelete') && ['uploading', 'parsing', 'queued'].includes(doc.status))} onClick={() => action(() => scope.call(request(`/documents/${doc.id}`, undefined, { method: 'DELETE' }), () => { setDocs(old => old.filter(d => d.id !== doc.id)); onSelect(selectedIds.filter(id => id !== doc.selection?.id)); }))}>移除资料</Button></div>
      </section>;
    })}
    {styleMode && styles.map(s => <section className="novel-item" key={s.profile.id}><b>{s.profile.title}</b><small>文风编号：{s.profile.id}</small>{!s.external && <><p>{s.statistics?.characters ?? 0}字 · {s.statistics?.paragraphs ?? 0}段 · 分析版本 {s.analysisVersion ?? 0}</p><p>{s.userDescription}</p><span>{s.status === 'statistics_ready' ? '统计已保存（未调用模型）' : '分析任务：' + (s.status || '等待处理')}</span><Button onClick={() => onMaintenance('style_analysis', s.profile.id)}>AI分析文风</Button></>}</section>)}
    <Modal title="编辑与预览分段" open={!!edit} width={720} okText="保存新版本" cancelText="取消" onCancel={() => setEdit(null)} onOk={() => action(() => scope.call(request(`/documents/${edit.id}/chapters`, { expectedVersion: edit.version, chapters: edit.chapters }, { method: 'PUT' }), r => { setDocs(old => old.map(d => d.id === edit.id ? r.document : d)); onSelect(selectedIds.filter(id => id !== edit.selection?.id)); setEdit(null); }))}>
      {edit?.chapters.map((c, i) => <div className="novel-item" key={c.id}><Input value={c.title} onChange={e => setEdit(old => ({ ...old, chapters: old.chapters.map((v, j) => j === i ? { ...v, title: e.target.value } : v) }))} /><Input.TextArea rows={4} value={c.content} onChange={e => setEdit(old => ({ ...old, chapters: old.chapters.map((v, j) => j === i ? { ...v, content: e.target.value } : v) }))} /></div>)}
    </Modal>
  </div>;
}
