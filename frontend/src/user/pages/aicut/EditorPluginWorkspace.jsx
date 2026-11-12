import { Button, Empty, Input } from 'antd';
import { useState } from 'react';
import './workspace-panels.css';

export default function EditorPluginWorkspace({ tools = [] }) {
  const [query, setQuery] = useState('');
  const visible = tools.filter(tool => typeof tool.action === 'function' && `${tool.title} ${tool.detail}`.toLowerCase().includes(query.trim().toLowerCase()));
  return <div className="aicut-workspace-panel"><header><strong>剪辑插件</strong><p>对当前工程使用已接入的处理工具。</p></header><Input aria-label="搜索剪辑插件" placeholder="搜索工具或功能" value={query} onChange={event => setQuery(event.target.value)} allowClear />
    {visible.length ? visible.map(tool => <section key={tool.id}><h3>{tool.title}</h3><p>{tool.detail}</p><Button block onClick={tool.action}>打开{tool.title}</Button></section>) : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="没有匹配的剪辑插件" />}
  </div>;
}
