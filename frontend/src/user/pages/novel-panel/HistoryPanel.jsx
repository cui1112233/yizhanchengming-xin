import { Empty, List, Tag } from 'antd';
import { useMemo } from 'react';
import { normalizeNovelHistory } from './historyPanelModel';

export function HistoryPanel({ history }) {
  const records = useMemo(() => normalizeNovelHistory(history), [history]);
  if (!records.length) return <Empty description="暂无历史记录" />;
  return <List dataSource={records} renderItem={item => <List.Item><span>{item.id}</span><Tag>{item.status}</Tag><time>{item.updatedAt}</time></List.Item>} />;
}

export default HistoryPanel;
