function formatBytes(value) {
  if (!Number.isFinite(value) || value <= 0) return '0 B'
  if (value >= 1024 ** 3) return `${(value / 1024 ** 3).toFixed(1)} GB`
  if (value >= 1024 ** 2) return `${(value / 1024 ** 2).toFixed(1)} MB`
  if (value >= 1024) return `${(value / 1024).toFixed(1)} KB`
  return `${value} B`
}

function formatDuration(durationMs = 0) {
  const total = Math.max(0, Math.round(durationMs))
  const minutes = Math.floor(total / 60000)
  const seconds = Math.floor((total % 60000) / 1000)
  const millis = total % 1000
  return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}.${String(millis).padStart(3, '0')}`
}

export function materialSummary(material) {
  const metadata = material?.metadata || {}
  const parts = [formatBytes(material?.sizeBytes || 0), formatDuration(metadata.durationMs)]
  if (metadata.width && metadata.height) parts.push(`${metadata.width}×${metadata.height}`)
  return parts.join(' · ')
}

const errors = {
  FORMAT_NOT_SUPPORTED: '暂不支持这个文件格式，请选择常用的视频、音频或图片',
  FILE_TOO_LARGE: '素材文件太大，请压缩后重试',
  FILE_PERMISSION_DENIED: '无法写入素材目录，请检查这台电脑的目录权限',
  DISK_SPACE_INSUFFICIENT: '素材电脑空间不足，请清理磁盘后重试',
  MEDIA_PROBE_FAILED: '素材解析失败，请确认文件可以正常播放',
}

export function materialErrorMessage(code) {
  return errors[code] || '素材导入失败，请稍后重试'
}
