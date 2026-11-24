import { Alert, Button, Progress, Skeleton, Tag } from 'antd';
import { Activity, Coins, Layers3, Sparkles, Zap } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { getMemberUsage, getTeamMembers } from '../../shared/api/member';
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
  const amount = summary?.actualCost ?? summary?.estimatedCost;
  if (!summary?.costCurrency || amount === null || amount === undefined) return '—（待核账）';
  return `${summary.costCurrency} ${Number(amount).toFixed(4)}`;
}

const knownNumber = value => value !== null && value !== undefined && Number.isFinite(Number(value));
function summaryTokens(summary, field = 'totalTokens') {
  if (!knownNumber(summary?.[field]) || summary?.usageKnown === false || (Number(summary?.calls) > 0 && summary?.knownUsageCalls === 0)) return '—';
  return formatTokens(summary[field]);
}

export default function UsageStatsPage() {
  const [loading, setLoading] = useState(true);
  const [center, setCenter] = useState(null);
  const [team, setTeam] = useState([]);
  const [teamError, setTeamError] = useState('');
  const [loadError, setLoadError] = useState('');
  const [reloadKey, setReloadKey] = useState(0);

  useEffect(() => {
    let alive = true;
    setLoading(true);
    setTeamError('');
    setLoadError('');
    (async () => {
      const nextCenter = await getMemberUsage();
      if (!alive) return;
      setCenter(nextCenter);
      setTeam([]);
      if (['dev', 'manager'].includes(nextCenter?.member?.role)) {
        try {
          const result = await getTeamMembers();
          if (alive) setTeam(result?.members || []);
        } catch (error) {
          if (alive) setTeamError(error.message || '团队成员用量读取失败');
        }
      }
    })()
      .catch(error => { if (alive) { setLoadError(error.message || '用量账单加载失败'); setCenter(null); } })
      .finally(() => { if (alive) setLoading(false); });
    return () => { alive = false; };
  }, [reloadKey]);

  const self = center?.member;
  const day = center?.usage?.day || {};
  const month = center?.usage?.month || {};
  const recent = center?.usage?.recent || [];
  const members = useMemo(() => team.filter(item => item.role === 'member'), [team]);
  const isAdmin = ['dev', 'manager'].includes(self?.role);
  const teamMonth = useMemo(() => members.reduce((sum, item) => sum + Number(item.usage?.month?.totalTokens || 0), 0), [members]);
  const teamCalls = useMemo(() => members.reduce((sum, item) => sum + Number(item.usage?.month?.calls || 0), 0), [members]);
  const pricedMembers = members.filter(item => knownNumber(item.usage?.month?.actualCost ?? item.usage?.month?.estimatedCost));
  const teamCost = pricedMembers.reduce((sum, item) => sum + Number(item.usage.month.actualCost ?? item.usage.month.estimatedCost), 0);
  const teamCurrency = members.find(item => item.usage?.month?.costCurrency)?.usage?.month?.costCurrency || null;
  const productionCalls = contentProductionCalls(month);
  const teamProductionCalls = members.reduce((sum, item) => sum + contentProductionCalls(item.usage?.month), 0);
  const ranking = useMemo(() => [...members].sort((a, b) => (b.usage?.month?.totalTokens || 0) - (a.usage?.month?.totalTokens || 0)).slice(0, 6), [members]);
  const chartRecent = recent.slice(0, 12).filter(item => item.usageKnown !== false && knownNumber(item.totalTokens));
  const maxRecent = Math.max(1, ...chartRecent.map(item => Number(item.totalTokens)));

  if (loading) return <div className="account-center-page"><Skeleton active paragraph={{ rows: 10 }} /></div>;
  if (!self) return <div className="account-center-page"><Alert type="error" showIcon message={loadError || '无法读取用量'} action={<Button onClick={() => setReloadKey(key => key + 1)}>重试</Button>} /></div>;

  return <div className="account-center-page usage-page">
    <PageHeader title="用量与制作" subtitle="查看 Token、调用次数、实际扣费与功能构成" />

    {center?.usage?.historyCoverage === 'tracked_requests' ? <Alert type="info" showIcon style={{ marginBottom: 16 }} message="YZCM API 真实账单" description="仅统计接入后能关联到当前成员的文本、图片和音频请求。费用来自实际扣费记录；未匹配到账单显示待核账。历史共享 Key 账单及异步视频结算暂不计入。" /> : null}
    <Button onClick={() => setReloadKey(key => key + 1)} style={{ marginBottom: 16 }}>刷新账单</Button>
    <div className="ac-metrics-grid five">
      <MetricCard label="今日消耗" value={summaryTokens(day)} suffix="Tokens" icon={Zap} accent="coral" hint={`${day.calls || 0} 次调用${day.pendingCalls ? ` · ${day.pendingCalls} 次待核账` : ''}`} />
      <MetricCard label="本月消耗" value={summaryTokens(month)} suffix="Tokens" icon={Sparkles} accent="blue" hint={`${month.calls || 0} 次调用${month.pendingCalls ? ` · ${month.pendingCalls} 次待核账` : ''}`} />
      <MetricCard label="输入 Tokens" value={summaryTokens(month, 'inputTokens')} suffix="Tokens" icon={Layers3} accent="violet" hint="本月已知输入" />
      <MetricCard label="输出 Tokens" value={summaryTokens(month, 'outputTokens')} suffix="Tokens" icon={Activity} accent="green" hint="本月已知输出" />
      <MetricCard label="实际扣费" value={month.costCurrency ? (month.actualCost == null ? '—' : Number(month.actualCost).toFixed(4)) : '—'} suffix={month.costCurrency || ''} icon={Coins} accent="gold" hint={`${month.pricedCalls || 0}/${month.billableCalls || 0} 次调用已核账`} />
    </div>

    {center?.memberQuota?.level && center.memberQuota.level !== 'unlimited' ? <div className={`ac-quota-alert is-${center.memberQuota.level}`}>
      <strong>个人月额度：{center.memberQuota.percent}%</strong><span>{formatTokens(center.memberQuota.used)} / {formatTokens(center.memberQuota.limit)} Tokens</span>
    </div> : null}
    {center?.teamGovernance?.quota?.level && center.teamGovernance.quota.level !== 'unlimited' ? <div className={`ac-quota-alert is-${center.teamGovernance.quota.level}`}>
      <strong>团队月额度：{center.teamGovernance.quota.percent}%</strong><span>{formatTokens(center.teamGovernance.quota.used)} / {formatTokens(center.teamGovernance.quota.limit)} Tokens · 70/90/100% 分级预警</span>
    </div> : null}

    <div className={`ac-production-summary${isAdmin ? ' is-team' : ''}`}><div><span>本月内容功能调用</span><strong>{productionCalls.toLocaleString('zh-CN')} 次</strong></div>{isAdmin ? <div><span>本月组员功能调用</span><strong>{teamProductionCalls.toLocaleString('zh-CN')} 次</strong></div> : null}<small>按功能调用记录汇总，不等同于作品数量。</small></div>

    <div className="ac-usage-layout">
      <Panel title="最近调用强度" eyebrow="RECENT CALLS" className="ac-usage-chart-panel">
        <div className="ac-bar-chart">
          {[...chartRecent].reverse().map((item, index) => {
            const height = Math.round((Number(item.totalTokens) / maxRecent) * 100);
            return <div className="ac-bar-column" key={item.id || index} title={`${item.feature || '调用'} · ${formatTokens(item.totalTokens)} Tokens`}>
              <span style={{ height: `${height}%` }} />
              <small>{new Date(item.at).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', hour12: false })}</small>
            </div>;
          })}
          {!chartRecent.length ? <div className="ac-empty">{recent.length ? '调用用量待核账' : '暂无最近调用'}</div> : null}
        </div>
        <div className="ac-chart-legend"><span><i className="tone-0" /> 单次总 Tokens</span><small>真实 usage 优先；缺失时明确标记估算。</small></div>
      </Panel>

      <Panel title="功能消耗构成" eyebrow="BREAKDOWN">
        <div className="ac-donut-summary"><div className="ac-donut"><strong>{summaryTokens(month)}</strong><small>Tokens</small></div></div>
        <FeatureBars byFeature={month.byFeature} total={month.totalTokens} />
      </Panel>

      {isAdmin ? <Panel title="团队成员用量 TOP6" eyebrow="TEAM RANKING">
        {teamError ? <Alert
          type="error"
          showIcon
          message="团队成员用量读取失败"
          description={teamError}
          action={<Button size="small" onClick={() => setReloadKey(key => key + 1)}>重试</Button>}
          style={{ marginBottom: 12 }}
        /> : null}
        <div className="ac-ranking-list usage-ranking">
          {ranking.map((member, index) => {
            const percent = teamMonth ? Math.round((Number(member.usage?.month?.totalTokens || 0) / teamMonth) * 100) : 0;
            return <div className="ac-ranking-row detailed" key={member.username}>
              <span className={`rank rank-${index + 1}`}>{index + 1}</span>
              <MemberIdentity member={member} size={34} />
              <Progress percent={percent} showInfo={false} />
              <div className="ac-ranking-value"><strong>{summaryTokens(member.usage?.month)}</strong><small>{contentProductionCalls(member.usage?.month)} 次功能调用 · {formatCost(member.usage?.month)}</small></div>
            </div>;
          })}
          {!ranking.length && !teamError ? <div className="ac-empty">暂无团队用量</div> : null}
        </div>
        <div className="ac-side-summary"><span>团队本月已知合计</span><strong>{teamError ? '读取失败，暂不展示合计' : `${members.some(item => summaryTokens(item.usage?.month) !== '—') ? formatTokens(teamMonth) : '—'} Tokens · ${teamCalls} 次调用 · ${teamProductionCalls} 次功能调用${teamCurrency && pricedMembers.length ? ` · 已核账 ${teamCurrency} ${teamCost.toFixed(4)}` : ' · 费用待核账'}`}</strong></div>
      </Panel> : null}
    </div>

    <Panel title="最近用量明细" eyebrow="USAGE LEDGER">
      <div className="ac-usage-table">
        <div className="ac-usage-table-head"><span>时间</span><span>功能</span><span>模型</span><span>输入</span><span>输出</span><span>总 Tokens</span><span>口径 / 费用</span></div>
        {recent.slice(0, 20).map(item => <div className="ac-usage-table-row" key={item.id}>
          <span>{formatDate(item.at)}</span>
          <strong>{item.feature || 'unknown'}</strong>
          <span>{item.model || '—'}</span>
          <span>{item.usageKnown === false ? '—' : summaryTokens(item, 'inputTokens')}</span>
          <span>{item.usageKnown === false ? '—' : summaryTokens(item, 'outputTokens')}</span>
          <b>{item.usageKnown === false ? '—' : summaryTokens(item)}</b>
          <span className="ac-usage-cost-cell"><Tag color={item.metadata?.usageEstimated ? 'gold' : item.usageKnown ? 'green' : 'default'}>{item.source === 'yzcm_billing_log' ? (item.costKnown ? '已核账' : '待核账') : item.metadata?.usageEstimated ? '估算用量' : item.usageKnown ? '真实 usage' : '调用记录'}</Tag><small>{item.costKnown ? `${item.costCurrency} ${Number(item.estimatedCost).toFixed(6)}` : '待核账'}</small></span>
        </div>)}
        {!recent.length ? <div className="ac-empty">暂无用量明细</div> : null}
      </div>
    </Panel>
  </div>;
}
