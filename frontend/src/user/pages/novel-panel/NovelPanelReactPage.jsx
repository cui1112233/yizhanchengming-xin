import { Alert, Button, Spin } from 'antd';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { getNovelPanelProject } from '../../../shared/api/novelPanel';
import { normalizeNovelPanelError } from './novelPanelContracts';
import { createNovelPanelViewState } from './novelPanelPageState';
import './novel-panel-react.css';

export function NovelPanelReactPage({ theme = 'dark', projectId = '', onFallback }) {
  const [loading, setLoading] = useState(Boolean(projectId));
  const [project, setProject] = useState(null);
  const [error, setError] = useState(null);

  const loadProject = useCallback(async () => {
    if (!String(projectId || '').trim()) {
      setLoading(false);
      setProject(null);
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const response = await getNovelPanelProject(projectId);
      setProject(response?.project || response);
    } catch (requestError) {
      setProject(null);
      setError(normalizeNovelPanelError(requestError));
    } finally {
      setLoading(false);
    }
  }, [projectId]);

  useEffect(() => { loadProject(); }, [loadProject]);

  const state = useMemo(() => createNovelPanelViewState({ projectId, loading, project, error }), [projectId, loading, project, error]);

  if (state.kind === 'loading') {
    return <main className={`novel-panel-react novel-panel-react--${theme}`}><Spin tip="正在加载小说面板…" /></main>;
  }
  if (state.kind === 'error') {
    return (
      <main className={`novel-panel-react novel-panel-react--${theme}`}>
        <Alert type="error" showIcon message="小说面板暂时无法打开" description={state.error.message} />
        <div className="novel-panel-react__actions">
          {state.error.retryable && <Button onClick={loadProject}>重试</Button>}
          <Button onClick={onFallback}>打开兼容工作台</Button>
        </div>
      </main>
    );
  }
  return (
    <main className={`novel-panel-react novel-panel-react--${theme}`}>
      <header className="novel-panel-react__header">
        <div>
          <p className="novel-panel-react__eyebrow">小说面板</p>
          <h1>{state.project.name || '未命名项目'}</h1>
        </div>
        <Button onClick={onFallback}>兼容工作台</Button>
      </header>
      <section className="novel-panel-react__empty">
        <h2>React 小说面板</h2>
        <p>项目已加载，资产、剧本、分镜和生成模块将在后续迁移阶段接入。</p>
      </section>
    </main>
  );
}

export default NovelPanelReactPage;
