import { useEffect, useId, useState } from 'react';
import { normalizeDirectorProject } from '../../pages/agent-director-desk.js';
import { applyDirectorInspectorChange } from './director-inspector-state.js';

function NumericField({ label, value, onCommit }) {
  const [draft, setDraft] = useState(String(value));
  const [error, setError] = useState(false);
  const errorId = useId();
  useEffect(() => { setDraft(String(value)); setError(false); }, [value]);
  const commit = () => {
    if (!draft.trim() || !Number.isFinite(Number(draft))) {
      setDraft(String(value));
      setError(true);
      return;
    }
    setError(false);
    setDraft(String(onCommit(draft)));
  };
  return <label className="agent-director-inspector__field">
    <span>{label}</span>
    <input aria-label={label} aria-invalid={error} aria-describedby={error ? errorId : undefined} type="text" inputMode="decimal" value={draft}
      onChange={event => { setDraft(event.target.value); setError(false); }} onBlur={commit}
      onKeyDown={event => { if (event.key === 'Enter') event.currentTarget.blur(); }} />
    {error ? <small id={errorId} role="alert">请输入有效数值</small> : null}
  </label>;
}

export default function DirectorInspector({ project, images = [], selection, onProjectChange, onSetActiveCamera, onChangeScene }) {
  const current = normalizeDirectorProject(project);
  const object = selection?.kind === 'actor' ? current.actors.find(item => item.id === selection.id)
    : selection?.kind === 'camera' ? current.cameras.find(item => item.id === selection.id) : null;
  if (selection?.kind === 'scene' && selection.id === 'scene') return <section className="agent-director-inspector" aria-label="场景检查器">
    <h3>场景</h3><p>{current.scene.name || '未选择场景图'}</p><p>画幅比例：{current.aspectRatio}</p>
    <button type="button" onClick={onChangeScene}>更换场景图片</button>
  </section>;
  if (!object) return <section className="agent-director-inspector"><p>请选择场景对象</p></section>;
  const actor = selection.kind === 'actor';
  const prefix = actor ? '角色' : '机位';
  const fields = actor ? [['位置', 'position'], ['旋转', 'rotation'], ['缩放', 'scale']]
    : [['位置', 'position'], ['旋转', 'rotation'], ['目标', 'target']];
  const field = (label, path, value) => <NumericField key={`${selection.id}-${path}`} label={label} value={value}
    onCommit={value => {
      const next = applyDirectorInspectorChange(current, selection, path, value);
      onProjectChange?.(next);
      const updated = (actor ? next.actors : next.cameras).find(item => item.id === selection.id);
      const [property, index] = path.split('.');
      return property === 'uniformScale' ? updated.scale[0] : index === undefined ? updated[property] : updated[property][Number(index)];
    }} />;
  return <section className="agent-director-inspector" aria-label={`${prefix}检查器`}>
    <h3>{actor ? (images.find(image => image.id === object.assetId)?.name?.trim() || object.assetId) : object.name}</h3>
    {fields.map(([name, path]) => <fieldset key={path}><legend>{name}</legend><div className="agent-director-inspector__axes">
      {['X', 'Y', 'Z'].map((axis, index) => field(`${prefix}${name} ${axis}`, `${path}.${index}`, object[path][index]))}
    </div></fieldset>)}
    {actor ? field('角色统一缩放', 'uniformScale', object.scale[0]) : <>
      {field('机位 FOV', 'fov', object.fov)}
      <button type="button" onClick={() => onSetActiveCamera?.(object.id)}>设为活动机位</button>
    </>}
  </section>;
}
