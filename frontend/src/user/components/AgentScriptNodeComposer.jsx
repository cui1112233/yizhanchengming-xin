import { Button, Checkbox, Input, InputNumber, Segmented } from 'antd';
import { AudioLines, Image, Scissors, Video } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { buildShotBatchPayload, scriptSplitReady, selectedShotIDs } from '../pages/agent-script-shots';
import { reconcileNodeModel } from '../pages/agent-model-capabilities';
import { AgentBatchModelPicker } from './AgentBatchModelPicker';
import { AgentBatchFactoryPromptSelector } from './AgentBatchFactoryPromptSelector';
import { AgentModelPicker } from './AgentModelPicker';
import { AgentNodeComposerTopActions } from './AgentNodeComposerTopActions';
import { compactComposerActionPlan } from '../pages/agent-node-composer-actions';
import { batchFactoryPreviewText } from '../pages/shuihuo/batchFactoryContentRange';

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
  compact = false,
  node, modelsByKind = {}, modelLoadError = '', scriptState = {},
  onConfigChange, onSourceTextChange, onCommitSource, onRunBatchFactoryStage, onSplitScript, onBatchShots, onBatchFactoryPromptSaved
}) {
  const config = node?.data?.config || {};
  const isBatchFactoryScript = config.source === 'batch_factory' && config.resourceType === 'book_script';
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
  const [batchFactoryStage, setBatchFactoryStage] = useState('');
  const [batchFactoryError, setBatchFactoryError] = useState('');
  const [batchFactoryRangeLines, setBatchFactoryRangeLines] = useState(() => Math.min(500, Math.max(1, Number(config.contentRangeLines) || 5)));
  const [viralCandidate, setViralCandidate] = useState('');
  useEffect(() => setDraftSource(sourceText), [node?.id, sourceText]);
  useEffect(() => setBatchFactoryRangeLines(Math.min(500, Math.max(1, Number(config.contentRangeLines) || 5))), [node?.id, config.contentRangeLines]);
  useEffect(() => setViralCandidate(''), [node?.id, sourceText]);

  const batchFactoryPreview = isBatchFactoryScript ? batchFactoryPreviewText(draftSource, { sourceMetadata: { contentRangeLines: batchFactoryRangeLines } }) : draftSource;

  const shots = Array.isArray(scriptState.shots) ? scriptState.shots : emptyShots;
  useEffect(() => {
    const available = new Set(shots.map(shot => shot.id));
    setSelection(current => Object.fromEntries(Object.entries(current).filter(([id, checked]) => checked && available.has(id))));
  }, [shots]);

  const selectedIDs = useMemo(() => selectedShotIDs(shots, selection), [selection, shots]);
  const currentSplitState = splitStateValue(config, scriptState);
  const stateName = String(currentSplitState.state || 'idle').toLowerCase();
  const splitBusy = ['pending', 'queued', 'running', 'processing'].includes(stateName);
  const sourceStatus = isBatchFactoryScript
    ? '批量工厂正文已同步'
    : currentSplitState.errorMessage || scriptState.errorMessage || modelLoadError || splitStateLabels[stateName] || '等待拆分';
  const textModels = modelsByKind.text || [];
  const selectedTextModelID = config.splitModelId || (isBatchFactoryScript ? config.textModelId || config.modelId || '' : '');
  const splitModel = reconcileNodeModel({ ...config, modelId: selectedTextModelID }, textModels, 'text');
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
    const modelId = String(selectedTextModelID || '').trim();
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

  async function runBatchFactoryStage(stage) {
    if (!isBatchFactoryScript || !draftSource.trim() || batchFactoryStage) return;
    setBatchFactoryError('');
    setBatchFactoryStage(stage);
    try {
      const result = await onRunBatchFactoryStage?.(node.id, stage, splitModel.modelId, {
        contentRangeLines: batchFactoryRangeLines,
        sourceText: draftSource
      });
      if (stage === 'viral') setViralCandidate(String(result?.candidate || '').trim());
    } catch (error) {
      setBatchFactoryError(error?.message || '批量工厂步骤提交失败');
    } finally {
      setBatchFactoryStage('');
    }
  }

  async function adoptViralCandidate() {
    const nextSource = viralCandidate.trim();
    if (!nextSource || batchFactoryStage) return;
    setBatchFactoryError('');
    setBatchFactoryStage('adopt-viral');
    try {
      setDraftSource(nextSource);
      onSourceTextChange?.(node.id, nextSource);
      await onCommitSource?.(node.id, nextSource);
      setViralCandidate('');
    } catch (error) {
      setBatchFactoryError(error?.message || '采用爆款候选失败');
    } finally {
      setBatchFactoryStage('');
    }
  }

  if (isBatchFactoryScript) {
    const actionPlan = compactComposerActionPlan({
      kind: 'script',
      source: config.source,
      resourceType: config.resourceType,
      hasViralCandidate: Boolean(viralCandidate)
    });
    const actionDisabled = Boolean(batchFactoryStage) || !draftSource.trim();
    const directActionIDs = compact
      ? actionPlan.primary.filter(actionID => actionID === 'prompt-templates')
      : actionPlan.primary;
    const moreActionIDs = compact
      ? [...actionPlan.more, ...actionPlan.primary.filter(actionID => actionID !== 'prompt-templates')]
      : actionPlan.more;
    const actionFor = actionID => {
      if (actionID === 'prompt-templates') return {
        id: actionID,
        content: <AgentBatchFactoryPromptSelector config={config} disabled={Boolean(batchFactoryStage)} onSaved={onBatchFactoryPromptSaved} display="popover" />
      };
      if (actionID === 'extract-assets') return {
        id: actionID,
        label: '提取图片资产',
        disabled: actionDisabled,
        onClick: () => runBatchFactoryStage('assets')
      };
      if (actionID === 'generate-storyboard') return {
        id: actionID,
        label: '生成视频分镜',
        disabled: actionDisabled,
        onClick: () => runBatchFactoryStage('director')
      };
      if (actionID === 'generate-viral') return {
        id: actionID,
        label: '生成爆款候选',
        disabled: actionDisabled,
        onClick: () => runBatchFactoryStage('viral')
      };
      if (actionID === 'save-line-count') return {
        id: actionID,
        label: '保存行数',
        disabled: actionDisabled,
        onClick: () => runBatchFactoryStage('settings')
      };
      if (actionID === 'adopt-viral-candidate') return {
        id: actionID,
        label: '采用候选',
        disabled: Boolean(batchFactoryStage),
        onClick: adoptViralCandidate
      };
      return null;
    };
    const primaryActions = directActionIDs.map(actionFor).filter(Boolean);
    const moreActions = moreActionIDs.map(actionFor).filter(Boolean);

    return <section className={`agent-node-composer agent-script-node-composer${compact ? ' agent-script-node-composer--compact' : ''} nodrag nopan`} aria-label="剧本工作流" onMouseDown={stop} onDoubleClick={stop}>
      <AgentNodeComposerTopActions primaryActions={primaryActions} moreActions={moreActions} />
      <div className="agent-node-composer__body">
        <Input.TextArea
          aria-label={compact ? `本次生产正文（${batchFactoryRangeLines}行）` : '剧本正文'}
          className="agent-node-composer__prompt"
          value={compact ? batchFactoryPreview : draftSource}
          maxLength={60000}
          autoSize={{ minRows: compact ? 3 : 4, maxRows: compact ? 5 : 8 }}
          placeholder="粘贴剧本正文"
          readOnly={compact}
          onChange={compact ? undefined : changeSource}
          onBlur={compact ? undefined : () => {
            const save = onCommitSource?.(node.id, draftSource);
            if (save?.catch) void save.catch(() => {});
          }}
        />
        {viralCandidate ? <div className="agent-script-step__candidate">
          <Input.TextArea aria-label="爆款候选正文" value={viralCandidate} readOnly autoSize={{ minRows: 3, maxRows: 8 }} />
        </div> : null}
        {batchFactoryError ? <span className="agent-node-composer__state is-error" role="alert">{batchFactoryError}</span> : null}
      </div>
      <div className="agent-node-composer__bottom-bar agent-node-composer__toolbar">
        <AgentModelPicker
          kind="text"
          config={{ ...config, modelId: selectedTextModelID }}
          models={textModels}
          disabled={Boolean(batchFactoryStage)}
          onChange={patch => update({ modelId: patch.modelId, textModelId: patch.modelId })}
        />
        <label className="agent-script-node-composer__line-count">
          <span>生产行数</span>
          <InputNumber
            aria-label="生产正文行数"
            min={1}
            max={500}
            precision={0}
            value={batchFactoryRangeLines}
            disabled={Boolean(batchFactoryStage)}
            onChange={value => setBatchFactoryRangeLines(Math.min(500, Math.max(1, Number(value) || 5)))}
          />
        </label>
        <span className="agent-node-composer__state" aria-live="polite">{batchFactoryStage ? '正在提交生产步骤…' : sourceStatus}</span>
      </div>
    </section>;
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
        onBlur={() => {
          const save = onCommitSource?.(node.id, draftSource);
          if (save?.catch) void save.catch(() => {});
        }}
      />
      <div className="agent-script-step__controls">
        <AgentModelPicker
          kind="text"
          config={{ ...config, modelId: selectedTextModelID }}
          models={textModels}
          disabled={splitBusy}
          onChange={patch => update(isBatchFactoryScript
            ? { modelId: patch.modelId, textModelId: patch.modelId }
            : { splitModelId: patch.modelId })}
        />
        <span className={`agent-node-composer__state${!isBatchFactoryScript && stateName === 'failed' ? ' is-error' : ''}`} aria-live="polite">
          {sourceStatus}
        </span>
        {!isBatchFactoryScript ? <Button
          type="primary"
          size="small"
          icon={<Scissors size={14} />}
          loading={splitBusy}
          disabled={!draftSource.trim() || !splitModel.modelId || !splitModel.available}
          onClick={splitScript}
        >拆分镜头</Button> : null}
      </div>
    </section>

    {isBatchFactoryScript ? <section className="agent-script-step" aria-label="批量工厂生产">
      <div className="agent-script-step__heading">
        <h4><span>批</span>批量工厂</h4>
        <span>结果会同步到本书的图片和视频卡</span>
      </div>
      <div className="agent-script-step__generate-controls agent-script-step__batch-factory-controls">
        <label>
          <span>生产正文行数</span>
          <InputNumber
            aria-label="生产正文行数"
            min={1}
            max={500}
            precision={0}
            value={batchFactoryRangeLines}
            disabled={Boolean(batchFactoryStage)}
            onChange={value => setBatchFactoryRangeLines(Math.min(500, Math.max(1, Number(value) || 5)))}
          />
        </label>
        <Button
          size="small"
          loading={batchFactoryStage === 'settings'}
          disabled={Boolean(batchFactoryStage) || !draftSource.trim()}
          onClick={() => runBatchFactoryStage('settings')}
        >保存行数</Button>
        <Button
          type="primary"
          size="small"
          loading={batchFactoryStage === 'assets'}
          disabled={Boolean(batchFactoryStage) || !draftSource.trim()}
          onClick={() => runBatchFactoryStage('assets')}
        >提取图片资产</Button>
        <Button
          size="small"
          loading={batchFactoryStage === 'director'}
          disabled={Boolean(batchFactoryStage) || !draftSource.trim()}
          onClick={() => runBatchFactoryStage('director')}
        >生成视频分镜</Button>
        <Button
          size="small"
          loading={batchFactoryStage === 'viral'}
          disabled={Boolean(batchFactoryStage) || !draftSource.trim()}
          onClick={() => runBatchFactoryStage('viral')}
        >生成爆款候选</Button>
      </div>
      {viralCandidate ? <div className="agent-script-step__candidate">
        <Input.TextArea aria-label="爆款候选正文" value={viralCandidate} readOnly autoSize={{ minRows: 3, maxRows: 8 }} />
        <Button size="small" type="primary" loading={batchFactoryStage === 'adopt-viral'} disabled={Boolean(batchFactoryStage)} onClick={adoptViralCandidate}>采用候选</Button>
      </div> : null}
      {batchFactoryError ? <span className="agent-node-composer__state is-error" role="alert">{batchFactoryError}</span> : null}
      <AgentBatchFactoryPromptSelector config={config} disabled={Boolean(batchFactoryStage)} onSaved={onBatchFactoryPromptSaved} />
    </section> : null}

    {!isBatchFactoryScript ? <section className="agent-script-step" aria-labelledby="agent-script-step-shots">
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
    </section> : null}

    {!isBatchFactoryScript ? <section className="agent-script-step" aria-labelledby="agent-script-step-generate">
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
    </section> : null}
  </section>;
}
