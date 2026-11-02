import { Alert, Button, Drawer, Form, Input, Modal, Popconfirm, Select, Space, Table, Tag, Typography, message } from 'antd';
import { useEffect, useMemo, useState } from 'react';
import {
  createAdminAccount,
  createAdminGrant,
  listAdminAccounts,
  listAdminApplications,
  listAdminGrants,
  resetAdminAccountPassword,
  reviewAdminApplication,
  revokeAdminGrant,
  setAdminAccountStatus
} from '../../shared/api/admin';
import { getCurrentAccount } from '../../shared/api/auth';

const capabilityOptions = [
  { value: 'account:review', label: '账号审核' },
  { value: 'preset:draft', label: '预设词草稿' },
  { value: 'preset:publish', label: '预设词发布' }
];

const scopeOptions = [
  { value: '*', label: '全部模块' },
  { value: 'script', label: '剧本生成' }
];

function errorMessage(error) {
  try {
    return JSON.parse(error.message).error || error.message;
  } catch (_) {
    return error.message || '操作失败';
  }
}

function formatDate(value) {
  return value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '-';
}

function capabilityLabel(capability) {
  return capabilityOptions.find(item => item.value === capability)?.label || capability;
}

function scopeLabel(scope) {
  return scopeOptions.find(item => item.value === scope)?.label || scope;
}

export function AccountGovernancePage() {
  const [accounts, setAccounts] = useState([]);
  const [applications, setApplications] = useState([]);
  const [grants, setGrants] = useState([]);
  const [session, setSession] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [creating, setCreating] = useState(false);
  const [accountDrawerOpen, setAccountDrawerOpen] = useState(false);
  const [grantDrawerOpen, setGrantDrawerOpen] = useState(false);
  const [passwordAccount, setPasswordAccount] = useState(null);
  const [submitting, setSubmitting] = useState(false);
  const [accountForm] = Form.useForm();
  const [grantForm] = Form.useForm();
  const [passwordForm] = Form.useForm();

  const isOwner = session?.isOwner === true;
  const pendingApplications = applications.filter(item => item.status === 'pending');
  const grantsBySubject = useMemo(() => grants.reduce((map, grant) => {
    map[grant.subject] = [...(map[grant.subject] || []), grant];
    return map;
  }, {}), [grants]);

  async function load() {
    setLoading(true);
    setError('');
    try {
      const current = await getCurrentAccount();
      setSession(current);
      const [accountResult, applicationResult, grantResult] = await Promise.all([
        listAdminAccounts(),
        listAdminApplications(),
        current.isOwner ? listAdminGrants() : Promise.resolve({ grants: [] })
      ]);
      setAccounts(accountResult.accounts || []);
      setApplications(applicationResult.applications || []);
      setGrants(grantResult.grants || []);
    } catch (requestError) {
      setError(errorMessage(requestError));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => { load(); }, []);

  async function createAccount(values) {
    setSubmitting(true);
    try {
      await createAdminAccount(values);
      message.success('账号已创建。');
      setAccountDrawerOpen(false);
      accountForm.resetFields();
      await load();
    } catch (requestError) {
      message.error(errorMessage(requestError));
    } finally {
      setSubmitting(false);
    }
  }

  async function reviewApplication(application, approved) {
    try {
      await reviewAdminApplication(application.id, approved);
      message.success(approved ? '已通过申请并创建账号。' : '已拒绝申请。');
      await load();
    } catch (requestError) {
      message.error(errorMessage(requestError));
    }
  }

  async function toggleAccount(account) {
    try {
      await setAdminAccountStatus(account.username, !account.active);
      message.success(account.active ? '账号已停用。' : '账号已启用。');
      await load();
    } catch (requestError) {
      message.error(errorMessage(requestError));
    }
  }

  async function resetPassword(values) {
    setSubmitting(true);
    try {
      await resetAdminAccountPassword(passwordAccount.username, values.password);
      message.success('密码已重置。');
      setPasswordAccount(null);
      passwordForm.resetFields();
    } catch (requestError) {
      message.error(errorMessage(requestError));
    } finally {
      setSubmitting(false);
    }
  }

  async function createGrant(values) {
    setSubmitting(true);
    try {
      await createAdminGrant(values);
      message.success('权限已授予。');
      setGrantDrawerOpen(false);
      grantForm.resetFields();
      await load();
    } catch (requestError) {
      message.error(errorMessage(requestError));
    } finally {
      setSubmitting(false);
    }
  }

  async function revokeGrant(id) {
    try {
      await revokeAdminGrant(id);
      message.success('权限已撤销。');
      await load();
    } catch (requestError) {
      message.error(errorMessage(requestError));
    }
  }

  const columns = [
    { title: '账号', dataIndex: 'username', width: 180, render: (username, account) => <div><strong>{username}</strong>{account.isOwner ? <Tag color="blue">主管理员</Tag> : null}</div> },
    { title: '状态', dataIndex: 'active', width: 92, render: active => <Tag color={active ? 'green' : 'default'}>{active ? '正常' : '已停用'}</Tag> },
    {
      title: '后台权限', key: 'permissions',
      render: (_, account) => account.isOwner ? '系统所有权限' : (grantsBySubject[account.username] || []).map(grant => `${capabilityLabel(grant.capability)} · ${scopeLabel(grant.scope)}`).join('；') || '无'
    },
    { title: '最近变更', dataIndex: 'updatedAt', width: 176, render: formatDate },
    {
      title: '操作', key: 'actions', width: 220,
      render: (_, account) => <Space size="small" wrap>
        {!account.isOwner ? <Button size="small" onClick={() => toggleAccount(account)}>{account.active ? '停用' : '启用'}</Button> : null}
        <Button size="small" onClick={() => setPasswordAccount(account)}>重置密码</Button>
      </Space>
    }
  ];

  return (
    <section className="admin-governance-page">
      <div className="admin-page-heading">
        <div>
          <Typography.Title level={2}>账号与授权</Typography.Title>
          <Typography.Paragraph>创建和维护成员账号；授权调整由服务端立即生效。</Typography.Paragraph>
        </div>
        {isOwner ? <Button type="primary" onClick={() => setAccountDrawerOpen(true)}>新建账号</Button> : null}
      </div>
      {error ? <Alert type="error" showIcon message="无法读取账号管理数据" description={error} /> : null}
      <div className="admin-stat-grid">
        <div className="admin-stat"><span>活跃账号</span><strong>{accounts.filter(account => account.active).length}</strong></div>
        <div className="admin-stat"><span>待审核申请</span><strong>{pendingApplications.length}</strong></div>
        <div className="admin-stat"><span>已授权成员</span><strong>{new Set(grants.map(grant => grant.subject)).size}</strong></div>
      </div>
      <div className="admin-governance-grid">
        <section className="admin-surface admin-account-table">
          <div className="admin-surface-heading"><strong>账号列表</strong>{isOwner ? <Button size="small" onClick={() => setGrantDrawerOpen(true)}>授予权限</Button> : null}</div>
          <Table rowKey="username" columns={columns} dataSource={accounts} loading={loading} pagination={false} size="middle" scroll={{ x: 860 }} />
        </section>
        <aside className="admin-side-stack">
          <section className="admin-surface">
            <div className="admin-surface-heading"><strong>待审核申请</strong><span>{pendingApplications.length} 项</span></div>
            {pendingApplications.length ? pendingApplications.map(application => <div className="admin-application" key={application.id}>
              <strong>{application.username}</strong><p>{application.reason}</p><small>{formatDate(application.createdAt)}</small>
              <Space size="small"><Popconfirm title="通过后将创建账号" onConfirm={() => reviewApplication(application, true)}><Button size="small" type="primary">通过</Button></Popconfirm><Popconfirm title="确认拒绝此申请" onConfirm={() => reviewApplication(application, false)}><Button size="small">拒绝</Button></Popconfirm></Space>
            </div>) : <div className="admin-empty">暂无待审核申请</div>}
          </section>
          {isOwner ? <section className="admin-surface">
            <div className="admin-surface-heading"><strong>当前授权</strong><span>{grants.length} 项</span></div>
            {grants.length ? grants.map(grant => <div className="admin-grant" key={grant.id}><div><strong>{grant.subject}</strong><p>{capabilityLabel(grant.capability)} · {scopeLabel(grant.scope)}</p></div><Popconfirm title="确认撤销此权限" onConfirm={() => revokeGrant(grant.id)}><Button type="link" danger size="small">撤销</Button></Popconfirm></div>) : <div className="admin-empty">暂无额外授权</div>}
          </section> : null}
        </aside>
      </div>

      <Drawer title="新建账号" open={accountDrawerOpen} onClose={() => setAccountDrawerOpen(false)} width={420} extra={<Button type="primary" loading={submitting} onClick={() => accountForm.submit()}>创建</Button>}>
        <Form form={accountForm} layout="vertical" onFinish={createAccount} initialValues={{ active: true }}>
          <Form.Item name="username" label="账号" rules={[{ required: true, message: '请输入账号' }, { pattern: /^[A-Za-z0-9][A-Za-z0-9._-]{1,31}$/, message: '2-32 位字母、数字、点、下划线或短横线' }]}><Input autoComplete="off" /></Form.Item>
          <Form.Item name="password" label="初始密码" rules={[{ required: true, min: 8, message: '密码至少 8 位' }]}><Input.Password autoComplete="new-password" /></Form.Item>
          <Alert type="info" showIcon message="初始密码仅在本次提交时发送，不会保存到浏览器或回显。" />
        </Form>
      </Drawer>

      <Drawer title="授予权限" open={grantDrawerOpen} onClose={() => setGrantDrawerOpen(false)} width={420} extra={<Button type="primary" loading={submitting} onClick={() => grantForm.submit()}>保存授权</Button>}>
        <Form form={grantForm} layout="vertical" onFinish={createGrant}>
          <Form.Item name="subject" label="账号" rules={[{ required: true, message: '请选择账号' }]}><Select options={accounts.filter(account => account.active && !account.isOwner).map(account => ({ value: account.username, label: account.username }))} /></Form.Item>
          <Form.Item name="capability" label="权限" rules={[{ required: true, message: '请选择权限' }]}><Select options={capabilityOptions} /></Form.Item>
          <Form.Item name="scope" label="作用模块" rules={[{ required: true, message: '请选择作用模块' }]}><Select options={scopeOptions} /></Form.Item>
        </Form>
      </Drawer>

      <Modal title={`重置密码${passwordAccount ? `：${passwordAccount.username}` : ''}`} open={Boolean(passwordAccount)} onCancel={() => setPasswordAccount(null)} onOk={() => passwordForm.submit()} confirmLoading={submitting} okText="重置">
        <Form form={passwordForm} layout="vertical" onFinish={resetPassword}>
          <Form.Item name="password" label="新密码" rules={[{ required: true, min: 8, message: '密码至少 8 位' }]}><Input.Password autoComplete="new-password" /></Form.Item>
        </Form>
      </Modal>
    </section>
  );
}
