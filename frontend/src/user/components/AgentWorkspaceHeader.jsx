import { Dropdown } from 'antd';
import { Bot, ChevronDown, Library, ListTodo, PanelsTopLeft, Pencil, Workflow } from 'lucide-react';
import { BrandLogo } from '../../shared/components/BrandLogo';

export function AgentWorkspaceHeader({
  projectTitle,
  activeView,
  saveStatus,
  nodeCount = 0,
  materialCount = 0,
  zoom = 100,
  isRunning = false,
  onRename,
  onBackHome,
  onShowProjects,
  onCreateProject,
  onDeleteProject,
  onViewChange,
  onRetrySync,
  onOpenProxy
}) {
  const projectMenu = {
    items: [
      { key: 'home', label: '回到主页', onClick: onBackHome },
      { key: 'projects', label: '全部项目', onClick: onShowProjects },
      { type: 'divider' },
      { key: 'new', label: '创建新项目', onClick: onCreateProject },
      { key: 'delete', label: '删除项目', danger: true, disabled: !projectTitle, onClick: onDeleteProject }
    ]
  };

  return (
    <header className="agent-canvas-toolbar">
      <div className="agent-canvas-toolbar__project">
        <Dropdown menu={projectMenu} trigger={['click']} placement="bottomLeft">
          <button className="agent-canvas-brand-menu" type="button" aria-label="项目菜单">
            <BrandLogo />
          </button>
        </Dropdown>
        <button className="agent-canvas-project-title" type="button" onClick={onRename}>
          <span>{projectTitle || '未选择项目'}</span>
          <ChevronDown className="agent-canvas-project-title__chevron" size={12} />
          <Pencil className="agent-canvas-project-title__edit" size={12} />
        </button>
        <div className="agent-canvas-view-switch" role="group" aria-label="项目视图">
          <button type="button" title="画布" aria-label="画布" aria-pressed={activeView === 'workflow'} className={activeView === 'workflow' ? 'is-active' : ''} onClick={() => onViewChange('workflow')}>
            <Workflow size={15} />
          </button>
          <button type="button" title="故事板" aria-label="故事板" aria-pressed={activeView === 'storyboard'} className={activeView === 'storyboard' ? 'is-active' : ''} onClick={() => onViewChange('storyboard')}>
            <PanelsTopLeft size={15} />
          </button>
          <button type="button" title="素材" aria-label="素材" aria-pressed={activeView === 'assets'} className={activeView === 'assets' ? 'is-active' : ''} onClick={() => onViewChange('assets')}>
            <Library size={15} />
          </button>
          <button type="button" title="任务" aria-label="任务" aria-pressed={activeView === 'tasks'} className={activeView === 'tasks' ? 'is-active' : ''} onClick={() => onViewChange('tasks')}>
            <ListTodo size={15} />
          </button>
        </div>
      </div>

      <div className="agent-canvas-toolbar__status">
        {onOpenProxy ? (
          <button className="agent-proxy-entry-btn" type="button" onClick={onOpenProxy} title="打开 Agent 代理工作区">
            <Bot size={13} />
            <span>Agent 代理</span>
          </button>
        ) : null}
        {isRunning ? <span className="agent-canvas-save-status is-running">正在执行…</span> : null}
        {!isRunning && saveStatus === 'error' ? <button className="agent-canvas-save-status is-error" type="button" onClick={onRetrySync}>内容已暂存在本机，重试同步</button> : null}
        <span className="agent-canvas-workspace-counts">{nodeCount} 个节点 · {materialCount} 项素材</span>
        <span className="agent-canvas-view-info">{zoom}%</span>
      </div>
    </header>
  );
}
