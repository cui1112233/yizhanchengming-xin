import { Alert, Button, Empty, Input, Space, Tag } from 'antd';
import { useMemo, useState } from 'react';
import { buildOutlineViewModel } from './storyOutlineModel';
import './story-outline-workspace.css';

export function StoryOutlineWorkspace({ outline, onSave, onGenerate, error }) {
  const initial = useMemo(() => buildOutlineViewModel(outline), [outline]);
  const [model, setModel] = useState(initial);
  const [busy, setBusy] = useState(false);
  const scenes = model.scenes || [];

  async function handleGenerate() {
    if (!onGenerate || busy) return;
    setBusy(true);
    try {
      const result = await onGenerate(model);
      const next = buildOutlineViewModel(result, model);
      setModel(next);
    } finally {
      setBusy(false);
    }
  }

  if (!scenes.length) return <Empty description="暂无分镜" />;
  return (
    <section className="novel-panel-outline" aria-label="故事剧本与分镜">
      {error && <Alert type="error" showIcon message={error} />}
      <div className="novel-panel-outline__toolbar">
        <Tag color="blue">{scenes.length} 个场景</Tag>
        <Space>
          <Button loading={busy} onClick={handleGenerate}>重新生成分镜</Button>
          <Button type="primary" onClick={() => onSave?.(model)}>保存分镜</Button>
        </Space>
      </div>
      {scenes.map(scene => (
        <article className="novel-panel-outline__scene" key={scene.index}>
          <h3>场景 {scene.index}</h3>
          {scene.shots.map(shot => (
            <div className="novel-panel-outline__shot" key={`${scene.index}-${shot.shotIndex}`}>
              <div className="novel-panel-outline__shot-title">镜头 {shot.shotIndex}</div>
              <Input.TextArea value={shot.prompt} autoSize={{ minRows: 2, maxRows: 8 }} onChange={event => setModel(current => ({ ...current, dirty: true, scenes: current.scenes.map(item => item.index !== scene.index ? item : { ...item, shots: item.shots.map(entry => entry.shotIndex !== shot.shotIndex ? entry : { ...entry, prompt: event.target.value }) }) }))} />
            </div>
          ))}
        </article>
      ))}
    </section>
  );
}

export default StoryOutlineWorkspace;
