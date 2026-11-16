import { Button, Checkbox, Input, message } from 'antd';
import { useEffect, useState } from 'react';
import { novelProjectRequest } from '../../../shared/api/novelCreation.js';
import useWorkspaceScope from './useWorkspaceScope.js';
export default function NotesPanel({ projectId, selectedIds, onSelect }) {
  const scope = useWorkspaceScope(projectId); const [notes, setNotes] = useState([]); const [edit, setEdit] = useState({ title: '', content: '' });
  const request = (path, body, method) => novelProjectRequest(projectId, path, body, { signal: scope.signal, ...(method ? { method } : {}) });
  const safe = fn => !scope.isCurrent() ? Promise.resolve() : fn().catch(e => { if (scope.isCurrent()) message.error(e.message); });
  useEffect(() => { safe(() => scope.call(request('/notes'), r => setNotes(r.notes || []))); }, [projectId]);
  return <div className="novel-panel"><p>选中的备忘录会作为本次创作的资料保存。</p>
    <Input aria-label="备忘录标题" placeholder="标题" value={edit.title} onChange={e => setEdit(old => ({ ...old, title: e.target.value }))} />
    <Input.TextArea aria-label="备忘录内容" rows={4} value={edit.content} onChange={e => setEdit(old => ({ ...old, content: e.target.value }))} />
    <div className="novel-actions"><Button disabled={!edit.content.trim()} onClick={() => safe(() => scope.call(request(`/notes${edit.id ? '/' + edit.id : ''}`, { title: edit.title, content: edit.content }, edit.id ? 'PUT' : 'POST'), r => { setNotes(old => edit.id ? old.map(n => n.id === edit.id ? r.note : n) : [...old, r.note]); setEdit({ title: '', content: '' }); }))}>{edit.id ? '保存备忘录' : '添加备忘录'}</Button>{edit.id && <Button onClick={() => setEdit({ title: '', content: '' })}>取消编辑</Button>}</div>
    {notes.map(note => <section key={note.id} className="novel-item"><Checkbox checked={selectedIds.includes(note.id)} onChange={e => onSelect(e.target.checked ? [...selectedIds, note.id] : selectedIds.filter(id => id !== note.id))}>{note.title}</Checkbox><p>{note.content}</p><div className="novel-actions"><Button onClick={() => setEdit({ ...note })}>编辑</Button><Button danger onClick={() => safe(() => scope.call(request('/notes/' + note.id, undefined, 'DELETE'), () => { setNotes(old => old.filter(n => n.id !== note.id)); onSelect(selectedIds.filter(id => id !== note.id)); }))}>删除备忘录</Button></div></section>)}
  </div>;
}
