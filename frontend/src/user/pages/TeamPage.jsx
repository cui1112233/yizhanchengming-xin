import { Button, Checkbox, Form, Input, InputNumber, Modal, Progress, Select, Skeleton, Tag, Tooltip, message } from 'antd';
import { Copy, KeyRound, Link2, MoreHorizontal, Pencil, Plus, RefreshCw, ShieldCheck, UserPlus, UsersRound } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import {
  createTeamInvite,
  createTeamMember,
  archiveTeamMember,
  getMemberCenter,
  getTeamInvites,
  getTeamMembers,
  getTeamMeta,
  getTeamGovernance,
  renameTeam,
  resetTeamMemberPassword,
  restoreArchivedMember,
  setTeamMemberApiScopes,
  setTeamMemberStatus,
  updateTeamGovernance,
  updateTeamMember
} from '../../shared/api/member';
import { listGlobalAccounts } from '../../shared/api/accountAdmin';
import { MemberIdentity, PageHeader, Panel, RoleBadge, formatDate, formatTokens } from './accountCenterShared';

const SCOPE_OPTIONS = [
  { label: '文本', value: 'text' },
  { label: '生图', value: 'image' },
  { label: '配音', value: 'tts' },
  { label: '视频生成', value: 'video' }
];
const CONTENT_FEATURES = new Set(['script', 'novel-fetch', 'image', 'shuihuo-production']);

function scopesOf(member) {
  const scopes = Array.isArray(member?.apiScopes) ? member.apiScopes : [];
  return scopes.includes('*') ? SCOPE_OPTIONS.map(item => item.value) : scopes;
}

function memberStatus(member) {
  if (member?.archive?.archivedAt) return { color: 'default', text: '已归档' };
  return member?.active ? { color: 'green', text: '正常' } : { color: 'red', text: '已停用' };
}

function contentProductionCalls(usage) {
  return Object.entries(usage?.callsByFeature || {})
    .filter(([feature]) => CONTENT_FEATURES.has(feature))
    .reduce((sum, [, calls]) => sum + Number(calls || 0), 0);
}

export default function TeamPage() {
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [center, setCenter] = useState(null);
  const [teamResult, setTeamResult] = useState(null);
  const [meta, setMeta] = useState(null);
  const [managerUsername, setManagerUsername] = useState('');
  const [invites, setInvites] = useState([]);
  const [governance, setGovernance] = useState([]);
  const [managerGrantCandidates, setManagerGrantCandidates] = useState([]);
  const [latestInviteUrl, setLatestInviteUrl] = useState('');
  const [memberOpen, setMemberOpen] = useState(false);
  const [memberManageOpen, setMemberManageOpen] = useState(false);
  const [inviteOpen, setInviteOpen] = useState(false);
  const [renameOpen, setRenameOpen] = useState(false);
  const [quotaOpen, setQuotaOpen] = useState(false);
  const [managerGrantOpen, setManagerGrantOpen] = useState(false);
  const [transferOpen, setTransferOpen] = useState(false);
  const [transferPreview, setTransferPreview] = useState(null);
  const [passwordOpen, setPasswordOpen] = useState(false);
  const [selectedMember, setSelectedMember] = useState(null);
  const [memberForm] = Form.useForm();
  const [memberManageForm] = Form.useForm();
  const [inviteForm] = Form.useForm();
  const [renameForm] = Form.useForm();
  const [quotaForm] = Form.useForm();
  const [managerGrantForm] = Form.useForm();
  const [transferForm] = Form.useForm();
  const [passwordForm] = Form.useForm();

  const self = center?.member;
  const availableTeams = useMemo(() => meta?.teams || [], [meta]);
  const activeTeam = useMemo(() => {
    if (self?.role === 'dev') return availableTeams.find(item => item.manager?.username === managerUsername)?.team || null;
    return meta?.team || teamResult?.team || null;
  }, [availableTeams, managerUsername, meta, self?.role, teamResult?.team]);
  const activeManager = self?.role === 'dev'
    ? availableTeams.find(item => item.manager?.username === managerUsername)?.manager || null
    : meta?.manager || self || null;
  const members = useMemo(() => {
    if (self?.role === 'dev' && managerUsername) return (teamResult?.members || []).filter(member => member.role === 'manager' ? member.username === managerUsername : member.boundTo === managerUsername);
    return teamResult?.members || [];
  }, [managerUsername, self?.role, teamResult?.members]);
  const memberCount = members.filter(member => member.role === 'member').length;
  const totalTokens = members.reduce((sum, member) => sum + Number(member.usage?.month?.totalTokens || 0), 0);
  const productionCalls = members.reduce((sum, member) => sum + contentProductionCalls(member.usage?.month), 0);
  const activeGovernance = useMemo(() => governance.find(item => (item.manager?.username || item.managerUsername) === activeManager?.username) || teamResult?.teamGovernance || null, [activeManager?.username, governance, teamResult?.teamGovernance]);
  const teamLimit = activeGovernance?.monthlyTokenLimit ?? null;
  const quotaPercent = teamLimit === null ? 0 : Math.min(100, Math.round(totalTokens / Math.max(teamLimit, 1) * 100));

  async function load({ silent = false } = {}) {
    if (!silent) setLoading(true);
    try {
      const [nextCenter, nextTeam, nextMeta, nextGovernance] = await Promise.all([getMemberCenter(), getTeamMembers(), getTeamMeta(), getTeamGovernance().catch(() => ({ teams: [] }))]);
      const candidateResult = nextCenter.member?.role === 'dev'
        ? await listGlobalAccounts({ role: 'member', active: true })
        : { accounts: [] };
      const nextManager = nextCenter.member?.role === 'dev'
        ? (managerUsername || nextMeta.teams?.[0]?.manager?.username || '')
        : (nextCenter.member?.username || '');
      const inviteResult = nextManager ? await getTeamInvites(nextManager) : { invites: [] };
      setCenter(nextCenter);
      setTeamResult(nextTeam);
      setMeta(nextMeta);
      setGovernance(nextGovernance.teams || []);
      setManagerGrantCandidates(candidateResult.accounts || []);
      setManagerUsername(nextManager);
      setInvites(inviteResult.invites || []);
    } catch (error) { message.error(error.message || '团队管理加载失败'); }
    finally { if (!silent) setLoading(false); }
  }

  useEffect(() => { load(); }, []);

  async function changeManager(value) {
    setManagerUsername(value);
    try {
      const result = await getTeamInvites(value);
      setInvites(result.invites || []);
    } catch (error) { message.error(error.message || '邀请记录读取失败'); }
  }

  async function refresh() { await load({ silent: true }); message.success('团队数据已刷新'); }

  async function copyTeamId() {
    if (!activeTeam?.id) return;
    try { await navigator.clipboard.writeText(activeTeam.id); message.success('团队 ID 已复制'); }
    catch { message.info(`团队 ID：${activeTeam.id}`); }
  }

  async function saveRename(values) {
    if (!activeTeam?.id) return;
    setSaving(true);
    try {
      await renameTeam(activeTeam.id, values.name);
      message.success('团队名称已更新');
      setRenameOpen(false);
      await load({ silent: true });
    } catch (error) { message.error(error.message || '团队名称保存失败'); }
    finally { setSaving(false); }
  }

  async function saveMember(values) {
    setSaving(true);
    try {
      await createTeamMember({
        ...values,
        role: 'member',
        boundTo: activeManager?.username,
        apiEnabled: (values.apiScopes || []).length > 0,
        apiScopes: values.apiScopes || []
      });
      message.success('成员账号已创建');
      memberForm.resetFields();
      setMemberOpen(false);
      await load({ silent: true });
    } catch (error) { message.error(error.message || '成员创建失败'); }
    finally { setSaving(false); }
  }

  async function saveInvite(values) {
    if (!activeManager?.username) return;
    setSaving(true);
    try {
      const created = await createTeamInvite({ managerUsername: activeManager.username, ...values, apiScopes: values.apiScopes || [] });
      const inviteUrl = new URL(created.inviteUrl, window.location.origin).toString();
      setLatestInviteUrl(inviteUrl);
      try { await navigator.clipboard.writeText(inviteUrl); message.success('邀请已生成，链接已复制'); }
      catch { message.success('邀请已生成，请在邀请管理中复制链接'); }
      inviteForm.resetFields();
      setInviteOpen(false);
      setInvites(current => [{ ...created.invite, inviteUrl: created.inviteUrl }, ...current]);
    } catch (error) { message.error(error.message || '邀请生成失败'); }
    finally { setSaving(false); }
  }

  async function toggleMember(member) {
    setSaving(true);
    try {
      await setTeamMemberStatus(member.username, !member.active);
      message.success(member.active ? '成员账号已停用' : '成员账号已恢复');
      await load({ silent: true });
    } catch (error) { message.error(error.message || '成员状态更新失败'); }
    finally { setSaving(false); }
  }

  function openMemberManage(member) {
    setSelectedMember(member);
    memberManageForm.setFieldsValue({ displayName: member.displayName, monthlyTokenLimit: member.monthlyTokenLimit, apiScopes: scopesOf(member) });
    setMemberManageOpen(true);
  }

  async function saveMemberManage(values) {
    if (!selectedMember) return;
    setSaving(true);
    try {
      await updateTeamMember(selectedMember.username, { displayName: values.displayName, monthlyTokenLimit: values.monthlyTokenLimit ?? null });
      if (selectedMember.active && !selectedMember.archive?.archivedAt) await setTeamMemberApiScopes(selectedMember.username, values.apiScopes || []);
      message.success('成员资料、额度与 API 权限已更新');
      setMemberManageOpen(false);
      await load({ silent: true });
    } catch (error) { message.error(error.message || '成员保存失败'); }
    finally { setSaving(false); }
  }

  async function archiveOrRestoreMember() {
    if (!selectedMember) return;
    setSaving(true);
    try {
      if (selectedMember.archive?.archivedAt) {
        await restoreArchivedMember(selectedMember.username);
        message.success('成员已恢复，请重新确认 API 权限');
      } else {
        await archiveTeamMember(selectedMember.username, '团队管理归档');
        message.success('成员已归档，历史用量和审计记录保留');
      }
      setMemberManageOpen(false);
      await load({ silent: true });
    } catch (error) { message.error(error.message || '成员状态更新失败'); }
    finally { setSaving(false); }
  }

  async function savePassword(values) {
    if (!selectedMember) return;
    setSaving(true);
    try {
      await resetTeamMemberPassword(selectedMember.username, values.password);
      message.success('密码已重置，成员其他会话已撤销');
      passwordForm.resetFields();
      setPasswordOpen(false);
    } catch (error) { message.error(error.message || '密码重置失败'); }
    finally { setSaving(false); }
  }

  async function saveQuota(values) {
    if (!activeManager?.username) return;
    setSaving(true);
    try {
      await updateTeamGovernance(activeManager.username, values.monthlyTokenLimit ?? null);
      message.success('团队月度总额度已更新');
      setQuotaOpen(false);
      await load({ silent: true });
    } catch (error) { message.error(error.message || '团队额度保存失败'); }
    finally { setSaving(false); }
  }

  async function grantManager(values) {
    if (self?.role !== 'dev' || !values.username) return;
    setSaving(true);
    try {
      await updateTeamMember(values.username, { role: 'manager', boundTo: null });
      message.success(`@${values.username} 已获得 MANAGER 管理权限`);
      managerGrantForm.resetFields();
      setManagerGrantOpen(false);
      await load({ silent: true });
    } catch (error) { message.error(error.message || '管理者授权失败'); }
    finally { setSaving(false); }
  }

  const teamNameForManager = manager => availableTeams.find(item => item.manager?.username === manager)?.team?.name || (manager ? `@${manager} 的团队` : '未加入团队');

  function previewTransfer(values) {
    const member = (teamResult?.members || []).find(item => item.username === values.username);
    if (!member) return message.error('未找到该成员账号');
    if (member.boundTo === values.boundTo) return message.info('该成员已在所选团队中');
    setTransferPreview({ member, boundTo: values.boundTo });
  }

  async function confirmTransfer() {
    if (!transferPreview) return;
    setSaving(true);
    try {
      await updateTeamMember(transferPreview.member.username, { boundTo: transferPreview.boundTo });
      message.success(`@${transferPreview.member.username} 已转入「${teamNameForManager(transferPreview.boundTo)}」，请由新团队重新授权 API。`);
      transferForm.resetFields();
      setTransferPreview(null);
      setTransferOpen(false);
      await load({ silent: true });
    } catch (error) { message.error(error.message || '成员转移失败'); }
    finally { setSaving(false); }
  }

  if (loading) return <div className="account-center-page"><Skeleton active paragraph={{ rows: 11 }} /></div>;
  if (!self || !['dev', 'manager'].includes(self.role)) return <div className="account-center-page"><div className="ac-empty">当前身份没有团队管理权限</div></div>;

  return <div className="account-center-page team-page">
    <PageHeader title="组员管理" subtitle="普通组员的 API 授权、额度、用量与内容制作调用管理" actions={<Button icon={<RefreshCw size={15} />} loading={saving} onClick={refresh}>刷新</Button>} />

    {self.role === 'dev' ? <div className="ac-team-admin-toolbar">
      {availableTeams.length > 1 ? <div className="ac-team-selector"><span>管理团队</span><Select value={managerUsername} onChange={changeManager} options={availableTeams.map(item => ({ value: item.manager.username, label: `${item.team.name} · @${item.manager.username}` }))} /></div> : <span className="ac-team-toolbar-note">DEV 可管理所有团队</span>}
      <Button onClick={() => { transferForm.resetFields(); setTransferOpen(true); }}>调整已有账号</Button>
      <Button type="primary" icon={<ShieldCheck size={15} />} onClick={() => { managerGrantForm.resetFields(); setManagerGrantOpen(true); }}>授权管理者</Button>
    </div> : null}

    {!activeTeam ? <Panel title="尚未选择团队"><div className="ac-empty">当前没有可管理团队。</div></Panel> : <>
      <div className="ac-team-reference-top">
        <Panel title="团队概览" className="ac-team-overview" action={<Button type="text" size="small" onClick={() => { quotaForm.setFieldsValue({ monthlyTokenLimit: teamLimit }); setQuotaOpen(true); }}>管理额度</Button>}>
          <div className="ac-team-overview-grid">
            <div className="ac-team-mark"><UsersRound size={25} /></div>
            <div><small>团队名称</small><strong>{activeTeam.name}</strong><button type="button" className="ac-inline-icon-button" onClick={() => { renameForm.setFieldsValue({ name: activeTeam.name }); setRenameOpen(true); }} aria-label="重命名团队"><Pencil size={14} /></button></div>
            <div className="ac-team-member-total"><small>成员数量</small><strong>{memberCount}<i> / {teamLimit === null ? '—' : formatTokens(teamLimit)}</i></strong></div>
          </div>
          <div className="ac-team-id-line"><span>Team ID</span><strong>{activeTeam.id}</strong><button type="button" onClick={copyTeamId}><Copy size={13} /></button></div>
          <div className="ac-team-owner-line"><span>主要负责人</span><MemberIdentity member={activeManager} size={28} /></div>
          <div className="ac-team-quota-line"><div><span>本月团队用量</span><strong>{formatTokens(totalTokens)} <i>/ {teamLimit === null ? '不限额' : `${formatTokens(teamLimit)} Tokens`}</i></strong></div><b>{teamLimit === null ? '—' : `${quotaPercent}%`}</b></div>
          <Progress percent={quotaPercent} showInfo={false} status={quotaPercent >= 100 ? 'exception' : 'active'} />
          <div className="ac-team-production-note"><div><span>本月内容制作调用</span><strong>{productionCalls.toLocaleString('zh-CN')} 次</strong></div><small>按已计费模型调用汇总，不等同于去重书籍数。</small></div>
        </Panel>
        <Panel title="快捷操作" className="ac-team-quick-panel">
          <div className="ac-team-quick-actions"><Button icon={<UserPlus size={16} />} onClick={() => { memberForm.resetFields(); setMemberOpen(true); }}>创建成员</Button><Button icon={<Link2 size={16} />} onClick={() => { inviteForm.setFieldsValue({ expiresInHours: 72, apiScopes: ['text'] }); setInviteOpen(true); }}>生成邀请</Button><Button icon={<Pencil size={16} />} onClick={() => { renameForm.setFieldsValue({ name: activeTeam.name }); setRenameOpen(true); }}>重命名团队</Button></div>
          <div className="ac-invite-default"><div><strong>邀请策略</strong><small>每张邀请在生成时单独配置权限、额度和有效期。</small></div><span><b>可分配 API</b><Tag>文本</Tag><Tag>生图</Tag><Tag>TTS</Tag></span><span><b>团队月度额度</b><em>{teamLimit === null ? '不限额' : `${formatTokens(teamLimit)} Tokens`}</em></span>{latestInviteUrl ? <Button size="small" type="link" icon={<Copy size={13} />} onClick={() => navigator.clipboard.writeText(latestInviteUrl).then(() => message.success('邀请链接已复制')).catch(() => message.info(latestInviteUrl))}>复制最新邀请链接</Button> : null}</div>
        </Panel>
      </div>

      <div className="ac-team-reference-bottom">
        <Panel title="成员列表" className="ac-team-members-panel" action={<span className="ac-panel-count">共 {memberCount} 位成员</span>}>
          <div className="ac-reference-table"><div className="ac-reference-table-head"><span>成员</span><span>显示名称</span><span>角色</span><span>账号状态</span><span>API 范围</span><span>月度额度</span><span>已用 Tokens</span><span>操作</span></div>
            {members.filter(member => member.role !== 'manager').map(member => { const status = memberStatus(member); const scopes = scopesOf(member); return <div className="ac-reference-table-row" key={member.username}><MemberIdentity member={member} size={30} showUsername /><strong>{member.displayName}</strong><RoleBadge role={member.role} compact /><Tag color={status.color}>{status.text}</Tag><span>{scopes.length ? scopes.map(scope => <Tag key={scope}>{SCOPE_OPTIONS.find(item => item.value === scope)?.label || scope}</Tag>) : '未授权'}</span><span>{member.monthlyTokenLimit === null ? '不限额' : `${formatTokens(member.monthlyTokenLimit)} Tokens`}</span><strong>{formatTokens(member.usage?.month?.totalTokens)}</strong><Tooltip title="管理成员"><Button type="text" icon={<MoreHorizontal size={18} />} loading={saving} onClick={() => openMemberManage(member)} /></Tooltip></div>; })}
            {!memberCount ? <div className="ac-empty">尚无成员，先创建成员或生成邀请。</div> : null}
          </div>
        </Panel>
        <Panel title="邀请管理" className="ac-team-invites-panel" action={<Button type="text" size="small" icon={<RefreshCw size={14} />} onClick={() => getTeamInvites(activeManager?.username).then(result => setInvites(result.invites || [])).catch(error => message.error(error.message || '邀请读取失败'))}>刷新</Button>}>
          <div className="ac-invite-cards">{invites.slice(0, 5).map(invite => <div className="ac-invite-card" key={invite.id}><div><strong>{invite.usedBy ? `已加入 · @${invite.usedBy}` : '成员邀请'}</strong><Tag color={invite.available ? 'green' : invite.expired ? 'red' : 'blue'}>{invite.available ? '可用' : invite.expired ? '已过期' : '已使用'}</Tag></div><small>有效期至 {formatDate(invite.expiresAt)}</small><small>已使用/上限 · {invite.usedBy ? '1 / 1' : '0 / 1'}</small><span>{(invite.apiScopes || []).length ? invite.apiScopes.map(scope => <Tag key={scope}>{SCOPE_OPTIONS.find(item => item.value === scope)?.label || scope}</Tag>) : <Tag>无 API</Tag>}<em>{invite.monthlyTokenLimit === null ? '不限额' : `${formatTokens(invite.monthlyTokenLimit)} Tokens`}</em></span></div>)}{!invites.length ? <div className="ac-empty">还没有创建邀请</div> : null}</div>
          <Button block icon={<Plus size={16} />} onClick={() => { inviteForm.setFieldsValue({ expiresInHours: 72, apiScopes: ['text'] }); setInviteOpen(true); }}>生成新邀请</Button>
        </Panel>
      </div>
    </>}

    <Modal title="创建成员" open={memberOpen} onCancel={() => setMemberOpen(false)} onOk={() => memberForm.submit()} okText="创建成员" confirmLoading={saving}>
      <Form form={memberForm} layout="vertical" onFinish={saveMember}><Form.Item name="username" label="登录账号" rules={[{ required: true, pattern: /^[A-Za-z0-9_-]{3,32}$/, message: '3-32 位字母、数字、下划线或连字符' }]}><Input /></Form.Item><Form.Item name="displayName" label="显示名称" rules={[{ required: true }]}><Input /></Form.Item><Form.Item name="password" label="初始密码" rules={[{ required: true, min: 8 }]}><Input.Password autoComplete="new-password" /></Form.Item><Form.Item name="monthlyTokenLimit" label="月度额度"><InputNumber min={0} style={{ width: '100%' }} addonAfter="Tokens" placeholder="不限额" /></Form.Item><Form.Item name="apiScopes" label="AI 能力权限"><Checkbox.Group options={SCOPE_OPTIONS} /></Form.Item></Form>
    </Modal>
    <Modal title="生成成员邀请" open={inviteOpen} onCancel={() => setInviteOpen(false)} onOk={() => inviteForm.submit()} okText="生成邀请" confirmLoading={saving}>
      <Form form={inviteForm} layout="vertical" onFinish={saveInvite}><Form.Item name="expiresInHours" label="有效期" rules={[{ required: true }]}><InputNumber min={1} max={720} style={{ width: '100%' }} addonAfter="小时" /></Form.Item><Form.Item name="monthlyTokenLimit" label="成员月度额度"><InputNumber min={0} style={{ width: '100%' }} addonAfter="Tokens" placeholder="不限额" /></Form.Item><Form.Item name="apiScopes" label="AI 能力权限"><Checkbox.Group options={SCOPE_OPTIONS} /></Form.Item></Form>
    </Modal>
    <Modal title={selectedMember ? `管理 ${selectedMember.displayName}` : '管理成员'} open={memberManageOpen} onCancel={() => setMemberManageOpen(false)} onOk={() => memberManageForm.submit()} okText="保存成员" confirmLoading={saving}>
      <Form form={memberManageForm} layout="vertical" onFinish={saveMemberManage}><Form.Item name="displayName" label="显示名称" rules={[{ required: true }]}><Input /></Form.Item><Form.Item name="monthlyTokenLimit" label="月度额度"><InputNumber min={0} style={{ width: '100%' }} addonAfter="Tokens" placeholder="不限额" /></Form.Item><Form.Item name="apiScopes" label="AI 能力权限"><Checkbox.Group options={SCOPE_OPTIONS} disabled={!selectedMember?.active || Boolean(selectedMember?.archive?.archivedAt)} /></Form.Item></Form>
      <div className="ac-member-actions"><Button danger={selectedMember?.active} loading={saving} onClick={() => selectedMember && toggleMember(selectedMember)}>{selectedMember?.active ? '停用账号' : '恢复账号'}</Button><Button loading={saving} onClick={() => { passwordForm.resetFields(); setPasswordOpen(true); }}>重置密码</Button><Button danger={!selectedMember?.archive?.archivedAt} loading={saving} onClick={archiveOrRestoreMember}>{selectedMember?.archive?.archivedAt ? '从归档恢复' : '归档成员'}</Button></div>
    </Modal>
    <Modal title={selectedMember ? `重置 ${selectedMember.displayName} 的密码` : '重置密码'} open={passwordOpen} onCancel={() => setPasswordOpen(false)} onOk={() => passwordForm.submit()} okText="确认重置" confirmLoading={saving}><Form form={passwordForm} layout="vertical" onFinish={savePassword}><Form.Item name="password" label="新密码" rules={[{ required: true, min: 8 }]}><Input.Password /></Form.Item><Form.Item name="confirm" label="确认新密码" dependencies={['password']} rules={[{ required: true }, ({ getFieldValue }) => ({ validator(_, value) { return value === getFieldValue('password') ? Promise.resolve() : Promise.reject(new Error('两次密码不一致')); } })]}><Input.Password /></Form.Item></Form></Modal>
    <Modal title="团队月度总额度" open={quotaOpen} onCancel={() => setQuotaOpen(false)} onOk={() => quotaForm.submit()} okText="保存额度" confirmLoading={saving}><Form form={quotaForm} layout="vertical" onFinish={saveQuota}><Form.Item name="monthlyTokenLimit" label="自然月总额度"><InputNumber min={0} style={{ width: '100%' }} addonAfter="Tokens" placeholder="不限额" /></Form.Item><p className="ac-muted-copy">额度为 0 时团队调用会被立即拦截；留空表示不限额。</p></Form></Modal>
    <Modal title="授权管理者" open={managerGrantOpen} onCancel={() => setManagerGrantOpen(false)} onOk={() => managerGrantForm.submit()} okText="确认授权" confirmLoading={saving}>
      <p className="ac-muted-copy">开发者可以从全站已启用的 MEMBER 账号中选择目标，提升为 MANAGER 后，他可以创建和管理自己的团队、成员与 API 权限。新账号加入系统后，刷新本页即可出现在列表中。只有 DEV 可以执行此操作。</p>
      <Form form={managerGrantForm} layout="vertical" onFinish={grantManager}>
        <Form.Item name="username" label="目标账号" rules={[{ required: true, message: '请选择要授权的账号' }]}>
          <Select showSearch optionFilterProp="label" placeholder="搜索并选择任意已启用 MEMBER 账号" options={managerGrantCandidates.map(member => ({ value: member.username, label: `${member.displayName} · @${member.username}` }))} />
        </Form.Item>
      </Form>
    </Modal>
    <Modal title="调整已有账号的团队" open={transferOpen} onCancel={() => setTransferOpen(false)} onOk={() => transferForm.submit()} okText="下一步" confirmLoading={saving}>
      <p className="ac-muted-copy">仅 DEV 可调整 MEMBER 的所属团队。转移会立即撤销原团队 API 授权，新团队需要在成员管理中重新授权。</p>
      <Form form={transferForm} layout="vertical" onFinish={previewTransfer}>
        <Form.Item name="username" label="现有成员账号" rules={[{ required: true, message: '请选择要调整的成员' }]}>
          <Select showSearch optionFilterProp="label" placeholder="选择 MEMBER 账号" options={(teamResult?.members || []).filter(member => member.role === 'member' && member.active && !member.archive?.archivedAt).map(member => ({ value: member.username, label: `${member.displayName} · @${member.username} · ${teamNameForManager(member.boundTo)}` }))} />
        </Form.Item>
        <Form.Item name="boundTo" label="转入团队" rules={[{ required: true, message: '请选择目标团队' }]}>
          <Select showSearch optionFilterProp="label" placeholder="选择目标团队" options={availableTeams.map(item => ({ value: item.manager.username, label: `${item.team.name} · @${item.manager.username}` }))} />
        </Form.Item>
      </Form>
    </Modal>
    <Modal title="确认转移团队" open={Boolean(transferPreview)} onCancel={() => setTransferPreview(null)} onOk={confirmTransfer} okText="确认转移" confirmLoading={saving} okButtonProps={{ danger: true }}>
      {transferPreview ? <div className="ac-session-control"><span><UsersRound size={22} /></span><div><h3>{transferPreview.member.displayName} · @{transferPreview.member.username}</h3><p>所属团队将从「{teamNameForManager(transferPreview.member.boundTo)}」调整为「{teamNameForManager(transferPreview.boundTo)}」。原团队的 API 授权会被撤销，新团队需重新授权。</p></div></div> : null}
    </Modal>
    <Modal title="重命名团队" open={renameOpen} onCancel={() => setRenameOpen(false)} onOk={() => renameForm.submit()} okText="保存" confirmLoading={saving}><Form form={renameForm} layout="vertical" onFinish={saveRename}><Form.Item name="name" label="团队名称" rules={[{ required: true, max: 60 }]}><Input autoFocus /></Form.Item></Form></Modal>
  </div>;
}
