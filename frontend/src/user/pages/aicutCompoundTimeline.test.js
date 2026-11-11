import test from 'node:test';
import assert from 'node:assert/strict';
import { buildCompoundTimelineView, collapseCompoundTimelineView, compoundBatchTransferTargets, compoundBlockTransferCommand, compoundBlockTransferTargets, compoundMemberMoveCommand, compoundMemberResizeCommand, compoundMemberTrackMoveCommand, compoundMemberTransferCommand, compoundMembersTransferCommand, compoundTransferTargets, moveCompoundTrackByOffset, reorderCompoundTrackIds } from './aicutCompoundTimeline.js';

const timeline = {
  tracks: [
    { id: 'video', kind: 'video', name: '画面', clips: [
      { id: 'v1', kind: 'video', startMs: 2000, durationMs: 3000, compoundId: 'compound-1', compoundName: '开场组合' },
      { id: 'outside', kind: 'video', startMs: 8000, durationMs: 1000 }
    ] },
    { id: 'audio', kind: 'audio', name: '声音', clips: [
      { id: 'a1', kind: 'audio', startMs: 2500, durationMs: 4000, compoundId: 'compound-1', compoundName: '开场组合' }
    ] }
  ]
};

test('复合时间线只投影组内成员并把绝对时间换成组内相对时间', () => {
  const view = buildCompoundTimelineView(timeline, 'compound-1');
  assert.equal(view.name, '开场组合');
  assert.equal(view.originMs, 2000);
  assert.equal(view.durationMs, 4500);
  assert.deepEqual(view.memberKeys, ['video::v1', 'audio::a1']);
  assert.deepEqual(view.tracks.map(track => track.clips.map(clip => [clip.id, clip.startMs])), [[['v1', 0]], [['a1', 500]]]);
  assert.equal(view.tracks.some(track => track.clips.some(clip => clip.id === 'outside')), false);
});

test('复合成员移动和裁剪命令把组内相对时间还原成主时间线绝对时间', () => {
  const view = buildCompoundTimelineView(timeline, 'compound-1');
  assert.deepEqual(compoundMemberMoveCommand(view, 'audio::a1', 1200), {
    type: 'clip.move', trackId: 'audio', clipId: 'a1', startMs: 3200, magnetic: false, linked: false
  });
  assert.deepEqual(compoundMemberResizeCommand(view, 'audio::a1', 1200, 2600), {
    type: 'clip.resize', trackId: 'audio', clipId: 'a1', edge: 'right', edgeTimeMs: 5800, ripple: false
  });
});

test('复合轨道拖放和上下移动都生成完整且稳定的轨道顺序', () => {
  const view = buildCompoundTimelineView(timeline, 'compound-1');
  assert.deepEqual(reorderCompoundTrackIds(view, 'video', 'audio'), ['audio', 'video']);
  assert.deepEqual(reorderCompoundTrackIds(view, 'audio', 'video'), ['audio', 'video']);
  assert.deepEqual(moveCompoundTrackByOffset(view, 'audio', -1), ['audio', 'video']);
  assert.deepEqual(moveCompoundTrackByOffset(view, 'video', -1), ['video', 'audio']);
  assert.deepEqual(moveCompoundTrackByOffset(view, 'audio', 1), ['video', 'audio']);
  assert.throws(() => reorderCompoundTrackIds(view, 'missing', 'audio'), /轨道不存在/);
});

test('复合成员跨轨道拖放生成保留绝对时间的统一命令', () => {
  const view = buildCompoundTimelineView(timeline, 'compound-1');
  assert.deepEqual(compoundMemberTrackMoveCommand(view, 'video::v1', 'audio'), {
    type: 'compound.member_track_move', compoundId: 'compound-1', memberKey: 'video::v1', targetTrackId: 'audio', startMs: 2000
  });
  assert.throws(() => compoundMemberTrackMoveCommand(view, 'video::missing', 'audio'), /复合成员不存在/);
  assert.throws(() => compoundMemberTrackMoveCommand(view, 'video::v1', 'missing'), /轨道不存在/);
});

test('跨复合投放目标只列出其他复合组的兼容未锁定轨道', () => {
  const extended = structuredClone(timeline);
  extended.tracks.push(
    { id: 'target-video', kind: 'video', name: '目标画面', clips: [
      { id: 't1', kind: 'video', startMs: 7000, durationMs: 500, compoundId: 'target', compoundName: '目标组' },
      { id: 't2', kind: 'video', startMs: 7600, durationMs: 500, compoundId: 'target', compoundName: '目标组' }
    ] },
    { id: 'target-audio', kind: 'audio', name: '目标声音', settings: { compoundId: 'target' }, clips: [] },
    { id: 'locked-video', kind: 'video', name: '锁定画面', locked: true, settings: { compoundId: 'target' }, clips: [] }
  );
  const targets = compoundTransferTargets(extended, 'compound-1', 'video::v1');
  assert.equal(targets.length, 1);
  assert.equal(targets[0].compoundId, 'target');
  assert.equal(targets[0].name, '目标组');
  assert.deepEqual(targets[0].tracks.map(track => track.id), ['target-video']);
  assert.deepEqual(compoundMemberTransferCommand(buildCompoundTimelineView(extended, 'compound-1'), 'video::v1', 'target', 'target-video'), {
    type: 'compound.member_transfer', sourceCompoundId: 'compound-1', targetCompoundId: 'target', memberKey: 'video::v1', targetTrackId: 'target-video', startMs: 2000
  });
});

test('复合块目标排除自己、当前父组和会形成循环的后代', () => {
  const nested = { tracks: [{ id: 'video', kind: 'video', clips: [
    { id: 'a', startMs: 0, durationMs: 500, compoundPath: [{ id: 'source', name: '来源' }, { id: 'parent', name: '父组' }], compoundId: 'parent' },
    { id: 'b', startMs: 600, durationMs: 500, compoundPath: [{ id: 'descendant', name: '后代' }, { id: 'source', name: '来源' }, { id: 'parent', name: '父组' }], compoundId: 'parent' },
    { id: 'p2', startMs: 1200, durationMs: 500, compoundPath: [{ id: 'parent', name: '父组' }], compoundId: 'parent' },
    { id: 't1', startMs: 1800, durationMs: 500, compoundPath: [{ id: 'target', name: '目标' }], compoundId: 'target' },
    { id: 't2', startMs: 2400, durationMs: 500, compoundPath: [{ id: 'target', name: '目标' }], compoundId: 'target' }
  ] }] };
  assert.deepEqual(compoundBlockTransferTargets(nested, 'source'), [{ compoundId: 'target', name: '目标' }]);
  assert.deepEqual(compoundBlockTransferCommand('source', 'target'), { type: 'compound.block_transfer', sourceCompoundId: 'source', targetCompoundId: 'target' });
});

test('批量成员迁移命令保留选择并列出非后代目标组', () => {
  const nested = { tracks: [{ id: 'video', kind: 'video', clips: [
    { id: 'a', startMs: 0, durationMs: 500, compoundPath: [{ id: 'source', name: '来源' }, { id: 'parent', name: '父组' }], compoundId: 'parent' },
    { id: 'b', startMs: 600, durationMs: 500, compoundPath: [{ id: 'descendant', name: '后代' }, { id: 'source', name: '来源' }, { id: 'parent', name: '父组' }], compoundId: 'parent' },
    { id: 'c', startMs: 1200, durationMs: 500, compoundPath: [{ id: 'source', name: '来源' }, { id: 'parent', name: '父组' }], compoundId: 'parent' },
    { id: 'p2', startMs: 1800, durationMs: 500, compoundPath: [{ id: 'parent', name: '父组' }], compoundId: 'parent' },
    { id: 't1', startMs: 2400, durationMs: 500, compoundPath: [{ id: 'target', name: '目标' }], compoundId: 'target' },
    { id: 't2', startMs: 3000, durationMs: 500, compoundPath: [{ id: 'target', name: '目标' }], compoundId: 'target' }
  ] }] };
  assert.deepEqual(compoundBatchTransferTargets(nested, 'source', ['video::a', 'video::c']), [
    { compoundId: 'parent', name: '父组', trackKinds: ['video'] }, { compoundId: 'target', name: '目标', trackKinds: ['video'] }
  ]);
  const view = buildCompoundTimelineView(nested, 'source');
  assert.deepEqual(compoundMembersTransferCommand(view, ['video::a', 'video::c'], 'target'), {
    type: 'compound.members_transfer', sourceCompoundId: 'source', targetCompoundId: 'target', memberKeys: ['video::a', 'video::c'], autoMapTracks: true
  });
  assert.deepEqual(compoundMembersTransferCommand(view, ['video::a', 'video::c'], 'target', { 'video::a': '__new__' }), {
    type: 'compound.members_transfer', sourceCompoundId: 'source', targetCompoundId: 'target', memberKeys: ['video::a', 'video::c'], autoMapTracks: true,
    targetTrackByMember: { 'video::a': '__new__' }
  });
});

test('复合时间线拒绝不存在的组、陌生成员和负数输入', () => {
  assert.throws(() => buildCompoundTimelineView(timeline, 'missing'), /复合片段不存在/);
  const view = buildCompoundTimelineView(timeline, 'compound-1');
  assert.throws(() => compoundMemberMoveCommand(view, 'audio::missing', 0), /复合成员不存在/);
  assert.throws(() => compoundMemberMoveCommand(view, 'audio::a1', -1), /不能小于零/);
  assert.throws(() => compoundMemberResizeCommand(view, 'audio::a1', 0, 0), /时长必须大于零/);
});

test('主时间线把跨轨复合成员折叠成一个覆盖完整跨度的片段块', () => {
  const before = structuredClone(timeline);
  const collapsed = collapseCompoundTimelineView(timeline);
  const clips = collapsed.tracks.flatMap(track => track.clips.map(clip => ({ trackId: track.id, ...clip })));
  const container = clips.find(clip => clip.isCompoundContainer);

  assert.ok(container);
  assert.equal(container.id, 'compound-container:compound-1');
  assert.equal(container.trackId, 'video');
  assert.equal(container.name, '开场组合');
  assert.equal(container.startMs, 2000);
  assert.equal(container.durationMs, 4500);
  assert.equal(container.memberCount, 2);
  assert.deepEqual(container.memberKeys, ['video::v1', 'audio::a1']);
  assert.equal(clips.some(clip => clip.id === 'v1' || clip.id === 'a1'), false);
  assert.equal(clips.some(clip => clip.id === 'outside'), true);
  assert.deepEqual(timeline, before, '显示投影不得修改保存和导出使用的原始时间线');
});

test('成员不足的损坏复合组保持原片段可见而不是被折叠隐藏', () => {
  const broken = {
    tracks: [{ id: 'video', kind: 'video', clips: [{ id: 'v1', startMs: 100, durationMs: 500, compoundId: 'broken' }] }]
  };
  const collapsed = collapseCompoundTimelineView(broken);
  assert.equal(collapsed.tracks[0].clips.length, 1);
  assert.equal(collapsed.tracks[0].clips[0].id, 'v1');
  assert.equal(Boolean(collapsed.tracks[0].clips[0].isCompoundContainer), false);
});

test('多层复合在主线只显示最外层，进入上层后显示子复合和组内空轨道', () => {
  const nested = {
    tracks: [
      { id: 'video', kind: 'video', name: '画面', clips: [
        { id: 'a', kind: 'video', startMs: 0, durationMs: 1000, compoundId: 'parent', compoundName: '上层', compoundPath: [{ id: 'child', name: '子层' }, { id: 'parent', name: '上层' }] },
        { id: 'b', kind: 'video', startMs: 1000, durationMs: 1000, compoundId: 'parent', compoundName: '上层', compoundPath: [{ id: 'child', name: '子层' }, { id: 'parent', name: '上层' }] },
        { id: 'c', kind: 'video', startMs: 2000, durationMs: 1000, compoundId: 'parent', compoundName: '上层', compoundPath: [{ id: 'parent', name: '上层' }] }
      ] },
      { id: 'inside-audio', kind: 'audio', name: '组内声音', settings: { compoundId: 'parent', compoundName: '上层' }, clips: [] }
    ]
  };
  const main = collapseCompoundTimelineView(nested);
  const mainContainers = main.tracks.flatMap(track => track.clips).filter(clip => clip.isCompoundContainer);
  assert.equal(mainContainers.length, 1);
  assert.equal(mainContainers[0].compoundId, 'parent');
  assert.equal(mainContainers[0].memberCount, 3);

  const parent = buildCompoundTimelineView(nested, 'parent');
  assert.equal(parent.parentCompoundId, '');
  assert.equal(parent.tracks.some(track => track.id === 'inside-audio' && track.clips.length === 0), true);
  const childContainer = parent.tracks.flatMap(track => track.clips).find(clip => clip.isCompoundContainer);
  assert.equal(childContainer.compoundId, 'child');
  assert.equal(childContainer.memberCount, 2);
  assert.equal(parent.tracks.flatMap(track => track.clips).some(clip => clip.id === 'c'), true);

  const child = buildCompoundTimelineView(nested, 'child');
  assert.equal(child.parentCompoundId, 'parent');
  assert.deepEqual(child.memberKeys, ['video::a', 'video::b']);
  assert.equal(child.tracks.flatMap(track => track.clips).some(clip => clip.isCompoundContainer), false);
});
