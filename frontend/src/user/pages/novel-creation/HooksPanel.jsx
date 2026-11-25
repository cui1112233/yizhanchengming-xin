import { Button, Empty, Spin, Tag } from 'antd';
import { useEffect, useState } from 'react';
import { getNovelCreationHooks } from '../../../shared/api/novelCreation.js';
import useWorkspaceScope from './useWorkspaceScope.js';

const levelText = { error: '严重', warning: '注意', info: '提示' };
const levelColor = { error: 'red', warning: 'orange', info: 'blue' };

function hookAgeText(hook, current) {
  if (!hook.startChapter) return '埋下章节未知';
  if (hook.lastAdvancedChapter) return `第 ${hook.lastAdvancedChapter} 章推进`;
  return `第 ${hook.startChapter} 章埋下`;
}

// 伏笔面板：只读取最新章节结算结果，不触发模型调用。
export default function HooksPanel({ projectId }) {
  const scope = useWorkspaceScope(projectId);
  const [health, setHealth] = useState(null), [loading, setLoading] = useState(false), [errorText, setErrorText] = useState('');
  async function load() {
    if (!projectId) return;
    setLoading(true); setErrorText('');
    try {
      await scope.call(getNovelCreationHooks(projectId, scope.signal), r => setHealth(r.hooks || null));
    } catch (e) {
      if (scope.isCurrent()) setErrorText(e.message || '伏笔状态暂时无法读取');
    } finally {
      if (scope.isCurrent()) setLoading(false);
    }
  }
  useEffect(() => { load(); }, [projectId]);
  if (!projectId) return <Empty description="请先选择作品" />;
  if (loading && !health) return <div className="novel-panel"><Spin /><span>正在读取伏笔状态</span></div>;
  if (errorText) {
    return <div className="novel-panel"><p role="alert">{errorText}</p><Button onClick={load}>重新读取</Button></div>;
  }
  if (!health) return <Empty description="暂无伏笔" />;
  const active = health.activeHooks || [], stale = health.staleHooks || [], alerts = health.alerts || [];
  return <div className="novel-panel" aria-label="伏笔管理">
    <p>基于最新章节结算结果统计，不调用模型。沉寂判定为 {8} 章未推进，旱灾阈值为 {12} 章。</p>
    <dl className="novel-hook-summary">
      {[['活跃', health.activeCount], ['沉寂', health.staleCount], ['已解决', health.resolvedCount], ['章节', health.currentChapter]].map(([label, value]) => <div key={label}><dt>{label}</dt><dd>{value ?? 0}</dd></div>)}
    </dl>
    <Button size="small" onClick={load} loading={loading}>刷新伏笔状态</Button>
    {!!alerts.length && <section className="novel-hook-alerts"><h4>健康提醒</h4>{alerts.map((alert, index) => <p key={index}><Tag color={levelColor[alert.level] || 'default'}>{levelText[alert.level] || alert.level}</Tag>{alert.message}</p>)}</section>}
    {!active.length && !stale.length && <Empty description="还没有埋下伏笔" />}
    {!!active.length && <section className="novel-hook-list"><h4>活跃伏笔 {active.length}</h4>{active.map(hook => <article key={hook.thread} className={stale.some(s => s.thread === hook.thread) ? 'novel-item stale' : 'novel-item'}><b>{hook.thread}</b><p>{hookAgeText(hook, health.currentChapter)}{hook.advancingCount ? ` · 已推进 ${hook.advancingCount} 次` : ''}{stale.some(s => s.thread === hook.thread) ? ' · 已沉寂' : ''}</p>{hook.expectedPayoff && <p>预期回收：{hook.expectedPayoff}</p>}</article>)}</section>}
    {!!health.unknownStartCount && <p role="status">有 {health.unknownStartCount} 个历史伏笔缺少埋下章号，暂不参与沉寂判定。</p>}
  </div>;
}
