import React, { useState } from 'react';
import { createRoot } from 'react-dom/client';
import '../../../shared/styles/global.css';
import '../../../shared/styles/agent-workspace.css';
import DirectorInspector from './DirectorInspector.jsx';
import { normalizeDirectorProject } from '../../pages/agent-director-desk.js';

function Fixture() {
  const [project, setProject] = useState(() => normalizeDirectorProject({ actors: [{ id: 'actor-1', assetId: 'image-1', position: [1, 2, 3] }] }));
  return <div className="agent-page"><section className="agent-director-desk-fullscreen"><div className="agent-director-shell">
    <div className="agent-director-shell__right"><DirectorInspector project={project} images={[{ id: 'image-1', name: '角色素材' }]} selection={{ kind: 'actor', id: 'actor-1' }} onProjectChange={setProject} /></div>
  </div></section></div>;
}
createRoot(document.getElementById('root')).render(<Fixture />);
