import { Crosshair, Trash2 } from 'lucide-react';
import { directorSceneObjects } from '../../pages/agent-director-desk.js';

export default function DirectorSceneTree({ project, selectedObject, onSelect, onLocate, onRemove }) {
  return <section className="agent-director-scene-tree" aria-label="场景对象">
    <h2>场景对象</h2>
    {directorSceneObjects(project).map(item => {
      const object = { kind: item.kind, id: item.id };
      const name = item.name || (item.kind === 'scene' ? '场景' : item.kind === 'camera' ? '机位' : item.kind === 'prop' ? '道具' : '角色');
      return <div className="agent-director-scene-tree__item" key={`${item.kind}:${item.id}`}>
        <button type="button" className="agent-director-scene-tree__select" aria-pressed={selectedObject?.kind === item.kind && selectedObject?.id === item.id} onClick={() => onSelect(object)}>{name}</button>
        <button type="button" aria-label={`定位${name}`} title={`定位${name}`} onClick={() => onLocate(object)}><Crosshair size={15} /></button>
        {item.id !== 'scene' && item.id !== 'camera-main' && <button type="button" aria-label={`删除${name}`} title={`删除${name}`} onClick={() => onRemove(object)}><Trash2 size={15} /></button>}
      </div>;
    })}
  </section>;
}
