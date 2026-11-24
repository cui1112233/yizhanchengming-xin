import { Button, Input, Text, View } from '@tarojs/components';
import { useEffect, useRef, useState } from 'react';
import { apiRequest } from '../../services/api';
export default function ScriptHistoryPanel({ onSelect, disabled = false }: { onSelect: (id: string) => void; disabled?: boolean }) {
  const [entries, setEntries] = useState<any[]>([]), [query, setQuery] = useState(''), [busy, setBusy] = useState(false), [error, setError] = useState('');
  const active = useRef(true), lock = useRef(false);
  const load = async () => {
    if (lock.current) return; lock.current = true; setBusy(true); setError('');
    try { const response = await apiRequest<any>('/api/history'); if (!Array.isArray(response?.entries)) throw new Error('历史列表格式不正确，请重试'); if (active.current) setEntries(response.entries); }
    catch (cause: any) { if (active.current) setError(cause.message || '历史读取失败'); }
    finally { lock.current = false; if (active.current) setBusy(false); }
  };
  useEffect(() => { void load(); return () => { active.current = false; }; }, []);
  const name = (item: any) => String(item.title || item.novelText?.split(/\r?\n/)[0]?.slice(0, 48) || '未命名剧本');
  const visible = entries.filter(item => name(item).includes(query.trim()));
  return <View className="studio-history-panel"><Input className="mp-input" value={query} onInput={event => setQuery(event.detail.value)} placeholder="搜索剧本原文或名称" /><View className="studio-section-head"><Text>剧本创作历史</Text><Button className="studio-link" disabled={busy || disabled} onClick={() => void load()}>刷新</Button></View>{busy ? <Text className="studio-caption">正在读取历史记录…</Text> : error ? <Text className="mp-error">{error}</Text> : !visible.length ? <Text className="studio-caption">{query ? '没有找到匹配的剧本' : '暂无生成历史'}</Text> : visible.map(item => <Button key={item.id} className="studio-history-row" disabled={disabled || !item.id} onClick={() => onSelect(String(item.id))}><View><Text>{name(item)}</Text><Text className="studio-caption">{item.createdAt ? new Date(item.createdAt).toLocaleString() : item.mode || '剧本创作'} · {item.format || ''}</Text></View><Text>打开</Text></Button>)}</View>;
}
