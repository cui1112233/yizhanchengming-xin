import { PanelLeftClose, PanelLeftOpen, PanelRightClose, PanelRightOpen, X } from 'lucide-react';

export default function DirectorDeskShell({ left, viewport, right, top, bottom, leftCollapsed, rightCollapsed, onToggleLeft, onToggleRight, onClose }) {
  return <div className={`agent-director-shell${leftCollapsed ? ' is-left-collapsed' : ''}${rightCollapsed ? ' is-right-collapsed' : ''}`}>
    <header className="agent-director-shell__top">{top}<button type="button" aria-label="关闭导演台" title="关闭导演台" onClick={onClose}><X size={16} /></button></header>
    <aside className={`agent-director-shell__left${leftCollapsed ? ' is-collapsed' : ''}`}>
      <button type="button" className="agent-director-shell__toggle" aria-label={leftCollapsed ? '展开左侧面板' : '收起左侧面板'} title={leftCollapsed ? '展开左侧面板' : '收起左侧面板'} aria-expanded={!leftCollapsed} onClick={onToggleLeft}>{leftCollapsed ? <PanelLeftOpen size={16} /> : <PanelLeftClose size={16} />}</button>
      {!leftCollapsed && <div className="agent-director-shell__left-content">{left}</div>}
    </aside>
    <main className="agent-director-shell__viewport">{viewport}</main>
    <aside className={`agent-director-shell__right${rightCollapsed ? ' is-collapsed' : ''}`}>
      <button type="button" className="agent-director-shell__toggle" aria-label={rightCollapsed ? '展开右侧面板' : '收起右侧面板'} title={rightCollapsed ? '展开右侧面板' : '收起右侧面板'} aria-expanded={!rightCollapsed} onClick={onToggleRight}>{rightCollapsed ? <PanelRightOpen size={16} /> : <PanelRightClose size={16} />}</button>
      {!rightCollapsed && <div className="agent-director-shell__right-content">{right}</div>}
    </aside>
    <footer className="agent-director-shell__bottom">{bottom}</footer>
  </div>;
}
