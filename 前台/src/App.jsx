import React from 'react';
import { Button, ConfigProvider, Empty, Space, Typography, theme } from 'antd';
import AgentWorkspace from './AgentWorkspace.jsx';
import ApiConfigPage from './ApiConfigPage.jsx';
import BatchFactoryPage from './BatchFactoryPage.jsx';

const { Paragraph, Title } = Typography;

const MODULE_NAMES = {
  '/novel-fetch': '小说获取',
  '/script': '剧本生成',
  '/shuihuo-production': '水货生产',
  '/shuihuo-production/creative': '创作漫剧',
  '/settings': '设置',
};

function normalizedPathname() {
  const value = String(window.location.pathname || '/').replace(/\/+$/, '');
  return value || '/';
}

function navigate(path) {
  window.location.assign(path);
}

function MigrationPage({ pathname }) {
  const name = MODULE_NAMES[pathname] || '该功能';
  return (
    <main className="migration-shell">
      <div className="migration-card">
        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={null} />
        <Title level={2}>{name}</Title>
        <Paragraph>
          这个入口已经保留，但当前新 Go 主线还没有完成该模块的真实迁移。
          这里不会放一个假页面冒充功能已经可用。
        </Paragraph>
        <Paragraph type="secondary">
          下一步会按旧 v88 的真实链路迁移页面、API、MySQL/Redis/TOS 与任务状态，再替换这个过渡提示。
        </Paragraph>
        <Space wrap>
          <Button type="primary" onClick={() => navigate('/agent')}>返回 Agent</Button>
          <Button onClick={() => navigate('/batch-factory')}>打开批量工厂</Button>
        </Space>
      </div>
    </main>
  );
}

function AppRoute() {
  const pathname = normalizedPathname();
  if (pathname === '/' || pathname === '/agent') return <AgentWorkspace />;
  if (pathname === '/api-config') return <ApiConfigPage />;
  if (pathname === '/batch-factory') return <BatchFactoryPage />;
  if (MODULE_NAMES[pathname]) return <MigrationPage pathname={pathname} />;
  return <MigrationPage pathname={pathname} />;
}

export default function App() {
  return (
    <ConfigProvider
      theme={{
        algorithm: theme.darkAlgorithm,
        token: {
          borderRadius: 12,
          fontSize: 15,
          colorPrimary: '#7c82ff',
        },
      }}
    >
      <AppRoute />
    </ConfigProvider>
  );
}
