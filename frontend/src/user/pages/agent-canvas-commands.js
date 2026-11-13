function nodeWidth(node) {
  return Number(node.measured?.width ?? node.width ?? 168);
}

export function retainCanvasNodeSelection(nodes, selectedIds) {
  const selected = new Set(selectedIds || []);
  return (nodes || []).map(node => ({ ...node, selected: selected.has(node.id) }));
}

export const CANVAS_SHORTCUTS = Object.freeze([
  { command: 'undo', keys: 'Ctrl/⌘ + Z', label: '撤销' },
  { command: 'redo', keys: 'Ctrl/⌘ + Shift + Z', label: '重做' },
  { command: 'copy', keys: 'Ctrl/⌘ + C', label: '复制节点' },
  { command: 'paste', keys: 'Ctrl/⌘ + V', label: '粘贴节点' },
  { command: 'duplicate', keys: 'Ctrl/⌘ + D', label: '创建副本' },
  { command: 'delete', keys: 'Delete', label: '删除选中' },
  { command: 'group', keys: 'Ctrl/⌘ + G', label: '编组' },
  { command: 'ungroup', keys: 'Ctrl/⌘ + Shift + G', label: '解组' },
  { command: 'fit_canvas', keys: 'Shift + 1', label: '适合屏幕' },
  { command: 'focus_selection', keys: 'Shift + 2', label: '聚焦选中' },
  { command: 'zoom_in', keys: 'Ctrl/⌘ + +', label: '放大' },
  { command: 'zoom_out', keys: 'Ctrl/⌘ + -', label: '缩小' }
]);

function nodeHeight(node) {
  return Number(node.measured?.height ?? node.height ?? 88);
}

function selectedNodes(nodes, selectedIds) {
  const selected = new Set(selectedIds || []);
  return nodes.filter(node => selected.has(node.id));
}

export function sameCanvasSelection(current = [], next = []) {
  return current.length === next.length && current.every((id, index) => id === next[index]);
}

export function isCanvasMultiSelectionEvent(event) {
  return Boolean(event?.shiftKey || event?.ctrlKey || event?.metaKey);
}

function selectionBounds(nodes) {
  if (!nodes.length) return null;
  const left = Math.min(...nodes.map(node => node.position.x));
  const top = Math.min(...nodes.map(node => node.position.y));
  const right = Math.max(...nodes.map(node => node.position.x + nodeWidth(node)));
  const bottom = Math.max(...nodes.map(node => node.position.y + nodeHeight(node)));
  return { left, top, right, bottom, centerX: (left + right) / 2, centerY: (top + bottom) / 2 };
}

export function selectCanvasNodes(nodes, clickedId, additive = false, selectedIds = null) {
  if (!clickedId || !nodes.some(node => node.id === clickedId)) return nodes;
  const selected = new Set(selectedIds || nodes.filter(node => node.selected).map(node => node.id));
  const clickedWasSelected = selected.has(clickedId);
  return nodes.map(node => ({
    ...node,
    selected: additive
      ? (node.id === clickedId ? !clickedWasSelected : selected.has(node.id))
      : node.id === clickedId
  }));
}

export function commandAvailability(nodes, selectedIds) {
  const selected = selectedNodes(nodes, selectedIds);
  return {
    canGroup: selected.length >= 2,
    canUngroup: selected.some(node => Boolean(node.data?.groupId)),
    canAlign: selected.length >= 2,
    canDistribute: selected.length >= 3,
    groupReason: selected.length >= 2 ? '' : '至少选择 2 个节点',
    ungroupReason: selected.some(node => Boolean(node.data?.groupId)) ? '' : '请选择已分组节点',
    alignReason: selected.length >= 2 ? '' : '至少选择 2 个节点',
    distributeReason: selected.length >= 3 ? '' : '至少选择 3 个节点'
  };
}

export function connectCanvasNodes(edges, connection) {
  const source = String(connection?.source || '');
  const target = String(connection?.target || '');
  const relation = String(connection?.data?.relation || 'reference');
  if (!source || !target || source === target) return edges;
  if (edges.some(edge => edge.source === source && edge.target === target && (edge.data?.relation || 'reference') === relation)) return edges;
  return [...edges, { id: `edge-${source}-${target}-${relation}`, source, target, data: { relation } }];
}

export function removeCanvasEdges(edges, edgeIds) {
  const removed = new Set(edgeIds || []);
  if (!removed.size) return edges;
  return edges.filter(edge => !removed.has(edge.id));
}

export function groupCanvasNodes(nodes, selectedIds, groupId) {
  const selected = new Set(selectedIds || []);
  if (selected.size < 2 || !groupId) return nodes;
  return nodes.map(node => selected.has(node.id)
    ? { ...node, data: { ...(node.data || {}), groupId } }
    : node);
}

export function ungroupCanvasNodes(nodes, selectedIds) {
  const selected = new Set(selectedIds || []);
  if (!selected.size) return nodes;
  return nodes.map(node => {
    if (!selected.has(node.id) || !node.data?.groupId) return node;
    const data = { ...node.data };
    delete data.groupId;
    return { ...node, data };
  });
}

export function alignCanvasNodes(nodes, selectedIds, mode) {
  const selected = selectedNodes(nodes, selectedIds);
  if (selected.length < 2) return nodes;
  const bounds = selectionBounds(selected);
  const supported = new Set(['left', 'centerX', 'right', 'top', 'centerY', 'bottom']);
  if (!supported.has(mode)) return nodes;
  const selectedSet = new Set(selectedIds);
  return nodes.map(node => {
    if (!selectedSet.has(node.id)) return node;
    let x = node.position.x;
    let y = node.position.y;
    if (mode === 'left') x = bounds.left;
    if (mode === 'centerX') x = bounds.centerX - nodeWidth(node) / 2;
    if (mode === 'right') x = bounds.right - nodeWidth(node);
    if (mode === 'top') y = bounds.top;
    if (mode === 'centerY') y = bounds.centerY - nodeHeight(node) / 2;
    if (mode === 'bottom') y = bounds.bottom - nodeHeight(node);
    return { ...node, position: { x, y } };
  });
}

export function distributeCanvasNodes(nodes, selectedIds, axis) {
  const selected = selectedNodes(nodes, selectedIds);
  if (selected.length < 3 || !['horizontal', 'vertical'].includes(axis)) return nodes;
  const horizontal = axis === 'horizontal';
  const ordered = [...selected].sort((a, b) => horizontal ? a.position.x - b.position.x : a.position.y - b.position.y);
  const first = ordered[0];
  const last = ordered.at(-1);
  const start = horizontal ? first.position.x : first.position.y;
  const end = horizontal ? last.position.x + nodeWidth(last) : last.position.y + nodeHeight(last);
  const occupied = ordered.reduce((sum, node) => sum + (horizontal ? nodeWidth(node) : nodeHeight(node)), 0);
  const gap = (end - start - occupied) / (ordered.length - 1);
  const positions = new Map();
  let cursor = start;
  for (const node of ordered) {
    positions.set(node.id, cursor);
    cursor += (horizontal ? nodeWidth(node) : nodeHeight(node)) + gap;
  }
  return nodes.map(node => {
    if (!positions.has(node.id)) return node;
    return {
      ...node,
      position: horizontal
        ? { x: positions.get(node.id), y: node.position.y }
        : { x: node.position.x, y: positions.get(node.id) }
    };
  });
}

export function canvasGroupBounds(nodes, groupId, padding = 16) {
  const members = nodes.filter(node => node.data?.groupId === groupId);
  const bounds = selectionBounds(members);
  if (!bounds) return null;
  return {
    x: bounds.left - padding,
    y: bounds.top - padding,
    width: bounds.right - bounds.left + padding * 2,
    height: bounds.bottom - bounds.top + padding * 2
  };
}

export function canvasKeyboardCommand(event) {
  const tag = String(event?.targetTagName || '').toLowerCase();
  if (event?.isContentEditable || ['input', 'textarea', 'select'].includes(tag)) return null;
  const key = String(event?.key || '').toLowerCase();
  const modifier = Boolean(event?.ctrlKey || event?.metaKey);
  if (key === 'delete' || key === 'backspace') return 'delete';
  if (event?.shiftKey && key === '1') return 'fit_canvas';
  if (event?.shiftKey && key === '2') return 'focus_selection';
  if (!modifier) return null;
  if (key === 'a') return 'select_all';
  if (key === 'c') return 'copy';
  if (key === 'v') return 'paste';
  if (key === 'd') return 'duplicate';
  if (key === 'g' && event?.shiftKey) return 'ungroup';
  if (key === 'g') return 'group';
  if (key === '+' || key === '=') return 'zoom_in';
  if (key === '-') return 'zoom_out';
  if (key === 'z' && event?.shiftKey) return 'redo';
  if (key === 'z') return 'undo';
  if (key === 'y') return 'redo';
  return null;
}

export function createCanvasNode(id, position) {
  return {
    id,
    type: 'agent',
    position: { x: Number(position?.x) || 0, y: Number(position?.y) || 0 },
    selected: true,
    data: {
      type: 'text',
      title: '新节点',
      summary: '',
      version: 1,
      sourceNodeIds: [],
      status: 'draft',
      assetRef: null
    }
  };
}

export function moveCanvasGroup(nodes, groupId, delta) {
  const dx = Number(delta?.x) || 0;
  const dy = Number(delta?.y) || 0;
  if (!groupId || (!dx && !dy)) return nodes;
  return nodes.map(node => node.data?.groupId === groupId
    ? { ...node, position: { x: node.position.x + dx, y: node.position.y + dy } }
    : node);
}
