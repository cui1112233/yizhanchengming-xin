import React, { useState } from 'react';
import { createRoot } from 'react-dom/client';
import '../../../shared/styles/global.css';
import '../../../shared/styles/agent-workspace.css';
import DirectorDeskShell from './DirectorDeskShell.jsx';
import DirectorToolRail from './DirectorToolRail.jsx';
import DirectorSceneTree from './DirectorSceneTree.jsx';

function Fixture() {
  const [leftCollapsed, setLeft] = useState(false);
  const [rightCollapsed, setRight] = useState(false);
  const [selected, setSelected] = useState({ kind: 'scene', id: 'scene' });
  const [tool, setTool] = useState('scene');
  return <div className="agent-page"><section className="agent-director-desk-fullscreen">
    <DirectorDeskShell leftCollapsed={leftCollapsed} rightCollapsed={rightCollapsed} onToggleLeft={() => setLeft(!leftCollapsed)} onToggleRight={() => setRight(!rightCollapsed)} onClose={() => {}} top={<span>导演台</span>} viewport={<div data-testid="viewport">视口</div>} right={<button type="button">右侧操作</button>} left={<><DirectorToolRail activeTool={tool} onToolChange={setTool} /><DirectorSceneTree project={{}} selectedObject={selected} onSelect={setSelected} onLocate={() => {}} onRemove={() => {}} /></>} />
  </section></div>;
}
createRoot(document.getElementById('root')).render(<Fixture />);
