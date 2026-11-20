export const EDITOR_REGIONS = ['素材区', '播放器', '检查器', '工具栏', '时间线']

export function commandForShortcut(event) {
  if (event.code === 'Space') return 'toggle-playback'
  if (event.code === 'Delete') return 'delete-selection'
  if (event.ctrlKey && event.code === 'KeyZ') return event.shiftKey ? 'redo' : 'undo'
  if (event.ctrlKey && event.code === 'KeyC') return 'copy'
  if (event.ctrlKey && event.code === 'KeyV') return 'paste'
  return ''
}

export function projectIDFromPath(pathname) {
  const matched = /^\/projects\/([^/]+)\/edit$/.exec(pathname)
  if (!matched || !/^mcp_[a-zA-Z0-9_-]+$/.test(matched[1])) return ''
  return matched[1]
}
