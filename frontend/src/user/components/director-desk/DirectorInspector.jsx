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

export default function DirectorInspector({ project, images = [], selection, onProjectChange, onSetActiveCamera, onChangeScene, onSemanticCameraChange }) {
  const current = normalizeDirectorProject(project);
  const object = selection?.kind === 'actor' ? current.actors.find(item => item.id === selection.id)
    : selection?.kind === 'prop' ? current.props.find(item => item.id === selection.id)
    : selection?.kind === 'camera' ? current.cameras.find(item => item.id === selection.id) : null;
  if (selection?.kind === 'scene' && selection.id === 'scene') return <section className="agent-director-inspector" aria-label="场景检查器">
    <h3>场景</h3><p>{current.scene.name || '未选择场景图'}</p><p>画幅比例：{current.aspectRatio}</p>
    <button type="button" onClick={onChangeScene}>更换场景图片</button>
  </section>;
  if (!object) return <section className="agent-director-inspector"><p>请选择场景对象</p></section>;
  const actor = selection.kind === 'actor';
  const prop = selection.kind === 'prop';
  const prefix = actor ? '角色' : prop ? '道具' : '机位';
  const fields = actor || prop ? [['位置', 'position'], ['旋转', 'rotation'], ['缩放', 'scale']]
    : [['位置', 'position'], ['目标', 'target']];
  const field = (label, path, value) => <NumericField key={`${selection.id}-${path}`} label={label} value={value}
    onCommit={value => {
      const next = applyDirectorInspectorChange(current, selection, path, value);
      onProjectChange?.(next);
      const updated = (actor ? next.actors : prop ? next.props : next.cameras).find(item => item.id === selection.id);
      const [property, index] = path.split('.');
      return property === 'uniformScale' ? updated.scale[0] : index === undefined ? updated[property] : updated[property][Number(index)];
    }} />;
  return <section className="agent-director-inspector" aria-label={`${prefix}检查器`}>
    <h3>{actor ? (images.find(image => image.id === object.assetId)?.name?.trim() || object.assetId) : object.name}</h3>
    {selection.kind === 'camera' ? <fieldset><legend>语义机位</legend><div className="agent-director-inspector__semantic">
      <label>景别<select aria-label="景别" value={current.shotSize} onChange={event => onSemanticCameraChange?.({ shotSize: event.target.value })}>{[['extremeClose', '大特写'], ['close', '特写'], ['mediumClose', '近景'], ['medium', '中景'], ['full', '全景'], ['wide', '大全景']].map(([value, label]) => <option key={value} value={value}>{label}</option>)}</select></label>
      {['focalLength', 'azimuth', 'elevation'].map(key => <NumericField key={key} label={key === 'focalLength' ? '焦段 mm' : key === 'azimuth' ? '方位角' : '俯仰角'} value={current[key]} onCommit={value => { onSemanticCameraChange?.({ [key]: Number(value) }); return Number(value); }} />)}
    </div></fieldset> : null}
    {fields.map(([name, path]) => <fieldset key={path}><legend>{name}</legend><div className="agent-director-inspector__axes">
      {['X', 'Y', 'Z'].map((axis, index) => field(`${prefix}${name} ${axis}`, `${path}.${index}`, object[path][index]))}
    </div></fieldset>)}
    {actor ? field('角色统一缩放', 'uniformScale', object.scale[0]) : prop ? null : <>
      <fieldset><legend>镜头滚转</legend><div className="agent-director-inspector__axes">
        {field('机位滚转', 'rotation.2', object.rotation[2])}
      </div></fieldset>
      {field('机位 FOV', 'fov', object.fov)}
      <button type="button" onClick={() => onSetActiveCamera?.(object.id)}>设为活动机位</button>
    </>}
  </section>;
}
