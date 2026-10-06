import React from 'react'
import { Button, Empty, Result, Spin } from 'antd'

export default function PageState({ state = 'empty', title, description, onRetry, className }) {
  if (state === 'loading') {
    return <div className={className || 'page-state'}><Spin size="large" tip={title || '正在加载…'} /></div>
  }
  if (state === 'failed') {
    return <Result className={className} status="error" title={title || '加载失败'} subTitle={description} extra={onRetry ? <Button type="primary" onClick={onRetry}>重试</Button> : null} />
  }
  return <Empty className={className} image={Empty.PRESENTED_IMAGE_SIMPLE} description={title || description || '暂无数据'} />
}
