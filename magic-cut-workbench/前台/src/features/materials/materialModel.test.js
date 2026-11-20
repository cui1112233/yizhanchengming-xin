import assert from 'node:assert/strict'
import test from 'node:test'

import { materialErrorMessage, materialSummary } from './materialModel.js'

test('material summary includes duration and resolution', () => {
  assert.equal(materialSummary({ sizeBytes: 1048576, metadata: { durationMs: 2500, width: 1920, height: 1080 } }), '1.0 MB · 00:02.500 · 1920×1080')
})

test('material errors keep user-facing recovery copy', () => {
  assert.equal(materialErrorMessage('FORMAT_NOT_SUPPORTED'), '暂不支持这个文件格式，请选择常用的视频、音频或图片')
  assert.equal(materialErrorMessage('DISK_SPACE_INSUFFICIENT'), '素材电脑空间不足，请清理磁盘后重试')
  assert.equal(materialErrorMessage('MEDIA_PROBE_FAILED'), '素材解析失败，请确认文件可以正常播放')
})
