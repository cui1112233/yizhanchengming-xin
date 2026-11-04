import { Avatar, Button, Form, Input, Modal, Skeleton, Tag, message } from 'antd';
import { Camera, CheckCircle2, CircleGauge, Crown, Mail, MailCheck, Pencil, Phone, Save, ShieldCheck, Sparkles, UsersRound } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { getMemberCenter, updateMemberProfile, uploadMemberAvatar } from '../../shared/api/member';
import { getEmailVerificationStatus, requestEmailVerification } from '../../shared/api/accountRecovery';
import { PageHeader, Panel, RoleBadge, avatarFallback, formatDate } from './accountCenterShared';
import { Link } from '../../shared/components/Link';

// 上传接口返回成功只代表文件已落盘；还要确认浏览器真的能取到这张图。
// 取不到时 Antd Avatar 会静默回退到首字母占位，所以这里先回读一次。
function verifyAvatarImage(url) {
  return new Promise((resolve, reject) => {
    if (!url) {
      reject(new Error('头像已保存，但服务端未返回头像地址，请刷新页面重试'));
      return;
    }
    const probe = new Image();
    const timer = window.setTimeout(() => {
      probe.onload = null;
      probe.onerror = null;
      reject(new Error('头像已保存，但图片加载超时，请刷新页面重试'));
    }, 15000);
    probe.onload = () => {
      window.clearTimeout(timer);
      resolve();
    };
    probe.onerror = () => {
      window.clearTimeout(timer);
      reject(new Error('头像已保存，但图片加载失败，请刷新页面重试'));
    };
    probe.src = url;
  });
}

export default function ProfilePage() {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [mailing, setMailing] = useState(false);
  const [editing, setEditing] = useState(false);
  const [center, setCenter] = useState(null);
  const [emailStatus, setEmailStatus] = useState(null);
  const fileRef = useRef(null);

  async function load() {
    setLoading(true);
    try {
      const [result, status] = await Promise.all([
        getMemberCenter(),
        getEmailVerificationStatus().catch(() => null)
      ]);
      setCenter(result);
      setEmailStatus(status);
      form.setFieldsValue({
        displayName: result.member?.displayName,
        bio: result.member?.bio || '',
        phone: result.member?.phone || '',
        email: result.member?.email || '',
        teamTitle: result.member?.teamTitle || ''
      });
    } catch (error) {
      message.error(error.message || '个人资料加载失败');
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => { load(); }, []);

  async function save(values) {
    setSaving(true);
    try {
      const previousEmail = center?.member?.email || '';
      const result = await updateMemberProfile(values);
      setCenter(current => ({ ...current, member: result.member }));
      setEditing(false);
      window.dispatchEvent(new CustomEvent('qiantie:profile-updated'));
      if ((result.member?.email || '') !== previousEmail) {
        setEmailStatus(await getEmailVerificationStatus().catch(() => null));
        message.success('个人资料已保存；邮箱发生变化，需要重新验证');
      } else {
        message.success('个人资料已保存');
      }
    } catch (error) {
      message.error(error.message || '保存失败');
    } finally {
      setSaving(false);
    }
  }

  async function sendVerification() {
    setMailing(true);
    try {
      const result = await requestEmailVerification();
      message.success(`验证邮件已发送到 ${result.email}`);
    } catch (error) {
      message.error(error.message || '验证邮件发送失败');
    } finally { setMailing(false); }
  }

  async function upload(event) {
    const file = event.target.files?.[0];
    event.target.value = '';
    if (!file) return;
    if (!['image/png', 'image/jpeg', 'image/webp'].includes(file.type)) return message.error('请选择 PNG、JPG 或 WebP 图片');
    if (file.size > 2 * 1024 * 1024) return message.error('头像不能超过 2MB');
    const reader = new FileReader();
    reader.onload = async () => {
      setSaving(true);
      try {
        const result = await uploadMemberAvatar(reader.result);
        const avatarUrl = result?.member?.avatarUrl;
        await verifyAvatarImage(avatarUrl);
        setCenter(current => ({ ...current, member: result.member }));
        window.dispatchEvent(new CustomEvent('qiantie:profile-updated'));
        message.success('头像已更新');
      } catch (error) {
        message.error(error.message || '头像上传失败，请重试');
      } finally {
        setSaving(false);
      }
    };
    reader.readAsDataURL(file);
  }

  const member = center?.member;
  const completion = useMemo(() => {
    if (!member) return 0;
    const fields = [member.avatarUrl, member.displayName, member.bio, member.phone, member.email, member.teamTitle];
    return Math.round((fields.filter(Boolean).length / fields.length) * 100);
  }, [member]);

  if (loading) return <div className="account-center-page"><Skeleton active paragraph={{ rows: 9 }} /></div>;
  if (!member) return <div className="account-center-page"><div className="ac-empty">无法读取资料</div></div>;

  const monthTokens = Number(center?.usage?.month?.totalTokens || 0);
  const monthlyLimit = member.monthlyTokenLimit;
  const remainingTokens = monthlyLimit === null || monthlyLimit === undefined ? '不限额' : Math.max(0, Number(monthlyLimit) - monthTokens).toLocaleString('zh-CN');
  const securityScore = Math.min(100, 65 + (emailStatus?.verified ? 15 : 0) + (member.mfaEnabled ? 20 : 0));
  const isGovernanceEnabled = ['dev', 'manager'].includes(member.role);

  return <div className="account-center-page profile-page">
    <PageHeader title="个人资料" subtitle="管理您的个人信息、会员权益、安全状态与治理权限" />
    <div className="ac-profile-summary-grid">
      <Panel title="我的信息" className="ac-profile-information">
        <div className="ac-profile-information-copy"><div><span>显示名称</span><strong>{member.displayName}</strong><Button type="text" size="small" icon={<Pencil size={14} />} onClick={() => setEditing(true)} /></div><div><span>账号</span><strong>@{member.username}</strong></div><div><span>角色</span><RoleBadge role={member.role} compact /></div><div><span>邮箱</span><strong>{member.email || '未填写'}</strong></div><div><span>加入时间</span><strong>{formatDate(member.createdAt).slice(0, 10)}</strong></div><div><span>所在团队</span><strong>{center?.team?.name || 'qiantie 核心团队'}</strong></div></div>
        <button className="ac-profile-summary-avatar" type="button" onClick={() => fileRef.current?.click()} disabled={saving} aria-label="更换头像"><Avatar size={118} src={member.avatarUrl}>{avatarFallback(member)}</Avatar><i><Camera size={15} /></i></button>
      </Panel>
      <Panel title="账户概览" className="ac-account-overview"><div className="ac-account-overview-grid"><div><span>本月 Token 使用</span><strong>{monthTokens.toLocaleString('zh-CN')}</strong><small>{monthlyLimit === null || monthlyLimit === undefined ? '当前没有设置上限' : `额度上限 ${Number(monthlyLimit).toLocaleString('zh-CN')}`}</small></div><div><span>可用剩余额度</span><strong>{remainingTokens}</strong><small>{monthlyLimit === null || monthlyLimit === undefined ? 'Tokens' : 'Tokens'}</small></div><div><span>资料完整度</span><strong>{completion}%</strong><small>可在编辑资料中补充</small></div><div><span>安全评分</span><strong>{securityScore}</strong><small>{emailStatus?.verified ? '邮箱已验证' : '建议验证邮箱'}</small></div></div></Panel>
    </div>
    <div className="ac-profile-cards-grid">
      <Panel title="会员权益" className="ac-profile-summary-card"><div className="ac-profile-card-title"><Crown size={23} /><div><strong>{member.role === 'dev' ? 'DEV 会员' : member.role === 'manager' ? 'MANAGER 会员' : 'MEMBER 会员'}</strong><small>{monthlyLimit === null || monthlyLimit === undefined ? 'Token 不限额' : `每月 ${Number(monthlyLimit).toLocaleString('zh-CN')} Tokens`}</small></div></div><div className="ac-profile-checks"><span><CheckCircle2 size={15} />账号资料与团队名片</span><span><CheckCircle2 size={15} />{member.role === 'member' ? '已授权 AI 服务范围' : '团队 AI 服务管理'}</span><span><CheckCircle2 size={15} />会员中心与用量统计</span></div><Link href="/member" className="ac-profile-card-link">查看会员中心 <span>›</span></Link></Panel>
      <Panel title="安全状态" className="ac-profile-summary-card"><div className="ac-profile-security-score"><span><ShieldCheck size={25} /></span><div><small>安全评分</small><strong>{securityScore}<i>/100</i></strong><b>{securityScore >= 90 ? '优秀' : '建议完善'}</b></div></div><div className="ac-profile-checks"><span><CheckCircle2 size={15} />登录保护</span><span><CheckCircle2 size={15} />邮箱验证 {emailStatus?.verified ? '已验证' : '待验证'}</span><span><CheckCircle2 size={15} />MFA {member.mfaEnabled ? '已启用' : '未启用'}</span></div>{member.email && !emailStatus?.verified ? <Button type="link" loading={mailing} disabled={!emailStatus?.deliveryConfigured} onClick={sendVerification}>发送邮箱验证</Button> : <Link href="/security" className="ac-profile-card-link">查看安全详情 <span>›</span></Link>}</Panel>
      <Panel title="联合治理权限" className="ac-profile-summary-card"><div className="ac-profile-governance"><UsersRound size={26} /><div><strong>{isGovernanceEnabled ? '已开启联合治理' : '当前为成员权限'}</strong><small>{isGovernanceEnabled ? '可以进入团队与成员治理工作区' : '由所属团队管理员统一治理'}</small></div></div><div className="ac-profile-checks"><span><CheckCircle2 size={15} />{isGovernanceEnabled ? '团队治理' : '团队协作'}</span><span><CheckCircle2 size={15} />{isGovernanceEnabled ? '成员授权管理' : '个人资料管理'}</span><span><CheckCircle2 size={15} />安全审计</span></div>{isGovernanceEnabled ? <Link href="/advanced-team-admin" className="ac-profile-card-link">前往联合治理中心 <span>›</span></Link> : <Link href="/member" className="ac-profile-card-link">查看团队身份与权益 <span>›</span></Link>}</Panel>
    </div>
    <input ref={fileRef} className="member-hidden-file" type="file" accept="image/png,image/jpeg,image/webp" onChange={upload} />
    <Modal title="编辑个人资料" open={editing} onCancel={() => setEditing(false)} onOk={() => form.submit()} okText="保存修改" confirmLoading={saving}>
      <Form form={form} layout="vertical" onFinish={save}><Form.Item label="显示名称" name="displayName" rules={[{ required: true, whitespace: true, max: 40 }]}><Input maxLength={40} /></Form.Item><Form.Item label="手机号" name="phone"><Input prefix={<Phone size={15} />} maxLength={32} /></Form.Item><Form.Item label="邮箱" name="email"><Input prefix={<Mail size={15} />} maxLength={120} /></Form.Item><Form.Item label="团队职务" name="teamTitle"><Input maxLength={60} /></Form.Item><Form.Item label="个人简介" name="bio"><Input.TextArea rows={4} maxLength={240} showCount /></Form.Item></Form>
    </Modal>
  </div>;
}
