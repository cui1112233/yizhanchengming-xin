import { Check, ChevronDown, Layers } from 'lucide-react';
import { Dropdown, message } from 'antd';
import { useState } from 'react';
import { WORK_MODES, WORK_MODE_DUET, workModeForPath } from './workMode.js';
import './work-mode.css';

// 品牌行即工作模式入口：样式跟我们的主题走，交互形态参考洛水的三段式下拉。
export default function WorkModeMenu({ pathname, className = '', children, ariaLabel = '选择工作模式' }) {
  const [open, setOpen] = useState(false);
  const current = workModeForPath(pathname);

  const items = WORK_MODES.map(mode => ({
    key: mode.id,
    className: [
      'work-mode-item',
      mode.unavailable ? 'is-unavailable' : '',
      current === mode.id ? 'is-current' : ''
    ].filter(Boolean).join(' '),
    label: (
      <span className="work-mode-item-label">
        <span className="work-mode-item-name">{mode.label}</span>
        {mode.unavailable
          ? <em className="work-mode-badge">未实现</em>
          : current === mode.id ? <Check size={14} aria-hidden="true" /> : null}
      </span>
    )
  }));

  function handleMenuClick({ key }) {
    if (key === WORK_MODE_DUET) {
      message.info('Duet 模式尚未实现');
      return;
    }
    if (key === current) return;
    const mode = WORK_MODES.find(item => item.id === key);
    if (!mode?.href) return;
    // 与 shared/components/Link 使用同一套导航方式：应用没有 React Router。
    window.history.pushState({}, '', mode.href);
    window.dispatchEvent(new PopStateEvent('popstate'));
  }

  return (
    <Dropdown
      open={open}
      onOpenChange={setOpen}
      trigger={['click']}
      placement="bottomLeft"
      overlayClassName="work-mode-dropdown"
      menu={{ items, onClick: handleMenuClick, selectedKeys: [current] }}
    >
      <button
        type="button"
        className={`work-mode-trigger ${className}`.trim()}
        aria-haspopup="menu"
        aria-expanded={open}
        aria-label={ariaLabel}
      >
        <span className="work-mode-collapsed-icon" aria-hidden="true"><Layers size={18} strokeWidth={1.8} /></span>
        {children}
        <ChevronDown className="work-mode-chevron" size={14} aria-hidden="true" />
      </button>
    </Dropdown>
  );
}
