import { Alert, Button, Table, Tag, Typography, message } from 'antd';
import { RefreshCw } from 'lucide-react';
import { useEffect, useState } from 'react';
import { listAdminErrorLogs } from '../../shared/api/admin';

function formatTime(value) {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? '-' : date.toLocaleString('zh-CN', { hour12: false });
}

export function ErrorLogPage() {
  const [entries, setEntries] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  async function loadLogs() {
    setLoading(true);
    setError('');
    try {
      const data = await listAdminErrorLogs();
      setEntries(Array.isArray(data.entries) ? data.entries : []);
    } catch (failure) {
      // 读取失败要单独标记，不能和“确实没有错误日志”共用同一个空表格。
      setError(failure.message || '读取错误日志失败');
      message.error(failure.message || '读取错误日志失败');
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => { void loadLogs(); }, []);

  return <section className="admin-dashboard">
    <div className="admin-page-heading">
      <div><Typography.Title level={2}>错误日志</Typography.Title><Typography.Paragraph>记录浏览器页面异常、接口失败和服务端未处理异常。敏感凭据会自动脱敏。</Typography.Paragraph></div>
      <Button icon={<RefreshCw size={16} aria-hidden="true" />} onClick={loadLogs} loading={loading}>刷新</Button>
    </div>
    {error ? <Alert
      type="error"
      showIcon
      message="错误日志读取失败"
      description={error}
      action={<Button size="small" onClick={loadLogs} loading={loading}>重试</Button>}
      style={{ marginBottom: 12 }}
    /> : null}
    <Table rowKey="id" loading={loading} dataSource={entries} pagination={{ pageSize: 20 }} locale={{ emptyText: error ? '读取失败，请点击上方「重试」' : '暂无错误日志' }} columns={[
      { title: '时间', dataIndex: 'at', width: 190, render: formatTime },
      { title: '来源', dataIndex: 'kind', width: 190, render: value => <Tag color={String(value).startsWith('server.') ? 'volcano' : 'gold'}>{value}</Tag> },
      { title: '用户', dataIndex: 'username', width: 140, render: value => value || '匿名' },
      { title: '错误信息', dataIndex: 'message', ellipsis: true },
      { title: '页面 / 接口', width: 250, render: (_, entry) => [entry.path, entry.method && `${entry.method}${entry.status ? ` ${entry.status}` : ''}`, entry.source].filter(Boolean).join(' | ') },
      { title: '用户说明', dataIndex: 'context', width: 260, ellipsis: true },
      { title: '堆栈', dataIndex: 'stack', width: 260, ellipsis: true }
    ]} />
  </section>;
}
