import React from 'react'
import { Tag } from 'antd'
import { getStatusPresentation } from './statusPresentation.js'

export default function StatusTag({ status, ...props }) {
  const presentation = getStatusPresentation(status)
  return <Tag color={presentation.color} {...props}>{presentation.label}</Tag>
}
