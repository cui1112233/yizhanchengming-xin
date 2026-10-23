const labels = Object.freeze({
  create_node: '创建节点',
  update_node: '修改节点',
  connect_nodes: '连接节点',
  disconnect_nodes: '断开连接',
  move_nodes: '移动节点',
  remove_nodes: '移除画布节点',
  start_generation: '开始生成',
  apply_result: '应用生成结果'
});

const nodeTypeLabels = Object.freeze({ text: '文本', image: '图片', video: '视频', audio: '音频' });
const operationalCommands = new Set(['start_generation', 'apply_result']);

function commandLabel(command = {}) {
  const type = String(command.type || '');
  if (type === 'create_node') {
    const node = command.node || {};
    return `创建${nodeTypeLabels[node.type] || ''}节点${node.title ? `「${node.title}」` : ''}`;
  }
  if (type === 'update_node') return `修改节点「${command.nodeId || '未知节点'}」`;
  if (type === 'connect_nodes') return `连接「${command.sourceNodeId || '未知节点'}」到「${command.targetNodeId || '未知节点'}」`;
  if (type === 'disconnect_nodes') return command.edgeId ? `断开连接「${command.edgeId}」` : `断开「${command.sourceNodeId || '未知节点'}」与「${command.targetNodeId || '未知节点'}」`;
  if (type === 'move_nodes') return `移动节点「${command.nodeId || '未知节点'}」`;
  if (type === 'remove_nodes') {
    const count = new Set([...(command.nodeIds || []), command.nodeId].filter(Boolean)).size;
    return `移除 ${count} 个画布节点`;
  }
  if (type === 'start_generation') return `使用节点「${command.nodeId || '未知节点'}」开始生成`;
  if (type === 'apply_result') return `把结果应用到节点「${command.nodeId || '未知节点'}」`;
  return labels[type] || '未知操作';
}

export function isDirectorPlanExecutable(plan = {}) {
  const commands = Array.isArray(plan.commands) ? plan.commands : [];
  return commands.length > 0 && commands.every(command => !operationalCommands.has(command?.type));
}

export function directorPlanPresentation(plan = {}) {
  const commands = Array.isArray(plan.commands) ? plan.commands : [];
  return {
    id: String(plan.id || ''),
    title: '导演方案',
    summary: String(plan.summary || '请确认本次画布改动。'),
    executable: isDirectorPlanExecutable(plan),
    items: commands.map((command, index) => ({
      key: `${command?.type || 'unknown'}-${index}`,
      type: String(command?.type || ''),
      label: commandLabel(command),
      destructive: command?.type === 'remove_nodes',
      requiresExecution: operationalCommands.has(command?.type)
    }))
  };
}
