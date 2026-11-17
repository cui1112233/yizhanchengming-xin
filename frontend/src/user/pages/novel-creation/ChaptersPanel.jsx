import { Button, Input, Modal, Select, message } from 'antd';
import { useEffect, useState } from 'react';
import { novelProjectRequest, novelRequest } from '../../../shared/api/novelCreation.js';
import useWorkspaceScope from './useWorkspaceScope.js';
import { downloadBlob, validRun } from './contracts.js';

export default function ChaptersPanel({ projectId, sessionId, chapters, run, canExportShort, onChapter, onMaintenance, onBatch }) {
  const scope = useWorkspaceScope(projectId + ':' + sessionId); const [selected, setSelected] = useState(''); const [edit, setEdit] = useState(null);
  const [versions, setVersions] = useState(null); const [continuity, setContinuity] = useState(null); const [format, setFormat] = useState('md'); const [error, setError] = useState('');
  const chapter = chapters.find(c => c.id === selected) || chapters[0];
  const savedShort = canExportShort === true && !!validRun(run) && ['succeeded', 'succeeded_with_warnings'].includes(run.status);
  const request = (path, body, options = {}) => novelProjectRequest(projectId, path, body, { signal: scope.signal, ...options });
  const safe = async fn => { if (!scope.isCurrent()) return; try { await fn(); } catch (e) { if (scope.isCurrent()) { setError(e.message); message.error(e.message); } } };
  function refreshContinuity() { return scope.call(request('/continuity'), setContinuity); }
  useEffect(() => { safe(refreshContinuity); }, [projectId, chapters]);
  async function exportSaved(short = false, persisted = false) {
    if (short && !savedShort) return;
    await safe(() => scope.call(persisted ? request('/exports', { format }) : short
      ? novelRequest(`/runs/${run.id}/short-result/export?format=${format}`, undefined, { responseType: 'blob', signal: scope.signal })
      : request(`/export?format=${format}`, undefined, { responseType: 'blob' }), r => {
      if (persisted) { if (r.downloadUrl) window.open(r.downloadUrl, '_blank', 'noopener'); }
      else downloadBlob(r, `${short ? '短篇' : '小说'}.${format}`);
    }));
  }
  return <div className="novel-panel">
    <div className="novel-actions"><b>已保存 {chapters.length} 章</b><Button size="small" onClick={onBatch}>批量续写</Button></div>
    {continuity?.dirtyFrom > 0 && <div className="novel-notice" role="status">从第{continuity.dirtyFrom}章起事实与记忆待刷新。<Button onClick={() => onMaintenance('continuity_refresh')}>刷新事实与记忆</Button></div>}
    <div className="novel-actions"><Select aria-label="导出格式" value={format} onChange={setFormat} options={[{ value: 'txt', label: 'TXT' }, { value: 'md', label: 'Markdown' }]} /><Button onClick={() => exportSaved()}>导出整本</Button>{savedShort && <Button onClick={() => exportSaved(true)}>导出本次短篇</Button>}</div>
    <small>只导出已保存正文。{chapters.length === 0 ? '暂无可导出正文' : ''}</small>{error && <p role="alert">{error}</p>}
    <div className="novel-chapter-list">{chapters.map(c => <Button block type={chapter?.id === c.id ? 'primary' : 'default'} key={c.id} onClick={() => setSelected(c.id)}>{c.title} · 版本{c.version}</Button>)}</div>
    {chapter && <article className="novel-chapter"><h3>{chapter.title}</h3><small>版本 {chapter.version} · {chapter.wordCount ?? chapter.content?.length ?? 0} 字</small><div className="novel-actions"><Button onClick={() => { setError(''); setEdit({ id: chapter.id, expectedVersion: chapter.version, title: chapter.title, content: chapter.content }); }}>编辑正文</Button><Button onClick={() => safe(() => scope.call(request(`/chapters/${chapter.id}/versions`), r => setVersions({ chapterId: chapter.id, expectedVersion: chapter.version, entries: r.versions || [] })))}>版本历史</Button></div><p>{chapter.content}</p></article>}
    <Modal open={!!edit} title="编辑已保存正文" okText="保存新版本" cancelText="取消" width={720} onCancel={() => setEdit(null)} onOk={() => safe(() => scope.call(request(`/chapters/${edit.id}`, { expectedVersion: edit.expectedVersion, title: edit.title, content: edit.content }, { method: 'PUT' }), r => { onChapter(r.chapter); setEdit(null); refreshContinuity(); }))}>
      <p>修改会生成新版本；版本冲突时保留此草稿，请先核对最新正文。</p>{error && <p role="alert">{error}</p>}<Input aria-label="章节标题" value={edit?.title} onChange={e => setEdit(old => ({ ...old, title: e.target.value }))} /><Input.TextArea aria-label="章节正文" rows={14} value={edit?.content} onChange={e => setEdit(old => ({ ...old, content: e.target.value }))} />
    </Modal>
    <Modal open={!!versions} title="章节版本历史" footer={null} onCancel={() => setVersions(null)} width={720}>
      <p>恢复会创建一个新版本，原来的历史会保留。</p>{versions?.entries.map(v => <section key={v.version} className="novel-item"><b>版本{v.version} · {v.title}</b><p>{v.content}</p><Button onClick={() => safe(() => scope.call(request(`/chapters/${versions.chapterId}/restore`, { expectedVersion: versions.expectedVersion, version: v.version }), r => { onChapter(r.chapter); setVersions(null); refreshContinuity(); }))}>恢复为新版本</Button></section>)}
    </Modal>
  </div>;
}
