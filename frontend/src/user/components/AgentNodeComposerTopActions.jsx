import { Dropdown, Tooltip } from 'antd';
import { Ellipsis } from 'lucide-react';

function stopCanvasEvent(event) {
  event.preventDefault();
  event.stopPropagation();
}

function visibleActions(actions) {
  return (Array.isArray(actions) ? actions : []).filter(action => action && action.visible !== false);
}

export function AgentNodeComposerTopActions({ primaryActions = [], moreActions = [] }) {
  const primary = visibleActions(primaryActions);
  const more = visibleActions(moreActions);
  const menu = more.map(action => ({
    key: action.id,
    label: action.label,
    disabled: Boolean(action.disabled),
    icon: action.icon,
    onClick: ({ domEvent }) => {
      stopCanvasEvent(domEvent);
      action.onClick?.(domEvent);
    }
  }));

  if (!primary.length && !menu.length) return null;

  return <div className="agent-node-composer__top-actions" onPointerDown={event => event.stopPropagation()} onClick={event => event.stopPropagation()}>
    {primary.map(action => action.content ? <span className="agent-node-composer__top-action-content" key={action.id}>{action.content}</span> : <button
      type="button"
      key={action.id}
      className="agent-node-composer__top-action"
      disabled={Boolean(action.disabled)}
      onClick={event => {
        stopCanvasEvent(event);
        action.onClick?.(event);
      }}
    >{action.icon}<span>{action.label}</span></button>)}
    {menu.length ? <Dropdown
      menu={{ items: menu }}
      trigger={['click']}
      // 节点和输入框都有 overflow:hidden，菜单挂在父节点里会被裁掉。
      // 挂到 body 后由 Ant Design 统一处理定位和层级，避免被上方视频节点遮挡。
      getPopupContainer={() => document.body}
      zIndex={1400}
    >
      <Tooltip title="更多功能">
        <button type="button" className="agent-node-composer__more-action" aria-label="更多功能" onClick={stopCanvasEvent}>
          <Ellipsis size={16} />
        </button>
      </Tooltip>
    </Dropdown> : null}
  </div>;
}
