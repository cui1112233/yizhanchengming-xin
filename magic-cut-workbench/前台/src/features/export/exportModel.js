export const EXPORT_FIELDS = ['输出名称', '分辨率', '帧率', '视频编码', '音频编码', '封装格式']

export const DEFAULT_EXPORT_VALUES = {
  outputName: '未命名工程',
  width: 1920,
  height: 1080,
  frameRate: 25,
  videoCodec: 'h264',
  audioCodec: 'aac',
  container: 'mp4',
}

export function buildExportPayload(projectId, values) {
  return {
    projectId,
    outputName: String(values.outputName || '').trim(),
    width: Number(values.width),
    height: Number(values.height),
    frameRate: Number(values.frameRate),
    videoCodec: values.videoCodec,
    audioCodec: values.audioCodec,
    container: values.container,
  }
}

const messages = {
  EXPORT_TIMELINE_EMPTY: '时间线中没有可导出的视频',
  EXPORT_MATERIAL_MISSING: '时间线中的素材已经丢失，请重新导入后再试',
  EXPORT_FFMPEG_UNAVAILABLE: '本机没有找到 FFmpeg，请检查运行环境',
  EXPORT_OUTPUT_UNAVAILABLE: '输出目录不可用，请检查磁盘空间和写入权限',
  EXPORT_FORMAT_NOT_SUPPORTED: '当前只支持 H.264 + AAC 的 MP4 导出',
}

export function exportErrorMessage(code) {
  return messages[code] || '创建导出任务失败，请稍后重试'
}
