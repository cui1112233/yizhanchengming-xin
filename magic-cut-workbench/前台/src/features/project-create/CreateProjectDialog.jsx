import { Alert, Form, Input, Modal, Radio, Select } from 'antd'
import { useEffect, useState } from 'react'

import { apiRequest } from '../../shared/api/client.js'
import { buildCreateProjectPayload, createIdempotencyKey } from './createProjectModel.js'

const defaults = () => ({
  name: '未命名工程',
  ratio: '16:9',
  resolution: '1920x1080',
  frameRate: 25,
  idempotencyKey: createIdempotencyKey(),
})

export function CreateProjectDialog({ open, onCancel, onCreated }) {
  const [form] = Form.useForm()
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (open) {
      form.setFieldsValue(defaults())
      setError('')
    }
  }, [form, open])

  async function submit() {
    const values = await form.validateFields()
    setSubmitting(true)
    setError('')
    try {
      const result = await apiRequest('/api/magic-cut/projects', {
        method: 'POST',
        body: JSON.stringify(buildCreateProjectPayload(values)),
      })
      onCreated?.(result.project)
    } catch (requestError) {
      setError(requestError.message || '工程创建失败')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Modal title="新建自由创作" open={open} onCancel={onCancel} onOk={submit} okText="创建" cancelText="取消" confirmLoading={submitting} destroyOnClose={false}>
      {error ? <Alert className="create-project-error" type="error" showIcon message={error} /> : null}
      <Form form={form} layout="vertical" requiredMark={false}>
        <Form.Item name="name" label="工程名称" rules={[{ required: true, whitespace: true, message: '请输入工程名称' }]}>
          <Input maxLength={80} placeholder="请输入工程名称" />
        </Form.Item>
        <Form.Item name="ratio" label="画面比例">
          <Radio.Group optionType="button" buttonStyle="solid" options={['16:9', '9:16', '1:1']} />
        </Form.Item>
        <Form.Item name="resolution" label="分辨率">
          <Select options={[
            { value: '1920x1080', label: '1920 × 1080' },
            { value: '1080x1920', label: '1080 × 1920' },
            { value: '1080x1080', label: '1080 × 1080' },
          ]} />
        </Form.Item>
        <Form.Item name="frameRate" label="帧率">
          <Select options={[24, 25, 30, 50, 60].map(value => ({ value, label: `${value} fps` }))} />
        </Form.Item>
        <Form.Item name="idempotencyKey" hidden><Input /></Form.Item>
      </Form>
    </Modal>
  )
}
