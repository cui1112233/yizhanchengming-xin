import NovelPanelReactPage from './novel-panel/NovelPanelReactPage';

/** 小说面板唯一入口：React 页面直接读取 Go API。 */
export function NovelPanelPage({ theme }) {
  const projectId = new URLSearchParams(window.location.search).get('project') || '';
  return <NovelPanelReactPage theme={theme} projectId={projectId} />;
}

export default NovelPanelPage;
