export const CREATE_PROJECT_FIELDS = [
  { key: 'name', label: '工程名称' },
  { key: 'ratio', label: '画面比例' },
  { key: 'resolution', label: '分辨率' },
  { key: 'frameRate', label: '帧率' },
]

export function buildCreateProjectPayload(values) {
  const [width, height] = String(values.resolution).split('x').map(Number)
  return {
    name: String(values.name || '').trim(),
    width,
    height,
    frameRate: Number(values.frameRate),
    idempotencyKey: values.idempotencyKey,
  }
}

export function createIdempotencyKey() {
  if (globalThis.crypto?.randomUUID) return globalThis.crypto.randomUUID()
  return `create-${Date.now()}-${Math.random().toString(16).slice(2)}`
}
