const PRODUCT_ROUTES = new Map([
  ['/novel-fetch', '小说获取'],
  ['/script', '剧本生成'],
  ['/shuihuo-production', '水货生产'],
  ['/shuihuo-production/creative', '创作漫剧'],
  ['/batch-factory', '批量工厂'],
  ['/agent', 'Agent 工作区'],
  ['/settings', '设置'],
  ['/api-config', 'API 配置'],
]);

const TASK_META = {
  in_progress: { label: '进行中', tone: 'processing' },
  waiting: { label: '等待中', tone: 'default' },
  needs_decision: { label: '待我决定', tone: 'warning' },
  completed: { label: '已完成', tone: 'success' },
  failed: { label: '失败', tone: 'error' },
};

export function taskStatusMeta(status) {
  return TASK_META[status] || { label: String(status || '未知'), tone: 'default' };
}

export function groupAgentTasks(tasks = []) {
  const grouped = { active: [], needsDecision: [], completed: [], failed: [] };
  for (const task of Array.isArray(tasks) ? tasks : []) {
    if (task?.status === 'needs_decision') grouped.needsDecision.push(task);
    else if (task?.status === 'completed') grouped.completed.push(task);
    else if (task?.status === 'failed') grouped.failed.push(task);
    else grouped.active.push(task);
  }
  return grouped;
}

function toolArguments(toolCall) {
  if (!toolCall) return {};
  if (toolCall.arguments && typeof toolCall.arguments === 'object') return toolCall.arguments;
  if (typeof toolCall.arguments === 'string') {
    try { return JSON.parse(toolCall.arguments); } catch { return {}; }
  }
  return {};
}

export function buildNavigationTarget(toolCall) {
  if (toolCall?.tool_name !== 'ui.navigate') return '';
  const args = toolArguments(toolCall);
  const path = String(args.path || '').trim();
  if (!PRODUCT_ROUTES.has(path)) return '';
  const section = String(args.section || '').trim();
  if (!section) return path;
  if (!/^[a-z0-9-]{1,64}$/i.test(section)) return path;
  return `${path}?section=${encodeURIComponent(section)}`;
}

export function toolCardMeta(toolCall) {
  const args = toolArguments(toolCall);
  if (toolCall?.tool_name === 'ui.navigate') {
    const name = PRODUCT_ROUTES.get(String(args.path || '')) || '系统页面';
    const status = toolCall?.status || 'proposed';
    return {
      title: `打开${name}`,
      detail: status === 'failed' ? '导航执行失败' : status === 'completed' ? '已经打开' : '已准备好导航',
      actionLabel: '打开',
      tone: status === 'failed' ? 'error' : 'processing',
    };
  }
  return { title: 'Agent 工具', detail: '工具调用', actionLabel: '查看', tone: 'default' };
}

export function normalizeMediaAssetIds(values = []) {
  const output = [];
  const seen = new Set();
  for (const raw of Array.isArray(values) ? values : []) {
    const id = String(raw || '').trim();
    if (!/^[A-Za-z0-9_-]{1,64}$/.test(id) || seen.has(id)) continue;
    seen.add(id);
    output.push(id);
  }
  return output;
}

export function routeLabel(path) {
  return PRODUCT_ROUTES.get(path) || '工作区';
}

export const agentProductRoutes = [...PRODUCT_ROUTES.entries()].map(([path, label]) => ({ path, label }));
