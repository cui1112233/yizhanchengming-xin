import { Progress, Skeleton, Tag, message } from 'antd';
import { Activity, Coins, Layers3, Sparkles, Zap } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { getMemberCenter, getTeamMembers } from '../../shared/api/member';
import {
  FeatureBars,
  MemberIdentity,
  MetricCard,
  PageHeader,
  Panel,
  formatDate,
  formatTokens
} from './accountCenterShared';

const CONTENT_FEATURES = new Set(['script', 'novel-fetch', 'image', 'shuihuo-production']);

function contentProductionCalls(usage) {
  return Object.entries(usage?.callsByFeature || {})
    .filter(([feature]) => CONTENT_FEATURES.has(feature))
    .reduce((sum, [, calls]) => sum + Number(calls || 0), 0);
}

function formatCost(summary) {
  if (!summary?.costCurrency || summary.estimatedCost === null || summary.estimatedCost === undefined) return '未计价';
  return `${summary.costCurrency} ${Number(summary.estimatedCost).toFixed(4)}`;
}

export default function UsageStatsPage() {
  const [loading, setLoading] = useState(true);
  const [center, setCenter] = useState(null);
  const [team, setTeam] = useState([]);

  useEffect(() => {
    let alive = true;
    Promise.all([getMemberCenter(), getTeamMembers().catch(() => ({ members: [] }))])
      .then(([nextCenter, teamResult]) => {
        if (!alive) return;
        setCenter(nextCenter);
        setTeam(teamResult.members || []);
      })
      .catch(error => message.error(error.message || '用量统计加载失败'))
      .finally(() => { if (alive) setLoading(false); });
    return () => { alive = false; };
  }, []);

  const self = center?.member;
  const day = center?.usage?.day || {};
  const month = center?.usage?.month || {};
  const recent = center?.usage?.recent || [];
  const members = useMemo(() => team.filter(item => item.role === 'member'), [team]);
  const isAdmin = ['dev', 'manager'].includes(self?.role);
  const teamMonth = useMemo(() => members.reduce((sum, item) => sum + Number(item.usage?.month?.totalTokens || 0), 0), [members]);
  const teamCalls = useMemo(() => members.reduce((sum, item) => sum + Number(item.usage?.month?.calls || 0), 0), [members]);
  const teamCost = useMemo(() => members.reduce((sum, item) => sum + Number(item.usage?.month?.estimatedCost || 0), 0), [members]);
  const teamCurrency = members.find(item => item.usage?.month?.costCurrency)?.usage?.month?.costCurrency || null;
  const productionCalls = contentProductionCalls(month);
  const teamProductionCalls = members.reduce((sum, item) => sum + contentProductionCalls(item.usage?.month), 0);
  const ranking = useMemo(() => [...members].sort((a, b) => (b.usage?.month?.totalTokens || 0) - (a.usage?.month?.totalTokens || 0)).slice(0, 6), [members]);
  const maxRecent = Math.max(1, ...recent.slice(0, 12).map(item => Number(item.totalTokens || 0)));

  if (loading) return <div className="account-center-page"><Skeleton active paragraph={{ rows: 10 }} /></div>;
  if (!self) return <div className="account-center-page"><div className="ac-empty">无法读取用量</div></div>;

  return <div className="account-center-page usage-page">
    <PageHeader title="用量与制作" subtitle="查看 Token、调用次数、费用估算、功能构成与内容制作调用" />

    <div className="ac-metrics-grid five">
      <MetricCard label="今日消耗" value={formatTokens(day.totalTokens)} suffix="Tokens" icon={Zap} accent="coral" hint={`${day.calls || 0} 次调用`} />
      <MetricCard label="本月消耗" value={formatTokens(month.totalTokens)} suffix="Tokens" icon={Sparkles} accent="blue" hint={`${month.calls || 0} 次调用`} />
      <MetricCard label="输入 Tokens" value={formatTokens(month.inputTokens)} suffix="Tokens" icon={Layers3} accent="violet" hint="本月输入" />
      <MetricCard label="输出 Tokens" value={formatTokens(month.outputTokens)} suffix="Tokens" icon={Activity} accent="green" hint="本月输出" />
      <MetricCard label="估算费用" value={month.costCurrency ? Number(month.estimatedCost || 0).toFixed(4) : '—'} suffix={month.costCurrency || ''} icon={Coins} accent="gold" hint={`${month.pricedCalls || 0}/${month.billableCalls || 0} 次调用有价格快照`} />
    </div>

    {center?.memberQuota?.level && center.memberQuota.level !== 'unlimited' ? <div className={`ac-quota-alert is-${center.memberQuota.level}`}>
      <strong>个人月额度：{center.memberQuota.percent}%</strong><span>{formatTokens(center.memberQuota.used)} / {formatTokens(center.memberQuota.limit)} Tokens</span>
    </div> : null}
    {center?.teamGovernance?.quota?.level && center.teamGovernance.quota.level !== 'unlimited' ? <div className={`ac-quota-alert is-${center.teamGovernance.quota.level}`}>
      <strong>团队月额度：{center.teamGovernance.quota.percent}%</strong><span>{formatTokens(center.teamGovernance.quota.used)} / {formatTokens(center.teamGovernance.quota.limit)} Tokens · 70/90/100% 分级预警</span>
    </div> : null}

    <div className={`ac-production-summary${isAdmin ? ' is-team' : ''}`}><div><span>本月内容制作调用</span><strong>{productionCalls.toLocaleString('zh-CN')} 次</strong></div>{isAdmin ? <div><span>本月组员制作调用</span><strong>{teamProductionCalls.toLocaleString('zh-CN')} 次</strong></div> : null}<small>按已计费模型调用汇总，当前不等同于按书籍去重的制作量。</small></div>

    <div className="ac-usage-layout">
      <Panel title="最近调用强度" eyebrow="RECENT CALLS" className="ac-usage-chart-panel">
        <div className="ac-bar-chart">
          {recent.slice(0, 12).reverse().map((item, index) => {
            const height = Math.max(8, Math.round((Number(item.totalTokens || 0) / maxRecent) * 100));
            return <div className="ac-bar-column" key={item.id || index} title={`${item.feature || '调用'} · ${formatTokens(item.totalTokens)} Tokens`}>
              <span style={{ height: `${height}%` }} />
              <small>{new Date(item.at).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', hour12: false })}</small>
            </div>;
          })}
          {!recent.length ? <div className="ac-empty">暂无最近调用</div> : null}
        </div>
        <div className="ac-chart-legend"><span><i className="tone-0" /> 单次总 Tokens</span><small>真实 usage 优先；缺失时明确标记估算。</small></div>
      </Panel>

      <Panel title="功能消耗构成" eyebrow="BREAKDOWN">
        <div className="ac-donut-summary"><div className="ac-donut"><strong>{formatTokens(month.totalTokens)}</strong><small>Tokens</small></div></div>
        <FeatureBars byFeature={month.byFeature} total={month.totalTokens} />
      </Panel>

      {isAdmin ? <Panel title="团队成员用量 TOP6" eyebrow="TEAM RANKING">
        <div className="ac-ranking-list usage-ranking">
          {ranking.map((member, index) => {
            const percent = teamMonth ? Math.round((Number(member.usage?.month?.totalTokens || 0) / teamMonth) * 100) : 0;
            return <div className="ac-ranking-row detailed" key={member.username}>
              <span className={`rank rank-${index + 1}`}>{index + 1}</span>
              <MemberIdentity member={member} size={34} />
              <Progress percent={percent} showInfo={false} />
              <div className="ac-ranking-value"><strong>{formatTokens(member.usage?.month?.totalTokens)}</strong><small>{contentProductionCalls(member.usage?.month)} 次制作 · {member.usage?.month?.costCurrency ? `${member.usage.month.costCurrency} ${Number(member.usage.month.estimatedCost || 0).toFixed(4)}` : `${percent}%`}</small></div>
            </div>;
          })}
          {!ranking.length ? <div className="ac-empty">暂无团队用量</div> : null}
        </div>
        <div className="ac-side-summary"><span>团队本月合计</span><strong>{formatTokens(teamMonth)} Tokens · {teamCalls} 次调用 · {teamProductionCalls} 次制作{teamCurrency ? ` · 估算 ${teamCurrency} ${teamCost.toFixed(4)}` : ''}</strong></div>
      </Panel> : null}
    </div>

    <Panel title="最近用量明细" eyebrow="USAGE LEDGER">
      <div className="ac-usage-table">
        <div className="ac-usage-table-head"><span>时间</span><span>功能</span><span>模型</span><span>输入</span><span>输出</span><span>总 Tokens</span><span>口径 / 费用</span></div>
        {recent.slice(0, 20).map(item => <div className="ac-usage-table-row" key={item.id}>
          <span>{formatDate(item.at)}</span>
          <strong>{item.feature || 'unknown'}</strong>
          <span>{item.model || '—'}</span>
          <span>{formatTokens(item.inputTokens)}</span>
          <span>{formatTokens(item.outputTokens)}</span>
          <b>{formatTokens(item.totalTokens)}</b>
          <span className="ac-usage-cost-cell"><Tag color={item.metadata?.usageEstimated ? 'gold' : item.usageKnown ? 'green' : 'default'}>{item.metadata?.usageEstimated ? '估算 usage' : item.usageKnown ? '真实 usage' : '调用记录'}</Tag><small>{item.costKnown ? `${item.costCurrency} ${Number(item.estimatedCost).toFixed(6)}` : '未配置价格'}</small></span>
        </div>)}
        {!recent.length ? <div className="ac-empty">暂无用量明细</div> : null}
      </div>
    </Panel>
  </div>;
}
