export function createWorkModeLinks(locationLike = {}) {
  const protocol = locationLike.protocol || 'http:'
  const rawHostname = locationLike.hostname || '127.0.0.1'
  const hostname = rawHostname.includes(':') ? `[${rawHostname}]` : rawHostname
  return {
    studio: `${protocol}//${hostname}:4001/`,
    agent: `${protocol}//${hostname}:4001/agent`,
    magicCut: '/',
  }
}
