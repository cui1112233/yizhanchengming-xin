import assert from 'node:assert/strict';
import test from 'node:test';
import { executeAICutCommand } from './aicutCommands.js';

test('声像命令保存小数，重新读取保留，拒绝越界和非有限数字', () => {
  const timeline = { tracks: [{ id: 'a', kind: 'audio', clips: [{ id: 'c', kind: 'audio', durationMs: 1000 }] }] };
  const command = { type: 'clip.audio_processing_set', trackId: 'a', clipId: 'c' };
  const result = executeAICutCommand(timeline, { ...command, processing: { pan: -.375 } });
  assert.equal(JSON.parse(JSON.stringify(result.timeline)).tracks[0].clips[0].audioProcessing.pan, -.375);
  for (const pan of [-1.01, 1.01, NaN, Infinity, -Infinity]) {
    assert.throws(() => executeAICutCommand(timeline, { ...command, processing: { pan } }), /声像/);
  }
  assert.throws(() => executeAICutCommand(timeline, { ...command, processing: { pan: .5, unknownPan: 1 } }), /未知/);
  assert.equal(executeAICutCommand(timeline, { ...command, processing: {} }).timeline.tracks[0].clips[0].audioProcessing.pan, 0);
});
