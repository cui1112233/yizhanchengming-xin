import { Button, Input, Modal, Popconfirm, Space, Table, Tag, Typography, message } from 'antd';
import { ExternalLink, Eye, Trash2 } from 'lucide-react';
import { useEffect, useState } from 'react';
import { getHistory, clearHistory, deleteHistory } from '../../shared/api/history';
import { listPlatformProjects } from '../../shared/api/platformProjects';

function formatTime(value) {
  if (!value) return '-';
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? '-' : date.toLocaleString('zh-CN', { hour12: false });
}

export function HistoryPage() {
  const [entries, setEntries] = useState([]);
  const [loading, setLoading] = useState(false);
  const [preview, setPreview] = useState({ open: false, title: '', content: '' });

  async function loadHistory() {
    setLoading(true);
    try {
      const data = await listPlatformProjects();
      setEntries(Array.isArray(data.entries) ? data.entries : []);
    } catch (error) {
      message.error(error.message || '读取历史失败');
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => { loadHistory(); }, []);

  useEffect(() => {
    const entryId = new URLSearchParams(window.location.search).get('entry');
    if (!entryId || !entries.length) return;
    const entry = entries.find(item => item.id === entryId);
    if (entry) void handleOpen(entry);
  }, [entries]);

  async function handleOpen(entry) {
    if (entry.kind !== 'script-history') {
      window.location.href = entry.route;
      return;
    }
    try {
      const content = await getHistory(entry.sourceId);
      setPreview({ open: true, title: entry.name || '生成记录', content });
    } catch (error) {
      message.error(error.message || '读取记录失败');
    }
  }

  async function handleDelete(entry) {
    if (entry.kind !== 'script-history') return;
    try {
      await deleteHistory(entry.sourceId);
      message.success('已删除剧本生成记录');
      await loadHistory();
    } catch (error) {
      message.error(error.message || '删除失败');
    }
  }

  async function handleClear() {
    try {
      await clearHistory();
      message.success('已清空剧本生成记录');
      await loadHistory();
    } catch (error) {
      message.error(error.message || '清空失败');
    }
  }

  return <Space className="utility-page history-page" direction="vertical" size={16} style={{ width: '100%' }}>
    <Space style={{ width: '100%', justifyContent: 'space-between' }}>
      <Typography.Title level={3} style={{ margin: 0 }}>项目与生成历史</Typography.Title>
      <Popconfirm title="清空所有剧本生成记录？" description="水货生产项目不会被删除。" onConfirm={handleClear}>
        <Button danger icon={<Trash2 size={16} aria-hidden="true" />} disabled={!entries.some(entry => entry.kind === 'script-history')}>清空剧本记录</Button>
      </Popconfirm>
    </Space>
    <Table rowKey="id" loading={loading} dataSource={entries} pagination={{ pageSize: 20 }} columns={[
      { title: '来源', dataIndex: 'typeLabel', width: 118, render: value => <Tag>{value}</Tag> },
      { title: '项目 / 记录', dataIndex: 'name', ellipsis: true },
      { title: '说明', dataIndex: 'summary', ellipsis: true },
      { title: '最后更新', dataIndex: 'updatedAt', width: 190, render: formatTime },
      { title: '操作', width: 180, render: (_, entry) => <Space><Button size="small" icon={entry.kind === 'script-history' ? <Eye size={15} aria-hidden="true" /> : <ExternalLink size={15} aria-hidden="true" />} onClick={() => handleOpen(entry)}>{entry.kind === 'script-history' ? '查看' : '打开项目'}</Button>{entry.kind === 'script-history' ? <Popconfirm title="删除这条剧本生成记录？" onConfirm={() => handleDelete(entry)}><Button size="small" danger icon={<Trash2 size={15} aria-hidden="true" />}>删除</Button></Popconfirm> : null}</Space> }
    ]} />
    <Modal title={preview.title} open={preview.open} width={860} footer={<Button onClick={() => setPreview({ open: false, title: '', content: '' })}>关闭</Button>} onCancel={() => setPreview({ open: false, title: '', content: '' })}><Input.TextArea value={preview.content} rows={18} readOnly /></Modal>
  </Space>;
}

export default HistoryPage;
