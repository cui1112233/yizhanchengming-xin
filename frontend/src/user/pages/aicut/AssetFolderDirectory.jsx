import { Button, Dropdown, Input, Modal } from 'antd';
import { Folder, MoreHorizontal, Plus } from 'lucide-react';
import { useState } from 'react';
import './asset-folder-directory.css';

export default function AssetFolderDirectory({ folders, assets, selectedId, onSelect, onCreate, onRename, onDelete, onMove }) {
  const [editing, setEditing] = useState(null);
  const [name, setName] = useState('');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const open = folder => { setEditing(folder || { id: '' }); setName(folder?.name || ''); setError(''); };
  const save = async () => {
    if (busy) return;
    if (!name.trim()) { setError('请输入文件夹名称'); return; }
    setBusy(true); setError('');
    try { if (editing.id) await onRename(editing.id, name.trim()); else await onCreate(name.trim()); setEditing(null); }
    catch (e) { setError(e.message || '保存失败，请重试'); } finally { setBusy(false); }
  };
  const row = (id, label, count, folder) => <div key={id || 'root'} className="aicut-folder-row" onDragOver={event => { if (onMove) { event.preventDefault(); event.dataTransfer.dropEffect = 'move'; } }} onDrop={event => { event.preventDefault(); const assetId = event.dataTransfer.getData('application/x-aicut-asset'); if (assetId && onMove) onMove(assetId, id); }}>
    <button type="button" aria-pressed={selectedId === id} className={selectedId === id ? 'active' : ''} onClick={() => onSelect(id)}><Folder size={12} /><span>{label}</span><small>{count}</small></button>
    {folder ? <Dropdown trigger={['click']} menu={{ items: [{ key: 'rename', label: '重命名' }, { key: 'delete', label: '删除文件夹', danger: true }], onClick: ({ key }) => { if (key === 'rename') open(folder); else Modal.confirm({ title: `删除“${folder.name}”？`, content: '其中的素材会移回未分类，素材文件和时间线不会被删除。', okText: '删除文件夹', cancelText: '取消', onOk: () => onDelete(folder.id) }); } }}><button type="button" className="aicut-folder-more" aria-label={`${label}文件夹操作`}><MoreHorizontal size={13} /></button></Dropdown> : null}
  </div>;
  return <div className="aicut-asset-folders">
    <div className="aicut-folder-heading"><strong>我的文件夹</strong><Button type="text" size="small" icon={<Plus size={12} />} aria-label="新建素材文件夹" onClick={() => open(null)} /></div>
    {row('', '未分类', assets.filter(a => !a.folderId).length)}
    {folders.map(folder => row(folder.id, folder.name, assets.filter(a => a.folderId === folder.id).length, folder))}
    <Modal title={editing?.id ? '重命名文件夹' : '新建素材文件夹'} open={Boolean(editing)} onCancel={() => setEditing(null)} onOk={save} confirmLoading={busy} okText="保存" cancelText="取消" destroyOnClose>
      <Input autoFocus maxLength={100} aria-label="素材文件夹名称" value={name} onChange={event => setName(event.target.value)} onPressEnter={save} placeholder="请输入文件夹名称" />
      {error ? <p role="alert">{error}</p> : null}
    </Modal>
  </div>;
}
