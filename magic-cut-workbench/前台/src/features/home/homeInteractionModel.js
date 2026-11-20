export function resolveHomeToolAction(tool) {
  if (!tool?.available) return 'unavailable'
  if (tool.action === 'create-project') return 'create-project'
  if (tool.action === 'navigate') return 'navigate'
  return 'unavailable'
}
