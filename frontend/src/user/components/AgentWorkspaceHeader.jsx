import { Segmented, Tooltip } from 'antd';
import { MessageSquare, PanelLeft, Pencil, Save } from 'lucide-react';

export function AgentWorkspaceHeader({
  projectTitle,
  activeView,
  drawerOpen,
  conversationOpen,
  saveStatus,
  saveLabel,
  saveDisabled,
  nodeCount = 0,
  materialCount = 0,
  zoom = 100,
  isRunning = false,
  onRename,
  onViewChange,
  onToggleDrawer,
  onToggleConversation,
  onSave
}) {
  return (
    <header className="agent-canvas-toolbar">
      <div className="agent-canvas-toolbar__project">
        <Tooltip title={drawerOpen ? '收起项目栏' : '打开项目栏'}>
          <button
            className={`agent-canvas-toolbar-button${drawerOpen ? ' is-active' : ''}`}
            type="button"
            aria-label={drawerOpen ? '收起项目栏' : '打开项目栏'}
            aria-expanded={drawerOpen}
            aria-controls="agent-project-drawer"
            onClick={onToggleDrawer}
          >
            <PanelLeft size={16} />
          </button>
        </Tooltip>
        <button className="agent-canvas-project-title" type="button" onClick={onRename}>
          <span>{projectTitle || '未选择项目'}</span>
          <Pencil size={13} />
        </button>
      </div>

      <Segmented
        className="agent-canvas-view-switch"
        aria-label="项目视图"
        value={activeView}
        options={[
          { label: '工作流', value: 'workflow' },
          { label: '故事板', value: 'storyboard' }
        ]}
        onChange={onViewChange}
      />

      <div className="agent-canvas-toolbar__status">
        <span className={`agent-canvas-save-status is-${saveStatus}`}>
          {isRunning ? '正在执行…' : saveLabel}
        </span>
        <span className="agent-canvas-workspace-counts">{nodeCount} 个节点 · {materialCount} 项素材</span>
        <span className="agent-canvas-view-info">{zoom}%</span>
      </div>

      <div className="agent-canvas-toolbar__actions">
        <Tooltip title={saveLabel}>
          <span>
            <button className="agent-canvas-toolbar-button" type="button" aria-label={saveLabel} disabled={saveDisabled} onClick={onSave}>
              <Save size={16} />
            </button>
          </span>
        </Tooltip>
        <Tooltip title={conversationOpen ? '收起项目对话' : '打开项目对话'}>
          <button
            className={`agent-canvas-toolbar-button${conversationOpen ? ' is-active' : ''}`}
            type="button"
            aria-label={conversationOpen ? '收起项目对话' : '打开项目对话'}
            aria-pressed={conversationOpen}
            onClick={onToggleConversation}
          >
            <MessageSquare size={16} />
          </button>
        </Tooltip>
      </div>
    </header>
  );
}
