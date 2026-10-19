import { Background, BaseEdge, Handle, MarkerType, Position, ReactFlow, ViewportPortal, applyEdgeChanges, applyNodeChanges, getBezierPath } from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { Button, Dropdown, Popover, Segmented, Switch, Tooltip } from 'antd';
import {
  AlignCenterHorizontal, AlignCenterVertical, AlignEndHorizontal, AlignEndVertical,
  AlignHorizontalDistributeCenter, AlignStartHorizontal, AlignStartVertical,
  AlignVerticalDistributeCenter, Focus, GitBranch, Group, Hand, Maximize2,
  MousePointer2, Plus, Redo2, SlidersHorizontal, Undo2, Ungroup, ZoomIn, ZoomOut
} from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import {
  alignCanvasNodes, canvasGroupBounds, canvasKeyboardCommand, commandAvailability,
  connectCanvasNodes, createCanvasNode, distributeCanvasNodes, groupCanvasNodes,
  moveCanvasGroup, selectCanvasNodes, ungroupCanvasNodes
} from '../pages/agent-canvas-commands';
import { createCanvasHistory } from '../pages/agent-canvas-history';
import {
  appearanceCssVariables, canvasEdgeColorOptions, canvasEdgeWidthOptions,
  edgeStyleForAppearance, normalizeCanvasAppearance
} from '../pages/agent-canvas-appearance';

function CanvasNode({ data }) {
  return (
    <article className={`agent-canvas-node agent-canvas-node--${data.type} agent-canvas-node--${data.status}`}>
      <Handle type="target" position={Position.Left} />
      <header><span>{data.type}</span><strong>{data.title}</strong></header>
      <p>{data.summary || '等待 Agent 写入结果'}</p>
      <footer><span>v{data.version}</span><span>{data.status}</span></footer>
      <Handle type="source" position={Position.Right} />
    </article>
  );
}

function CanvasEdge({ id, sourceX, sourceY, sourcePosition, targetX, targetY, targetPosition, data, markerEnd }) {
  const appearance = normalizeCanvasAppearance(data?.appearance);
  const [edgePath] = getBezierPath({ sourceX, sourceY, sourcePosition, targetX, targetY, targetPosition });
  const style = edgeStyleForAppearance(appearance);
  return <>
    <BaseEdge id={id} path={edgePath} markerEnd={markerEnd} style={style} />
    {appearance.edgeAnimated ? <path d={edgePath} className="agent-canvas-edge-flow" style={style} /> : null}
  </>;
}

const canvasEdgeTypes = { canvas: CanvasEdge };

function toFlowNode(node) {
  return {
    id: node.id,
    type: 'agent',
    position: node.position,
    selected: false,
    data: {
      type: node.type,
      title: node.title,
      summary: node.summary,
      version: node.version,
      sourceNodeIds: node.sourceNodeIds,
      status: node.status,
      assetRef: node.assetRef || null,
      ...(node.groupId ? { groupId: node.groupId } : {})
    }
  };
}

function toFlowEdge(edge, appearance) {
  const style = edgeStyleForAppearance(appearance);
  return {
    ...edge,
    type: 'canvas',
    markerEnd: { type: MarkerType.ArrowClosed, color: style.stroke },
    data: { relation: edge.relation || edge.data?.relation || 'reference', appearance }
  };
}

function toCanvas(nodes, edges, revision, viewport, appearance) {
  return {
    version: 1,
    revision,
    ...(viewport ? { viewport } : {}),
    appearance: normalizeCanvasAppearance(appearance),
    nodes: nodes.map(node => ({
      id: node.id,
      type: node.data.type,
      position: { x: node.position.x, y: node.position.y },
      title: node.data.title,
      version: node.data.version,
      sourceNodeIds: [...(node.data.sourceNodeIds || [])],
      status: node.data.status,
      summary: node.data.summary || '',
      ...(node.data.assetRef ? { assetRef: node.data.assetRef } : {}),
      ...(node.data.groupId ? { groupId: node.data.groupId } : {})
    })),
    edges: edges.map(edge => ({ id: edge.id, source: edge.source, target: edge.target, relation: edge.data?.relation || 'reference' }))
  };
}

function canvasSignature(nodes, edges, viewport, appearance) {
  return JSON.stringify(toCanvas(nodes, edges, 0, viewport, appearance));
}

function IconTool({ label, disabled = false, active = false, onClick, children }) {
  return (
    <Tooltip title={label}>
      <span>
        <Button className={active ? 'is-active' : ''} type="text" aria-label={label} disabled={disabled} icon={children} onClick={onClick} />
      </span>
    </Tooltip>
  );
}

export function AgentCanvas({ canvas, onChange, onAppearanceChange, onSelectionChange, onViewportChange }) {
  const currentCanvas = canvas || { version: 1, revision: 0, nodes: [], edges: [] };
  const [nodes, setNodes] = useState([]);
  const [edges, setEdges] = useState([]);
  const [selectedIds, setSelectedIds] = useState([]);
  const [interactionMode, setInteractionMode] = useState('select');
  const [flowInstance, setFlowInstance] = useState(null);
  const [viewport, setViewport] = useState(currentCanvas.viewport || { x: 0, y: 0, zoom: 1 });
  const [appearance, setAppearance] = useState(() => normalizeCanvasAppearance(currentCanvas.appearance));
  const [nodeDragActive, setNodeDragActive] = useState(false);
  const [, setHistoryVersion] = useState(0);
  const rootRef = useRef(null);
  const nodesRef = useRef([]);
  const edgesRef = useRef([]);
  const viewportRef = useRef(viewport);
  const appearanceRef = useRef(appearance);
  const signatureRef = useRef('');
  const groupDragRef = useRef(null);
  const nodeSequenceRef = useRef(0);
  const historyRef = useRef(createCanvasHistory({ nodes: [], edges: [] }));

  const canvasNodeTypes = useMemo(() => ({ agent: CanvasNode }), []);
  const availability = useMemo(() => commandAvailability(nodes, selectedIds), [nodes, selectedIds]);
  const groups = useMemo(() => [...new Set(nodes.map(node => node.data?.groupId).filter(Boolean))]
    .map(groupId => ({ groupId, bounds: canvasGroupBounds(nodes, groupId) }))
    .filter(group => group.bounds), [nodes]);

  function updateRefs(nextNodes, nextEdges, nextViewport = viewportRef.current, nextAppearance = appearanceRef.current) {
    nodesRef.current = nextNodes;
    edgesRef.current = nextEdges;
    viewportRef.current = nextViewport;
    appearanceRef.current = nextAppearance;
    signatureRef.current = canvasSignature(nextNodes, nextEdges, nextViewport, nextAppearance);
  }

  useEffect(() => {
    const nextNodes = (currentCanvas.nodes || []).map(toFlowNode);
    const nextAppearance = normalizeCanvasAppearance(currentCanvas.appearance);
    const nextEdges = (currentCanvas.edges || []).map(edge => toFlowEdge(edge, nextAppearance));
    const nextViewport = currentCanvas.viewport || { x: 0, y: 0, zoom: 1 };
    if (canvasSignature(nextNodes, nextEdges, nextViewport, nextAppearance) === signatureRef.current) return;
    setNodes(nextNodes);
    setEdges(nextEdges);
    setSelectedIds([]);
    setViewport(nextViewport);
    setAppearance(nextAppearance);
    updateRefs(nextNodes, nextEdges, nextViewport, nextAppearance);
    historyRef.current.reset({ nodes: nextNodes, edges: nextEdges });
    setHistoryVersion(value => value + 1);
  }, [canvas]);

  function emitChange(nextNodes, nextEdges, nextViewport = viewportRef.current, nextAppearance = appearanceRef.current) {
    updateRefs(nextNodes, nextEdges, nextViewport, nextAppearance);
    onChange?.(toCanvas(nextNodes, nextEdges, currentCanvas.revision, nextViewport, nextAppearance));
  }

  function commit(nextNodes, nextEdges, { record = true, nextViewport = viewportRef.current } = {}) {
    setNodes(nextNodes);
    setEdges(nextEdges);
    if (record) historyRef.current.push({ nodes: nextNodes, edges: nextEdges });
    setHistoryVersion(value => value + 1);
    emitChange(nextNodes, nextEdges, nextViewport);
  }

  function handleNodesChange(changes) {
    const next = applyNodeChanges(changes, nodesRef.current);
    nodesRef.current = next;
    setNodes(next);
    if (changes.some(change => change.type === 'remove')) {
      const remaining = new Set(next.map(node => node.id));
      const nextEdges = edgesRef.current.filter(edge => remaining.has(edge.source) && remaining.has(edge.target));
      commit(next, nextEdges);
    }
  }

  function handleEdgesChange(changes) {
    const next = applyEdgeChanges(changes, edgesRef.current);
    edgesRef.current = next;
    setEdges(next);
    if (changes.some(change => change.type === 'remove' || change.type === 'add')) commit(nodesRef.current, next);
  }

  function handleSelectionChange(selection) {
    const ids = (selection?.nodes || []).map(node => node.id);
    setSelectedIds(ids);
    onSelectionChange?.(ids);
  }

  function handleNodeClick(event, node) {
    const next = selectCanvasNodes(
      nodesRef.current,
      node.id,
      event.ctrlKey || event.metaKey || event.shiftKey,
      selectedIds
    );
    const ids = next.filter(item => item.selected).map(item => item.id);
    nodesRef.current = next;
    setNodes(next);
    setSelectedIds(ids);
    onSelectionChange?.(ids);
  }

  function handleConnect(connection) {
    const next = connectCanvasNodes(edgesRef.current, connection);
    if (next !== edgesRef.current) commit(nodesRef.current, next.map(edge => toFlowEdge(edge, appearanceRef.current)));
  }

  function runNodeCommand(command) {
    const next = command(nodesRef.current);
    if (next !== nodesRef.current) commit(next, edgesRef.current);
  }

  function addNode() {
    const bounds = rootRef.current?.getBoundingClientRect();
    const screenPoint = bounds ? { x: bounds.left + bounds.width / 2, y: bounds.top + bounds.height / 2 } : { x: 0, y: 0 };
    const position = flowInstance?.screenToFlowPosition(screenPoint) || { x: 0, y: 0 };
    nodeSequenceRef.current += 1;
    const nextNode = createCanvasNode(`node-${Date.now()}-${nodeSequenceRef.current}`, position);
    const next = nodesRef.current.map(node => ({ ...node, selected: false })).concat(nextNode);
    setSelectedIds([nextNode.id]);
    onSelectionChange?.([nextNode.id]);
    commit(next, edgesRef.current);
  }

  function applyHistory(direction) {
    const snapshot = direction === 'undo' ? historyRef.current.undo() : historyRef.current.redo();
    commit(snapshot.nodes, snapshot.edges, { record: false });
  }

  function handleKeyDown(event) {
    const command = canvasKeyboardCommand({
      ctrlKey: event.ctrlKey, metaKey: event.metaKey, shiftKey: event.shiftKey, key: event.key,
      targetTagName: event.target?.tagName, isContentEditable: event.target?.isContentEditable
    });
    if (!command) return;
    if (command === 'undo' && !historyRef.current.canUndo()) return;
    if (command === 'redo' && !historyRef.current.canRedo()) return;
    event.preventDefault();
    applyHistory(command);
  }

  function handleMoveEnd(_event, nextViewport) {
    setViewport(nextViewport);
    viewportRef.current = nextViewport;
    onViewportChange?.(nextViewport);
    emitChange(nodesRef.current, edgesRef.current, nextViewport);
  }

  function updateAppearance(nextValue) {
    const nextAppearance = normalizeCanvasAppearance({ ...appearanceRef.current, ...nextValue });
    const nextEdges = edgesRef.current.map(edge => toFlowEdge(edge, nextAppearance));
    setAppearance(nextAppearance);
    setEdges(nextEdges);
    onAppearanceChange?.(nextAppearance);
    emitChange(nodesRef.current, nextEdges, viewportRef.current, nextAppearance);
  }

  function startGroupDrag(event, groupId) {
    event.stopPropagation();
    event.currentTarget.setPointerCapture(event.pointerId);
    groupDragRef.current = { groupId, pointerId: event.pointerId, startX: event.clientX, startY: event.clientY, startNodes: nodesRef.current };
  }

  function moveGroupDrag(event) {
    const drag = groupDragRef.current;
    if (!drag || drag.pointerId !== event.pointerId) return;
    const next = moveCanvasGroup(drag.startNodes, drag.groupId, { x: event.clientX - drag.startX, y: event.clientY - drag.startY });
    nodesRef.current = next;
    setNodes(next);
  }

  function endGroupDrag(event) {
    const drag = groupDragRef.current;
    if (!drag || drag.pointerId !== event.pointerId) return;
    groupDragRef.current = null;
    commit(nodesRef.current, edgesRef.current);
  }

  const alignItems = [
    { key: 'left', label: '左对齐', icon: <AlignStartVertical size={15} /> },
    { key: 'centerX', label: '水平居中', icon: <AlignCenterVertical size={15} /> },
    { key: 'right', label: '右对齐', icon: <AlignEndVertical size={15} /> },
    { key: 'top', label: '顶对齐', icon: <AlignStartHorizontal size={15} /> },
    { key: 'centerY', label: '垂直居中', icon: <AlignCenterHorizontal size={15} /> },
    { key: 'bottom', label: '底对齐', icon: <AlignEndHorizontal size={15} /> }
  ];

  return (
    <div ref={rootRef} className="agent-canvas" aria-label="创作画布" tabIndex={0} onKeyDown={handleKeyDown} style={appearanceCssVariables(appearance)}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={canvasNodeTypes}
        edgeTypes={canvasEdgeTypes}
        onInit={setFlowInstance}
        onNodesChange={handleNodesChange}
        onEdgesChange={handleEdgesChange}
        onSelectionChange={handleSelectionChange}
        onNodeClick={handleNodeClick}
        onConnect={handleConnect}
        onNodeDragStart={() => setNodeDragActive(true)}
        onNodeDragStop={() => { setNodeDragActive(false); commit(nodesRef.current, edgesRef.current); }}
        minZoom={0.25}
        maxZoom={2}
        fitView={!currentCanvas.viewport}
        defaultViewport={currentCanvas.viewport}
        onMove={(_event, nextViewport) => {
          setViewport(nextViewport);
          viewportRef.current = nextViewport;
          onViewportChange?.(nextViewport);
        }}
        onMoveEnd={handleMoveEnd}
        panOnDrag={interactionMode === 'pan'}
        nodesDraggable={interactionMode !== 'pan'}
        elementsSelectable={interactionMode !== 'pan'}
        selectionOnDrag={interactionMode === 'select'}
        snapToGrid={appearance.alignmentGuides}
        snapGrid={[20, 20]}
        multiSelectionKeyCode="Shift"
      >
        {appearance.backgroundDots ? <Background variant="dots" gap={20} size={1} /> : null}
        <ViewportPortal>
          {groups.map(({ groupId, bounds }) => (
            <section className="agent-canvas-group" key={groupId} style={{ left: bounds.x, top: bounds.y, width: bounds.width, height: bounds.height }}>
              <button
                type="button"
                className="agent-canvas-group-handle"
                onPointerDown={event => startGroupDrag(event, groupId)}
                onPointerMove={moveGroupDrag}
                onPointerUp={endGroupDrag}
                onPointerCancel={endGroupDrag}
              >节点分组</button>
            </section>
          ))}
        </ViewportPortal>
      </ReactFlow>

      {appearance.alignmentGuides && nodeDragActive ? <><span className="agent-canvas-guide agent-canvas-guide--vertical" /><span className="agent-canvas-guide agent-canvas-guide--horizontal" /></> : null}

      <nav className="agent-canvas-toolrail" aria-label="画布工具">
        <IconTool label="选择节点" active={interactionMode === 'select'} onClick={() => setInteractionMode('select')}><MousePointer2 size={17} /></IconTool>
        <IconTool label="拖动画布" active={interactionMode === 'pan'} onClick={() => setInteractionMode('pan')}><Hand size={17} /></IconTool>
        <IconTool label="连接节点" active={interactionMode === 'connect'} onClick={() => setInteractionMode('connect')}><GitBranch size={17} /></IconTool>
        <IconTool label="添加节点" onClick={addNode}><Plus size={17} /></IconTool>
        <Popover trigger="click" placement="rightTop" overlayClassName="agent-canvas-appearance-popover" content={<section className="agent-canvas-appearance-panel" aria-label="画布外观设置">
          <div className="agent-canvas-appearance-row"><span>连线颜色</span><div className="agent-canvas-color-options">{canvasEdgeColorOptions.map(option => <button key={option.value} type="button" className={`agent-canvas-color-swatch is-${option.value}${appearance.edgeColor === option.value ? ' is-active' : ''}`} aria-label={option.label} aria-pressed={appearance.edgeColor === option.value} onClick={() => updateAppearance({ edgeColor: option.value })} />)}</div></div>
          <div className="agent-canvas-appearance-row"><span>连线动效</span><Switch size="small" checked={appearance.edgeAnimated} onChange={edgeAnimated => updateAppearance({ edgeAnimated })} /></div>
          <div className="agent-canvas-appearance-row agent-canvas-appearance-row--stack"><span>连线粗细</span><Segmented size="small" value={appearance.edgeWidth} options={canvasEdgeWidthOptions} onChange={edgeWidth => updateAppearance({ edgeWidth })} /></div>
          <div className="agent-canvas-appearance-row"><span>背景网点</span><Switch size="small" checked={appearance.backgroundDots} onChange={backgroundDots => updateAppearance({ backgroundDots })} /></div>
          <div className="agent-canvas-appearance-row"><span>辅助线对齐</span><Switch size="small" checked={appearance.alignmentGuides} onChange={alignmentGuides => updateAppearance({ alignmentGuides })} /></div>
        </section>}>
          <span><IconTool label="画布外观"><SlidersHorizontal size={17} /></IconTool></span>
        </Popover>
      </nav>

      <nav className="agent-canvas-editbar" aria-label="节点编辑工具">
        <IconTool label="撤销 Ctrl+Z" disabled={!historyRef.current.canUndo()} onClick={() => applyHistory('undo')}><Undo2 size={17} /></IconTool>
        <IconTool label="重做 Ctrl+Shift+Z" disabled={!historyRef.current.canRedo()} onClick={() => applyHistory('redo')}><Redo2 size={17} /></IconTool>
        <span className="agent-canvas-tool-divider" />
        <IconTool label={availability.canGroup ? '分组' : availability.groupReason} disabled={!availability.canGroup} onClick={() => runNodeCommand(current => groupCanvasNodes(current, selectedIds, `group-${Date.now()}`))}><Group size={17} /></IconTool>
        <IconTool label={availability.canUngroup ? '解组' : availability.ungroupReason} disabled={!availability.canUngroup} onClick={() => runNodeCommand(current => ungroupCanvasNodes(current, selectedIds))}><Ungroup size={17} /></IconTool>
        <Dropdown disabled={!availability.canAlign} menu={{ items: alignItems, onClick: ({ key }) => runNodeCommand(current => alignCanvasNodes(current, selectedIds, key)) }} trigger={['click']}>
          <span><IconTool label={availability.canAlign ? '对齐' : availability.alignReason} disabled={!availability.canAlign}><AlignCenterHorizontal size={17} /></IconTool></span>
        </Dropdown>
        <IconTool label={availability.canDistribute ? '水平等距' : availability.distributeReason} disabled={!availability.canDistribute} onClick={() => runNodeCommand(current => distributeCanvasNodes(current, selectedIds, 'horizontal'))}><AlignHorizontalDistributeCenter size={17} /></IconTool>
        <IconTool label={availability.canDistribute ? '垂直等距' : availability.distributeReason} disabled={!availability.canDistribute} onClick={() => runNodeCommand(current => distributeCanvasNodes(current, selectedIds, 'vertical'))}><AlignVerticalDistributeCenter size={17} /></IconTool>
      </nav>

      <nav className="agent-canvas-zoombar" aria-label="画布缩放工具">
        <IconTool label="缩小" onClick={() => flowInstance?.zoomOut()}><ZoomOut size={17} /></IconTool>
        <span className="agent-canvas-zoom-value">{Math.round((viewport.zoom || 1) * 100)}%</span>
        <IconTool label="放大" onClick={() => flowInstance?.zoomIn()}><ZoomIn size={17} /></IconTool>
        <IconTool label="适合屏幕" onClick={() => flowInstance?.fitView({ padding: 0.18 })}><Maximize2 size={17} /></IconTool>
        <IconTool label={selectedIds.length ? '聚焦选中' : '请先选择节点'} disabled={!selectedIds.length} onClick={() => flowInstance?.fitView({ nodes: nodesRef.current.filter(node => selectedIds.includes(node.id)), padding: 0.3 })}><Focus size={17} /></IconTool>
      </nav>
    </div>
  );
}
