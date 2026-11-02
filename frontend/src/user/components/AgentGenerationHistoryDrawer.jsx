import { Button, Drawer, Empty, Input, Select, Tag, message } from 'antd';
import { History, RefreshCw } from 'lucide-react';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { insertAgentGenerationHistoryResult, listAgentGenerationHistory } from '../../shared/api/agent';
import { generationHistoryKindOptions, generationHistoryRows } from '../pages/agent-generation-history';

const STATE_OPTIONS = [
  { value: '', label: '全部状态' },
  { value: 'awaiting_confirmation', label: '待确认' },
  { value: 'completed', label: '已完成' },
  { value: 'failed', label: '失败' },
  { value: 'cancelled', label: '已取消' }
];

const STATE_LABELS = {
  awaiting_confirmation: '待确认', completed: '已完成', failed: '失败', cancelled: '已取消'
};

export function AgentGenerationHistoryDrawer({ open, onClose, currentProjectId, onInserted }) {
  const [filters, setFilters] = useState({ kind: '', state: '', query: '' });
  const [tasks, setTasks] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [insertingKey, setInsertingKey] = useState('');

  const load = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const result = await listAgentGenerationHistory(filters);
      setTasks(Array.isArray(result?.tasks) ? result.tasks : []);
    } catch (caught) {
      setError(caught?.message || '生成历史读取失败');
    } finally {
      setLoading(false);
    }
  }, [filters]);

  useEffect(() => {
    if (open) void load();
  }, [open, load]);

  const rows = useMemo(() => generationHistoryRows(tasks), [tasks]);
  const updateFilter = (key, value) => setFilters(current => ({ ...current, [key]: value }));
  const insert = async row => {
    if (!currentProjectId) return;
    const key = `${row.taskId}:${row.resultId}`;
    setInsertingKey(key);
    try {
      const result = await insertAgentGenerationHistoryResult(currentProjectId, row.taskId, row.resultId);
      message.success('已放入当前画布');
      onInserted?.(result);
    } catch (caught) {
      message.error(caught?.message || '放入画布失败');
    } finally {
      setInsertingKey('');
    }
  };

  return <Drawer
    title={<span className="agent-generation-history-drawer__title"><History size={17} />生成历史</span>}
    open={open}
    onClose={onClose}
    width={440}
    className="agent-generation-history-drawer"
    extra={<Button type="text" aria-label="刷新生成历史" icon={<RefreshCw size={16} />} loading={loading} onClick={() => void load()} />}
  >
    <div className="agent-generation-history-drawer__filters">
      <Input.Search allowClear placeholder="搜索项目、模型或素材" value={filters.query} onChange={event => updateFilter('query', event.target.value)} onSearch={() => void load()} />
      <Select value={filters.kind} options={generationHistoryKindOptions} onChange={value => updateFilter('kind', value)} />
      <Select value={filters.state} options={STATE_OPTIONS} onChange={value => updateFilter('state', value)} />
    </div>
    {error ? <p className="agent-generation-history-drawer__error" role="alert">{error}</p> : null}
    <div className="agent-generation-history-drawer__list" aria-label="生成历史结果">
      {!loading && !rows.length ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无可复用的生成结果" /> : null}
      {rows.map(row => {
        const key = `${row.taskId}:${row.resultId}`;
        return <article className="agent-generation-history-drawer__item" key={key}>
          <div className="agent-generation-history-drawer__item-heading">
            <strong title={row.name}>{row.name}</strong>
            <Tag>{row.kind === 'image' ? '图片' : row.kind === 'video' ? '视频' : '音频'}</Tag>
          </div>
          <small>来源项目：{row.projectId}</small>
          <small>模型：{row.modelId || '未记录'} · {STATE_LABELS[row.state] || row.state || '未知状态'}</small>
          <Button size="small" type="primary" disabled={!currentProjectId} loading={insertingKey === key} onClick={() => void insert(row)}>放入画布</Button>
        </article>;
      })}
    </div>
  </Drawer>;
}
