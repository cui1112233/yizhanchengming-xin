import test from 'node:test';
import assert from 'node:assert/strict';
import { addEffectLayer, addTimelineTrack, draggedTimelineStart, markTimelineClip, mergeTimelineClipWithNext, moveEffectLayer, moveTimelineClip, normalizeSubtitleRhythm, planTimelineBGM, removeTimelineObject, removeTimelineTrack, renameTimelineTrack, resizeEffectLayer, resolveAICutAgentTool, setClipPlaybackRate, setClipTransition, setTimelineClipText, splitTimelineClip, splitTimelineClipAt, subtitleRhythmLabel, timelinePixelsToTime, timelineTimeToPixels, toggleTimelineTrackFlag, trimTimelineClipEdge, updateTimelineTrackSettings } from './aicutTimeline.js';

const base = {
  revision: 3,
  tracks: [{ id: 'video-1', kind: 'video', name: '视频1', order: 0, muted: false, locked: false, clips: [{ id: 'clip-1', kind: 'video', name: '片段', startMs: 1000, durationMs: 4000, sourceStartMs: 200, sourceDurationMs: 8000 }] }]
};

test('新增轨道按同类轨道编号且不修改原时间线', () => {
  const result = addTimelineTrack(base, 'video', () => 'video-2');
  assert.equal(result.tracks.length, 2);
  assert.equal(result.tracks[1].name, '视频2');
  assert.equal(base.tracks.length, 1);
});

test('锁定和静音只切换目标轨道', () => {
  const locked = toggleTimelineTrackFlag(base, 'video-1', 'locked');
  const muted = toggleTimelineTrackFlag(locked, 'video-1', 'muted');
  assert.equal(muted.tracks[0].locked, true);
  assert.equal(muted.tracks[0].muted, true);
});

test('刀片在片段中点生成前后连续的两个片段', () => {
  const result = splitTimelineClip(base, 'video-1', 'clip-1', () => 'clip-2');
  const [first, second] = result.tracks[0].clips;
  assert.equal(first.durationMs, 2000);
  assert.equal(second.id, 'clip-2');
  assert.equal(second.startMs, 3000);
  assert.equal(second.sourceStartMs, 2200);
  assert.equal(second.durationMs, 2000);
});

test('剃刀模式按点击位置分割并保持源素材时间连续', () => {
  const result = splitTimelineClipAt(base, 'video-1', 'clip-1', 1250, () => 'clip-razor');
  const [first, second] = result.tracks[0].clips;
  assert.equal(first.durationMs, 1250);
  assert.equal(second.startMs, 2250);
  assert.equal(second.sourceStartMs, 1450);
  assert.equal(second.durationMs, 2750);
  assert.throws(() => splitTimelineClipAt(base, 'video-1', 'clip-1', 20, () => 'bad'), /分割点距离片段边缘太近/);
});

test('字幕可拆分、编辑并与下一条合并，锁定轨道拒绝修改', () => {
  const source = { revision: 0, tracks: [{ id: 'sub', kind: 'subtitle', locked: false, clips: [{ id: 's1', kind: 'subtitle', name: '第一句第二句', text: '第一句 第二句', startMs: 0, durationMs: 2000, sourceStartMs: 0, sourceDurationMs: 2000 }] }] };
  const split = splitTimelineClip(source, 'sub', 's1', () => 's2');
  assert.equal(split.tracks[0].clips[0].text, '第一句');
  assert.equal(split.tracks[0].clips[1].text, '第二句');
  const edited = setTimelineClipText(split, 'sub', 's2', '修改后');
  const merged = mergeTimelineClipWithNext(edited, 'sub', 's1');
  assert.equal(merged.tracks[0].clips.length, 1);
  assert.equal(merged.tracks[0].clips[0].text, '第一句 修改后');
  assert.throws(() => setTimelineClipText({ ...source, tracks: [{ ...source.tracks[0], locked: true }] }, 'sub', 's1', '不允许'), /轨道已锁定/);
});

test('删除有片段的轨道必须明确允许，防止误删', () => {
  assert.throws(() => removeTimelineTrack(base, 'video-1'), /轨道中还有片段/);
  const result = removeTimelineTrack({ ...base, tracks: [{ ...base.tracks[0], clips: [] }] }, 'video-1');
  assert.equal(result.tracks.length, 0);
});

test('轨道菜单重命名会清理空白并拒绝空名称', () => {
  const renamed = renameTimelineTrack(base, 'video-1', '  主画面  ');
  assert.equal(renamed.tracks[0].name, '主画面');
  assert.equal(base.tracks[0].name, '视频1');
  assert.throws(() => renameTimelineTrack(base, 'video-1', '   '), /轨道名称不能为空/);
  assert.throws(() => renameTimelineTrack(base, 'missing', '画面'), /轨道不存在/);
});

test('左裁和右裁保持源区间有效并拒绝裁成空片段', () => {
  const left = trimTimelineClipEdge(base, 'video-1', 'clip-1', 'left', 500);
  assert.equal(left.tracks[0].clips[0].startMs, 1500);
  assert.equal(left.tracks[0].clips[0].sourceStartMs, 700);
  assert.equal(left.tracks[0].clips[0].durationMs, 3500);
  const right = trimTimelineClipEdge(left, 'video-1', 'clip-1', 'right', 500);
  assert.equal(right.tracks[0].clips[0].durationMs, 3000);
  assert.throws(() => trimTimelineClipEdge(base, 'video-1', 'clip-1', 'left', 4000), /片段至少保留/);
});

test('变速按源时长计算画面时长并保留原对象', () => {
  const result = setClipPlaybackRate(base, 'video-1', 'clip-1', 2);
  assert.equal(result.tracks[0].clips[0].playbackRate, 2);
  assert.equal(result.tracks[0].clips[0].durationMs, 2000);
  assert.equal(base.tracks[0].clips[0].durationMs, 4000);
  assert.throws(() => setClipPlaybackRate(base, 'video-1', 'clip-1', 8), /变速范围/);
});

test('红蓝绿建议和字幕轨设置写入时间线', () => {
  const marked = markTimelineClip(base, 'video-1', 'clip-1', 'review', '需要人工确认');
  assert.equal(marked.tracks[0].clips[0].reviewStatus, 'review');
  assert.equal(marked.tracks[0].clips[0].reviewReason, '需要人工确认');
  const subtitleTimeline = { revision: 0, tracks: [{ id: 'sub-1', kind: 'subtitle', clips: [] }] };
  const styled = updateTimelineTrackSettings(subtitleTimeline, 'sub-1', { subtitleStyleId: '黄字大标题', subtitleRhythm: 'word', subtitleAnimation: 'pop', previewVisible: true });
  assert.equal(styled.tracks[0].settings.subtitleStyleId, '黄字大标题');
  assert.equal(styled.tracks[0].settings.subtitleRhythm, 'word');
});

test('字幕节奏只向后端保存内部值并把内部值显示为中文', () => {
  assert.equal(normalizeSubtitleRhythm('按句/短语'), 'sentence');
  assert.equal(normalizeSubtitleRhythm('逐词高亮'), 'word');
  assert.equal(normalizeSubtitleRhythm('word'), 'word');
  assert.equal(normalizeSubtitleRhythm('unexpected'), 'sentence');
  assert.equal(subtitleRhythmLabel('sentence'), '按句/短语');
  assert.equal(subtitleRhythmLabel('word'), '逐词高亮');
});

test('Agent 指令只匹配明确支持的工具且视觉理解不会被冒充为文字工具', () => {
  assert.deepEqual(resolveAICutAgentTool('识别全片字幕'), { tool: 'caption_asr' });
  assert.deepEqual(resolveAICutAgentTool('去掉口误和超过一秒的停顿'), { tool: 'oral_clean' });
  assert.deepEqual(resolveAICutAgentTool('按镜头粗剪并删除废镜头'), { tool: 'smart_split' });
  assert.deepEqual(resolveAICutAgentTool('给高潮段落配 BGM'), { tool: 'bgm_plan' });
  assert.deepEqual(resolveAICutAgentTool('理解画面并判断保留镜头'), { tool: 'visual_analyze' });
  assert.equal(resolveAICutAgentTool('随便处理一下'), null);
});

test('自动配乐按参考画面范围铺满并给首尾添加淡化', () => {
  const timeline = {
    revision: 2,
    tracks: [{ id: 'video-1', kind: 'video', name: '视频1', clips: [{ id: 'v1', startMs: 1000, durationMs: 5000 }] }]
  };
  let next = 0;
  const result = planTimelineBGM(timeline, { id: 'music-1', kind: 'audio', name: '钢琴.mp3', durationMs: 2000 }, {
    referenceTrackId: 'video-1', align: true, fill: true, continue: false, fade: true
  }, () => `generated-${++next}`);
  const track = result.tracks.find(item => item.name === 'BGM');
  assert.deepEqual(track.clips.map(clip => [clip.startMs, clip.durationMs, clip.sourceStartMs]), [[1000, 2000, 0], [3000, 2000, 0], [5000, 1000, 0]]);
  assert.equal(track.clips[0].fadeInMs, 500);
  assert.equal(track.clips[0].fadeOutMs, 0);
  assert.equal(track.clips[2].fadeOutMs, 500);
  assert.equal(track.clips[2].volume, 0.25);
});

test('自动配乐续接从已有 BGM 末尾开始且不铺满时只加入原时长', () => {
  const timeline = {
    revision: 1,
    tracks: [
      { id: 'video-1', kind: 'video', clips: [{ id: 'v1', startMs: 0, durationMs: 9000 }] },
      { id: 'bgm-1', kind: 'audio', name: 'BGM', clips: [{ id: 'old', startMs: 0, durationMs: 3000 }] }
    ]
  };
  const result = planTimelineBGM(timeline, { id: 'music-2', kind: 'audio', name: '鼓点.mp3', durationMs: 1500 }, {
    referenceTrackId: 'video-1', align: true, fill: false, continue: true, fade: false
  }, () => 'new');
  const clips = result.tracks.find(item => item.id === 'bgm-1').clips;
  assert.equal(clips.length, 2);
  assert.deepEqual([clips[1].startMs, clips[1].durationMs, clips[1].fadeInMs, clips[1].fadeOutMs], [3000, 1500, 0, 0]);
});

test('自动配乐拒绝空音频和没有画面的工程', () => {
  assert.throws(() => planTimelineBGM({ tracks: [] }, { id: 'music', kind: 'audio', durationMs: 1000 }, {}, () => 'id'), /参考画面轨道/);
  assert.throws(() => planTimelineBGM({ tracks: [{ id: 'v', kind: 'video', clips: [{ startMs: 0, durationMs: 1000 }] }] }, { id: 'bad', kind: 'video', durationMs: 1000 }, { referenceTrackId: 'v' }, () => 'id'), /音频素材/);
});

test('移动片段会吸附到相邻边界并联动同起点的未锁定片段', () => {
  const timeline = {
    tracks: [
      { id: 'video-1', kind: 'video', locked: false, clips: [
        { id: 'moving', startMs: 1000, durationMs: 1000 },
        { id: 'anchor', startMs: 4000, durationMs: 1000 }
      ] },
      { id: 'sub-1', kind: 'subtitle', locked: false, clips: [{ id: 'linked', startMs: 1000, durationMs: 500 }] },
      { id: 'audio-1', kind: 'audio', locked: true, clips: [{ id: 'locked-linked', startMs: 1000, durationMs: 500 }] }
    ]
  };
  const result = moveTimelineClip(timeline, 'video-1', 'moving', 3940, { magnetic: true, linked: true, thresholdMs: 100 });
  assert.equal(result.tracks[0].clips[0].startMs, 4000);
  assert.equal(result.tracks[1].clips[0].startMs, 4000);
  assert.equal(result.tracks[2].clips[0].startMs, 1000);
});

test('关闭吸附与联动时只移动目标片段并拒绝负数和锁定轨道', () => {
  const plain = moveTimelineClip(base, 'video-1', 'clip-1', 1234, { magnetic: false, linked: false });
  assert.equal(plain.tracks[0].clips[0].startMs, 1234);
  assert.throws(() => moveTimelineClip(base, 'video-1', 'clip-1', -1), /开始时间不能小于 0/);
  const locked = { ...base, tracks: [{ ...base.tracks[0], locked: true }] };
  assert.throws(() => moveTimelineClip(locked, 'video-1', 'clip-1', 2000), /轨道已锁定/);
});

test('时间和像素换算支持缩放，拖动结果不会小于零', () => {
  assert.equal(timelineTimeToPixels(2000, 1), 144);
  assert.equal(timelineTimeToPixels(2000, 2), 288);
  assert.equal(timelinePixelsToTime(72, 1), 1000);
  assert.equal(timelinePixelsToTime(72, 2), 500);
  assert.equal(draggedTimelineStart(1000, -200, 1), 0);
  assert.equal(draggedTimelineStart(1000, 72, 1), 2000);
});

test('特效作为独立区间添加、移动、裁剪和删除且不修改原时间线', () => {
  const added = addEffectLayer(base, { resourceId: 'effect-glow', startMs: 1200, durationMs: 800, parameters: { intensity: .5 } }, () => 'effect-1');
  assert.equal(added.ok, true);
  const effectTrack = added.timeline.tracks.find(track => track.kind === 'effect');
  assert.deepEqual(effectTrack.clips[0], { id: 'effect-1', kind: 'effect', resourceId: 'effect-glow', name: 'effect-glow', startMs: 1200, durationMs: 800, parameters: { intensity: .5 } });
  assert.equal(base.tracks.length, 1);
  const moved = moveEffectLayer(added.timeline, effectTrack.id, 'effect-1', 2000);
  const resized = resizeEffectLayer(moved.timeline, effectTrack.id, 'effect-1', 1200);
  const removed = removeTimelineObject(resized.timeline, effectTrack.id, 'effect-1');
  assert.equal(removed.timeline.tracks.find(track => track.id === effectTrack.id).clips.length, 0);
});

test('锁定特效轨拒绝修改并返回中文原因', () => {
  const source = { tracks: [{ id: 'fx', kind: 'effect', locked: true, clips: [{ id: 'e1', kind: 'effect', startMs: 0, durationMs: 500 }] }] };
  assert.deepEqual(moveEffectLayer(source, 'fx', 'e1', 1000), { ok: false, timeline: source, reasonZh: '轨道已锁定' });
});

test('转场只允许相邻片段并把过长时长截断到双方一半', () => {
  const source = { tracks: [{ id: 'v', kind: 'video', locked: false, clips: [
    { id: 'a', startMs: 0, durationMs: 1000 }, { id: 'b', startMs: 1000, durationMs: 600 },
  ] }] };
  const result = setClipTransition(source, 'v', 'b', 'transition-cross-dissolve', 900);
  assert.equal(result.ok, true);
  assert.equal(result.timeline.tracks[0].clips[1].transitionMs, 300);
  const gap = { tracks: [{ ...source.tracks[0], clips: [{ id: 'a', startMs: 0, durationMs: 500 }, { id: 'b', startMs: 1000, durationMs: 600 }] }] };
  assert.equal(setClipTransition(gap, 'v', 'b', 'transition-cross-dissolve', 300).reasonZh, '两个片段之间有空隙');
});
