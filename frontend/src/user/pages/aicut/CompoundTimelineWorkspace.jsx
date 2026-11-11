import { Alert, Button, Checkbox, InputNumber, Modal, Select } from 'antd';
import { ArrowDown, ArrowLeft, ArrowUp, GripVertical, Layers3, Plus, Scissors, Trash2 } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { executeAICutCommand } from '../aicutCommands.js';
import { buildCompoundTimelineView, compoundBatchTransferTargets, compoundBlockTransferCommand, compoundBlockTransferTargets, compoundMemberMoveCommand, compoundMemberResizeCommand, compoundMemberTrackMoveCommand, compoundMemberTransferCommand, compoundMembersTransferCommand, compoundTransferTargets, moveCompoundTrackByOffset, reorderCompoundTrackIds } from '../aicutCompoundTimeline.js';

const compoundTrackDragType = 'application/x-aicut-compound-track';
const compoundMemberDragType = 'application/x-aicut-compound-member';
const compoundMembersDragType = 'application/x-aicut-compound-members';
const compoundBlockDragType = 'application/x-aicut-compound-block';

function trackKindLabel(kind) {
  return kind === 'video' ? '画面' : kind === 'audio' ? '音频' : kind === 'subtitle' ? '字幕' : '特效';
}

function timelineMember(timeline, memberKey) {
  const [trackId, clipId] = String(memberKey || '').split('::');
  const track = (timeline?.tracks || []).find(item => item.id === trackId);
  const clip = track?.clips?.find(item => item.id === clipId);
  return track && clip ? { track, clip } : null;
}

function MemberEditor({ view, track, clip, onCommand, selected, selectedBatchKeys, onToggleSelection, onMemberDragStart, onBatchDragStart, onMemberDragEnd }) {
  const memberKey = `${track.id}::${clip.id}`;
  const [startOffsetMs, setStartOffsetMs] = useState(Number(clip.startMs) || 0);
  const [durationMs, setDurationMs] = useState(Number(clip.durationMs) || 1);
  const startOffsetRef = useRef(startOffsetMs);
  const durationRef = useRef(durationMs);

  useEffect(() => {
    const nextStart = Number(clip.startMs) || 0;
    const nextDuration = Number(clip.durationMs) || 1;
    startOffsetRef.current = nextStart;
    durationRef.current = nextDuration;
    setStartOffsetMs(nextStart);
    setDurationMs(nextDuration);
  }, [clip.durationMs, clip.startMs]);

  function commitMove() {
    onCommand(compoundMemberMoveCommand(view, memberKey, startOffsetRef.current));
  }

  function commitDuration() {
    onCommand(compoundMemberResizeCommand(view, memberKey, startOffsetRef.current, durationRef.current));
  }

  return <div className={`aicut-compound-member kind-${track.kind}`}>
    <div className="aicut-compound-member-fields">
      <span className="aicut-compound-member-title">
        <button
          type="button"
          className="aicut-compound-member-grip"
          draggable={!track.locked}
          disabled={track.locked}
          aria-label={`拖动${clip.name || clip.id}到其他兼容轨道`}
          title={track.locked ? '轨道已锁定' : '拖动到其他兼容轨道'}
          onDragStart={(event) => {
            event.dataTransfer.effectAllowed = 'move';
            if (selected && selectedBatchKeys.length > 1) {
              event.dataTransfer.setData(compoundMembersDragType, JSON.stringify(selectedBatchKeys));
              onBatchDragStart(selectedBatchKeys);
            } else {
              event.dataTransfer.setData(compoundMemberDragType, memberKey);
              onMemberDragStart(memberKey);
            }
          }}
          onDragEnd={onMemberDragEnd}
        ><GripVertical size={14} /></button>
        <Checkbox checked={selected} aria-label={`选择${clip.name || clip.id}用于批量迁移`} onChange={() => onToggleSelection(memberKey)} />
        <strong>{clip.name || clip.id}</strong>
      </span>
      <label>组内开始<InputNumber min={0} value={startOffsetMs} onChange={value => { const next = Number(value) || 0; startOffsetRef.current = next; setStartOffsetMs(next); }} onBlur={commitMove} addonAfter="ms" /></label>
      <label>片段时长<InputNumber min={1} value={durationMs} onChange={value => { const next = Math.max(1, Number(value) || 1); durationRef.current = next; setDurationMs(next); }} onBlur={commitDuration} addonAfter="ms" /></label>
    </div>
    <div className="aicut-compound-member-rail" aria-label={`${clip.name || clip.id} 组内时间位置`}>
      <i style={{ left: `${startOffsetMs / view.durationMs * 100}%`, width: `${Math.max(2, durationMs / view.durationMs * 100)}%` }} />
    </div>
  </div>;
}

function NestedCompoundEditor({ view, clip, onOpenCompound, onBlockDragStart, onBlockDragEnd }) {
  return <div className="aicut-compound-member nested">
    <div className="aicut-compound-member-fields">
      <span className="aicut-compound-member-title">
        <button
          type="button"
          className="aicut-compound-member-grip"
          draggable
          aria-label={`拖动${clip.name || '子复合'}复合块到其他复合片段`}
          title="拖动整个复合块到其他复合片段"
          onDragStart={(event) => {
            event.dataTransfer.effectAllowed = 'move';
            event.dataTransfer.setData(compoundBlockDragType, clip.compoundId);
            onBlockDragStart(clip.compoundId);
          }}
          onDragEnd={onBlockDragEnd}
        ><GripVertical size={14} /></button>
        <strong>{clip.name || '子复合片段'}</strong>
      </span>
      <span>{clip.memberCount} 个成员 · {(Number(clip.durationMs) / 1000).toFixed(2)} 秒</span>
      <Button size="small" icon={<Layers3 size={13} />} onClick={() => onOpenCompound(clip.compoundId)}>进入子复合片段</Button>
    </div>
    <div className="aicut-compound-member-rail" aria-label={`${clip.name || '子复合片段'} 组内时间位置`}>
      <i style={{ left: `${Number(clip.startMs) / view.durationMs * 100}%`, width: `${Math.max(6, Number(clip.durationMs) / view.durationMs * 100)}%` }} />
    </div>
  </div>;
}

export default function CompoundTimelineWorkspace({ timeline, compoundId, playheadMs, selectedClipKeys = [], onToggleSelection, onCommand, onClose, onDissolve, onSplit, onOpenCompound, onAddTrack, onRemoveTrack, onReorderTrack, onDropAsset }) {
  const view = useMemo(() => buildCompoundTimelineView(timeline, compoundId), [compoundId, timeline]);
  const [draggedTrackId, setDraggedTrackId] = useState('');
  const [draggedMemberKey, setDraggedMemberKey] = useState('');
  const [draggedMemberKeys, setDraggedMemberKeys] = useState([]);
  const [draggedBlockId, setDraggedBlockId] = useState('');
  const [batchTransferDraft, setBatchTransferDraft] = useState(null);
  const transferTargets = useMemo(() => draggedMemberKey ? compoundTransferTargets(timeline, view.id, draggedMemberKey) : [], [draggedMemberKey, timeline, view.id]);
  const blockTransferTargets = useMemo(() => draggedBlockId ? compoundBlockTransferTargets(timeline, draggedBlockId) : [], [draggedBlockId, timeline]);
  const batchTransferTargets = useMemo(() => draggedMemberKeys.length > 1 ? compoundBatchTransferTargets(timeline, view.id, draggedMemberKeys) : [], [draggedMemberKeys, timeline, view.id]);
  const visibleMemberKeys = useMemo(() => view.tracks.flatMap(track => track.clips.filter(clip => !clip.isCompoundContainer).map(clip => `${track.id}::${clip.id}`)), [view]);
  const selectedBatchKeys = selectedClipKeys.filter(key => visibleMemberKeys.includes(key));
  const batchTransferPreview = useMemo(() => {
    if (!batchTransferDraft) return null;
    try {
      const command = compoundMembersTransferCommand(
        view,
        batchTransferDraft.memberKeys,
        batchTransferDraft.targetCompoundId,
        batchTransferDraft.targetTrackByMember
      );
      return { command, result: executeAICutCommand(timeline, command), error: '' };
    } catch (error) {
      return { command: null, result: null, error: error?.message || '无法生成轨道映射预览' };
    }
  }, [batchTransferDraft, timeline, view]);
  const defaultSplitTime = Number(playheadMs) > view.originMs && Number(playheadMs) < view.originMs + view.durationMs
    ? Math.round(Number(playheadMs))
    : view.originMs + Math.round(view.durationMs / 2);
  const [splitTimeMs, setSplitTimeMs] = useState(defaultSplitTime);
  const splitTimeRef = useRef(defaultSplitTime);
  useEffect(() => {
    splitTimeRef.current = defaultSplitTime;
    setSplitTimeMs(defaultSplitTime);
  }, [defaultSplitTime, view.id]);
  const canSplitAtPlayhead = splitTimeMs > view.originMs && splitTimeMs < view.originMs + view.durationMs;
  return <section className="aicut-compound-workspace" aria-label="复合片段内部时间线">
    <header>
      <Button type="text" icon={<ArrowLeft size={16} />} onClick={() => onClose(view.parentCompoundId)}>{view.parentCompoundId ? '返回上层复合' : '返回主时间线'}</Button>
      <div><Layers3 size={18} /><strong>{view.name}</strong><small>{view.memberKeys.length} 个成员 · {(view.durationMs / 1000).toFixed(2)} 秒</small></div>
      <Button size="small" icon={<Plus size={13} />} onClick={() => onAddTrack(view.id, 'video')}>新增画面轨</Button>
      <Button size="small" icon={<Plus size={13} />} onClick={() => onAddTrack(view.id, 'audio')}>新增音频轨</Button>
      <Button size="small" icon={<Plus size={13} />} onClick={() => onAddTrack(view.id, 'subtitle')}>新增字幕轨</Button>
      <label>分割位置<InputNumber aria-label="复合片段分割位置" min={view.originMs + 1} max={view.originMs + view.durationMs - 1} value={splitTimeMs} addonAfter="ms" onChange={value => { const next = Math.round(Number(value) || 0); splitTimeRef.current = next; setSplitTimeMs(next); }} /></label>
      <Button icon={<Scissors size={14} />} title="默认在播放头分割；也可输入精确位置" disabled={!canSplitAtPlayhead} onClick={() => onSplit(view.id, splitTimeRef.current)}>执行分割</Button>
      <Button danger onClick={() => onDissolve(view.id)}>拆散复合片段</Button>
    </header>
    <div className="aicut-compound-ruler">组内时间：0.00 秒 — {(view.durationMs / 1000).toFixed(2)} 秒</div>
    {draggedMemberKey || draggedMemberKeys.length || draggedBlockId ? <div className="aicut-compound-transfer-dock">
      <strong>{draggedMemberKeys.length > 1 ? `已选择 ${draggedMemberKeys.length} 个成员，拖到目标复合片段` : '拖到其他复合片段'}</strong>
      {draggedMemberKey && transferTargets.length ? transferTargets.flatMap(target => target.tracks.map(track => <div
        key={`${target.compoundId}:${track.id}`}
        className="aicut-compound-transfer-target"
        role="button"
        tabIndex={0}
        onDragOver={(event) => {
          if (event.dataTransfer.types.includes(compoundMemberDragType)) event.preventDefault();
        }}
        onDrop={(event) => {
          const memberKey = event.dataTransfer.getData(compoundMemberDragType);
          if (!memberKey) return;
          event.preventDefault();
          onCommand(compoundMemberTransferCommand(view, memberKey, target.compoundId, track.id));
          setDraggedMemberKey('');
        }}
      >
        <span>{target.name}</span><small>{track.name} · {track.kind === 'video' ? '画面' : track.kind === 'audio' ? '音频' : '字幕'}</small>
      </div>)) : null}
      {draggedBlockId && blockTransferTargets.length ? blockTransferTargets.map(target => <div
        key={`block:${target.compoundId}`}
        className="aicut-compound-transfer-target"
        role="button"
        tabIndex={0}
        onDragOver={(event) => {
          if (event.dataTransfer.types.includes(compoundBlockDragType)) event.preventDefault();
        }}
        onDrop={(event) => {
          const sourceCompoundId = event.dataTransfer.getData(compoundBlockDragType);
          if (!sourceCompoundId) return;
          event.preventDefault();
          onCommand(compoundBlockTransferCommand(sourceCompoundId, target.compoundId));
          setDraggedBlockId('');
        }}
      ><span>{target.name}</span><small>迁移整个复合块</small></div>) : null}
      {draggedMemberKeys.length > 1 && batchTransferTargets.length ? batchTransferTargets.map(target => <div
        key={`batch:${target.compoundId}`}
        className="aicut-compound-transfer-target"
        role="button"
        tabIndex={0}
        onDragOver={(event) => {
          if (event.dataTransfer.types.includes(compoundMembersDragType)) event.preventDefault();
        }}
        onDrop={(event) => {
          const payload = event.dataTransfer.getData(compoundMembersDragType);
          if (!payload) return;
          event.preventDefault();
          let memberKeys = [];
          try { memberKeys = JSON.parse(payload); } catch { return; }
          if (!Array.isArray(memberKeys) || memberKeys.length < 2) return;
          setBatchTransferDraft({
            targetCompoundId: target.compoundId,
            targetName: target.name,
            memberKeys,
            targetTrackByMember: {}
          });
          setDraggedMemberKeys([]);
        }}
      ><span>{target.name}</span><small>批量迁移 {draggedMemberKeys.length} 个成员 · 自动匹配{target.trackKinds.map(kind => kind === 'video' ? '画面' : kind === 'audio' ? '音频' : kind === 'subtitle' ? '字幕' : '特效').join('/')}轨</small></div>) : null}
      {(draggedMemberKey && !transferTargets.length) || (draggedBlockId && !blockTransferTargets.length) || (draggedMemberKeys.length > 1 && !batchTransferTargets.length) ? <span className="aicut-compound-transfer-empty">没有兼容的目标复合片段</span> : null}
    </div> : null}
    <Modal
      open={Boolean(batchTransferDraft)}
      title="批量迁移 · 轨道映射预览"
      okText="确认迁移"
      cancelText="取消"
      width={720}
      maskClosable={false}
      okButtonProps={{ disabled: !batchTransferPreview?.result }}
      onCancel={() => setBatchTransferDraft(null)}
      onOk={() => {
        if (!batchTransferPreview?.command) return;
        onCommand(batchTransferPreview.command);
        setBatchTransferDraft(null);
      }}
    >
      <p className="aicut-compound-mapping-summary">目标：<strong>{batchTransferDraft?.targetName}</strong>。自动匹配会避开锁定、类型不符和时间冲突；你也可以逐项指定轨道。</p>
      {batchTransferPreview?.error ? <Alert type="error" showIcon message={batchTransferPreview.error} /> : null}
      <div className="aicut-compound-mapping-list">
        {(batchTransferDraft?.memberKeys || []).map(memberKey => {
          const member = timelineMember(timeline, memberKey);
          const available = member ? compoundTransferTargets(timeline, view.id, memberKey).find(target => target.compoundId === batchTransferDraft.targetCompoundId)?.tracks || [] : [];
          const mappedTrackId = batchTransferPreview?.result?.metadata?.targetTrackByMember?.[memberKey] || '';
          const mappedTrack = batchTransferPreview?.result?.timeline?.tracks?.find(track => track.id === mappedTrackId);
          const created = batchTransferPreview?.result?.metadata?.createdTrackIds?.includes(mappedTrackId);
          return <div className="aicut-compound-mapping-row" key={memberKey}>
            <span><strong>{member?.clip?.name || member?.clip?.id || memberKey}</strong><small>{trackKindLabel(member?.track?.kind)}</small></span>
            <Select
              aria-label={`${member?.clip?.name || memberKey}目标轨道`}
              value={batchTransferDraft?.targetTrackByMember?.[memberKey] || '__auto__'}
              options={[
                { value: '__auto__', label: '自动匹配' },
                ...available.map(track => ({ value: track.id, label: `${track.name} · ${trackKindLabel(track.kind)}` })),
                { value: '__new__', label: `新建${trackKindLabel(member?.track?.kind)}专用轨道` }
              ]}
              onChange={value => setBatchTransferDraft(current => ({
                ...current,
                targetTrackByMember: { ...(current?.targetTrackByMember || {}), [memberKey]: value }
              }))}
            />
            <span className="aicut-compound-mapping-result">将进入：{mappedTrack?.name || '等待有效选择'}{created ? '（新建）' : ''}</span>
          </div>;
        })}
      </div>
    </Modal>
    <div className="aicut-compound-tracks">
      {view.tracks.map((track, trackIndex) => <section
        className={`aicut-compound-track${track.settings?.compoundId === view.id ? ' internal' : ''}${draggedTrackId && draggedTrackId !== track.id ? ' reorder-target' : ''}${draggedMemberKey && !draggedMemberKey.startsWith(`${track.id}::`) ? ' member-target' : ''}`}
        key={track.id}
        onDragOver={(event) => {
          if (!track.locked && (event.dataTransfer.types.includes(compoundTrackDragType) || event.dataTransfer.types.includes(compoundMemberDragType) || event.dataTransfer.types.includes('application/x-aicut-asset'))) event.preventDefault();
        }}
        onDrop={(event) => {
          const sourceTrackId = event.dataTransfer.getData(compoundTrackDragType);
          if (sourceTrackId) {
            event.preventDefault();
            setDraggedTrackId('');
            if (sourceTrackId !== track.id) onReorderTrack(view.id, reorderCompoundTrackIds(view, sourceTrackId, track.id));
            return;
          }
          const memberKey = event.dataTransfer.getData(compoundMemberDragType);
          if (memberKey) {
            event.preventDefault();
            setDraggedMemberKey('');
            onCommand(compoundMemberTrackMoveCommand(view, memberKey, track.id));
            return;
          }
          onDropAsset(event, track, view);
        }}
      >
        <h3>
          <span>{track.name || track.id}<small>{track.kind === 'video' ? '画面' : track.kind === 'audio' ? '音频' : '字幕'}</small></span>
          <span className="aicut-compound-track-actions">
            <button
              type="button"
              className="aicut-compound-track-grip"
              draggable={!track.locked}
              disabled={track.locked}
              aria-label={`拖动${track.name || track.id}调整顺序`}
              title={track.locked ? '轨道已锁定' : '拖动调整轨道顺序'}
              onDragStart={(event) => {
                event.dataTransfer.effectAllowed = 'move';
                event.dataTransfer.setData(compoundTrackDragType, track.id);
                setDraggedTrackId(track.id);
              }}
              onDragEnd={() => setDraggedTrackId('')}
            ><GripVertical size={14} /></button>
            <Button size="small" type="text" aria-label="上移轨道" title="上移轨道" disabled={track.locked || trackIndex === 0} icon={<ArrowUp size={13} />} onClick={() => onReorderTrack(view.id, moveCompoundTrackByOffset(view, track.id, -1))} />
            <Button size="small" type="text" aria-label="下移轨道" title="下移轨道" disabled={track.locked || trackIndex === view.tracks.length - 1} icon={<ArrowDown size={13} />} onClick={() => onReorderTrack(view.id, moveCompoundTrackByOffset(view, track.id, 1))} />
            {track.settings?.compoundId === view.id && !track.clips.length ? <Button size="small" danger icon={<Trash2 size={12} />} onClick={() => onRemoveTrack(view.id, track.id)}>删除空轨道</Button> : null}
          </span>
        </h3>
        {!track.clips.length ? <div className="aicut-compound-empty-track">从左侧素材库拖入匹配素材</div> : null}
        {track.clips.map(clip => clip.isCompoundContainer
          ? <NestedCompoundEditor key={clip.id} view={view} clip={clip} onOpenCompound={onOpenCompound} onBlockDragStart={setDraggedBlockId} onBlockDragEnd={() => setDraggedBlockId('')} />
          : <MemberEditor
            key={clip.id}
            view={view}
            track={track}
            clip={clip}
            onCommand={onCommand}
            selected={selectedClipKeys.includes(`${track.id}::${clip.id}`)}
            selectedBatchKeys={selectedBatchKeys}
            onToggleSelection={onToggleSelection}
            onMemberDragStart={(key) => { setDraggedMemberKeys([]); setDraggedMemberKey(key); }}
            onBatchDragStart={(keys) => { setDraggedMemberKey(''); setDraggedMemberKeys(keys); }}
            onMemberDragEnd={() => { setDraggedMemberKey(''); setDraggedMemberKeys([]); }}
          />)}
      </section>)}
    </div>
    <footer>内部编辑直接修改同一份工程时间线；保存、预览和导出不会生成另一份假数据。{selectedBatchKeys.length > 1 ? ` 已选择 ${selectedBatchKeys.length} 个成员，拖动任一已选成员可批量迁移。` : ''}</footer>
  </section>;
}
