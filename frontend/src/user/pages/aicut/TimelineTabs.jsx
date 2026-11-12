import { Button, Dropdown, Input, Modal } from 'antd';
import { MoreHorizontal, Plus } from 'lucide-react';
import { useState } from 'react';

export default function TimelineTabs({ items, activeId, busy, onSelect, onCreate, onRename, onDuplicate, onDelete, onReorder }) {
  const [dialog, setDialog] = useState(null);
  const [title, setTitle] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [dragId, setDragId] = useState('');
  const open = (action, item) => { setDialog({ action, item }); setTitle(action === 'create' ? `时间线 ${items.length + 1}` : action === 'duplicate' ? `${item.title} 副本` : item.title); };
  const submit = async () => {
    setSubmitting(true);
    try {
      if (dialog.action === 'create') await onCreate(title.trim());
      else if (dialog.action === 'rename') await onRename(dialog.item.id, title.trim());
      else if (dialog.action === 'duplicate') await onDuplicate(dialog.item.id, title.trim());
      else await onDelete(dialog.item.id);
      setDialog(null);
    } catch { /* The caller displays the concrete error and keeps the dialog open. */ }
    finally { setSubmitting(false); }
  };
  return <><div className="aicut-timeline-tabs" role="tablist" aria-label="工程时间线">
    {items.map(item => <div key={item.id} className={`aicut-timeline-tab${activeId === item.id ? ' active' : ''}`} draggable={!busy} onDragStart={() => setDragId(item.id)} onDragOver={event => event.preventDefault()} onDrop={event => {
      event.preventDefault(); if (!dragId || dragId === item.id || busy) return;
      const ids = items.map(value => value.id).filter(id => id !== dragId); ids.splice(ids.indexOf(item.id), 0, dragId); setDragId(''); onReorder(ids);
    }} onDragEnd={() => setDragId('')}>
      <button type="button" role="tab" aria-selected={activeId === item.id} disabled={busy} title={item.title} onClick={() => onSelect(item.id)}>{item.title}</button>
      <Dropdown trigger={['click']} menu={{ items: [{ key: 'rename', label: '重命名', disabled: item.default, onClick: () => open('rename', item) }, { key: 'duplicate', label: '复制时间线', disabled: items.length >= 50, onClick: () => open('duplicate', item) }, { key: 'delete', label: '删除时间线', danger: true, disabled: item.default, onClick: () => open('delete', item) }] }}><button type="button" disabled={busy} aria-label={`${item.title}的时间线菜单`}><MoreHorizontal size={12} /></button></Dropdown>
    </div>)}
    <Button type="text" size="small" aria-label="新增时间线" title="新增时间线" icon={<Plus size={13} />} disabled={busy || items.length >= 50} onClick={() => open('create')} />
  </div>
  {dialog ? <Modal open title={dialog.action === 'delete' ? '删除时间线' : dialog.action === 'create' ? '新增时间线' : dialog.action === 'duplicate' ? '复制时间线' : '重命名时间线'} onCancel={() => !submitting && setDialog(null)} onOk={submit} confirmLoading={submitting} okText={dialog.action === 'delete' ? '删除' : '确定'} okButtonProps={{ danger: dialog.action === 'delete', disabled: dialog.action !== 'delete' && !title.trim() }}>
    {dialog.action === 'delete' ? <p>删除“{dialog.item.title}”及其编辑历史。当前工程的原素材和其他时间线会保留，此操作无法撤销。</p> : <Input aria-label="时间线名称" maxLength={80} value={title} onChange={event => setTitle(event.target.value)} onPressEnter={() => !submitting && title.trim() && submit()} />}
  </Modal> : null}</>;
}
