import { Alert, Button, Form, Input, InputNumber, Modal, Select, Space } from 'antd'
import { useState } from 'react'

import { apiRequest } from '../../shared/api/client.js'
import { buildExportPayload, DEFAULT_EXPORT_VALUES, exportErrorMessage } from './exportModel.js'

export function ExportDialog({ open, projectID, onClose }) {
  const [form] = Form.useForm()
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [taskID, setTaskID] = useState('')

  async function submit(values) {
    setSubmitting(true)
    setError('')
    try {
      const payload = await apiRequest('/api/magic-cut/exports', { method: 'POST', body: JSON.stringify(buildExportPayload(projectID, values)) })
      setTaskID(payload.taskId || '')
    } catch (caught) {
      setError(exportErrorMessage(caught.code))
    } finally {
      setSubmitting(false)
    }
  }

  function close() {
    setError('')
    setTaskID('')
    onClose?.()
  }

  return (
    <Modal title="导出视频" open={open} onCancel={close} footer={null} destroyOnHidden width={520}>
      {taskID ? (
        <Space direction="vertical" size="middle" style={{ width: '100%' }}>
          <Alert type="success" showIcon message="导出任务已创建" description={`任务编号：${taskID}`} />
          <Button type="primary" href="/tasks">前往任务中心</Button>
        </Space>
      ) : (
        <Form form={form} layout="vertical" initialValues={DEFAULT_EXPORT_VALUES} onFinish={submit}>
          {error ? <Alert type="error" showIcon message={error} style={{ marginBottom: 16 }} /> : null}
          <Form.Item name="outputName" label="输出名称" rules={[{ required: true, whitespace: true, message: '请输入输出名称' }, { pattern: /^[^\\/:*?"<>|]+$/, message: '输出名称不能包含路径或特殊字符' }]}><Input maxLength={120} /></Form.Item>
          <Form.Item label="分辨率" required>
            <Space.Compact block>
              <Form.Item name="width" noStyle rules={[{ required: true }]}><InputNumber min={16} max={7680} addonBefore="宽" style={{ width: '50%' }} /></Form.Item>
              <Form.Item name="height" noStyle rules={[{ required: true }]}><InputNumber min={16} max={4320} addonBefore="高" style={{ width: '50%' }} /></Form.Item>
            </Space.Compact>
          </Form.Item>
          <Form.Item name="frameRate" label="帧率"><Select options={[24, 25, 30, 50, 60].map(value => ({ value, label: `${value} fps` }))} /></Form.Item>
          <Form.Item name="videoCodec" label="视频编码"><Select options={[{ value: 'h264', label: 'H.264' }]} /></Form.Item>
          <Form.Item name="audioCodec" label="音频编码"><Select options={[{ value: 'aac', label: 'AAC' }]} /></Form.Item>
          <Form.Item name="container" label="封装格式"><Select options={[{ value: 'mp4', label: 'MP4' }]} /></Form.Item>
          <Space style={{ width: '100%', justifyContent: 'flex-end' }}><Button onClick={close}>取消</Button><Button type="primary" htmlType="submit" loading={submitting}>开始导出</Button></Space>
        </Form>
      )}
    </Modal>
  )
}
