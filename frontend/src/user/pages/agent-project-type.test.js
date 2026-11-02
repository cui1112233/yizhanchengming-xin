import test from 'node:test';
import assert from 'node:assert/strict';
import { agentProjectCapabilities, agentProjectTypeLabel, normalizeAgentProjectType } from './agent-project-type.js';

test('项目类型缺失或未知时按视频兼容', () => {
  assert.equal(normalizeAgentProjectType(''), 'video');
  assert.equal(normalizeAgentProjectType('unknown'), 'video');
  assert.equal(agentProjectTypeLabel('video'), '视频作品');
});

test('漫画项目只开放剧本、文本和图片链路', () => {
  assert.deepEqual(agentProjectCapabilities('comic'), {
    nodeTypes: ['script', 'text', 'image'],
    modelKinds: ['text', 'image'],
    allowVideoGeneration: false,
    allowAudio: false
  });
});

test('视频项目保留完整媒体链路', () => {
  assert.equal(agentProjectCapabilities('video').allowVideoGeneration, true);
  assert.deepEqual(agentProjectCapabilities('video').nodeTypes, ['script', 'text', 'image', 'video', 'audio']);
});
