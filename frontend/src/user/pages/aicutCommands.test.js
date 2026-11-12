import test from 'node:test';
import assert from 'node:assert/strict';
import { createAICutClipboard, executeAICutCommand, findTimelineEditPoint, reconcileTimelineSave, restoreTimelineSnapshot } from './aicutCommands.js';
import { buildAICutPreviewFrame } from './aicutPreview.js';

const source = {
  revision: 7,
  tracks: [
    { id: 'video', kind: 'video', locked: false, clips: [
      { id: 'a', kind: 'video', name: 'A', startMs: 0, durationMs: 2000, sourceStartMs: 0, sourceDurationMs: 4000 },
      { id: 'b', kind: 'video', name: 'B', startMs: 4000, durationMs: 1000, sourceStartMs: 0, sourceDurationMs: 1000 }
    ] },
    { id: 'audio', kind: 'audio', locked: false, clips: [
      { id: 'voice', kind: 'audio', name: '声音', startMs: 0, durationMs: 2000, sourceStartMs: 0, sourceDurationMs: 2000 }
    ] }
  ]
};

test('统一命令层执行复制、删除、分割、移动和裁剪并返回统一结果', () => {
  const copied = executeAICutCommand(source, { type: 'clip.duplicate', trackId: 'video', clipId: 'a', newId: 'copy-a' });
  assert.equal(copied.changed, true);
  assert.equal(copied.selectionKey, 'video::copy-a');
  assert.equal(copied.timeline.tracks[0].clips.find(item => item.id === 'copy-a').startMs, 2000);

  const removed = executeAICutCommand(copied.timeline, { type: 'clip.delete', trackId: 'video', clipId: 'copy-a' });
  assert.equal(removed.timeline.tracks[0].clips.some(item => item.id === 'copy-a'), false);

  let next = 0;
  const split = executeAICutCommand(source, { type: 'clip.split_at_playhead', timeMs: 1000 }, { idFactory: () => `split-${++next}` });
  assert.equal(split.metadata.splitCount, 2);
  assert.deepEqual(split.timeline.tracks.map(track => track.clips.length), [3, 2]);

  const moved = executeAICutCommand(source, { type: 'clip.move', trackId: 'video', clipId: 'a', startMs: 1950, magnetic: true, thresholdMs: 100 });
  assert.equal(moved.timeline.tracks[0].clips[0].startMs, 2000);

  const resized = executeAICutCommand(source, { type: 'clip.resize', trackId: 'video', clipId: 'a', edge: 'right', edgeTimeMs: 1500 });
  assert.equal(resized.timeline.tracks[0].clips[0].durationMs, 1500);
  assert.equal(source.tracks[0].clips[0].durationMs, 2000);
});

test('字幕合并经过统一命令层并保持选中合并后的片段', () => {
  const timeline = {
    tracks: [{ id: 'subtitle', kind: 'subtitle', locked: false, clips: [
      { id: 'first', kind: 'subtitle', text: '第一句', startMs: 0, durationMs: 800 },
      { id: 'second', kind: 'subtitle', text: '第二句', startMs: 900, durationMs: 700 }
    ] }]
  };
  const merged = executeAICutCommand(timeline, { type: 'clip.merge_next', trackId: 'subtitle', clipId: 'first' });
  assert.equal(merged.selectionKey, 'subtitle::first');
  assert.equal(merged.timeline.tracks[0].clips.length, 1);
  assert.equal(merged.timeline.tracks[0].clips[0].text, '第一句 第二句');
  assert.equal(merged.timeline.tracks[0].clips[0].durationMs, 1600);
  assert.equal(timeline.tracks[0].clips.length, 2);
  assert.throws(() => executeAICutCommand(merged.timeline, { type: 'clip.merge_next', trackId: 'subtitle', clipId: 'first' }), /后面没有可合并的字幕/);
});

test('联动命令建立持久分组，移动时整组平移，取消联动后只移动目标', () => {
  const linked = executeAICutCommand(source, {
    type: 'clip.link',
    clipKeys: ['video::a', 'audio::voice'],
    groupId: 'link-1'
  });
  assert.equal(linked.timeline.tracks[0].clips[0].linkGroupId, 'link-1');
  assert.equal(linked.timeline.tracks[1].clips[0].linkGroupId, 'link-1');

  const moved = executeAICutCommand(linked.timeline, {
    type: 'clip.move', trackId: 'video', clipId: 'a', startMs: 1000, linked: true, magnetic: false
  });
  assert.equal(moved.timeline.tracks[0].clips[0].startMs, 1000);
  assert.equal(moved.timeline.tracks[1].clips[0].startMs, 1000);

  const unlinked = executeAICutCommand(moved.timeline, { type: 'clip.unlink', clipKeys: ['video::a'] });
  assert.equal(unlinked.timeline.tracks[0].clips[0].linkGroupId, undefined);
  assert.equal(unlinked.timeline.tracks[1].clips[0].linkGroupId, undefined);
});

test('统一命令层拒绝空命令、未知命令、锁定轨道和不足两个片段的联动', () => {
  assert.throws(() => executeAICutCommand(source), /缺少剪辑命令/);
  assert.throws(() => executeAICutCommand(source, { type: 'unknown' }), /不支持的剪辑命令/);
  assert.throws(() => executeAICutCommand(source, { type: 'clip.link', clipKeys: ['video::a'] }), /至少选择两个片段/);
  const locked = { ...source, tracks: source.tracks.map(track => track.id === 'video' ? { ...track, locked: true } : track) };
  assert.throws(() => executeAICutCommand(locked, { type: 'clip.delete', trackId: 'video', clipId: 'a' }), /轨道已锁定/);
});

test('取消命令和重复联动是幂等操作，不制造撤销记录', () => {
  const canceled = executeAICutCommand(source, { type: 'edit.cancel' });
  assert.equal(canceled.changed, false);
  assert.equal(canceled.timeline, source);
  const linked = executeAICutCommand(source, { type: 'clip.link', clipKeys: ['video::a', 'audio::voice'], groupId: 'same' });
  const repeated = executeAICutCommand(linked.timeline, { type: 'clip.link', clipKeys: ['video::a', 'audio::voice'], groupId: 'same' });
  assert.equal(repeated.changed, false);
});

test('复合片段建立持久成员关系，选择一个成员即可展开或拆散整组', () => {
  const created = executeAICutCommand(source, {
    type: 'clip.compound_create',
    clipKeys: ['video::a', 'audio::voice'],
    compoundId: 'compound-1',
    name: '开场复合片段'
  });
  assert.equal(created.timeline.tracks[0].clips[0].compoundId, 'compound-1');
  assert.equal(created.timeline.tracks[1].clips[0].compoundId, 'compound-1');
  assert.equal(created.timeline.tracks[0].clips[0].compoundName, '开场复合片段');
  assert.deepEqual(created.metadata.memberKeys, ['video::a', 'audio::voice']);

  const repeated = executeAICutCommand(created.timeline, {
    type: 'clip.compound_create', clipKeys: ['video::a', 'audio::voice'], compoundId: 'compound-1', name: '开场复合片段'
  });
  assert.equal(repeated.changed, false);

  const dissolved = executeAICutCommand(created.timeline, { type: 'clip.compound_dissolve', clipKeys: ['video::a'] });
  assert.equal(dissolved.timeline.tracks[0].clips[0].compoundId, undefined);
  assert.equal(dissolved.timeline.tracks[1].clips[0].compoundId, undefined);
  assert.deepEqual(dissolved.metadata.memberKeys, ['video::a', 'audio::voice']);
});

test('复合片段整体移动保持成员相对时间并拒绝越界、锁轨和轨道重叠', () => {
  const created = executeAICutCommand(source, {
    type: 'clip.compound_create', clipKeys: ['video::a', 'audio::voice'], compoundId: 'compound-1', name: '开场复合片段'
  }).timeline;
  const moved = executeAICutCommand(created, { type: 'clip.compound_move', compoundId: 'compound-1', startMs: 1000 });
  assert.equal(moved.timeline.tracks[0].clips.find(clip => clip.id === 'a').startMs, 1000);
  assert.equal(moved.timeline.tracks[1].clips.find(clip => clip.id === 'voice').startMs, 1000);
  assert.deepEqual(moved.metadata.memberKeys, ['video::a', 'audio::voice']);
  assert.equal(created.tracks[0].clips.find(clip => clip.id === 'a').startMs, 0);

  const repeated = executeAICutCommand(moved.timeline, { type: 'clip.compound_move', compoundId: 'compound-1', startMs: 1000 });
  assert.equal(repeated.changed, false);
  assert.throws(() => executeAICutCommand(created, { type: 'clip.compound_move', compoundId: 'compound-1', startMs: -1 }), /不能早于时间线零点/);
  assert.throws(() => executeAICutCommand(created, { type: 'clip.compound_move', compoundId: 'compound-1', startMs: 2500 }), /与现有片段重叠/);
  const locked = { ...created, tracks: created.tracks.map(track => track.id === 'audio' ? { ...track, locked: true } : track) };
  assert.throws(() => executeAICutCommand(locked, { type: 'clip.compound_move', compoundId: 'compound-1', startMs: 1000 }), /轨道已锁定/);
});

test('复合片段整体裁剪会裁短跨边界成员、移除窗外成员并保持源时间线不变', () => {
  const compound = {
    revision: 3,
    tracks: [
      { id: 'video', kind: 'video', locked: false, clips: [
        { id: 'lead', kind: 'video', startMs: 1000, durationMs: 2000, sourceStartMs: 0, sourceDurationMs: 4000, compoundId: 'compound-trim' },
        { id: 'late', kind: 'video', startMs: 3500, durationMs: 1500, sourceStartMs: 0, sourceDurationMs: 1500, compoundId: 'compound-trim' },
        { id: 'outside', kind: 'video', startMs: 6000, durationMs: 1000, sourceStartMs: 0, sourceDurationMs: 1000 }
      ] },
      { id: 'audio', kind: 'audio', locked: false, clips: [
        { id: 'bed', kind: 'audio', startMs: 500, durationMs: 5000, sourceStartMs: 100, sourceDurationMs: 6000, compoundId: 'compound-trim' }
      ] }
    ]
  };

  const right = executeAICutCommand(compound, { type: 'clip.compound_trim', compoundId: 'compound-trim', edge: 'right', edgeTimeMs: 4000 });
  assert.equal(right.timeline.tracks[0].clips.find(clip => clip.id === 'lead').durationMs, 2000);
  assert.equal(right.timeline.tracks[0].clips.find(clip => clip.id === 'late').durationMs, 500);
  assert.equal(right.timeline.tracks[1].clips.find(clip => clip.id === 'bed').durationMs, 3500);
  assert.equal(right.timeline.tracks[0].clips.find(clip => clip.id === 'outside').durationMs, 1000);
  assert.deepEqual(right.metadata.trimmedMemberKeys.sort(), ['audio::bed', 'video::late']);

  const left = executeAICutCommand(compound, { type: 'clip.compound_trim', compoundId: 'compound-trim', edge: 'left', edgeTimeMs: 1500 });
  const lead = left.timeline.tracks[0].clips.find(clip => clip.id === 'lead');
  const bed = left.timeline.tracks[1].clips.find(clip => clip.id === 'bed');
  assert.deepEqual({ startMs: lead.startMs, durationMs: lead.durationMs, sourceStartMs: lead.sourceStartMs }, { startMs: 1500, durationMs: 1500, sourceStartMs: 500 });
  assert.deepEqual({ startMs: bed.startMs, durationMs: bed.durationMs, sourceStartMs: bed.sourceStartMs }, { startMs: 1500, durationMs: 4000, sourceStartMs: 1100 });

  const removed = executeAICutCommand(compound, { type: 'clip.compound_trim', compoundId: 'compound-trim', edge: 'right', edgeTimeMs: 3200 });
  assert.equal(removed.timeline.tracks[0].clips.some(clip => clip.id === 'late'), false);
  assert.deepEqual(removed.metadata.removedMemberKeys, ['video::late']);
  assert.equal(compound.tracks[0].clips.find(clip => clip.id === 'late').durationMs, 1500);
});

test('复合片段整体裁剪拒绝无效边界、锁轨和裁到不足两个成员', () => {
  const compound = {
    tracks: [
      { id: 'video', kind: 'video', locked: false, clips: [
        { id: 'lead', startMs: 1000, durationMs: 2000, sourceStartMs: 0, compoundId: 'compound-1' }
      ] },
      { id: 'audio', kind: 'audio', locked: false, clips: [
        { id: 'bed', startMs: 500, durationMs: 5000, sourceStartMs: 0, compoundId: 'compound-1' }
      ] }
    ]
  };
  assert.throws(() => executeAICutCommand(compound, { type: 'clip.compound_trim', compoundId: 'compound-1', edge: 'middle', edgeTimeMs: 500 }), /不支持的裁剪方向/);
  assert.throws(() => executeAICutCommand(compound, { type: 'clip.compound_trim', compoundId: 'compound-1', edge: 'right', edgeTimeMs: -1 }), /裁剪时间无效/);
  assert.throws(() => executeAICutCommand(compound, { type: 'clip.compound_trim', compoundId: 'compound-1', edge: 'right', edgeTimeMs: 800 }), /至少保留两个成员/);
  const locked = { ...compound, tracks: compound.tracks.map(track => track.id === 'audio' ? { ...track, locked: true } : track) };
  assert.throws(() => executeAICutCommand(locked, { type: 'clip.compound_trim', compoundId: 'compound-1', edge: 'right', edgeTimeMs: 1000 }), /轨道已锁定/);
  const unchanged = executeAICutCommand(compound, { type: 'clip.compound_trim', compoundId: 'compound-1', edge: 'right', edgeTimeMs: 5500 });
  assert.equal(unchanged.changed, false);
});

test('复合片段整体分割会切开跨界成员并生成左右两个可用复合组', () => {
  const compound = {
    revision: 4,
    tracks: [
      { id: 'video', kind: 'video', locked: false, clips: [
        { id: 'lead', kind: 'video', name: '主画面', startMs: 0, durationMs: 4000, sourceStartMs: 0, sourceDurationMs: 5000, compoundId: 'compound-split', compoundName: '开场组合' },
        { id: 'outro', kind: 'video', name: '尾镜头', startMs: 2500, durationMs: 1500, sourceStartMs: 0, sourceDurationMs: 1500, compoundId: 'compound-split', compoundName: '开场组合' }
      ] },
      { id: 'audio', kind: 'audio', locked: false, clips: [
        { id: 'bed', kind: 'audio', name: '配乐', startMs: 0, durationMs: 4000, sourceStartMs: 100, sourceDurationMs: 5000, compoundId: 'compound-split', compoundName: '开场组合' },
        { id: 'intro', kind: 'audio', name: '片头声', startMs: 0, durationMs: 1500, sourceStartMs: 0, sourceDurationMs: 1500, compoundId: 'compound-split', compoundName: '开场组合' }
      ] }
    ]
  };
  let next = 0;
  const result = executeAICutCommand(compound, {
    type: 'clip.compound_split', compoundId: 'compound-split', timeMs: 2000, rightCompoundId: 'compound-right'
  }, { idFactory: prefix => `${prefix}-${++next}` });

  const left = result.timeline.tracks.flatMap(track => track.clips.map(clip => ({ track, clip }))).filter(({ clip }) => clip.compoundId === 'compound-split');
  const right = result.timeline.tracks.flatMap(track => track.clips.map(clip => ({ track, clip }))).filter(({ clip }) => clip.compoundId === 'compound-right');
  assert.equal(left.length, 3);
  assert.equal(right.length, 3);
  assert.equal(right.every(({ clip }) => clip.compoundName === '开场组合（续）'), true);
  assert.equal(right.find(({ clip }) => clip.name.includes('主画面')).clip.sourceStartMs, 2000);
  assert.equal(right.find(({ clip }) => clip.name.includes('配乐')).clip.sourceStartMs, 2100);
  assert.equal(result.metadata.splitMemberCount, 2);
  assert.equal(result.metadata.leftCompoundId, 'compound-split');
  assert.equal(result.metadata.rightCompoundId, 'compound-right');
  assert.equal(compound.tracks[0].clips[0].durationMs, 4000);
});

test('复合片段整体分割拒绝贴边、锁轨和任一侧不足两个成员', () => {
  const compound = {
    tracks: [
      { id: 'video', kind: 'video', locked: false, clips: [
        { id: 'main', startMs: 0, durationMs: 3000, sourceStartMs: 0, compoundId: 'compound-1' }
      ] },
      { id: 'audio', kind: 'audio', locked: false, clips: [
        { id: 'intro', startMs: 0, durationMs: 1000, sourceStartMs: 0, compoundId: 'compound-1' }
      ] }
    ]
  };
  assert.throws(() => executeAICutCommand(compound, { type: 'clip.compound_split', compoundId: 'compound-1', timeMs: 0 }), /必须在复合片段内部/);
  assert.throws(() => executeAICutCommand(compound, { type: 'clip.compound_split', compoundId: 'compound-1', timeMs: 1500 }), /左右两侧都至少保留两个成员/);
  const locked = { ...compound, tracks: compound.tracks.map(track => track.id === 'audio' ? { ...track, locked: true } : track) };
  assert.throws(() => executeAICutCommand(locked, { type: 'clip.compound_split', compoundId: 'compound-1', timeMs: 500 }), /轨道已锁定/);
});

test('创建上层复合片段会保留完整子复合路径，拆散上层后恢复子复合', () => {
  const nestedSource = {
    tracks: [
      { id: 'video', kind: 'video', locked: false, clips: [
        { id: 'a', kind: 'video', startMs: 0, durationMs: 1000, compoundId: 'child', compoundName: '子组合', compoundPath: [{ id: 'child', name: '子组合' }] },
        { id: 'b', kind: 'video', startMs: 1000, durationMs: 1000, compoundId: 'child', compoundName: '子组合', compoundPath: [{ id: 'child', name: '子组合' }] },
        { id: 'c', kind: 'video', startMs: 2000, durationMs: 1000 }
      ] }
    ]
  };
  const created = executeAICutCommand(nestedSource, {
    type: 'clip.compound_create', clipKeys: ['video::a', 'video::b', 'video::c'], compoundId: 'parent', name: '上层组合'
  });
  const [a, b, c] = created.timeline.tracks[0].clips;
  assert.deepEqual(a.compoundPath, [{ id: 'child', name: '子组合' }, { id: 'parent', name: '上层组合' }]);
  assert.deepEqual(b.compoundPath, a.compoundPath);
  assert.deepEqual(c.compoundPath, [{ id: 'parent', name: '上层组合' }]);
  assert.equal(a.compoundId, 'parent');
  assert.equal(c.compoundName, '上层组合');
  assert.throws(() => executeAICutCommand(nestedSource, {
    type: 'clip.compound_create', clipKeys: ['video::a', 'video::c'], compoundId: 'bad-parent', name: '不完整组合'
  }), /必须完整选择子复合片段/);

  const dissolved = executeAICutCommand(created.timeline, { type: 'clip.compound_dissolve', clipKeys: ['video::a'] });
  assert.deepEqual(dissolved.timeline.tracks[0].clips[0].compoundPath, [{ id: 'child', name: '子组合' }]);
  assert.equal(dissolved.timeline.tracks[0].clips[0].compoundId, 'child');
  assert.equal(dissolved.timeline.tracks[0].clips[2].compoundId, undefined);

  const childDissolved = executeAICutCommand(created.timeline, {
    type: 'clip.compound_dissolve', compoundId: 'child', clipKeys: ['video::a']
  });
  assert.deepEqual(childDissolved.timeline.tracks[0].clips[0].compoundPath, [{ id: 'parent', name: '上层组合' }]);
  assert.deepEqual(childDissolved.timeline.tracks[0].clips[1].compoundPath, [{ id: 'parent', name: '上层组合' }]);
  assert.equal(childDissolved.timeline.tracks[0].clips[2].compoundId, 'parent');
});

test('复合内部轨道经过统一命令新增和删除，并拒绝误删非空轨道', () => {
  const compound = {
    tracks: [{ id: 'video', kind: 'video', clips: [
      { id: 'a', kind: 'video', startMs: 0, durationMs: 1000, compoundId: 'compound-1' },
      { id: 'b', kind: 'video', startMs: 1000, durationMs: 1000, compoundId: 'compound-1' }
    ] }]
  };
  const added = executeAICutCommand(compound, {
    type: 'compound.track_add', compoundId: 'compound-1', kind: 'audio', trackId: 'inside-audio', name: '组内声音'
  });
  assert.deepEqual(added.timeline.tracks[1], {
    id: 'inside-audio', kind: 'audio', name: '组内声音', order: 1, muted: false, locked: false,
    settings: { compoundId: 'compound-1', compoundName: '复合片段' }, clips: []
  });
  assert.throws(() => executeAICutCommand({ ...added.timeline, tracks: added.timeline.tracks.map(track => track.id === 'inside-audio' ? { ...track, clips: [{ id: 'voice' }] } : track) }, {
    type: 'compound.track_remove', compoundId: 'compound-1', trackId: 'inside-audio'
  }), /轨道中还有片段/);
  const removed = executeAICutCommand(added.timeline, { type: 'compound.track_remove', compoundId: 'compound-1', trackId: 'inside-audio' });
  assert.equal(removed.timeline.tracks.some(track => track.id === 'inside-audio'), false);
  assert.throws(() => executeAICutCommand(compound, {
    type: 'compound.track_add', compoundId: 'missing', kind: 'audio', trackId: 'bad'
  }), /复合片段不存在/);
});

test('复合内部轨道重排经过统一命令并保留组外轨道位置', () => {
  const timeline = {
    tracks: [
      { id: 'video', kind: 'video', order: 0, clips: [
        { id: 'a', kind: 'video', startMs: 0, durationMs: 1000, compoundId: 'compound-1' },
        { id: 'b', kind: 'video', startMs: 1000, durationMs: 1000, compoundId: 'compound-1' }
      ] },
      { id: 'outside', kind: 'effect', order: 1, clips: [{ id: 'fx', startMs: 0, durationMs: 500 }] },
      { id: 'inside-audio', kind: 'audio', order: 2, settings: { compoundId: 'compound-1' }, clips: [] },
      { id: 'subtitle', kind: 'subtitle', order: 3, clips: [{ id: 's1', startMs: 0, durationMs: 1000, compoundId: 'compound-1' }] }
    ]
  };
  const result = executeAICutCommand(timeline, {
    type: 'compound.track_reorder',
    compoundId: 'compound-1',
    orderedTrackIds: ['subtitle', 'video', 'inside-audio']
  });
  assert.deepEqual(result.timeline.tracks.map(track => track.id), ['subtitle', 'outside', 'video', 'inside-audio']);
  assert.deepEqual(result.timeline.tracks.map(track => track.order), [0, 1, 2, 3]);
  assert.deepEqual(result.metadata.orderedTrackIds, ['subtitle', 'video', 'inside-audio']);
  assert.equal(timeline.tracks[0].id, 'video');

  const repeated = executeAICutCommand(result.timeline, {
    type: 'compound.track_reorder', compoundId: 'compound-1', orderedTrackIds: ['subtitle', 'video', 'inside-audio']
  });
  assert.equal(repeated.changed, false);
});

test('复合内部轨道重排拒绝缺失、重复、组外和锁定轨道', () => {
  const timeline = {
    tracks: [
      { id: 'video', kind: 'video', clips: [
        { id: 'a', startMs: 0, durationMs: 1000, compoundId: 'compound-1' },
        { id: 'b', startMs: 1000, durationMs: 1000, compoundId: 'compound-1' }
      ] },
      { id: 'inside', kind: 'audio', settings: { compoundId: 'compound-1' }, clips: [] },
      { id: 'outside', kind: 'audio', clips: [] }
    ]
  };
  const run = orderedTrackIds => executeAICutCommand(timeline, { type: 'compound.track_reorder', compoundId: 'compound-1', orderedTrackIds });
  assert.throws(() => run(['video']), /必须包含全部组内轨道/);
  assert.throws(() => run(['video', 'video']), /不能重复/);
  assert.throws(() => run(['video', 'outside']), /不属于当前复合片段/);
  assert.throws(() => executeAICutCommand({ ...timeline, tracks: timeline.tracks.map(track => track.id === 'inside' ? { ...track, locked: true } : track) }, {
    type: 'compound.track_reorder', compoundId: 'compound-1', orderedTrackIds: ['inside', 'video']
  }), /轨道已锁定/);
});

test('复合成员跨轨道移动保留组内时间和片段参数并改变真实轨道', () => {
  const timeline = {
    tracks: [
      { id: 'video-a', kind: 'video', clips: [
        { id: 'a', kind: 'video', assetId: 'asset-a', startMs: 1200, durationMs: 800, sourceStartMs: 50, compoundId: 'compound-1', compoundName: '组合' },
        { id: 'b', kind: 'video', startMs: 2200, durationMs: 600, compoundId: 'compound-1', compoundName: '组合' }
      ] },
      { id: 'video-b', kind: 'video', settings: { compoundId: 'compound-1' }, clips: [{ id: 'existing', kind: 'image', startMs: 3000, durationMs: 500, compoundId: 'compound-1' }] },
      { id: 'outside', kind: 'video', clips: [] }
    ]
  };
  const result = executeAICutCommand(timeline, {
    type: 'compound.member_track_move', compoundId: 'compound-1', memberKey: 'video-a::a', targetTrackId: 'video-b', startMs: 1200
  });
  assert.deepEqual(result.timeline.tracks[0].clips.map(clip => clip.id), ['b']);
  assert.deepEqual(result.timeline.tracks[1].clips.map(clip => clip.id), ['a', 'existing']);
  assert.equal(result.timeline.tracks[1].clips[0].assetId, 'asset-a');
  assert.equal(result.timeline.tracks[1].clips[0].sourceStartMs, 50);
  assert.equal(result.selectionKey, 'video-b::a');
  assert.equal(result.metadata.sourceTrackId, 'video-a');
  assert.equal(timeline.tracks[0].clips.length, 2);

  const repeated = executeAICutCommand(result.timeline, {
    type: 'compound.member_track_move', compoundId: 'compound-1', memberKey: 'video-b::a', targetTrackId: 'video-b', startMs: 1200
  });
  assert.equal(repeated.changed, false);
});

test('复合成员跨轨道移动拒绝类型不匹配、组外、锁定、重叠和编号冲突', () => {
  const base = {
    tracks: [
      { id: 'video', kind: 'video', clips: [
        { id: 'a', kind: 'video', startMs: 0, durationMs: 1000, compoundId: 'compound-1' },
        { id: 'b', kind: 'video', startMs: 1000, durationMs: 1000, compoundId: 'compound-1' }
      ] },
      { id: 'video-target', kind: 'video', settings: { compoundId: 'compound-1' }, clips: [{ id: 'busy', kind: 'video', startMs: 0, durationMs: 500 }] },
      { id: 'audio-target', kind: 'audio', settings: { compoundId: 'compound-1' }, clips: [] },
      { id: 'outside', kind: 'video', clips: [] }
    ]
  };
  const run = (overrides = {}, timeline = base) => executeAICutCommand(timeline, {
    type: 'compound.member_track_move', compoundId: 'compound-1', memberKey: 'video::a', targetTrackId: 'video-target', startMs: 0, ...overrides
  });
  assert.throws(() => run({ targetTrackId: 'audio-target' }), /轨道类型不匹配/);
  assert.throws(() => run({ targetTrackId: 'outside' }), /不属于当前复合片段/);
  assert.throws(() => run(), /目标轨道已有片段占用/);
  assert.throws(() => run({}, { ...base, tracks: base.tracks.map(track => track.id === 'video-target' ? { ...track, locked: true, clips: [] } : track) }), /轨道已锁定/);
  assert.throws(() => run({}, { ...base, tracks: base.tracks.map(track => track.id === 'video-target' ? { ...track, clips: [{ id: 'a', startMs: 2000, durationMs: 100 }] } : track) }), /片段编号已存在/);
});

test('复合成员跨组迁移同时更新真实轨道和复合路径', () => {
  const timeline = {
    tracks: [
      { id: 'source-video', kind: 'video', clips: [
        { id: 's1', kind: 'video', assetId: 'asset-s1', startMs: 0, durationMs: 800, sourceStartMs: 120, compoundId: 'source', compoundName: '来源组' },
        { id: 's2', kind: 'video', startMs: 1000, durationMs: 800, compoundId: 'source', compoundName: '来源组' },
        { id: 's3', kind: 'video', startMs: 2000, durationMs: 800, compoundId: 'source', compoundName: '来源组' }
      ] },
      { id: 'target-video', kind: 'video', settings: { compoundId: 'target', compoundName: '目标组' }, clips: [
        { id: 't1', kind: 'video', startMs: 3000, durationMs: 800, compoundId: 'target', compoundName: '目标组' },
        { id: 't2', kind: 'video', startMs: 4000, durationMs: 800, compoundId: 'target', compoundName: '目标组' }
      ] }
    ]
  };
  const result = executeAICutCommand(timeline, {
    type: 'compound.member_transfer', sourceCompoundId: 'source', targetCompoundId: 'target', memberKey: 'source-video::s1', targetTrackId: 'target-video', startMs: 0
  });
  assert.deepEqual(result.timeline.tracks[0].clips.map(clip => clip.id), ['s2', 's3']);
  const moved = result.timeline.tracks[1].clips.find(clip => clip.id === 's1');
  assert.equal(moved.assetId, 'asset-s1');
  assert.equal(moved.sourceStartMs, 120);
  assert.deepEqual(moved.compoundPath, [{ id: 'target', name: '目标组' }]);
  assert.equal(moved.compoundId, 'target');
  assert.equal(result.selectionKey, 'target-video::s1');
  assert.equal(result.metadata.sourceCompoundId, 'source');
  assert.equal(result.metadata.targetCompoundId, 'target');
  assert.equal(timeline.tracks[0].clips.length, 3);
});

test('嵌套复合成员可从子组迁移到父组且父组成员关系保持', () => {
  const path = [{ id: 'child', name: '子组' }, { id: 'parent', name: '父组' }];
  const timeline = { tracks: [{ id: 'video', kind: 'video', clips: [
    { id: 'a', kind: 'video', startMs: 0, durationMs: 500, compoundPath: path, compoundId: 'parent', compoundName: '父组' },
    { id: 'b', kind: 'video', startMs: 600, durationMs: 500, compoundPath: path, compoundId: 'parent', compoundName: '父组' },
    { id: 'c', kind: 'video', startMs: 1200, durationMs: 500, compoundPath: path, compoundId: 'parent', compoundName: '父组' },
    { id: 'p2', kind: 'video', startMs: 1800, durationMs: 500, compoundPath: [{ id: 'parent', name: '父组' }], compoundId: 'parent', compoundName: '父组' }
  ] }] };
  const result = executeAICutCommand(timeline, {
    type: 'compound.member_transfer', sourceCompoundId: 'child', targetCompoundId: 'parent', memberKey: 'video::a', targetTrackId: 'video', startMs: 0
  });
  assert.deepEqual(result.timeline.tracks[0].clips.find(clip => clip.id === 'a').compoundPath, [{ id: 'parent', name: '父组' }]);
  assert.equal(result.timeline.tracks[0].clips.filter(clip => clip.compoundPath?.some(ref => ref.id === 'child')).length, 2);
  assert.equal(result.timeline.tracks[0].clips.filter(clip => clip.compoundPath?.some(ref => ref.id === 'parent')).length, 4);
});

test('跨复合迁移拒绝破坏来源组、相同目标、锁定、重叠和类型错误', () => {
  const base = {
    tracks: [
      { id: 'source', kind: 'video', clips: [
        { id: 's1', kind: 'video', startMs: 0, durationMs: 800, compoundId: 'source-group' },
        { id: 's2', kind: 'video', startMs: 1000, durationMs: 800, compoundId: 'source-group' }
      ] },
      { id: 'target', kind: 'video', settings: { compoundId: 'target-group' }, clips: [
        { id: 't1', kind: 'video', startMs: 0, durationMs: 800, compoundId: 'target-group' },
        { id: 't2', kind: 'video', startMs: 1000, durationMs: 800, compoundId: 'target-group' }
      ] },
      { id: 'audio-target', kind: 'audio', settings: { compoundId: 'target-group' }, clips: [] }
    ]
  };
  const run = (overrides = {}, timeline = base) => executeAICutCommand(timeline, {
    type: 'compound.member_transfer', sourceCompoundId: 'source-group', targetCompoundId: 'target-group', memberKey: 'source::s1', targetTrackId: 'target', startMs: 2000, ...overrides
  });
  assert.throws(() => run(), /来源复合片段至少保留两个成员/);
  assert.throws(() => run({ targetCompoundId: 'source-group', targetTrackId: 'source' }), /不能相同/);
  const threeMembers = { ...base, tracks: base.tracks.map(track => track.id === 'source' ? { ...track, clips: [...track.clips, { id: 's3', kind: 'video', startMs: 2000, durationMs: 800, compoundId: 'source-group' }] } : track) };
  assert.throws(() => run({ targetTrackId: 'audio-target' }, threeMembers), /轨道类型不匹配/);
  assert.throws(() => run({ startMs: 0 }, threeMembers), /目标轨道已有片段占用/);
  assert.throws(() => run({}, { ...threeMembers, tracks: threeMembers.tracks.map(track => track.id === 'target' ? { ...track, locked: true } : track) }), /轨道已锁定/);
});

test('整个嵌套复合块跨组迁移保留子层级、轨道和时间', () => {
  const timeline = { tracks: [
    { id: 'video', kind: 'video', clips: [
      { id: 'a', kind: 'video', startMs: 0, durationMs: 500, compoundPath: [{ id: 'source', name: '来源块' }, { id: 'old-parent', name: '旧父组' }], compoundId: 'old-parent' },
      { id: 'b', kind: 'video', startMs: 600, durationMs: 500, compoundPath: [{ id: 'source', name: '来源块' }, { id: 'old-parent', name: '旧父组' }], compoundId: 'old-parent' },
      { id: 'c', kind: 'video', startMs: 1200, durationMs: 500, compoundPath: [{ id: 'child', name: '内部子组' }, { id: 'source', name: '来源块' }, { id: 'old-parent', name: '旧父组' }], compoundId: 'old-parent' },
      { id: 'old-extra', kind: 'video', startMs: 1800, durationMs: 500, compoundPath: [{ id: 'old-parent', name: '旧父组' }], compoundId: 'old-parent' },
      { id: 't1', kind: 'video', startMs: 2400, durationMs: 500, compoundPath: [{ id: 'target', name: '目标组' }], compoundId: 'target' },
      { id: 't2', kind: 'video', startMs: 3000, durationMs: 500, compoundPath: [{ id: 'target', name: '目标组' }], compoundId: 'target' }
    ] }
  ] };
  const result = executeAICutCommand(timeline, { type: 'compound.block_transfer', sourceCompoundId: 'source', targetCompoundId: 'target' });
  const [a, b, c] = result.timeline.tracks[0].clips;
  assert.deepEqual(a.compoundPath, [{ id: 'source', name: '来源块' }, { id: 'target', name: '目标组' }]);
  assert.deepEqual(b.compoundPath, [{ id: 'source', name: '来源块' }, { id: 'target', name: '目标组' }]);
  assert.deepEqual(c.compoundPath, [{ id: 'child', name: '内部子组' }, { id: 'source', name: '来源块' }, { id: 'target', name: '目标组' }]);
  assert.equal(c.startMs, 1200);
  assert.equal(result.timeline.tracks[0].id, 'video');
  assert.deepEqual(result.metadata.memberKeys, ['video::a', 'video::b', 'video::c']);
  assert.equal(timeline.tracks[0].clips[0].compoundId, 'old-parent');
});

test('复合块迁移拒绝自身、后代、锁定轨道并对当前父组幂等', () => {
  const timeline = { tracks: [{ id: 'video', kind: 'video', clips: [
    { id: 'a', startMs: 0, durationMs: 500, compoundPath: [{ id: 'source', name: '来源' }, { id: 'parent', name: '父组' }], compoundId: 'parent' },
    { id: 'b', startMs: 600, durationMs: 500, compoundPath: [{ id: 'descendant', name: '后代' }, { id: 'source', name: '来源' }, { id: 'parent', name: '父组' }], compoundId: 'parent' },
    { id: 'p2', startMs: 1200, durationMs: 500, compoundPath: [{ id: 'parent', name: '父组' }], compoundId: 'parent' }
  ] }] };
  assert.throws(() => executeAICutCommand(timeline, { type: 'compound.block_transfer', sourceCompoundId: 'source', targetCompoundId: 'source' }), /不能相同/);
  assert.throws(() => executeAICutCommand(timeline, { type: 'compound.block_transfer', sourceCompoundId: 'source', targetCompoundId: 'descendant' }), /不能拖入自己的后代/);
  const sameParent = executeAICutCommand(timeline, { type: 'compound.block_transfer', sourceCompoundId: 'source', targetCompoundId: 'parent' });
  assert.equal(sameParent.changed, false);
  const locked = { ...timeline, tracks: timeline.tracks.map(track => ({ ...track, locked: true })) };
  assert.throws(() => executeAICutCommand(locked, { type: 'compound.block_transfer', sourceCompoundId: 'source', targetCompoundId: 'parent' }), /轨道已锁定/);
});

test('多个复合成员一次迁移并允许整组迁走', () => {
  const timeline = { tracks: [{ id: 'video', kind: 'video', clips: [
    { id: 'a', startMs: 0, durationMs: 500, compoundPath: [{ id: 'source', name: '来源' }], compoundId: 'source' },
    { id: 'b', startMs: 600, durationMs: 500, compoundPath: [{ id: 'source', name: '来源' }], compoundId: 'source' },
    { id: 'c', startMs: 1200, durationMs: 500, compoundPath: [{ id: 'source', name: '来源' }], compoundId: 'source' },
    { id: 'd', startMs: 1800, durationMs: 500, compoundPath: [{ id: 'source', name: '来源' }], compoundId: 'source' },
    { id: 't1', startMs: 2400, durationMs: 500, compoundPath: [{ id: 'target', name: '目标' }], compoundId: 'target' },
    { id: 't2', startMs: 3000, durationMs: 500, compoundPath: [{ id: 'target', name: '目标' }], compoundId: 'target' }
  ] }, { id: 'source-empty', kind: 'audio', settings: { compoundId: 'source', compoundName: '来源' }, clips: [] }] };
  const partial = executeAICutCommand(timeline, {
    type: 'compound.members_transfer', sourceCompoundId: 'source', targetCompoundId: 'target', memberKeys: ['video::a', 'video::b']
  });
  assert.deepEqual(partial.timeline.tracks[0].clips.slice(0, 2).map(clip => clip.compoundPath), [
    [{ id: 'target', name: '目标' }], [{ id: 'target', name: '目标' }]
  ]);
  assert.equal(partial.timeline.tracks[0].clips.slice(2, 4).every(clip => clip.compoundId === 'source'), true);
  assert.deepEqual(partial.metadata.memberKeys, ['video::a', 'video::b']);

  const all = executeAICutCommand(timeline, {
    type: 'compound.members_transfer', sourceCompoundId: 'source', targetCompoundId: 'target', memberKeys: ['video::a', 'video::b', 'video::c', 'video::d']
  });
  assert.equal(all.timeline.tracks[0].clips.some(clip => clip.compoundPath?.some(ref => ref.id === 'source')), false);
  assert.equal(all.timeline.tracks[0].clips.filter(clip => clip.compoundPath?.some(ref => ref.id === 'target')).length, 6);
  assert.equal(all.timeline.tracks.some(track => track.id === 'source-empty'), false);
});

test('批量迁移保留完整嵌套子组并拒绝半组选取、单成员残留、重复和锁定', () => {
  const timeline = { tracks: [{ id: 'video', kind: 'video', clips: [
    { id: 'a', startMs: 0, durationMs: 500, compoundPath: [{ id: 'child', name: '子组' }, { id: 'source', name: '来源' }], compoundId: 'source' },
    { id: 'b', startMs: 600, durationMs: 500, compoundPath: [{ id: 'child', name: '子组' }, { id: 'source', name: '来源' }], compoundId: 'source' },
    { id: 'c', startMs: 1200, durationMs: 500, compoundPath: [{ id: 'source', name: '来源' }], compoundId: 'source' },
    { id: 'd', startMs: 1800, durationMs: 500, compoundPath: [{ id: 'source', name: '来源' }], compoundId: 'source' },
    { id: 't1', startMs: 2400, durationMs: 500, compoundPath: [{ id: 'target', name: '目标' }], compoundId: 'target' },
    { id: 't2', startMs: 3000, durationMs: 500, compoundPath: [{ id: 'target', name: '目标' }], compoundId: 'target' }
  ] }] };
  const run = (memberKeys, source = timeline) => executeAICutCommand(source, {
    type: 'compound.members_transfer', sourceCompoundId: 'source', targetCompoundId: 'target', memberKeys
  });
  assert.throws(() => run(['video::a', 'video::c']), /必须完整选择子复合片段/);
  assert.throws(() => run(['video::a', 'video::b', 'video::c']), /来源复合片段不能只剩一个成员/);
  assert.throws(() => run(['video::a', 'video::a']), /不能重复/);
  assert.throws(() => run(['video::a']), /至少选择两个成员/);
  const locked = { ...timeline, tracks: timeline.tracks.map(track => ({ ...track, locked: true })) };
  assert.throws(() => run(['video::a', 'video::b'], locked), /轨道已锁定/);
  const moved = run(['video::a', 'video::b']);
  assert.deepEqual(moved.timeline.tracks[0].clips[0].compoundPath, [{ id: 'child', name: '子组' }, { id: 'target', name: '目标' }]);
});

test('批量跨组迁移自动映射目标专用轨道并为冲突成员新建轨道', () => {
  const timeline = { tracks: [
    { id: 'source-video-a', kind: 'video', name: '来源画面A', settings: { compoundId: 'source', compoundName: '来源' }, clips: [
      { id: 'a', kind: 'video', startMs: 0, durationMs: 500, compoundPath: [{ id: 'source', name: '来源' }], compoundId: 'source' },
      { id: 'c', kind: 'video', startMs: 2000, durationMs: 500, compoundPath: [{ id: 'source', name: '来源' }], compoundId: 'source' }
    ] },
    { id: 'source-video-b', kind: 'video', name: '来源画面B', settings: { compoundId: 'source', compoundName: '来源' }, clips: [
      { id: 'b', kind: 'video', startMs: 0, durationMs: 500, compoundPath: [{ id: 'source', name: '来源' }], compoundId: 'source' },
      { id: 'd', kind: 'video', startMs: 2600, durationMs: 500, compoundPath: [{ id: 'source', name: '来源' }], compoundId: 'source' }
    ] },
    { id: 'source-audio', kind: 'audio', name: '来源声音', settings: { compoundId: 'source', compoundName: '来源' }, clips: [
      { id: 'voice', kind: 'audio', startMs: 600, durationMs: 500, compoundPath: [{ id: 'source', name: '来源' }], compoundId: 'source' }
    ] },
    { id: 'target-video', kind: 'video', name: '目标画面', settings: { compoundId: 'target', compoundName: '目标' }, clips: [
      { id: 't1', kind: 'video', startMs: 1000, durationMs: 500, compoundPath: [{ id: 'target', name: '目标' }], compoundId: 'target' },
      { id: 't2', kind: 'video', startMs: 1600, durationMs: 500, compoundPath: [{ id: 'target', name: '目标' }], compoundId: 'target' }
    ] },
    { id: 'target-audio', kind: 'audio', name: '目标声音', settings: { compoundId: 'target', compoundName: '目标' }, clips: [] }
  ] };
  const result = executeAICutCommand(timeline, {
    type: 'compound.members_transfer',
    sourceCompoundId: 'source',
    targetCompoundId: 'target',
    memberKeys: ['source-video-a::a', 'source-video-b::b', 'source-audio::voice'],
    autoMapTracks: true
  });
  assert.equal(result.timeline.tracks.find(track => track.id === 'target-video').clips.some(clip => clip.id === 'a'), true);
  const createdVideo = result.timeline.tracks.find(track => track.id.startsWith('compound-target-video-auto-'));
  assert.ok(createdVideo);
  assert.deepEqual(createdVideo.clips.map(clip => clip.id), ['b']);
  assert.equal(result.timeline.tracks.find(track => track.id === 'target-audio').clips.some(clip => clip.id === 'voice'), true);
  assert.deepEqual(result.metadata.createdTrackIds, [createdVideo.id]);
  assert.deepEqual(result.metadata.targetTrackByMember, {
    'source-video-a::a': 'target-video',
    'source-video-b::b': createdVideo.id,
    'source-audio::voice': 'target-audio'
  });
  assert.equal(result.selectionKey, 'target-video::a');
  assert.equal(timeline.tracks.find(track => track.id === 'target-video').clips.length, 2, '命令不得修改源时间线');

  const all = executeAICutCommand(timeline, {
    type: 'compound.members_transfer',
    sourceCompoundId: 'source',
    targetCompoundId: 'target',
    memberKeys: ['source-video-a::a', 'source-video-a::c', 'source-video-b::b', 'source-video-b::d', 'source-audio::voice'],
    autoMapTracks: true
  });
  assert.equal(all.timeline.tracks.some(track => track.settings?.compoundId === 'source'), false);
  assert.equal(all.timeline.tracks.flatMap(track => track.clips).filter(clip => clip.compoundId === 'target').length, 7);

  const lockedTargets = { ...timeline, tracks: timeline.tracks.map(track => ['target-video', 'target-audio'].includes(track.id) ? { ...track, locked: true } : track) };
  const avoidsLocked = executeAICutCommand(lockedTargets, {
    type: 'compound.members_transfer', sourceCompoundId: 'source', targetCompoundId: 'target',
    memberKeys: ['source-video-a::a', 'source-video-b::b', 'source-audio::voice'], autoMapTracks: true
  });
  assert.equal(avoidsLocked.timeline.tracks.find(track => track.id === 'target-video').clips.some(clip => clip.id === 'a' || clip.id === 'b'), false);
  assert.equal(avoidsLocked.metadata.createdTrackIds.length, 3);

  const manual = executeAICutCommand(timeline, {
    type: 'compound.members_transfer', sourceCompoundId: 'source', targetCompoundId: 'target',
    memberKeys: ['source-video-a::a', 'source-video-b::b', 'source-audio::voice'], autoMapTracks: true,
    targetTrackByMember: {
      'source-video-a::a': '__new__',
      'source-video-b::b': 'target-video',
      'source-audio::voice': 'target-audio'
    }
  });
  assert.equal(manual.metadata.targetTrackByMember['source-video-b::b'], 'target-video');
  assert.match(manual.metadata.targetTrackByMember['source-video-a::a'], /^compound-target-video-auto-/);
  assert.throws(() => executeAICutCommand(lockedTargets, {
    type: 'compound.members_transfer', sourceCompoundId: 'source', targetCompoundId: 'target',
    memberKeys: ['source-video-a::a', 'source-video-b::b'], autoMapTracks: true,
    targetTrackByMember: { 'source-video-a::a': 'target-video' }
  }), /指定的目标轨道已锁定/);
});

test('自动配乐经过统一命令层并保留节拍标记与撤销快照', () => {
  const timeline = {
    revision: 3,
    tracks: [{ id: 'video-1', kind: 'video', name: '画面', clips: [{ id: 'v1', startMs: 0, durationMs: 3000 }] }],
    markers: []
  };
  let next = 0;
  const result = executeAICutCommand(timeline, {
    type: 'timeline.bgm_plan',
    asset: { id: 'audio-1', kind: 'audio', name: '鼓点.mp3', durationMs: 2000 },
    options: { referenceTrackId: 'video-1', fill: true, beatsMs: [500, 1000] }
  }, { idFactory: (prefix) => `${prefix}-${++next}` });
  assert.equal(result.changed, true);
  assert.equal(timeline.tracks.length, 1);
  assert.equal(result.timeline.tracks.find(track => track.name === 'BGM').clips.length, 2);
  assert.deepEqual(result.timeline.markers.filter(marker => marker.source === 'audio-beat').map(marker => marker.timeMs), [500, 1000, 2500, 3000]);
});

test('保存返回时有新编辑不会覆盖当前时间线，保存后撤销沿用最新服务端版本号', () => {
  const saved = { revision: 8, tracks: source.tracks };
  const current = { revision: 7, tracks: [...source.tracks, { id: 'subtitle', kind: 'subtitle', clips: [] }] };
  const raced = reconcileTimelineSave(current, saved, 4, 5);
  assert.equal(raced.dirty, true);
  assert.equal(raced.timeline.revision, 8);
  assert.equal(raced.timeline.tracks.length, 3);

  const settled = reconcileTimelineSave(source, saved, 5, 5);
  assert.equal(settled.dirty, false);
  assert.equal(settled.timeline, saved);

  const restored = restoreTimelineSnapshot({ revision: 3, tracks: [] }, 8);
  assert.equal(restored.revision, 8);
  assert.deepEqual(restored.tracks, []);
});

test('命令结果直接驱动预览，不存在另一份展示用假数据', () => {
  const assets = [{ id: 'asset-a', kind: 'video', name: 'A.mp4', url: '/a.mp4' }];
  const timeline = { tracks: [{ id: 'video', kind: 'video', clips: [{ id: 'a', assetId: 'asset-a', kind: 'video', startMs: 0, durationMs: 2000, sourceStartMs: 0 }] }] };
  assert.equal(buildAICutPreviewFrame(timeline, assets, 500).visuals.length, 1);
  const moved = executeAICutCommand(timeline, { type: 'clip.move', trackId: 'video', clipId: 'a', startMs: 1000, magnetic: false });
  assert.equal(buildAICutPreviewFrame(moved.timeline, assets, 500).visuals.length, 0);
  assert.equal(buildAICutPreviewFrame(moved.timeline, assets, 1500).visuals[0].clipId, 'a');
  const trimmed = executeAICutCommand(moved.timeline, { type: 'clip.resize', trackId: 'video', clipId: 'a', edge: 'right', edgeTimeMs: 1800 });
  assert.equal(buildAICutPreviewFrame(trimmed.timeline, assets, 1900).visuals.length, 0);
});

test('多选波纹删除会按每条轨道关闭被删除片段留下的空隙', () => {
  const timeline = { tracks: [{ id: 'v', kind: 'video', locked: false, clips: [
    { id: 'a', startMs: 0, durationMs: 1000 },
    { id: 'b', startMs: 1000, durationMs: 500 },
    { id: 'c', startMs: 1500, durationMs: 1000 },
    { id: 'd', startMs: 2500, durationMs: 500 }
  ] }] };
  const result = executeAICutCommand(timeline, { type: 'clip.delete_many', clipKeys: ['v::b', 'v::d'], ripple: true });
  assert.deepEqual(result.timeline.tracks[0].clips.map(clip => [clip.id, clip.startMs]), [['a', 0], ['c', 1000]]);
  assert.equal(result.metadata.deletedCount, 2);
});

test('复制板保留相对时间、轨道和联动关系，粘贴时生成新编号并从播放头开始', () => {
  const linked = executeAICutCommand(source, { type: 'clip.link', clipKeys: ['video::a', 'audio::voice'], groupId: 'old-group' }).timeline;
  const clipboard = createAICutClipboard(linked, ['video::a', 'audio::voice']);
  assert.equal(clipboard.items.length, 2);
  assert.deepEqual(clipboard.items.map(item => item.relativeStartMs), [0, 0]);
  let next = 0;
  const pasted = executeAICutCommand(linked, { type: 'clip.paste', clipboard, atMs: 6000 }, { idFactory: prefix => `${prefix}-${++next}` });
  const videoCopy = pasted.timeline.tracks[0].clips.find(clip => clip.id === 'clip-1');
  const audioCopy = pasted.timeline.tracks[1].clips.find(clip => clip.id === 'clip-2');
  assert.equal(videoCopy.startMs, 6000);
  assert.equal(audioCopy.startMs, 6000);
  assert.ok(videoCopy.linkGroupId && videoCopy.linkGroupId === audioCopy.linkGroupId);
  assert.notEqual(videoCopy.linkGroupId, 'old-group');
  assert.deepEqual(pasted.metadata.pastedKeys, ['video::clip-1', 'audio::clip-2']);
});

test('复制复合片段会生成新的复合编号且保持成员关系', () => {
  const compound = executeAICutCommand(source, {
    type: 'clip.compound_create', clipKeys: ['video::a', 'audio::voice'], compoundId: 'old-compound', name: '镜头组合'
  }).timeline;
  const clipboard = createAICutClipboard(compound, ['video::a', 'audio::voice']);
  let next = 0;
  const pasted = executeAICutCommand(compound, { type: 'clip.paste', clipboard, atMs: 6000 }, { idFactory: prefix => `${prefix}-${++next}` });
  const copies = pasted.metadata.pastedKeys.map(key => {
    const [trackId, clipId] = key.split('::');
    return pasted.timeline.tracks.find(track => track.id === trackId).clips.find(clip => clip.id === clipId);
  });
  assert.ok(copies[0].compoundId);
  assert.equal(copies[0].compoundId, copies[1].compoundId);
  assert.notEqual(copies[0].compoundId, 'old-compound');
  assert.equal(copies[0].compoundName, '镜头组合');
});

test('波纹右裁会移动后续片段，普通裁剪不移动后续片段', () => {
  const timeline = { tracks: [{ id: 'v', kind: 'video', locked: false, clips: [
    { id: 'a', startMs: 0, durationMs: 2000, sourceStartMs: 0, sourceDurationMs: 3000 },
    { id: 'b', startMs: 2000, durationMs: 1000, sourceStartMs: 0, sourceDurationMs: 1000 }
  ] }] };
  const ripple = executeAICutCommand(timeline, { type: 'clip.resize', trackId: 'v', clipId: 'a', edge: 'right', edgeTimeMs: 1500, ripple: true });
  assert.deepEqual(ripple.timeline.tracks[0].clips.map(clip => [clip.id, clip.startMs, clip.durationMs]), [['a', 0, 1500], ['b', 1500, 1000]]);
  const plain = executeAICutCommand(timeline, { type: 'clip.resize', trackId: 'v', clipId: 'a', edge: 'right', edgeTimeMs: 1500, ripple: false });
  assert.equal(plain.timeline.tracks[0].clips[1].startMs, 2000);
});

test('逐帧移动按项目帧率换算毫秒并遵守零点和联动组边界', () => {
  const linked = executeAICutCommand(source, { type: 'clip.link', clipKeys: ['video::a', 'audio::voice'], groupId: 'frame-group' }).timeline;
  const moved = executeAICutCommand(linked, { type: 'clip.nudge', trackId: 'video', clipId: 'a', frames: 3, fps: 30, linked: true });
  assert.equal(moved.timeline.tracks[0].clips[0].startMs, 100);
  assert.equal(moved.timeline.tracks[1].clips[0].startMs, 100);
  assert.throws(() => executeAICutCommand(linked, { type: 'clip.nudge', trackId: 'video', clipId: 'a', frames: -1, fps: 30 }), /不能早于时间线零点/);
});

test('编辑点跳转会去重所有片段起止边界并处理首尾', () => {
  assert.equal(findTimelineEditPoint(source, 0, 'next'), 2000);
  assert.equal(findTimelineEditPoint(source, 4000, 'previous'), 2000);
  assert.equal(findTimelineEditPoint(source, 5000, 'next'), 5000);
  assert.equal(findTimelineEditPoint({ tracks: [] }, 123, 'previous'), 0);
});

test('时间线标记点可以新增、同位置更新和删除', () => {
  let next = 0;
  const added = executeAICutCommand(source, { type: 'timeline.marker_add', timeMs: 1250, label: '情绪转折', color: '#f59e0b' }, { idFactory: prefix => `${prefix}-${++next}` });
  assert.deepEqual(added.timeline.markers, [{ id: 'marker-1', timeMs: 1250, label: '情绪转折', color: '#f59e0b' }]);
  const updated = executeAICutCommand(added.timeline, { type: 'timeline.marker_add', timeMs: 1250, label: '高潮', color: '#ef4444' }, { idFactory: prefix => `${prefix}-${++next}` });
  assert.equal(updated.timeline.markers.length, 1);
  assert.equal(updated.timeline.markers[0].label, '高潮');
  const removed = executeAICutCommand(updated.timeline, { type: 'timeline.marker_remove', markerId: 'marker-1' });
  assert.deepEqual(removed.timeline.markers, []);
});

test('片段关键帧在片段内部新增、同时间替换并可删除', () => {
  let next = 0;
  const added = executeAICutCommand(source, {
    type: 'clip.keyframe_upsert', trackId: 'video', clipId: 'a', offsetMs: 500,
    values: { x: .25, y: -.1, scaleX: 1.2, scaleY: 1.2, rotation: 12, opacity: .8 }
  }, { idFactory: prefix => `${prefix}-${++next}` });
  assert.equal(added.timeline.tracks[0].clips[0].keyframes[0].id, 'keyframe-1');
  const replaced = executeAICutCommand(added.timeline, {
    type: 'clip.keyframe_upsert', trackId: 'video', clipId: 'a', offsetMs: 500,
    values: { x: .5, volume: .75 }
  }, { idFactory: prefix => `${prefix}-${++next}` });
  assert.equal(replaced.timeline.tracks[0].clips[0].keyframes.length, 1);
  assert.equal(replaced.timeline.tracks[0].clips[0].keyframes[0].values.x, .5);
  assert.throws(() => executeAICutCommand(source, { type: 'clip.keyframe_upsert', trackId: 'video', clipId: 'a', offsetMs: 2500, values: { x: 0 } }), /片段范围/);
  assert.throws(() => executeAICutCommand(source, { type: 'clip.keyframe_upsert', trackId: 'video', clipId: 'a', offsetMs: 500, values: { opacity: 2 } }), /透明度必须在 0 到 1 之间/);
  const removed = executeAICutCommand(replaced.timeline, { type: 'clip.keyframe_remove', trackId: 'video', clipId: 'a', keyframeId: 'keyframe-1' });
  assert.deepEqual(removed.timeline.tracks[0].clips[0].keyframes, []);
});

test('区域跟踪关键帧一次写入且重复执行只替换旧跟踪结果', () => {
  const timeline = { tracks: [{ id: 'video', kind: 'video', locked: false, clips: [{
    id: 'a', kind: 'video', durationMs: 1200, keyframes: [
      { id: 'manual-opacity', offsetMs: 750, values: { opacity: .6 } },
      { id: 'tracking-keyframe-old', offsetMs: 300, values: { x: .2, y: 0 } }
    ]
  }] }] };
  let next = 0;
  const first = executeAICutCommand(timeline, {
    type: 'clip.tracking_keyframes_apply', trackId: 'video', clipId: 'a',
    frames: [
      { offsetMs: 0, x: 0, y: 0, confidence: 1 },
      { offsetMs: 500, x: .1, y: -.05, confidence: .8 },
      { offsetMs: 1000, x: .2, y: -.1, confidence: .7 }
    ]
  }, { idFactory: prefix => `${prefix}-${++next}` });
  const firstFrames = first.timeline.tracks[0].clips[0].keyframes;
  assert.deepEqual(firstFrames.map(frame => frame.offsetMs), [0, 500, 750, 1000]);
  assert.equal(firstFrames.some(frame => frame.id === 'tracking-keyframe-old'), false);
  assert.equal(firstFrames.find(frame => frame.id === 'manual-opacity').values.opacity, .6);
  assert.equal(first.metadata.trackingKeyframeCount, 3);

  const second = executeAICutCommand(first.timeline, {
    type: 'clip.tracking_keyframes_apply', trackId: 'video', clipId: 'a',
    frames: [{ offsetMs: 0, x: 0, y: 0 }, { offsetMs: 1200, x: -.2, y: .15 }]
  }, { idFactory: prefix => `${prefix}-${++next}` });
  const secondFrames = second.timeline.tracks[0].clips[0].keyframes;
  assert.deepEqual(secondFrames.map(frame => frame.offsetMs), [0, 750, 1200]);
  assert.equal(secondFrames.filter(frame => frame.id.startsWith('tracking-keyframe-')).length, 2);
  assert.equal(secondFrames.find(frame => frame.id === 'manual-opacity').values.opacity, .6);
});

test('区域跟踪命令拒绝空结果、非视频、锁定轨道、越界和人工位置关键帧冲突', () => {
  const timeline = { tracks: [{ id: 'video', kind: 'video', locked: false, clips: [{
    id: 'a', kind: 'video', durationMs: 1000, keyframes: [{ id: 'manual-position', offsetMs: 500, values: { x: .2 } }]
  }] }] };
  const base = { type: 'clip.tracking_keyframes_apply', trackId: 'video', clipId: 'a' };
  const dependencies = { idFactory: prefix => `${prefix}-1` };
  assert.throws(() => executeAICutCommand(timeline, { ...base, frames: [] }, dependencies), /没有可写入的跟踪结果/);
  assert.throws(() => executeAICutCommand(timeline, { ...base, frames: [{ offsetMs: 1100, x: 0, y: 0 }] }, dependencies), /片段范围/);
  assert.throws(() => executeAICutCommand(timeline, { ...base, frames: [{ offsetMs: 0, x: 2, y: 0 }] }, dependencies), /水平位置/);
  assert.throws(() => executeAICutCommand(timeline, { ...base, frames: [{ offsetMs: 500, x: 0, y: 0 }] }, dependencies), /人工位置关键帧冲突/);
  const locked = { ...timeline, tracks: [{ ...timeline.tracks[0], locked: true }] };
  assert.throws(() => executeAICutCommand(locked, { ...base, frames: [{ offsetMs: 0, x: 0, y: 0 }] }, dependencies), /轨道已锁定/);
  const image = { tracks: [{ id: 'video', kind: 'video', locked: false, clips: [{ id: 'a', kind: 'image', durationMs: 1000 }] }] };
  assert.throws(() => executeAICutCommand(image, { ...base, frames: [{ offsetMs: 0, x: 0, y: 0 }] }, dependencies), /只支持视频片段/);
});

test('关键帧缓动通过统一命令保存预设与自定义控制柄', () => {
  const timeline = { tracks: [{ id: 'video', kind: 'video', locked: false, clips: [{ id: 'a', durationMs: 2000, keyframes: [
    { id: 'k1', offsetMs: 0, values: { x: 0 } },
    { id: 'k2', offsetMs: 1000, values: { x: 1 } }
  ] }] }] };
  const eased = executeAICutCommand(timeline, { type: 'clip.keyframe_easing_set', trackId: 'video', clipId: 'a', keyframeId: 'k1', easing: { type: 'ease-in' } });
  assert.equal(eased.timeline.tracks[0].clips[0].keyframes[0].easing.type, 'ease-in');
  const custom = executeAICutCommand(eased.timeline, { type: 'clip.keyframe_easing_set', trackId: 'video', clipId: 'a', keyframeId: 'k1', easing: { type: 'custom', x1: .15, y1: .4, x2: .75, y2: .95 } });
  assert.deepEqual(custom.timeline.tracks[0].clips[0].keyframes[0].easing, { type: 'custom', x1: .15, y1: .4, x2: .75, y2: .95 });
  assert.throws(() => executeAICutCommand(timeline, { type: 'clip.keyframe_easing_set', trackId: 'video', clipId: 'a', keyframeId: 'missing', easing: { type: 'linear' } }), /关键帧不存在/);
});

test('批量关键帧编辑在一次命令中统一缓动、偏移或删除', () => {
  const timeline = { tracks: [{ id: 'video', kind: 'video', locked: false, clips: [{ id: 'a', durationMs: 1200, keyframes: [
    { id: 'k1', offsetMs: 0, values: { x: 0 } },
    { id: 'k2', offsetMs: 500, values: { x: .25 } },
    { id: 'k3', offsetMs: 1000, values: { x: .5 } }
  ] }] }] };
  const eased = executeAICutCommand(timeline, { type: 'clip.keyframes_batch_edit', action: 'easing', trackId: 'video', clipId: 'a', keyframeIds: ['k1', 'k2'], easing: { type: 'ease-out' } });
  assert.deepEqual(eased.timeline.tracks[0].clips[0].keyframes.slice(0, 2).map(frame => frame.easing?.type), ['ease-out', 'ease-out']);
  assert.deepEqual(eased.metadata.keyframeIds, ['k1', 'k2']);

  const shifted = executeAICutCommand(eased.timeline, { type: 'clip.keyframes_batch_edit', action: 'shift', trackId: 'video', clipId: 'a', keyframeIds: ['k1', 'k2'], deltaMs: 100 });
  assert.deepEqual(shifted.timeline.tracks[0].clips[0].keyframes.map(frame => frame.offsetMs), [100, 600, 1000]);

  const removed = executeAICutCommand(shifted.timeline, { type: 'clip.keyframes_batch_edit', action: 'remove', trackId: 'video', clipId: 'a', keyframeIds: ['k1', 'k3'] });
  assert.deepEqual(removed.timeline.tracks[0].clips[0].keyframes.map(frame => frame.id), ['k2']);
});

test('批量关键帧编辑拒绝空选择、未知帧、越界和时间碰撞', () => {
  const timeline = { tracks: [{ id: 'video', kind: 'video', locked: false, clips: [{ id: 'a', durationMs: 1200, keyframes: [
    { id: 'k1', offsetMs: 0, values: { x: 0 } },
    { id: 'k2', offsetMs: 500, values: { x: .25 } },
    { id: 'k3', offsetMs: 1000, values: { x: .5 } }
  ] }] }] };
  const command = { type: 'clip.keyframes_batch_edit', trackId: 'video', clipId: 'a' };
  assert.throws(() => executeAICutCommand(timeline, { ...command, action: 'remove', keyframeIds: [] }), /请至少选择一个关键帧/);
  assert.throws(() => executeAICutCommand(timeline, { ...command, action: 'remove', keyframeIds: ['missing'] }), /关键帧不存在/);
  assert.throws(() => executeAICutCommand(timeline, { ...command, action: 'shift', keyframeIds: ['k3'], deltaMs: 300 }), /片段范围/);
  assert.throws(() => executeAICutCommand(timeline, { ...command, action: 'shift', keyframeIds: ['k2'], deltaMs: 500 }), /时间位置冲突/);
  assert.throws(() => executeAICutCommand(timeline, { ...command, action: 'unknown', keyframeIds: ['k1'] }), /批量编辑操作无效/);
});

test('跨片段粘贴关键帧保留相对间隔、参数与缓动并生成新编号', () => {
  const timeline = { tracks: [{ id: 'video', kind: 'video', locked: false, clips: [{
    id: 'target', durationMs: 1200, keyframes: [{ id: 'old', offsetMs: 1000, values: { x: .8 } }]
  }] }] };
  let next = 0;
  const pasted = executeAICutCommand(timeline, {
    type: 'clip.keyframes_paste', trackId: 'video', clipId: 'target', anchorMs: 100,
    frames: [
      { offsetMs: 300, values: { x: .1, opacity: .8 }, easing: { type: 'ease-in' } },
      { offsetMs: 700, values: { x: .4, opacity: .9 }, easing: { type: 'custom', x1: .2, y1: .1, x2: .8, y2: .9 } }
    ]
  }, { idFactory: prefix => `${prefix}-${++next}` });
  const frames = pasted.timeline.tracks[0].clips[0].keyframes;
  assert.deepEqual(frames.map(frame => frame.offsetMs), [100, 500, 1000]);
  assert.deepEqual(frames.slice(0, 2).map(frame => frame.id), ['keyframe-1', 'keyframe-2']);
  assert.equal(frames[0].easing.type, 'ease-in');
  assert.deepEqual(frames[1].easing, { type: 'custom', x1: .2, y1: .1, x2: .8, y2: .9 });
  assert.deepEqual(pasted.metadata.keyframeIds, ['keyframe-1', 'keyframe-2']);
});

test('跨片段粘贴关键帧整批拒绝空剪贴板、非法锚点、越界和时间冲突', () => {
  const timeline = { tracks: [{ id: 'video', kind: 'video', locked: false, clips: [{
    id: 'target', durationMs: 1000, keyframes: [{ id: 'old', offsetMs: 500, values: { x: .8 } }]
  }] }] };
  const base = { type: 'clip.keyframes_paste', trackId: 'video', clipId: 'target' };
  const dependencies = { idFactory: prefix => `${prefix}-new` };
  assert.throws(() => executeAICutCommand(timeline, { ...base, anchorMs: 0, frames: [] }, dependencies), /关键帧剪贴板为空/);
  assert.throws(() => executeAICutCommand(timeline, { ...base, anchorMs: -1, frames: [{ offsetMs: 0, values: { x: 0 } }] }, dependencies), /粘贴位置/);
  assert.throws(() => executeAICutCommand(timeline, { ...base, anchorMs: 800, frames: [{ offsetMs: 0, values: { x: 0 } }, { offsetMs: 300, values: { x: 1 } }] }, dependencies), /片段范围/);
  assert.throws(() => executeAICutCommand(timeline, { ...base, anchorMs: 500, frames: [{ offsetMs: 0, values: { x: 0 } }] }, dependencies), /时间位置冲突/);
  assert.throws(() => executeAICutCommand(timeline, { ...base, anchorMs: 100, frames: [{ offsetMs: 0, values: {} }] }, dependencies), /缺少可用参数/);
});

test('转场命令只接受相邻视频片段并把时长限制在双方一半以内', () => {
  const timeline = { tracks: [{ id: 'v', kind: 'video', locked: false, clips: [
    { id: 'a', kind: 'video', startMs: 0, durationMs: 600 },
    { id: 'b', kind: 'video', startMs: 600, durationMs: 1000 }
  ] }] };
  const result = executeAICutCommand(timeline, {
    type: 'clip.transition_set', trackId: 'v', clipId: 'b', transition: 'transition-cross-dissolve', transitionMs: 900
  }, { idFactory: () => 'tr-1' });
  assert.equal(result.selectionKey, 'track-tr-1::tr-1');
  assert.equal(result.timeline.tracks[0].clips[1].transition, 'transition-cross-dissolve');
  assert.equal(result.timeline.tracks[0].clips[1].transitionMs, 300);
  assert.equal(result.timeline.tracks.find(track => track.kind === 'transition').clips[0].durationMs, 300);

  const withGap = { ...timeline, tracks: [{ ...timeline.tracks[0], clips: [timeline.tracks[0].clips[0], { ...timeline.tracks[0].clips[1], startMs: 700 }] }] };
  assert.throws(() => executeAICutCommand(withGap, {
    type: 'clip.transition_set', trackId: 'v', clipId: 'b', transition: 'transition-cross-dissolve', transitionMs: 300
  }, { idFactory: () => 'tr-2' }), /两个片段之间有空隙/);
});

test('速度曲线命令按曲线平均倍率重算片段时长并保留曲线点', () => {
  const timeline = { tracks: [{ id: 'v', kind: 'video', locked: false, clips: [
    { id: 'a', kind: 'video', startMs: 0, durationMs: 2000, sourceStartMs: 0, sourceDurationMs: 4000, playbackRate: 1 }
  ] }] };
  const points = [{ position: 0, rate: .5 }, { position: .5, rate: 2 }, { position: 1, rate: .5 }];
  const result = executeAICutCommand(timeline, { type: 'clip.speed_curve_set', trackId: 'v', clipId: 'a', points });
  const clip = result.timeline.tracks[0].clips[0];
  assert.equal(result.selectionKey, 'v::a');
  assert.equal(clip.durationMs, 1600);
  assert.equal(clip.playbackRate, 1.25);
  assert.deepEqual(clip.speedCurve, points);
  assert.equal(timeline.tracks[0].clips[0].durationMs, 2000);
});

test('音量包络命令新增、移动和删除音量点且保留同时间其他关键帧参数', () => {
  const timeline = { tracks: [{ id: 'a', kind: 'audio', clips: [{
    id: 'c', kind: 'audio', durationMs: 1000, volume: 1,
    keyframes: [{ id: 'k', offsetMs: 500, values: { x: .2 } }]
  }] }] };
  const added = executeAICutCommand(timeline, { type: 'clip.volume_point_upsert', trackId: 'a', clipId: 'c', offsetMs: 500, volume: .4 }, { idFactory: () => 'new' });
  assert.deepEqual(added.timeline.tracks[0].clips[0].keyframes[0].values, { x: .2, volume: .4 });
  const moved = executeAICutCommand(added.timeline, { type: 'clip.volume_point_move', trackId: 'a', clipId: 'c', keyframeId: 'k', offsetMs: 750, volume: 1.6 });
  assert.deepEqual(moved.timeline.tracks[0].clips[0].keyframes[0], { id: 'k', offsetMs: 750, values: { x: .2, volume: 1.6 } });
  const removed = executeAICutCommand(moved.timeline, { type: 'clip.volume_point_remove', trackId: 'a', clipId: 'c', keyframeId: 'k' });
  assert.deepEqual(removed.timeline.tracks[0].clips[0].keyframes, [{ id: 'k', offsetMs: 750, values: { x: .2 } }]);
  assert.throws(() => executeAICutCommand(timeline, { type: 'clip.volume_point_upsert', trackId: 'a', clipId: 'c', offsetMs: 400, volume: 3 }, { idFactory: () => 'bad' }), /音量必须在 0 到 2 之间/);
});

test('音频处理命令统一设置降噪、三段均衡器、声道混音和响度匹配', () => {
  const timeline = { tracks: [{ id: 'a', kind: 'audio', locked: false, clips: [{
    id: 'c', kind: 'audio', durationMs: 1000, volume: 1
  }] }] };
  const result = executeAICutCommand(timeline, {
    type: 'clip.audio_processing_set', trackId: 'a', clipId: 'c',
    processing: { denoise: .65, eqLowDB: 3, eqMidDB: -2, eqHighDB: 4, channelMode: 'mono', loudnessEnabled: true, targetLUFS: -14, voiceEffect: 'robot' }
  });
  assert.deepEqual(result.timeline.tracks[0].clips[0].audioProcessing, {
    denoise: .65, eqLowDB: 3, eqMidDB: -2, eqHighDB: 4, channelMode: 'mono', loudnessEnabled: true, targetLUFS: -14, voiceEffect: 'robot', pitchSemitones: 0, pan: 0
  });
  assert.equal(result.selectionKey, 'a::c');
  assert.equal(timeline.tracks[0].clips[0].audioProcessing, undefined);
});

test('扩展音频预设全部通过统一命令写入时间线', () => {
  const audio = { tracks: [{ id: 'a', kind: 'audio', locked: false, clips: [{ id: 'c', kind: 'audio', durationMs: 1000 }] }] };
  const effects = ['bright', 'bass', 'treble', 'radio', 'megaphone', 'deep', 'sweet', 'whisper', 'distorted', 'lofi', 'echo', 'room', 'hall', 'plate', 'cathedral', 'stadium', 'underwater', 'vinyl'];
  for (const voiceEffect of effects) {
    const result = executeAICutCommand(audio, { type: 'clip.audio_processing_set', trackId: 'a', clipId: 'c', processing: { voiceEffect, targetLUFS: -16 } });
    assert.equal(result.timeline.tracks[0].clips[0].audioProcessing.voiceEffect, voiceEffect);
  }
});

test('通用音调参数通过统一命令保存并拒绝越界输入', () => {
  const audio = { tracks: [{ id: 'a', kind: 'audio', locked: false, clips: [{ id: 'c', kind: 'audio', durationMs: 1000 }] }] };
  const result = executeAICutCommand(audio, { type: 'clip.audio_processing_set', trackId: 'a', clipId: 'c', processing: { voiceEffect: 'natural', pitchSemitones: 7, targetLUFS: -16 } });
  assert.equal(result.timeline.tracks[0].clips[0].audioProcessing.pitchSemitones, 7);
  assert.throws(() => executeAICutCommand(audio, { type: 'clip.audio_processing_set', trackId: 'a', clipId: 'c', processing: { pitchSemitones: 12.1 } }), /音调必须在 -12 到 12 个半音之间/);
});

test('音频处理命令拒绝非音频、锁轨和越界参数', () => {
  const audio = { tracks: [{ id: 'a', kind: 'audio', locked: false, clips: [{ id: 'c', kind: 'audio', durationMs: 1000 }] }] };
  assert.throws(() => executeAICutCommand(audio, { type: 'clip.audio_processing_set', trackId: 'a', clipId: 'c', processing: { denoise: 2 } }), /降噪强度/);
  assert.throws(() => executeAICutCommand(audio, { type: 'clip.audio_processing_set', trackId: 'a', clipId: 'c', processing: { eqLowDB: -13 } }), /均衡器/);
  assert.throws(() => executeAICutCommand(audio, { type: 'clip.audio_processing_set', trackId: 'a', clipId: 'c', processing: { channelMode: 'surround' } }), /声道模式/);
  assert.throws(() => executeAICutCommand(audio, { type: 'clip.audio_processing_set', trackId: 'a', clipId: 'c', processing: { loudnessEnabled: true, targetLUFS: -7 } }), /目标响度/);
  assert.throws(() => executeAICutCommand(audio, { type: 'clip.audio_processing_set', trackId: 'a', clipId: 'c', processing: { voiceEffect: 'alien' } }), /音频效果/);
  assert.throws(() => executeAICutCommand({ tracks: [{ id: 'v', kind: 'video', clips: [{ id: 'c', kind: 'video', durationMs: 1000 }] }] }, { type: 'clip.audio_processing_set', trackId: 'v', clipId: 'c', processing: {} }), /只支持音频/);
  assert.throws(() => executeAICutCommand({ ...audio, tracks: [{ ...audio.tracks[0], locked: true }] }, { type: 'clip.audio_processing_set', trackId: 'a', clipId: 'c', processing: {} }), /轨道已锁定/);
});

test('音频淡入淡出经过统一命令并保留原时间线不变', () => {
  const timeline = { tracks: [{ id: 'a', kind: 'audio', locked: false, clips: [{ id: 'c', kind: 'audio', durationMs: 2000, fadeInMs: 0, fadeOutMs: 0 }] }] };
  const result = executeAICutCommand(timeline, { type: 'clip.audio_fade_set', trackId: 'a', clipId: 'c', fadeInMs: 400, fadeOutMs: 600 });
  assert.equal(result.timeline.tracks[0].clips[0].fadeInMs, 400);
  assert.equal(result.timeline.tracks[0].clips[0].fadeOutMs, 600);
  assert.equal(result.selectionKey, 'a::c');
  assert.deepEqual([timeline.tracks[0].clips[0].fadeInMs, timeline.tracks[0].clips[0].fadeOutMs], [0, 0]);
});

test('音频淡入淡出拒绝画面、锁轨、负数、超长和非有限输入', () => {
  const audio = { tracks: [{ id: 'a', kind: 'audio', locked: false, clips: [{ id: 'c', kind: 'audio', durationMs: 1000 }] }] };
  assert.throws(() => executeAICutCommand(audio, { type: 'clip.audio_fade_set', trackId: 'a', clipId: 'c', fadeInMs: -1, fadeOutMs: 0 }), /淡入/);
  assert.throws(() => executeAICutCommand(audio, { type: 'clip.audio_fade_set', trackId: 'a', clipId: 'c', fadeInMs: 0, fadeOutMs: 1001 }), /淡出/);
  assert.throws(() => executeAICutCommand(audio, { type: 'clip.audio_fade_set', trackId: 'a', clipId: 'c', fadeInMs: Number.NaN, fadeOutMs: 0 }), /淡入/);
  assert.throws(() => executeAICutCommand({ ...audio, tracks: [{ ...audio.tracks[0], locked: true }] }, { type: 'clip.audio_fade_set', trackId: 'a', clipId: 'c', fadeInMs: 0, fadeOutMs: 0 }), /轨道已锁定/);
  const video = { tracks: [{ id: 'v', kind: 'video', clips: [{ id: 'c', kind: 'video', durationMs: 1000 }] }] };
  assert.throws(() => executeAICutCommand(video, { type: 'clip.audio_fade_set', trackId: 'v', clipId: 'c', fadeInMs: 0, fadeOutMs: 0 }), /只支持音频/);
});

test('高级画面命令统一设置调色、色键和蒙版', () => {
  const timeline = { tracks: [{ id: 'v', kind: 'video', locked: false, clips: [{ id: 'c', kind: 'video', durationMs: 1000 }] }] };
  const visualProcessing = {
    frameInterpolation: 'off',
    stabilization: { enabled: false, smoothness: 15, zoom: 0, tripod: false },
    personSegmentation: { enabled: true, matteAssetId: 'matte-1', threshold: .55, feather: .12, signature: 'asset-1|0|1000|1|0|none' },
    color: { exposure: .5, temperature: .2, tint: -.1, highlights: .3, shadows: -.2, gamma: 1.1, curve: [0, .2, .55, .85, 1], curvePoints: [{ x: 0, y: 0 }, { x: .25, y: .2 }, { x: .5, y: .55 }, { x: .75, y: .85 }, { x: 1, y: 1 }] },
    chromaKey: { enabled: true, removeBlack: false, color: '#00ff00', similarity: .25, blend: .1 },
    mask: { type: 'ellipse', x: .5, y: .5, width: .8, height: .7, feather: .1, invert: false, points: [], bezierPoints: [] },
    geometry: { cropLeft: .1, cropTop: .05, cropRight: .1, cropBottom: .05, flipHorizontal: true, flipVertical: false },
    details: { sharpen: .4, vignette: .3, grain: .2 },
    objectRemoval: { enabled: false, x: .5, y: .5, width: .2, height: .2 }
  };
  const result = executeAICutCommand(timeline, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: visualProcessing });
  assert.deepEqual(result.timeline.tracks[0].clips[0].visualProcessing, visualProcessing);
  assert.equal(result.selectionKey, 'v::c');
  assert.equal(timeline.tracks[0].clips[0].visualProcessing, undefined);
});

test('区域物体消除通过统一命令保存位置和启用状态', () => {
  const timeline = { tracks: [{ id: 'v', kind: 'video', locked: false, clips: [{ id: 'c', kind: 'video', durationMs: 1000 }] }] };
  const result = executeAICutCommand(timeline, {
    type: 'clip.visual_processing_set',
    trackId: 'v',
    clipId: 'c',
    processing: { objectRemoval: { enabled: true, x: .67, y: .61, width: .29, height: .26 } }
  });
  assert.deepEqual(result.timeline.tracks[0].clips[0].visualProcessing.objectRemoval, {
    enabled: true,
    x: .67,
    y: .61,
    width: .29,
    height: .26
  });
  assert.equal(timeline.tracks[0].clips[0].visualProcessing, undefined);
});

test('区域物体消除统一命令拒绝越界或无效区域', () => {
  const timeline = { tracks: [{ id: 'v', kind: 'video', locked: false, clips: [{ id: 'c', kind: 'video', durationMs: 1000 }] }] };
  const command = (objectRemoval) => ({ type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { objectRemoval } });
  assert.throws(() => executeAICutCommand(timeline, command({ enabled: true, x: 1.1, y: .5, width: .2, height: .2 })), /消除区域中心X/);
  assert.throws(() => executeAICutCommand(timeline, command({ enabled: true, x: .5, y: .5, width: .01, height: .2 })), /消除区域宽度/);
  assert.throws(() => executeAICutCommand(timeline, command({ enabled: true, x: .95, y: .5, width: .2, height: .2 })), /必须位于画面内/);
});

test('人物抠像统一命令拒绝图片和缺失蒙版信息', () => {
  const video = { tracks: [{ id: 'v', kind: 'video', locked: false, clips: [{ id: 'c', kind: 'video', durationMs: 1000 }] }] };
  const image = { tracks: [{ id: 'v', kind: 'video', locked: false, clips: [{ id: 'c', kind: 'image', durationMs: 1000 }] }] };
  assert.throws(() => executeAICutCommand(video, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { personSegmentation: { enabled: true } } }), /蒙版素材/);
  assert.throws(() => executeAICutCommand(image, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { personSegmentation: { enabled: true, matteAssetId: 'm', signature: 's' } } }), /只支持视频/);
});

test('视频补帧通过统一命令保存并对重复设置保持稳定', () => {
  const timeline = { tracks: [{ id: 'v', kind: 'video', locked: false, clips: [{ id: 'c', kind: 'video', durationMs: 1000 }] }] };
  const first = executeAICutCommand(timeline, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { frameInterpolation: 'optical-flow' } });
  assert.equal(first.timeline.tracks[0].clips[0].visualProcessing.frameInterpolation, 'optical-flow');
  const second = executeAICutCommand(first.timeline, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { frameInterpolation: 'optical-flow' } });
  assert.deepEqual(second.timeline, first.timeline);
  const disabled = executeAICutCommand(second.timeline, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { frameInterpolation: 'off' } });
  assert.equal(disabled.timeline.tracks[0].clips[0].visualProcessing.frameInterpolation, 'off');
});

test('视频补帧拒绝图片、未知模式和锁定轨道', () => {
  const video = { tracks: [{ id: 'v', kind: 'video', locked: false, clips: [{ id: 'c', kind: 'video', durationMs: 1000 }] }] };
  assert.throws(() => executeAICutCommand(video, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { frameInterpolation: 'fake' } }), /补帧模式/);
  const image = { tracks: [{ id: 'v', kind: 'video', locked: false, clips: [{ id: 'c', kind: 'image', durationMs: 1000 }] }] };
  assert.throws(() => executeAICutCommand(image, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { frameInterpolation: 'blend' } }), /只支持视频/);
  assert.throws(() => executeAICutCommand({ ...video, tracks: [{ ...video.tracks[0], locked: true }] }, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { frameInterpolation: 'blend' } }), /轨道已锁定/);
});

test('视频防抖参数通过统一命令保存并支持关闭', () => {
  const timeline = { tracks: [{ id: 'v', kind: 'video', locked: false, clips: [{ id: 'c', kind: 'video', durationMs: 1000 }] }] };
  const enabled = executeAICutCommand(timeline, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { stabilization: { enabled: true, smoothness: 60, zoom: 12, tripod: true } } });
  assert.deepEqual(enabled.timeline.tracks[0].clips[0].visualProcessing.stabilization, { enabled: true, smoothness: 60, zoom: 12, tripod: true });
  const disabled = executeAICutCommand(enabled.timeline, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { stabilization: { enabled: false, smoothness: 60, zoom: 12, tripod: true } } });
  assert.equal(disabled.timeline.tracks[0].clips[0].visualProcessing.stabilization.enabled, false);
});

test('视频防抖拒绝图片、越界参数和锁定轨道', () => {
  const video = { tracks: [{ id: 'v', kind: 'video', locked: false, clips: [{ id: 'c', kind: 'video', durationMs: 1000 }] }] };
  assert.throws(() => executeAICutCommand(video, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { stabilization: { enabled: true, smoothness: 0 } } }), /平滑度/);
  assert.throws(() => executeAICutCommand(video, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { stabilization: { enabled: true, smoothness: 15, zoom: 31 } } }), /缩放/);
  const image = { tracks: [{ id: 'v', kind: 'video', locked: false, clips: [{ id: 'c', kind: 'image', durationMs: 1000 }] }] };
  assert.throws(() => executeAICutCommand(image, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { stabilization: { enabled: true, smoothness: 15, zoom: 0 } } }), /只支持视频/);
  assert.throws(() => executeAICutCommand({ ...video, tracks: [{ ...video.tracks[0], locked: true }] }, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { stabilization: { enabled: true, smoothness: 15, zoom: 0 } } }), /轨道已锁定/);
});

test('批量画面命令一次修改多个片段并保留各自未修改的参数', () => {
  const timeline = { tracks: [
    { id: 'v', kind: 'video', locked: false, clips: [
      { id: 'a', kind: 'video', durationMs: 1000, visualProcessing: { color: { exposure: .4 }, details: { sharpen: .2 } } },
      { id: 'b', kind: 'image', durationMs: 1000, visualProcessing: { color: { exposure: -.3 }, mask: { type: 'ellipse', x: .5, y: .5, width: .8, height: .7 } } }
    ] }
  ] };
  const result = executeAICutCommand(timeline, {
    type: 'clip.visual_processing_batch_patch',
    clipKeys: ['v::a', 'v::b'],
    patch: { chromaKey: { enabled: true, removeBlack: true, color: '#112233', similarity: .3, blend: .12 } }
  });
  const [first, second] = result.timeline.tracks[0].clips;
  assert.equal(result.metadata.clipCount, 2);
  assert.equal(first.visualProcessing.color.exposure, .4);
  assert.equal(first.visualProcessing.details.sharpen, .2);
  assert.equal(second.visualProcessing.color.exposure, -.3);
  assert.equal(second.visualProcessing.mask.type, 'ellipse');
  assert.deepEqual(first.visualProcessing.chromaKey, { enabled: true, removeBlack: true, color: '#112233', similarity: .3, blend: .12 });
  assert.deepEqual(second.visualProcessing.chromaKey, first.visualProcessing.chromaKey);
  assert.equal(timeline.tracks[0].clips[0].visualProcessing.chromaKey, undefined);
});

test('批量画面命令对空选择、音频、锁轨和不存在片段保持原子失败', () => {
  const timeline = { tracks: [
    { id: 'v', kind: 'video', locked: false, clips: [{ id: 'a', kind: 'video', durationMs: 1000 }] },
    { id: 'a', kind: 'audio', locked: false, clips: [{ id: 'voice', kind: 'audio', durationMs: 1000 }] }
  ] };
  assert.throws(() => executeAICutCommand(timeline, { type: 'clip.visual_processing_batch_patch', clipKeys: [], patch: {} }), /至少选择一个/);
  assert.throws(() => executeAICutCommand(timeline, { type: 'clip.visual_processing_batch_patch', clipKeys: ['v::a', 'a::voice'], patch: {} }), /只支持画面/);
  assert.throws(() => executeAICutCommand(timeline, { type: 'clip.visual_processing_batch_patch', clipKeys: ['v::a', 'v::missing'], patch: {} }), /片段不存在/);
  assert.throws(() => executeAICutCommand({ ...timeline, tracks: [{ ...timeline.tracks[0], locked: true }, timeline.tracks[1]] }, { type: 'clip.visual_processing_batch_patch', clipKeys: ['v::a'], patch: {} }), /轨道已锁定/);
  assert.equal(timeline.tracks[0].clips[0].visualProcessing, undefined);
});

test('高级画面命令拒绝音频、锁轨和非法参数', () => {
  const video = { tracks: [{ id: 'v', kind: 'video', locked: false, clips: [{ id: 'c', kind: 'video', durationMs: 1000 }] }] };
  assert.throws(() => executeAICutCommand(video, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { color: { exposure: 3 } } }), /曝光/);
  assert.throws(() => executeAICutCommand(video, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { chromaKey: { color: 'green' } } }), /色键颜色/);
  assert.throws(() => executeAICutCommand(video, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { mask: { type: 'star' } } }), /蒙版类型/);
  assert.throws(() => executeAICutCommand(video, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { color: { curve: [0, .5, 1] } } }), /曲线/);
  assert.throws(() => executeAICutCommand(video, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { color: { curvePoints: [{ x: 0, y: 0 }, { x: .5, y: .4 }, { x: .4, y: .6 }, { x: 1, y: 1 }] } } }), /横坐标/);
  assert.throws(() => executeAICutCommand(video, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { color: { curvePoints: [{ x: .1, y: 0 }, { x: 1, y: 1 }] } } }), /首尾/);
  assert.throws(() => executeAICutCommand(video, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { mask: { feather: .6 } } }), /羽化/);
  assert.throws(() => executeAICutCommand(video, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { mask: { type: 'polygon', points: [{ x: 0, y: 0 }, { x: 1, y: 1 }] } } }), /3到12个/);
  assert.throws(() => executeAICutCommand(video, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { mask: { type: 'polygon', points: [{ x: 0, y: 0 }, { x: 2, y: 0 }, { x: 0, y: 1 }] } } }), /控制点/);
  assert.throws(() => executeAICutCommand(video, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { mask: { type: 'polygon', points: [{ x: .2, y: .2 }, { x: .4, y: .4 }, { x: .6, y: .6 }] } } }), /有效面积/);
  assert.throws(() => executeAICutCommand(video, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { mask: { type: 'bezier', bezierPoints: [{ x: .2, y: .2 }] } } }), /3到8个/);
  assert.throws(() => executeAICutCommand(video, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { mask: { type: 'bezier', bezierPoints: [{ x: .2, y: .2, inX: -1, inY: .2, outX: .3, outY: .2 }, { x: .8, y: .2, inX: .7, inY: .2, outX: .9, outY: .2 }, { x: .5, y: .8, inX: .4, inY: .7, outX: .6, outY: .7 }] } } }), /控制柄/);
  assert.throws(() => executeAICutCommand(video, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { geometry: { cropLeft: .6, cropRight: .5 } } }), /水平裁剪/);
  assert.throws(() => executeAICutCommand(video, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { geometry: { cropTop: .7, cropBottom: .4 } } }), /垂直裁剪/);
  assert.throws(() => executeAICutCommand(video, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { details: { sharpen: 1.1 } } }), /锐化/);
  assert.throws(() => executeAICutCommand(video, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { details: { vignette: -.1 } } }), /暗角/);
  assert.throws(() => executeAICutCommand(video, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { details: { grain: Number.NaN } } }), /颗粒/);
  const audio = { tracks: [{ id: 'a', kind: 'audio', clips: [{ id: 'c', kind: 'audio', durationMs: 1000 }] }] };
  assert.throws(() => executeAICutCommand(audio, { type: 'clip.visual_processing_set', trackId: 'a', clipId: 'c', processing: {} }), /只支持画面/);
  assert.throws(() => executeAICutCommand({ ...video, tracks: [{ ...video.tracks[0], locked: true }] }, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: {} }), /轨道已锁定/);
});

test('手绘多边形蒙版通过统一命令保存控制点、羽化和反转', () => {
  const timeline = { tracks: [{ id: 'v', kind: 'video', locked: false, clips: [{ id: 'c', kind: 'video', durationMs: 1000 }] }] };
  const points = [{ x: .1, y: .15 }, { x: .9, y: .2 }, { x: .7, y: .88 }, { x: .2, y: .75 }];
  const result = executeAICutCommand(timeline, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { mask: { type: 'polygon', points, feather: .06, invert: true } } });
  assert.deepEqual(result.timeline.tracks[0].clips[0].visualProcessing.mask.points, points);
  assert.equal(result.timeline.tracks[0].clips[0].visualProcessing.mask.feather, .06);
  assert.equal(result.timeline.tracks[0].clips[0].visualProcessing.mask.invert, true);
});

test('贝塞尔曲线蒙版通过统一命令保存锚点、双控制柄和采样轮廓', () => {
  const timeline = { tracks: [{ id: 'v', kind: 'video', locked: false, clips: [{ id: 'c', kind: 'video', durationMs: 1000 }] }] };
  const bezierPoints = [
    { x: .2, y: .2, inX: .1, inY: .3, outX: .35, outY: .1 },
    { x: .8, y: .2, inX: .65, inY: .1, outX: .9, outY: .35 },
    { x: .5, y: .85, inX: .7, inY: .8, outX: .3, outY: .8 }
  ];
  const result = executeAICutCommand(timeline, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { mask: { type: 'bezier', bezierPoints, feather: .03 } } });
  const mask = result.timeline.tracks[0].clips[0].visualProcessing.mask;
  assert.deepEqual(mask.bezierPoints, bezierPoints);
  assert.equal(mask.points.length, 18);
  assert.deepEqual(mask.points[0], { x: .2, y: .2 });
});

test('片段倒放和定格经过统一命令并保持互斥', () => {
  const timeline = { tracks: [{ id: 'v', kind: 'video', clips: [{ id: 'c', kind: 'video', startMs: 1000, durationMs: 2000, sourceStartMs: 500, sourceDurationMs: 5000, playbackRate: 1 }] }] };
  const reversed = executeAICutCommand(timeline, { type: 'clip.reverse_set', trackId: 'v', clipId: 'c', reversed: true });
  assert.equal(reversed.timeline.tracks[0].clips[0].reversed, true);
  assert.equal(reversed.selectionKey, 'v::c');
  const frozen = executeAICutCommand(reversed.timeline, { type: 'clip.freeze_set', trackId: 'v', clipId: 'c', timeMs: 1500 });
  assert.equal(frozen.timeline.tracks[0].clips[0].freezeSourceMs, 1999);
  assert.equal(frozen.timeline.tracks[0].clips[0].reversed, undefined);
  const cleared = executeAICutCommand(frozen.timeline, { type: 'clip.freeze_set', trackId: 'v', clipId: 'c', clear: true });
  assert.equal(cleared.timeline.tracks[0].clips[0].freezeSourceMs, undefined);
});

test('片段倒放和定格拒绝错误类型、锁轨和片段外时间', () => {
  const audio = { tracks: [{ id: 'a', kind: 'audio', clips: [{ id: 'c', kind: 'audio', startMs: 0, durationMs: 1000, sourceDurationMs: 1000 }] }] };
  assert.throws(() => executeAICutCommand(audio, { type: 'clip.reverse_set', trackId: 'a', clipId: 'c', reversed: true }), /只支持视频/);
  assert.throws(() => executeAICutCommand(audio, { type: 'clip.freeze_set', trackId: 'a', clipId: 'c', timeMs: 500 }), /只支持视频/);
  const video = { tracks: [{ id: 'v', kind: 'video', locked: true, clips: [{ id: 'c', kind: 'video', startMs: 1000, durationMs: 1000, sourceDurationMs: 2000 }] }] };
  assert.throws(() => executeAICutCommand(video, { type: 'clip.reverse_set', trackId: 'v', clipId: 'c', reversed: true }), /轨道已锁定/);
  const unlocked = { ...video, tracks: [{ ...video.tracks[0], locked: false }] };
  assert.throws(() => executeAICutCommand(unlocked, { type: 'clip.freeze_set', trackId: 'v', clipId: 'c', timeMs: 999 }), /播放头必须位于片段内/);
});

test('画面混合模式经过统一命令并可恢复正常模式', () => {
  const timeline = { tracks: [{ id: 'v', kind: 'video', locked: false, clips: [{ id: 'c', kind: 'video', durationMs: 1000 }] }] };
  const blended = executeAICutCommand(timeline, { type: 'clip.blend_mode_set', trackId: 'v', clipId: 'c', blendMode: 'multiply' });
  assert.equal(blended.timeline.tracks[0].clips[0].blendMode, 'multiply');
  assert.equal(blended.selectionKey, 'v::c');
  assert.equal(timeline.tracks[0].clips[0].blendMode, undefined);
  const normal = executeAICutCommand(blended.timeline, { type: 'clip.blend_mode_set', trackId: 'v', clipId: 'c', blendMode: 'normal' });
  assert.equal(normal.timeline.tracks[0].clips[0].blendMode, undefined);
});

test('画面混合模式拒绝音频、锁轨和未知模式', () => {
  const video = { tracks: [{ id: 'v', kind: 'video', locked: false, clips: [{ id: 'c', kind: 'image', durationMs: 1000 }] }] };
  for (const blendMode of ['multiply', 'screen', 'overlay', 'soft-light', 'difference', 'add']) {
    assert.doesNotThrow(() => executeAICutCommand(video, { type: 'clip.blend_mode_set', trackId: 'v', clipId: 'c', blendMode }));
  }
  assert.throws(() => executeAICutCommand(video, { type: 'clip.blend_mode_set', trackId: 'v', clipId: 'c', blendMode: 'unknown' }), /不支持的混合模式/);
  const audio = { tracks: [{ id: 'a', kind: 'audio', clips: [{ id: 'c', kind: 'audio', durationMs: 1000 }] }] };
  assert.throws(() => executeAICutCommand(audio, { type: 'clip.blend_mode_set', trackId: 'a', clipId: 'c', blendMode: 'screen' }), /只支持画面/);
  assert.throws(() => executeAICutCommand({ ...video, tracks: [{ ...video.tracks[0], locked: true }] }, { type: 'clip.blend_mode_set', trackId: 'v', clipId: 'c', blendMode: 'screen' }), /轨道已锁定/);
});
