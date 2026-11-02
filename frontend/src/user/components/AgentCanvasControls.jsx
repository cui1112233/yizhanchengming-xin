import { Button, Modal, Tooltip } from 'antd';
import { Eye, EyeOff, Grid2X2, Keyboard, Map, Maximize2, MessageCircle, Rows3, ZoomIn, ZoomOut } from 'lucide-react';
import { useState } from 'react';
import { CANVAS_SHORTCUTS } from '../pages/agent-canvas-commands';

function ControlButton({ label, active = false, disabled = false, onClick, children }) {
  return <Tooltip title={label}><span><Button className={active ? 'is-active' : ''} type="text" aria-label={label} disabled={disabled} icon={children} onClick={onClick} /></span></Tooltip>;
}

export function AgentCanvasControls({
  workspaceView,
  zoom,
  hasSelection,
  conversationOpen,
  onToggleConversation,
  onWorkspaceViewChange,
  onAutoLayout,
  onZoomIn,
  onZoomOut,
  onFitView,
  onFocusSelection
}) {
  const [shortcutsOpen, setShortcutsOpen] = useState(false);
  return <>
    <div className="agent-canvas-bottom-dock">
    <nav className="agent-canvas-display-controls" aria-label="画布显示与布局">
      <ControlButton label={workspaceView.showMinimap ? '隐藏小地图' : '显示小地图'} active={workspaceView.showMinimap} onClick={() => onWorkspaceViewChange({ showMinimap: !workspaceView.showMinimap })}>
        <Map size={16} />
      </ControlButton>
      <ControlButton label={workspaceView.showConnections ? '隐藏连线' : '显示连线'} active={workspaceView.showConnections} onClick={() => onWorkspaceViewChange({ showConnections: !workspaceView.showConnections })}>
        {workspaceView.showConnections ? <Eye size={16} /> : <EyeOff size={16} />}
      </ControlButton>
      <ControlButton label="网格吸附" active={workspaceView.snapToGrid} onClick={() => onWorkspaceViewChange({ snapToGrid: !workspaceView.snapToGrid })}><Grid2X2 size={16} /></ControlButton>
      <ControlButton label="自动横排" onClick={() => onAutoLayout('LR')}><Rows3 size={16} /></ControlButton>
      <ControlButton label="自动竖排" onClick={() => onAutoLayout('TB')}><Rows3 className="is-vertical" size={16} /></ControlButton>
      <ControlButton label="快捷键" onClick={() => setShortcutsOpen(true)}><Keyboard size={16} /></ControlButton>
    </nav>
    <span className="agent-canvas-tool-divider" />
    <nav className="agent-canvas-zoom-controls" aria-label="画布缩放">
      <ControlButton label="缩小" onClick={onZoomOut}><ZoomOut size={16} /></ControlButton>
      <span>{zoom}%</span>
      <ControlButton label="放大" onClick={onZoomIn}><ZoomIn size={16} /></ControlButton>
      <ControlButton label="适合屏幕" onClick={onFitView}><Maximize2 size={16} /></ControlButton>
      <ControlButton label={hasSelection ? '聚焦选中' : '请先选择节点'} disabled={!hasSelection} onClick={onFocusSelection}><Maximize2 size={16} /></ControlButton>
    </nav>
    <span className="agent-canvas-tool-divider" />
      <Tooltip title={conversationOpen ? '收起项目对话' : '打开项目对话'}>
        <button
          className={`agent-canvas-conversation-toggle${conversationOpen ? ' is-active' : ''}`}
          type="button"
          aria-label={conversationOpen ? '收起项目对话' : '打开项目对话'}
          aria-pressed={conversationOpen}
          onClick={onToggleConversation}
        >
          <MessageCircle size={18} />
        </button>
      </Tooltip>
    </div>

    <Modal className="agent-canvas-shortcuts-modal" title="画布快捷键" open={shortcutsOpen} footer={null} onCancel={() => setShortcutsOpen(false)}>
      <div className="agent-canvas-shortcuts-list">
        {CANVAS_SHORTCUTS.map(item => <div key={item.command}><span>{item.label}</span><kbd>{item.keys}</kbd></div>)}
      </div>
    </Modal>
  </>;
}
