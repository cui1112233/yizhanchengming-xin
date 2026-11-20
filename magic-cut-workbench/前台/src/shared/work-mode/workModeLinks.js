export function createWorkModeLinks(locationLike = {}) {
  const protocol = locationLike.protocol || 'http:'
  const rawHostname = locationLike.hostname || '127.0.0.1'
  const hostname = rawHostname.includes(':') ? `[${rawHostname}]` : rawHostname
  return {
    studio: `${protocol}//${hostname}:4001/`,
    agent: `${protocol}//${hostname}:4001/agent/chat`,
    magicCut: '/',
  }
}

export function createWorkModeModel(locationLike = {}) {
  const links = createWorkModeLinks(locationLike)
  return {
    brandLabel: '魔剪',
    logoSrc: '/assets/brand-logo-white.png',
    modes: [
      { id: 'studio', label: '一战晟铭', href: links.studio },
      { id: 'agent', label: 'Agent 代理', href: links.agent },
      { id: 'magic-cut', label: '魔剪', href: links.magicCut, active: true },
      { id: 'duet', label: 'Duet', disabled: true },
    ],
  }
}
