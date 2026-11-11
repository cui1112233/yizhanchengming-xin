// 工作模式由当前路由推导，不写 localStorage，避免出现“记住了但并没进入”的假状态。
export const WORK_MODE_STUDIO = 'studio';
export const WORK_MODE_AGENT = 'agent';
export const WORK_MODE_DUET = 'duet';

export const WORK_MODES = [
  { id: WORK_MODE_STUDIO, label: '一战晟铭', href: '/', summary: '短剧创作主工作台' },
  { id: WORK_MODE_AGENT, label: 'Agent 代理', href: '/agent/chat', summary: '对话式 Agent' },
  { id: WORK_MODE_DUET, label: 'Duet', href: '', summary: '双人协作模式', unavailable: true }
];

export function workModeForPath(pathname) {
  const path = String(pathname || '/').replace(/\/+$/, '') || '/';
  // Duet 尚未实现，永远不会成为当前项。
  if (path === '/agent/chat' || path.startsWith('/agent/chat/')) return WORK_MODE_AGENT;
  return WORK_MODE_STUDIO;
}

export function workModeById(id) {
  return WORK_MODES.find(mode => mode.id === id) || WORK_MODES[0];
}
