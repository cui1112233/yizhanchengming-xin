import assert from 'node:assert/strict';
import test from 'node:test';
import { audioEffectOptions, audioProcessingPreviewPlan, normalizeAudioProcessing } from './aicutAudioProcessing.js';

test('音频处理默认值和边界归一化', () => {
  assert.deepEqual(normalizeAudioProcessing(), { denoise: 0, eqLowDB: 0, eqMidDB: 0, eqHighDB: 0, channelMode: 'stereo', loudnessEnabled: false, targetLUFS: -16, voiceEffect: 'natural', pitchSemitones: 0 });
  assert.deepEqual(normalizeAudioProcessing({ denoise: 2, eqLowDB: -20, eqMidDB: 3, eqHighDB: 99, channelMode: 'bad' }), {
    denoise: 1, eqLowDB: -12, eqMidDB: 3, eqHighDB: 12, channelMode: 'stereo', loudnessEnabled: false, targetLUFS: -16, voiceEffect: 'natural', pitchSemitones: 0
  });
  assert.equal(normalizeAudioProcessing({ targetLUFS: 0 }).targetLUFS, -16, '旧工程的零值必须恢复为安全默认值');
  assert.equal(normalizeAudioProcessing({ voiceEffect: 'robot' }).voiceEffect, 'robot');
  assert.equal(normalizeAudioProcessing({ voiceEffect: 'unknown' }).voiceEffect, 'natural');
  assert.equal(normalizeAudioProcessing({ pitchSemitones: 99 }).pitchSemitones, 12);
  assert.equal(normalizeAudioProcessing({ pitchSemitones: -99 }).pitchSemitones, -12);
});

test('完整低内存音频预设目录中的每一项都能进入工程数据', () => {
  const expected = [
    ['natural', '原声'], ['enhance', '人声增强'], ['bright', '清亮'], ['bass', '低音增强'], ['treble', '高音增强'],
    ['telephone', '电话声'], ['radio', '收音机'], ['megaphone', '扩音器'], ['robot', '机器人声'], ['chipmunk', '花栗鼠声'],
    ['deep', '低沉声'], ['sweet', '甜美声'], ['whisper', '耳语声'], ['distorted', '失真声'], ['lofi', '复古低保真'],
    ['echo', '回声'], ['room', '房间混响'], ['hall', '大厅混响'], ['plate', '板式混响'], ['cathedral', '教堂混响'],
    ['stadium', '体育场混响'], ['underwater', '水下声'], ['vinyl', '黑胶唱片'],
  ];
  assert.deepEqual(audioEffectOptions.map(({ value, label }) => [value, label]), expected);
  for (const [value] of expected) assert.equal(normalizeAudioProcessing({ voiceEffect: value }).voiceEffect, value);
});

test('扩展音频预设生成不同的低耗试听计划', () => {
  assert.equal(audioProcessingPreviewPlan({ voiceEffect: 'bass' }).eqLowDB, 8);
  assert.equal(audioProcessingPreviewPlan({ voiceEffect: 'treble' }).eqHighDB, 7);
  assert.equal(audioProcessingPreviewPlan({ voiceEffect: 'radio' }).distortionAmount, 22);
  assert.equal(audioProcessingPreviewPlan({ voiceEffect: 'echo' }).echoDelayMS, 220);
  assert.equal(audioProcessingPreviewPlan({ voiceEffect: 'cathedral' }).ambienceDelayMS, 95);
  assert.equal(audioProcessingPreviewPlan({ voiceEffect: 'underwater' }).lowpassHz, 1100);
});

test('通用音调参数转换为稳定比例供试听提示和服务端导出', () => {
  assert.equal(audioProcessingPreviewPlan({ pitchSemitones: 12 }).pitchRatio, 2);
  assert.equal(audioProcessingPreviewPlan({ pitchSemitones: -12 }).pitchRatio, .5);
  assert.equal(audioProcessingPreviewPlan({ pitchSemitones: 0 }).pitchRatio, 1);
});

test('浏览器试听计划把降噪和响度匹配转换为低耗处理并保留均衡器与声道', () => {
  assert.deepEqual(audioProcessingPreviewPlan({ denoise: .5, eqLowDB: 2, eqMidDB: -1, eqHighDB: 4, channelMode: 'right', loudnessEnabled: true, targetLUFS: -14 }), {
    highpassHz: 130,
    lowpassHz: 17000,
    eqLowDB: 2,
    eqMidDB: -1,
    eqHighDB: 4,
    channelMode: 'right',
    loudnessEnabled: true,
    targetLUFS: -14,
    compressorThresholdDB: -24,
    compressorRatio: 3,
    makeupGainDB: 2,
    voiceEffect: 'natural',
    voicePresenceDB: 0,
    distortionAmount: 0,
    modulationHz: 0,
    modulationDepthMS: 0,
    echoDelayMS: 0,
    echoFeedback: 0,
    echoWet: 0,
    ambienceDelayMS: 0,
    ambienceFeedback: 0,
    ambienceWet: 0,
    pitchSemitones: 0,
    pitchRatio: 1
  });
});

test('五种人声预设产生稳定且不同的浏览器试听计划', () => {
  const enhance = audioProcessingPreviewPlan({ voiceEffect: 'enhance' });
  assert.equal(enhance.highpassHz, 80);
  assert.equal(enhance.lowpassHz, 14000);
  assert.equal(enhance.voicePresenceDB, 4);
  assert.equal(enhance.compressorRatio, 3);
  const telephone = audioProcessingPreviewPlan({ voiceEffect: 'telephone' });
  assert.equal(telephone.highpassHz, 300);
  assert.equal(telephone.lowpassHz, 3400);
  const robot = audioProcessingPreviewPlan({ voiceEffect: 'robot' });
  assert.ok(robot.distortionAmount > 0);
  assert.ok(robot.modulationHz > 0);
  const chipmunk = audioProcessingPreviewPlan({ voiceEffect: 'chipmunk' });
  assert.ok(chipmunk.voicePresenceDB > enhance.voicePresenceDB);
  assert.ok(chipmunk.modulationDepthMS > 0);
});
