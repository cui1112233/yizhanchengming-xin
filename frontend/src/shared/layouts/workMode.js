// 工作模式由当前路由推导，不写 localStorage，避免出现“记住了但并没进入”的假状态。
export const WORK_MODE_STUDIO = 'studio';
export const WORK_MODE_AGENT = 'agent';
export const WORK_MODE_MAGIC_CUT = 'magic-cut';
export const WORK_MODE_DUET = 'duet';

export const WORK_MODES = [
  { id: WORK_MODE_STUDIO, label: '一战晟铭', href: '/', summary: '短剧创作主工作台' },
  { id: WORK_MODE_AGENT, label: 'Agent 代理', href: '/agent/chat', summary: '对话式 Agent' },
  { id: WORK_MODE_MAGIC_CUT, label: '魔剪', href: '/magic-cut', summary: '本地视频剪辑工作台' },
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

export function workModeForLocation(locationLike = {}) {
  if (String(locationLike.port || '') === '4040') return WORK_MODE_MAGIC_CUT;
  if (String(locationLike.pathname || '').startsWith('/magic-cut')) return WORK_MODE_MAGIC_CUT;
  return workModeForPath(locationLike.pathname);
}

export function workModeHref(id, locationLike = {}) {
  const mode = workModeById(id);
  if (mode.unavailable || !mode.href) return '';

  const protocol = locationLike.protocol || 'http:';
  const hostnameValue = locationLike.hostname || '127.0.0.1';
  const hostname = hostnameValue.includes(':') ? `[${hostnameValue}]` : hostnameValue;
  const currentPort = String(locationLike.port || '');
  const isolatedDevelopment = currentPort === '4001' || currentPort === '4040';

  if (id === WORK_MODE_MAGIC_CUT) {
    return isolatedDevelopment
      ? `${protocol}//${hostname}:4040/`
      : `${protocol}//${hostname}${currentPort ? `:${currentPort}` : ''}/magic-cut`;
  }

  const targetPort = currentPort === '4040' ? '4001' : currentPort;
  const port = targetPort ? `:${targetPort}` : '';
  return `${protocol}//${hostname}${port}${mode.href}`;
}

export function magicCutBootstrapRequest(token, targetHref) {
  const value = String(token || '').trim();
  if (!value) return null;
  const target = new URL(targetHref);
  return {
    url: `${target.origin}/api/auth/bootstrap`,
    options: {
      method: 'POST',
      credentials: 'include',
      headers: { Authorization: `Bearer ${value}` }
    }
  };
}
