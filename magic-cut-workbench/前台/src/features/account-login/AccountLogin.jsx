import { Alert, Button, Checkbox, Form, Input } from 'antd'
import { useState } from 'react'

import { loginMagicCut } from './sessionLogin.js'

export function AccountLogin({ onAuthenticated }) {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  async function submit(values) {
    setLoading(true)
    setError('')
    try {
      const account = await loginMagicCut(values)
      onAuthenticated?.(account)
    } catch (caught) {
      setError(caught.message || '登录失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Form className="magic-login-form" layout="vertical" initialValues={{ remember: true }} onFinish={submit}>
      {error ? <Alert type="error" showIcon message={error} /> : null}
      <Form.Item label="账号" name="username" rules={[{ required: true, message: '请输入账号' }]}>
        <Input autoComplete="username" placeholder="请输入一战晟铭账号" />
      </Form.Item>
      <Form.Item label="密码" name="password" rules={[{ required: true, message: '请输入密码' }]}>
        <Input.Password autoComplete="current-password" placeholder="请输入密码" />
      </Form.Item>
      <Form.Item name="remember" valuePropName="checked"><Checkbox>30 天保持登录</Checkbox></Form.Item>
      <Button block type="primary" size="large" htmlType="submit" loading={loading}>登录并进入魔剪</Button>
    </Form>
  )
}
