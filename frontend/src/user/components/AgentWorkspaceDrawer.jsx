import { Button, Empty, Input, Select, Spin, Tabs, Tooltip } from 'antd';
import { CircleAlert, Clock3, FileText, Film, History, Image, LocateFixed, Music, Plus, Search, X } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import {
  buildWorkspaceGroupTree,
  filterWorkspaceAssets,
  filterWorkspaceExecutions,
  filterWorkspaceNodes,
  workspaceExecutionKind,
  workspaceDrawerWidth
} from '../pages/agent-workspace-navigation';
import { groupResultLibrary, lineageLabel } from '../pages/agent-result-lineage';
import { AuthenticatedAgentImage } from './AuthenticatedAgentImage';
import { AuthenticatedAgentVideo } from './AuthenticatedAgentVideo';

const kindLabels = { text: '文本', image: '图片', video: '视频', audio: '音频' };
const statusLabels = {
  draft: '草稿', pending: '等待中', queued: '等待中', running: '生成中', processing: '生成中',
  awaiting_confirmation: '等待选择', done: '已完成', completed: '已完成', failed: '失败', cancelled: '已取消'
};

const executionTypeLabels = {
  'text-generation': '文本生成',
  'image-generation': '图片生成',
  'video-generation': '视频生成',
  'audio-generation': '音频生成'
};

function executionLabel(execution) {
  return execution?.name || executionTypeLabels[execution?.toolName || execution?.tool] || execution?.toolName || execution?.tool || '生成任务';
}

function executionTime(execution) {
  const value = execution?.updatedAt || execution?.createdAt;
  if (!value) return '暂无时间';
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? String(value) : date.toLocaleString('zh-CN', { hour12: false });
}

function nodeValue(node, key) {
  return node?.[key] ?? node?.data?.[key];
}

function KindIcon({ kind, size = 14 }) {
  if (kind === 'image') return <Image size={size} />;
  if (kind === 'video') return <Film size={size} />;
  if (kind === 'audio') return <Music size={size} />;
  return <FileText size={size} />;
}

function AssetPreview({ asset }) {
  if (asset?.kind === 'image' && asset?.url) return <AuthenticatedAgentImage source={asset.url} alt="" />;
  if (asset?.kind === 'video' && asset?.url) return <AuthenticatedAgentVideo source={asset.url} aria-label="视频素材缩略图" preload="metadata" muted />;
  return <KindIcon kind={asset?.kind} size={16} />;
}

export function AgentWorkspaceDrawer({
  open,
  width,
  activeTab,
  nodes = [],
  groups = [],
  assets = [],
  executions = [],
  shots = [],
  generations = [],
  assetsLoading = false,
  assetsError = '',
  tasks = [],
  activeTaskId,
  onTabChange,
  onWidthChange,
  onLocateNode,
  onActivateGenerationResult,
  onCreateFromAsset,
  onSelectTask,
  onNewProject,
  onClose,
  footer
}) {
  const [query, setQuery] = useState('');
  const [kinds, setKinds] = useState([]);
  const [statuses, setStatuses] = useState([]);
  const [historyKinds, setHistoryKinds] = useState([]);
  const [localWidth, setLocalWidth] = useState(() => workspaceDrawerWidth(width));
  const resizeRef = useRef(null);

  useEffect(() => setLocalWidth(workspaceDrawerWidth(width)), [width]);

  const filteredNodes = useMemo(
    () => filterWorkspaceNodes(nodes, { query, kinds, statuses, ratings: [] }),
    [nodes, query, kinds, statuses]
  );
  const nodeTree = useMemo(() => buildWorkspaceGroupTree(filteredNodes, groups), [filteredNodes, groups]);
  const filteredAssets = useMemo(() => filterWorkspaceAssets(assets, { query, kinds }), [assets, query, kinds]);
  const filteredExecutions = useMemo(() => filterWorkspaceExecutions(executions, { query, kinds: historyKinds }), [executions, query, historyKinds]);
  const resultGroups = useMemo(() => {
    const needle = String(query || '').trim().toLocaleLowerCase();
    const allowedKinds = new Set(historyKinds.map(String));
    return groupResultLibrary({ shots, nodes, generations, assets }).map(group => ({
      ...group,
      items: group.items.filter(item => (
        (!allowedKinds.size || allowedKinds.has(item.kind)) &&
        (!needle || [item.title, item.label, item.status, group.shot?.scene].some(value => String(value || '').toLocaleLowerCase().includes(needle)))
      ))
    })).filter(group => group.items.length);
  }, [assets, generations, historyKinds, nodes, query, shots]);

  function startResize(event) {
    resizeRef.current = { pointerId: event.pointerId, startX: event.clientX, startWidth: localWidth };
    event.currentTarget.setPointerCapture(event.pointerId);
  }

  function moveResize(event) {
    const resize = resizeRef.current;
    if (!resize || resize.pointerId !== event.pointerId) return;
    setLocalWidth(workspaceDrawerWidth(resize.startWidth + event.clientX - resize.startX));
  }

  function finishResize(event) {
    const resize = resizeRef.current;
    if (!resize || resize.pointerId !== event.pointerId) return;
    const finalWidth = workspaceDrawerWidth(resize.startWidth + event.clientX - resize.startX);
    resizeRef.current = null;
    event.currentTarget.releasePointerCapture?.(event.pointerId);
    setLocalWidth(finalWidth);
    onWidthChange?.(finalWidth);
  }

  const filters = <div className="agent-workspace-drawer__filters">
    <Input allowClear prefix={<Search size={14} />} placeholder={activeTab === 'assets' ? '搜索项目素材' : '搜索画布节点'} value={query} onChange={event => setQuery(event.target.value)} />
    <Select
      mode="multiple"
      allowClear
      maxTagCount="responsive"
      placeholder="类型"
      value={kinds}
      options={Object.entries(kindLabels).map(([value, label]) => ({ value, label }))}
      onChange={setKinds}
    />
    {activeTab === 'nodes' ? <Select
      mode="multiple"
      allowClear
      maxTagCount="responsive"
      placeholder="状态"
      value={statuses}
      options={Object.entries(statusLabels).map(([value, label]) => ({ value, label }))}
      onChange={setStatuses}
    /> : null}
  </div>;

  const nodeContent = <div className="agent-workspace-drawer__body">
    {filters}
    <div className="agent-workspace-tree">
      {!nodeTree.length ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="没有符合条件的节点" /> : nodeTree.map(group => <section key={group.id} className="agent-workspace-tree__group">
        <header><strong>{group.title}</strong><span>{group.children.length}</span></header>
        {group.children.map(node => <button type="button" className="agent-workspace-tree__row" key={node.id} onClick={() => onLocateNode?.(node.id)}>
          <KindIcon kind={nodeValue(node, 'type')} />
          <span><strong>{nodeValue(node, 'title') || '未命名节点'}</strong><small>{nodeValue(node, 'summary') || statusLabels[nodeValue(node, 'status')] || '暂无内容'}</small></span>
          <Tooltip title="在画布中定位"><LocateFixed size={14} /></Tooltip>
        </button>)}
      </section>)}
    </div>
  </div>;

  const assetContent = <div className="agent-workspace-drawer__body">
    {filters}
    {assetsLoading ? <div className="agent-workspace-drawer__loading"><Spin size="small" />正在读取素材</div> : null}
    {assetsError ? <div className="agent-workspace-drawer__error">{assetsError}</div> : null}
    {!assetsLoading && !assetsError && !filteredAssets.length ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="当前项目还没有素材" /> : null}
    <div className="agent-workspace-assets">
      {filteredAssets.map(asset => <button
        type="button"
        className="agent-workspace-assets__row"
        key={asset.id}
        draggable
        onDragStart={event => {
          event.dataTransfer.effectAllowed = 'copy';
          event.dataTransfer.setData('application/x-agent-asset', JSON.stringify({ id: asset.id, kind: asset.kind }));
        }}
        onClick={() => onCreateFromAsset?.(asset)}
      >
        <span className={`agent-workspace-assets__preview is-${asset.kind || 'file'}`}><AssetPreview asset={asset} /></span>
        <span><strong>{asset.name || asset.title || '未命名素材'}</strong><small>{kindLabels[asset.kind] || asset.kind || '素材'}</small></span>
        <Plus size={14} />
      </button>)}
    </div>
  </div>;

  const historyContent = <div className="agent-workspace-drawer__body">
    <div className="agent-workspace-drawer__filters agent-workspace-drawer__history-filters">
      <Input allowClear prefix={<Search size={14} />} placeholder="搜索生成记录" value={query} onChange={event => setQuery(event.target.value)} />
      <Select
        mode="multiple"
        allowClear
        maxTagCount="responsive"
        placeholder="媒介"
        value={historyKinds}
        options={Object.entries(kindLabels).map(([value, label]) => ({ value, label }))}
        onChange={setHistoryKinds}
      />
    </div>
    {resultGroups.length ? <div className="agent-workspace-results" aria-label="镜头结果库">
      {resultGroups.map(group => <section className="agent-workspace-results__group" key={group.id}>
        <header><strong>{group.title}</strong><span>{group.shot?.scene || `${group.items.length} 个版本`}</span></header>
        {group.items.map(item => {
          const generation = generations.find(candidate => String(candidate?.id || '') === item.generationId) || {};
          const node = nodes.find(candidate => String(candidate?.id || '') === item.nodeId) || {};
          const nodeConfig = node?.config || node?.data?.config || {};
          const labelNode = { ...node, config: { ...nodeConfig, source: { ...(nodeConfig.source || {}), sequence: group.sequence } } };
          const canActivate = !item.active && item.kind !== 'video' && item.generationId && item.resultId;
          return <div className={`agent-workspace-results__item${item.active ? ' is-active' : ''}`} key={item.id}>
            <button type="button" className="agent-workspace-results__open" onClick={() => onLocateNode?.(item.nodeId)}>
              <span className={`agent-workspace-assets__preview is-${item.kind || 'file'}`}><AssetPreview asset={item.asset} /></span>
              <span><strong>{lineageLabel(labelNode, generation)}</strong><small>{item.title}</small></span>
              <em>{item.active ? '当前' : '历史'}</em>
            </button>
            {canActivate ? <Button size="small" type="link" className="agent-workspace-results__activate" onClick={() => onActivateGenerationResult?.(item)}>设为当前</Button> : null}
          </div>;
        })}
      </section>)}
    </div> : null}
    {!executions.length && !resultGroups.length ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="当前项目还没有生成记录" /> : executions.length && !filteredExecutions.length && !resultGroups.length ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="没有符合条件的生成记录" /> : <div className="agent-workspace-history">
      {filteredExecutions.map((execution, index) => {
        const status = execution.status || execution.state || 'pending';
        const error = execution?.errorMessage || execution?.error || '';
        return <article className={`agent-workspace-history__item is-${status}`} key={execution?.id || index}>
          <span className="agent-workspace-history__icon">{status === 'failed' ? <CircleAlert size={15} /> : <KindIcon kind={workspaceExecutionKind(execution)} size={15} />}</span>
          <span className="agent-workspace-history__content">
            <strong>{executionLabel(execution)}</strong>
            <small><Clock3 size={11} />{executionTime(execution)}</small>
            {error ? <em title={String(error)}>{String(error)}</em> : null}
          </span>
          <span className={`agent-workspace-history__status is-${status}`}>{statusLabels[status] || status}</span>
        </article>;
      })}
    </div>}
  </div>;

  return <aside id="agent-project-drawer" className={`agent-project-drawer${open ? ' is-open' : ''}`} aria-hidden={!open} style={{ width: localWidth }}>
    <header className="agent-workspace-drawer__header">
      <Select
        className="agent-workspace-drawer__project-select"
        value={activeTaskId || undefined}
        placeholder="选择项目"
        options={tasks.map(task => ({ value: task.id, label: task.title || '未命名项目' }))}
        onChange={onSelectTask}
      />
      <Tooltip title="新建项目"><Button type="text" aria-label="新建项目" icon={<Plus size={16} />} onClick={onNewProject} /></Tooltip>
      <Tooltip title="收起项目栏"><Button type="text" aria-label="收起项目栏" icon={<X size={16} />} onClick={onClose} /></Tooltip>
    </header>
    <Tabs
      className="agent-workspace-drawer__tabs"
      activeKey={activeTab}
      onChange={onTabChange}
      items={[
        { key: 'nodes', label: `画布节点 ${nodes.length}`, children: nodeContent },
        { key: 'assets', label: `项目素材 ${assets.length}`, children: assetContent },
        { key: 'history', label: `生成历史 ${executions.length}`, children: historyContent }
      ]}
    />
    {footer ? <footer className="agent-workspace-drawer__footer">{footer}</footer> : null}
    <div
      className="agent-workspace-drawer__resize"
      role="separator"
      aria-label="调整项目栏宽度"
      aria-orientation="vertical"
      onPointerDown={startResize}
      onPointerMove={moveResize}
      onPointerUp={finishResize}
      onPointerCancel={finishResize}
    />
  </aside>;
}
