import { PanelLeftClose, PanelLeftOpen, PanelRightClose, PanelRightOpen, X } from 'lucide-react';
import { useState, useSyncExternalStore } from 'react';

const narrowQuery = '(max-width: 1180px)';
const subscribeNarrow = callback => {
  const query = window.matchMedia(narrowQuery);
  query.addEventListener('change', callback);
  return () => query.removeEventListener('change', callback);
};
const isNarrow = () => window.matchMedia(narrowQuery).matches;

export default function DirectorDeskShell({ left, viewport, right, top, bottom, leftCollapsed, rightCollapsed, onToggleLeft, onToggleRight, onClose }) {
  const narrow = useSyncExternalStore(subscribeNarrow, isNarrow, () => false);
  const [preferredSide, setPreferredSide] = useState('left');
  const leftOpen = !leftCollapsed && !(narrow && !rightCollapsed && preferredSide === 'right');
  const rightOpen = !rightCollapsed && !(narrow && !leftCollapsed && preferredSide === 'left');
  const toggleLeft = () => {
    if (narrow && !leftOpen && !rightCollapsed) onToggleRight();
    if (leftOpen || leftCollapsed) onToggleLeft();
    setPreferredSide('left');
  };
  const toggleRight = () => {
    if (narrow && !rightOpen && !leftCollapsed) onToggleLeft();
    if (rightOpen || rightCollapsed) onToggleRight();
    setPreferredSide('right');
  };
  return <div className={`agent-director-shell${leftOpen ? '' : ' is-left-collapsed'}${rightOpen ? '' : ' is-right-collapsed'}`}>
    <header className="agent-director-shell__top">{top}<button type="button" aria-label="关闭导演台" title="关闭导演台" onClick={onClose}><X size={16} /></button></header>
    <aside className={`agent-director-shell__left${leftOpen ? '' : ' is-collapsed'}`}>
      <button type="button" className="agent-director-shell__toggle" aria-label={leftOpen ? '收起左侧面板' : '展开左侧面板'} title={leftOpen ? '收起左侧面板' : '展开左侧面板'} aria-expanded={leftOpen} onClick={toggleLeft}>{leftOpen ? <PanelLeftClose size={16} /> : <PanelLeftOpen size={16} />}</button>
      {leftOpen && <div className="agent-director-shell__left-content">{left}</div>}
    </aside>
    <main className="agent-director-shell__viewport">{viewport}</main>
    <aside className={`agent-director-shell__right${rightOpen ? '' : ' is-collapsed'}`}>
      <button type="button" className="agent-director-shell__toggle" aria-label={rightOpen ? '收起右侧面板' : '展开右侧面板'} title={rightOpen ? '收起右侧面板' : '展开右侧面板'} aria-expanded={rightOpen} onClick={toggleRight}>{rightOpen ? <PanelRightClose size={16} /> : <PanelRightOpen size={16} />}</button>
      {rightOpen && <div className="agent-director-shell__right-content">{right}</div>}
    </aside>
    <footer className="agent-director-shell__bottom">{bottom}</footer>
  </div>;
}
