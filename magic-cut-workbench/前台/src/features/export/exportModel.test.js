import assert from 'node:assert/strict'
import test from 'node:test'

import { DEFAULT_EXPORT_VALUES, EXPORT_FIELDS, buildExportPayload, exportErrorMessage } from './exportModel.js'

test('export dialog keeps the original-use fields in order', () => {
  assert.deepEqual(EXPORT_FIELDS, ['输出名称', '分辨率', '帧率', '视频编码', '音频编码', '封装格式'])
  assert.deepEqual(DEFAULT_EXPORT_VALUES, { outputName: '未命名工程', width: 1920, height: 1080, frameRate: 25, videoCodec: 'h264', audioCodec: 'aac', container: 'mp4' })
})

test('export payload includes the current project and safe defaults', () => {
  assert.deepEqual(buildExportPayload('mcp_1', { ...DEFAULT_EXPORT_VALUES, outputName: '第一集' }), { projectId: 'mcp_1', outputName: '第一集', width: 1920, height: 1080, frameRate: 25, videoCodec: 'h264', audioCodec: 'aac', container: 'mp4' })
})

test('export errors keep actionable Chinese recovery copy', () => {
  assert.equal(exportErrorMessage('EXPORT_TIMELINE_EMPTY'), '时间线中没有可导出的视频')
  assert.equal(exportErrorMessage('EXPORT_FFMPEG_UNAVAILABLE'), '本机没有找到 FFmpeg，请检查运行环境')
  assert.equal(exportErrorMessage('UNKNOWN'), '创建导出任务失败，请稍后重试')
})
