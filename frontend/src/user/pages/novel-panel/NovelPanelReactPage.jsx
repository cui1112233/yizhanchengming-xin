import { Alert, Button, List, Spin, Tabs } from 'antd';
import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  generateNovelPanelScenePrompts,
  getNovelPanelProject,
  createNovelPanelProject,
  listNovelPanelHistory,
  saveNovelPanelProject
} from '../../../shared/api/novelPanel';
import AssetWorkspace from './AssetWorkspace';
import GenerationPanel from './GenerationPanel';
import HistoryPanel from './HistoryPanel';
import StoryOutlineWorkspace from './StoryOutlineWorkspace';
import { normalizeNovelPanelError } from './novelPanelContracts';
import { createNovelPanelViewState } from './novelPanelPageState';
import './novel-panel-react.css';

export function NovelPanelReactPage({ theme = 'dark', projectId = '', onFallback }) {
  const [loading, setLoading] = useState(Boolean(projectId));
  const [project, setProject] = useState(null);
  const [error, setError] = useState(null);
  const [history, setHistory] = useState(null);
  const [projects, setProjects] = useState([]);
  const [projectListLoading, setProjectListLoading] = useState(!projectId);

  const loadProject = useCallback(async () => {
    if (!String(projectId || '').trim()) {
      setProjectListLoading(true);
      setError(null);
      listNovelPanelProjects()
        .then(response => setProjects(Array.isArray(response?.projects) ? response.projects : []))
        .catch(requestError => setError(normalizeNovelPanelError(requestError)))
        .finally(() => { setProjectListLoading(false); setLoading(false); });
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const response = await getNovelPanelProject(projectId);
      setProject(response?.project || response);
      try {
        setHistory(await listNovelPanelHistory());
      } catch {
        setHistory(null);
      }
    } catch (requestError) {
      setProject(null);
      setError(normalizeNovelPanelError(requestError));
    } finally {
      setLoading(false);
    }
  }, [projectId]);

  useEffect(() => { loadProject(); }, [loadProject]);

  const state = useMemo(() => createNovelPanelViewState({ projectId, loading, project, error, projects, projectListLoading }), [projectId, loading, project, error, projects, projectListLoading]);

  const saveProject = useCallback(async nextData => {
    const current = state.project;
    const response = await saveNovelPanelProject({
      project_id: current.id,
      name: current.name,
      data: nextData
    });
    setProject(response?.project || response);
  }, [state.project]);

  const generatePrompts = useCallback(async () => {
    const data = state.project.data || {};
    const response = await generateNovelPanelScenePrompts({
      project_id: state.project.id,
      outline_shots: data.outline_shots || [],
      assets: data.assets || data.referenceAssets || []
    });
    const nextData = { ...data, ...(response?.data || response) };
    setProject(current => ({ ...current, data: nextData }));
    return nextData;
  }, [state.project]);

  if (state.kind === 'loading') {
    return <main className={`novel-panel-react novel-panel-react--${theme}`}><Spin tip="正在加载小说面板…" /></main>;
  }
  if (state.kind === 'select') {
    const openProject = id => {
      const next = new URL(window.location.href);
      next.searchParams.set('project', id);
      window.location.assign(next.toString());
    };
    const createProject = async () => {
      const response = await createNovelPanelProject({ name: '未命名小说项目', data: { source_text: '', assets: [], outline_shots: [] } });
      const id = response?.project?.id || response?.id;
      if (id) openProject(id);
    };
    return (
      <main className={`novel-panel-react novel-panel-react--${theme}`}>
        <header className="novel-panel-react__header"><div><p className="novel-panel-react__eyebrow">小说面板</p><h1>选择项目</h1></div><Button type="primary" onClick={createProject}>新建小说项目</Button></header>
        <List bordered dataSource={projects} locale={{ emptyText: '暂无项目，请先新建' }} renderItem={item => <List.Item actions={[<Button key="open" type="link" onClick={() => openProject(item.id)}>打开</Button>]}><List.Item.Meta title={item.name || '未命名项目'} description={item.id} /></List.Item>} />
      </main>
    );
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
      <Tabs
        items={[
          { key: 'assets', label: '资产设置', children: <AssetWorkspace project={state.project} /> },
          { key: 'story', label: '故事与分镜', children: <StoryOutlineWorkspace outline={state.project.data?.outline || state.project.data} onSave={model => saveProject({ ...state.project.data, outline_shots: model.scenes.flatMap(scene => scene.shots.map(shot => ({ scene_index: scene.index, shot_index: shot.shotIndex, prompt: shot.prompt }))) })} onGenerate={generatePrompts} /> },
          { key: 'generation', label: '生成状态', children: <GenerationPanel onGenerate={async () => generatePrompts()} /> },
          { key: 'history', label: '历史记录', children: <HistoryPanel history={history} /> }
        ]}
      />
    </main>
  );
}

export default NovelPanelReactPage;
