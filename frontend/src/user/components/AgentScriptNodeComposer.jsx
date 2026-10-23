import { Button, Checkbox, Input, Segmented } from 'antd';
import { AudioLines, Image, Scissors, Video } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { buildShotBatchPayload, scriptSplitReady, selectedShotIDs } from '../pages/agent-script-shots';
import { reconcileNodeModel } from '../pages/agent-model-capabilities';
import { AgentBatchModelPicker } from './AgentBatchModelPicker';
import { AgentModelPicker } from './AgentModelPicker';

const splitStateLabels = {
  idle: '尚未拆分', stale: '剧本有修改，请重新拆分', pending: '等待拆分', queued: '等待拆分', running: '正在拆分', processing: '正在拆分',
  completed: '拆分完成', done: '拆分完成', failed: '拆分失败', canceled: '已取消', cancelled: '已取消'
};

const batchKinds = [
  { kind: 'image', label: '图片', action: '生成图片', icon: <Image size={14} /> },
  { kind: 'video', label: '视频', action: '生成视频', icon: <Video size={14} /> },
  { kind: 'audio', label: '配音', action: '生成配音', icon: <AudioLines size={14} /> }
];
const emptyShots = Object.freeze([]);

function splitStateValue(config, state) {
  if (state?.state) return state;
  const stored = config?.splitState;
  if (stored && typeof stored === 'object') return stored;
  return { state: String(stored || 'idle'), errorMessage: '' };
}

function shotDescription(shot) {
  return [shot.scene, shot.characters, shot.dialogue, shot.narration].filter(Boolean).join(' · ') || '暂无镜头描述';
}

function shotReadiness(shot) {
  const ready = [shot.imagePrompt && '图片', shot.videoPrompt && '视频', shot.ttsText && '配音'].filter(Boolean);
  return ready.length ? `${ready.join(' · ')}提示词已就绪` : '待完善镜头内容';
}

export default function AgentScriptNodeComposer({
  node, modelsByKind = {}, modelLoadError = '', scriptState = {},
  onConfigChange, onSourceTextChange, onSplitScript, onBatchShots
}) {
  const config = node?.data?.config || {};
  const sourceText = String(config.sourceText || '');
  const [draftSource, setDraftSource] = useState(sourceText);
  const [selection, setSelection] = useState({});
  const [activeKind, setActiveKind] = useState('image');
  const [batchModels, setBatchModels] = useState({
    image: { provider: 'personal_api', modelId: '' },
    video: { provider: 'personal_api', modelId: '' },
    audio: { provider: 'personal_api', modelId: '' }
  });
  const [batchError, setBatchError] = useState('');
  const [batchBusyKind, setBatchBusyKind] = useState('');
  useEffect(() => setDraftSource(sourceText), [node?.id, sourceText]);

  const shots = Array.isArray(scriptState.shots) ? scriptState.shots : emptyShots;
  useEffect(() => {
    const available = new Set(shots.map(shot => shot.id));
    setSelection(current => Object.fromEntries(Object.entries(current).filter(([id, checked]) => checked && available.has(id))));
  }, [shots]);

  const selectedIDs = useMemo(() => selectedShotIDs(shots, selection), [selection, shots]);
  const currentSplitState = splitStateValue(config, scriptState);
  const stateName = String(currentSplitState.state || 'idle').toLowerCase();
  const splitBusy = ['pending', 'queued', 'running', 'processing'].includes(stateName);
  const textModels = modelsByKind.text || [];
  const splitModel = reconcileNodeModel({ ...config, modelId: config.splitModelId || '' }, textModels, 'text');
  const splitReady = scriptSplitReady(config);
  const activeBatchKind = batchKinds.find(item => item.kind === activeKind) || batchKinds[0];
  const activeModel = batchModels[activeBatchKind.kind] || {};
  const stop = event => event.stopPropagation();
  const update = patch => onConfigChange?.(node.id, patch);
  const changeSource = event => {
    const nextSource = event.target.value;
    setDraftSource(nextSource);
    onSourceTextChange?.(node.id, nextSource);
  };

  async function splitScript() {
    const preparedSource = draftSource.trim();
    const modelId = String(config.splitModelId || '').trim();
    if (!preparedSource || !modelId || splitBusy) return;
    setBatchError('');
    await onSplitScript?.(node.id, { nodeId: node.id, nodeVersion: node.data?.version || 1, modelId, sourceText: draftSource });
  }

  async function submitBatch() {
    const payload = buildShotBatchPayload({
      scriptNodeId: node.id, selectedIDs, kind: activeBatchKind.kind,
      provider: activeModel.provider, modelId: activeModel.modelId
    });
    if (!payload.shotIds.length || !payload.modelId) return;
    setBatchError('');
    setBatchBusyKind(activeBatchKind.kind);
    try {
      const result = await onBatchShots?.(node.id, payload);
      const skipped = Array.isArray(result?.skipped) ? result.skipped : [];
      if (skipped.length) setBatchError(skipped.map(item => item.reason).filter(Boolean).join('；'));
    } catch (error) {
      setBatchError(error?.message || '批量生成提交失败');
    } finally {
      setBatchBusyKind('');
    }
  }

  return <section className="agent-node-composer agent-script-node-composer nodrag nopan" aria-label="剧本工作流" onMouseDown={stop} onDoubleClick={stop}>
    <section className="agent-script-step" aria-labelledby="agent-script-step-source">
      <h4 id="agent-script-step-source"><span>1</span>剧本</h4>
      <Input.TextArea
        aria-label="剧本正文"
        className="agent-node-composer__prompt"
        value={draftSource}
        maxLength={60000}
        autoSize={{ minRows: 4, maxRows: 8 }}
        placeholder="粘贴剧本正文"
        onChange={changeSource}
      />
      <div className="agent-script-step__controls">
        <AgentModelPicker
          kind="text"
          config={{ ...config, modelId: config.splitModelId || '' }}
          models={textModels}
          disabled={splitBusy}
          onChange={patch => update({ splitModelId: patch.modelId })}
        />
        <span className={`agent-node-composer__state${stateName === 'failed' ? ' is-error' : ''}`} aria-live="polite">
          {currentSplitState.errorMessage || scriptState.errorMessage || modelLoadError || splitStateLabels[stateName] || '等待拆分'}
        </span>
        <Button
          type="primary"
          size="small"
          icon={<Scissors size={14} />}
          loading={splitBusy}
          disabled={!draftSource.trim() || !splitModel.modelId || !splitModel.available}
          onClick={splitScript}
        >拆分镜头</Button>
      </div>
    </section>

    <section className="agent-script-step" aria-labelledby="agent-script-step-shots">
      <div className="agent-script-step__heading">
        <h4 id="agent-script-step-shots"><span>2</span>镜头</h4>
        <span>{splitReady ? `${shots.length} 个镜头 · 已选 ${selectedIDs.length}` : splitStateLabels[stateName] || '等待拆分'}</span>
      </div>
      {shots.length && splitReady ? <>
        <div className="agent-script-shot-list__select-all">
          <Checkbox
            checked={selectedIDs.length === shots.length}
            indeterminate={selectedIDs.length > 0 && selectedIDs.length < shots.length}
            onChange={event => setSelection(event.target.checked ? Object.fromEntries(shots.map(shot => [shot.id, true])) : {})}
          >全选镜头</Checkbox>
        </div>
        <div className="agent-script-shot-list">
          {shots.map(shot => <article key={shot.id} className="agent-script-shot-row">
            <Checkbox
              aria-label={`选择镜头 ${String(shot.sequence).padStart(2, '0')}`}
              checked={Boolean(selection[shot.id])}
              onChange={event => setSelection(current => ({ ...current, [shot.id]: event.target.checked }))}
            />
            <div className="agent-script-shot-row__content">
              <strong>镜头 {String(shot.sequence).padStart(2, '0')}</strong>
              <span title={shotDescription(shot)}>{shotDescription(shot)}</span>
              <small>{shotReadiness(shot)}</small>
            </div>
          </article>)}
        </div>
      </> : <p className="agent-script-step__empty">{splitBusy ? '镜头拆分完成后会显示在这里' : '先完成第 1 步，拆分出镜头'}</p>}
    </section>

    <section className="agent-script-step" aria-labelledby="agent-script-step-generate">
      <div className="agent-script-step__heading">
        <h4 id="agent-script-step-generate"><span>3</span>生成</h4>
        <span>{selectedIDs.length ? `将处理 ${selectedIDs.length} 个镜头` : '先勾选要处理的镜头'}</span>
      </div>
      <Segmented
        block
        aria-label="生成内容"
        value={activeBatchKind.kind}
        options={batchKinds.map(item => ({ value: item.kind, label: item.label }))}
        onChange={setActiveKind}
      />
      <div className="agent-script-step__generate-controls">
        <AgentBatchModelPicker
          kind={activeBatchKind.kind}
          config={activeModel}
          models={modelsByKind[activeBatchKind.kind] || []}
          onChange={patch => setBatchModels(current => ({ ...current, [activeBatchKind.kind]: patch }))}
          disabled={Boolean(batchBusyKind)}
        />
        <Button
          type="primary"
          size="small"
          icon={activeBatchKind.icon}
          loading={batchBusyKind === activeBatchKind.kind}
          disabled={Boolean(batchBusyKind) || !splitReady || !selectedIDs.length || !activeModel.modelId}
          onClick={submitBatch}
        >{activeBatchKind.action} {selectedIDs.length ? `(${selectedIDs.length})` : ''}</Button>
      </div>
      {batchError ? <span className="agent-node-composer__state is-error" role="alert">{batchError}</span> : null}
    </section>
  </section>;
}
