import { Button, Drawer, Empty, Input, Select, Tag, message } from 'antd';
import { FolderPlus, LibraryBig, RefreshCw } from 'lucide-react';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { createAgentMaterialLibraryFolder, insertAgentMaterialLibraryItem, listAgentAssets, listAgentMaterialLibraryFolders, listAgentMaterialLibraryItems, saveAgentMaterialLibraryItem } from '../../shared/api/agent';
import { materialLibraryRows } from '../pages/agent-material-library';

const kinds = [{ value: '', label: '全部类型' }, { value: 'image', label: '图片' }, { value: 'video', label: '视频' }, { value: 'audio', label: '音频' }];

export function AgentMaterialLibraryDrawer({ open, onClose, currentProjectId, onInserted }) {
  const [folders, setFolders] = useState([]);
  const [folderId, setFolderId] = useState('');
  const [items, setItems] = useState([]);
  const [projectAssets, setProjectAssets] = useState([]);
  const [filters, setFilters] = useState({ kind: '', query: '' });
  const [newFolderName, setNewFolderName] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [busyKey, setBusyKey] = useState('');

  const loadFolders = useCallback(async () => {
    const result = await listAgentMaterialLibraryFolders();
    const next = Array.isArray(result?.folders) ? result.folders : [];
    setFolders(next);
    setFolderId(current => next.some(folder => folder.id === current) ? current : String(next[0]?.id || ''));
  }, []);
  const load = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      await loadFolders();
      const assets = currentProjectId ? await listAgentAssets(currentProjectId) : { assets: [] };
      setProjectAssets(Array.isArray(assets?.assets) ? assets.assets : []);
    } catch (caught) {
      setError(caught?.message || '素材库读取失败');
    } finally {
      setLoading(false);
    }
  }, [currentProjectId, loadFolders]);
  const loadItems = useCallback(async () => {
    if (!folderId) {
      setItems([]);
      return;
    }
    try {
      const result = await listAgentMaterialLibraryItems(folderId, filters);
      setItems(Array.isArray(result?.items) ? result.items : []);
    } catch (caught) {
      setError(caught?.message || '素材库素材读取失败');
    }
  }, [folderId, filters]);
  useEffect(() => { if (open) void load(); }, [open, load]);
  useEffect(() => { if (open && folderId) void loadItems(); }, [open, folderId, loadItems]);
  const rows = useMemo(() => materialLibraryRows(items), [items]);
  const createFolder = async () => {
    const name = newFolderName.trim();
    if (!name) return;
    setBusyKey('folder');
    try {
      const result = await createAgentMaterialLibraryFolder(name);
      setNewFolderName('');
      await loadFolders();
      setFolderId(result?.folder?.id || '');
    } catch (caught) {
      message.error(caught?.message || '新建文件夹失败');
    } finally { setBusyKey(''); }
  };
  const saveAsset = async asset => {
    if (!folderId || !currentProjectId) return;
    setBusyKey(`save:${asset.id}`);
    try {
      await saveAgentMaterialLibraryItem(folderId, { projectId: currentProjectId, assetId: asset.id, tags: [] });
      message.success('已保存到素材库');
      await loadItems();
    } catch (caught) { message.error(caught?.message || '保存素材失败'); } finally { setBusyKey(''); }
  };
  const insert = async row => {
    if (!currentProjectId) return;
    setBusyKey(`insert:${row.id}`);
    try {
      const result = await insertAgentMaterialLibraryItem(currentProjectId, row.id);
      message.success('已放入当前画布');
      onInserted?.(result);
    } catch (caught) { message.error(caught?.message || '放入画布失败'); } finally { setBusyKey(''); }
  };

  return <Drawer title={<span className="agent-material-library-drawer__title"><LibraryBig size={17} />素材库</span>} open={open} onClose={onClose} width={460} className="agent-material-library-drawer" extra={<Button type="text" aria-label="刷新素材库" icon={<RefreshCw size={16} />} loading={loading} onClick={() => void load()} />}>
    <section className="agent-material-library-drawer__folders">
      <Input value={newFolderName} maxLength={120} placeholder="新建文件夹" onChange={event => setNewFolderName(event.target.value)} onPressEnter={() => void createFolder()} />
      <Button icon={<FolderPlus size={15} />} loading={busyKey === 'folder'} onClick={() => void createFolder()}>新建</Button>
    </section>
    <Select value={folderId || undefined} placeholder="选择素材库文件夹" options={folders.map(folder => ({ value: folder.id, label: folder.name }))} onChange={setFolderId} />
    <section className="agent-material-library-drawer__section">
      <header><strong>当前项目素材</strong><small>保存当前素材</small></header>
      {!projectAssets.length ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="当前项目还没有素材" /> : <div className="agent-material-library-drawer__project-assets">{projectAssets.map(asset => <article key={asset.id}><span>{asset.name}</span><Tag>{asset.kind === 'image' ? '图片' : asset.kind === 'video' ? '视频' : '音频'}</Tag><Button size="small" disabled={!folderId} loading={busyKey === `save:${asset.id}`} onClick={() => void saveAsset(asset)}>保存</Button></article>)}</div>}
    </section>
    <section className="agent-material-library-drawer__section">
      <header><strong>已保存素材</strong><small>来源项目可追溯</small></header>
      <div className="agent-material-library-drawer__filters"><Input.Search allowClear placeholder="搜索名称或标签" value={filters.query} onChange={event => setFilters(current => ({ ...current, query: event.target.value }))} /><Select value={filters.kind} options={kinds} onChange={kind => setFilters(current => ({ ...current, kind }))} /></div>
      {error ? <p className="agent-material-library-drawer__error" role="alert">{error}</p> : null}
      {!loading && folderId && !rows.length ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="这个文件夹还没有素材" /> : null}
      <div className="agent-material-library-drawer__list">{rows.map(row => <article key={row.id}><div><strong title={row.name}>{row.name}</strong><Tag>{row.kind === 'image' ? '图片' : row.kind === 'video' ? '视频' : '音频'}</Tag></div><small>来源项目：{row.sourceLabel.replace('项目 ', '')}</small><small>来源素材：{row.sourceAssetId} · 版本 v{row.version}</small>{row.tags.length ? <span>{row.tags.map(tag => <Tag key={tag}>{tag}</Tag>)}</span> : null}<Button size="small" type="primary" loading={busyKey === `insert:${row.id}`} disabled={!currentProjectId} onClick={() => void insert(row)}>放入当前画布</Button></article>)}</div>
    </section>
  </Drawer>;
}
