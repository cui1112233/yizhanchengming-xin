import React, { useEffect, useState } from 'react'
import { Alert, Button, Card, Form, Input, Result, Spin, Typography } from 'antd'
import * as defaultApi from './api.js'

export default function AuthBoundary({ children, api = defaultApi, onAuthenticated }) {
  const [status, setStatus] = useState('initializing')
  const [user, setUser] = useState(null)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    let active = true

    const markUnauthenticated = () => {
      if (!active) return
      setUser(null)
      setStatus('unauthenticated')
    }

    const handleExpired = () => {
      if (!active) return
      setError('当前登录状态已过期，请重新登录。')
      markUnauthenticated()
    }

    window.addEventListener('ycm:auth-unauthenticated', handleExpired)
    Promise.resolve(api.getCurrentUser())
      .then((payload) => {
        if (!active) return
        setUser(payload?.user || null)
        if (payload?.user) onAuthenticated?.(payload.user)
        setError('')
        setStatus(payload?.user ? 'authenticated' : 'unauthenticated')
      })
      .catch((reason) => {
        if (!active) return
        if (reason?.status === 403) {
          setUser(null)
          setError('')
          setStatus('forbidden')
          return
        }
        markUnauthenticated()
      })

    return () => {
      active = false
      window.removeEventListener('ycm:auth-unauthenticated', handleExpired)
    }
  }, [api, onAuthenticated])

  const submitLogin = async (values) => {
    setSubmitting(true)
    setError('')
    try {
      const payload = await api.login(values)
      if (!payload?.user) {
        throw new Error('登录响应缺少用户信息')
      }
      setUser(payload.user)
      onAuthenticated?.(payload.user)
      setStatus('authenticated')
    } catch (loginError) {
      setError(loginError instanceof Error ? loginError.message : '登录失败，请重试。')
      setStatus('unauthenticated')
    } finally {
      setSubmitting(false)
    }
  }

  const submitLogout = async () => {
    await api.logout()
    setUser(null)
    setStatus('unauthenticated')
  }

  if (status === 'initializing') {
    return (
      <main className="page-shell" style={{ minHeight: '100vh', display: 'grid', placeItems: 'center' }}>
        <div style={{ textAlign: 'center' }}>
          <Spin size="large" />
          <Typography.Paragraph style={{ marginTop: 16 }}>正在恢复登录状态…</Typography.Paragraph>
        </div>
      </main>
    )
  }

  if (status === 'forbidden') {
    return (
      <main className="page-shell" style={{ minHeight: '100vh', display: 'grid', placeItems: 'center' }}>
        <Result status="403" title="Forbidden" subTitle="无权限访问此页面。" />
      </main>
    )
  }

  if (status === 'unauthenticated') {
    return (
      <main className="login-page">
        <Card className="login-card" style={{ width: 'min(440px, 92vw)' }}>
          <div className="login-brand">
            <img src="/assets/brand-logo-black.png" alt="一战晟铭" />
          </div>
          <Typography.Title level={2}>一战晟铭登录</Typography.Title>
          <Typography.Paragraph type="secondary" className="login-subtitle">继续你的创作工作流</Typography.Paragraph>
          {error ? <Alert type="warning" showIcon message={error} style={{ marginBottom: 16 }} /> : null}
          <Form layout="vertical" onFinish={submitLogin}>
            <Form.Item name="username" label="用户名" rules={[{ required: true, message: '请输入用户名' }]}>
              <Input autoComplete="username" />
            </Form.Item>
            <Form.Item name="password" label="密码" rules={[{ required: true, message: '请输入密码' }]}>
              <Input.Password autoComplete="current-password" />
            </Form.Item>
            <Button type="primary" htmlType="submit" loading={submitting} block>登录并进入工作台</Button>
          </Form>
          <Typography.Paragraph type="secondary" style={{ marginTop: 16, marginBottom: 0 }}>
            登录成功后会继续停留在你原本访问的页面。
          </Typography.Paragraph>
        </Card>
      </main>
    )
  }

  // Test and embedding callers may supply a DOM element. Do not leak account
  // objects or event props onto it; only workspace components receive them.
  return React.isValidElement(children) && typeof children.type !== 'string'
    ? React.cloneElement(children, { currentUser: user, onLogout: submitLogout })
    : children
}
