import React, { useState } from 'react';
import { createRoot } from 'react-dom/client';
import '../../../shared/styles/global.css';
import '../../../shared/styles/agent-workspace.css';
import DirectorInspector from './DirectorInspector.jsx';
import { normalizeDirectorProject } from '../../pages/agent-director-desk.js';

function Fixture() {
  const [project, setProject] = useState(() => normalizeDirectorProject({ actors: [{ id: 'actor-1', assetId: 'image-1', position: [1, 2, 3] }] }));
  const [selection, setSelection] = useState({ kind: 'actor', id: 'actor-1' });
  return <div className="agent-page"><section className="agent-director-desk-fullscreen"><div className="agent-director-shell">
    <button type="button" onClick={() => setSelection({ kind: 'actor', id: 'actor-1' })}>检查演员</button>
    <button type="button" onClick={() => setSelection({ kind: 'camera', id: 'camera-main' })}>检查机位</button>
    <output data-testid="camera-roll">{project.cameras[0].rotation[2]}</output>
    <div className="agent-director-shell__right"><DirectorInspector project={project} images={[{ id: 'image-1', name: '角色素材' }]} selection={selection} onProjectChange={setProject} /></div>
  </div></section></div>;
}
createRoot(document.getElementById('root')).render(<Fixture />);
