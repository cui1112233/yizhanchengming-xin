import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

import {
  audioJobRequestForNode,
  audioOperationNeedsModel,
  audioSourceAsset
} from './agent-canvas-audio-jobs.js';
import { createAgentAudioGeneration } from '../../shared/api/agent.js';

const dirname = path.dirname(fileURLToPath(import.meta.url));
const canvasSource = fs.readFileSync(path.resolve(dirname, '../components/AgentCanvas.jsx'), 'utf8');
const composerSource = fs.readFileSync(path.resolve(dirname, '../components/AgentNodeComposer.jsx'), 'utf8');

function audioNode(operation, overrides = {}) {
  return {
    id: 'audio-1',
    data: {
      type: 'audio', version: 3, assetRef: null,
      config: {
        provider: 'personal_api', modelId: 'audio-model', prompt: '雨夜旁白',
        parameters: { operation, speed: 1, pitch: 0, emotion: 'neutral', format: 'mp3' },
        advanced: {}, referenceAssets: [],
        ...overrides
      }
    }
  };
}

test('音频生成请求携带当前节点模型和独立参数', () => {
  const payload = audioJobRequestForNode(audioNode('speech'), { taskId: 'project-1', attempt: 2 });
  assert.equal(payload.nodeId, 'audio-1');
  assert.equal(payload.nodeVersion, 3);
  assert.equal(payload.operation, 'speech');
  assert.equal(payload.prompt, '雨夜旁白');
  assert.equal(payload.provider, 'personal_api');
  assert.equal(payload.modelId, 'audio-model');
  assert.equal(payload.sourceAssetId, '');
  assert.equal(payload.settings.speed, 1);
  assert.equal(payload.idempotencyKey, 'agent-audio:project-1:audio-1:3:2');
  assert.equal(audioOperationNeedsModel('speech'), true);
  assert.equal(audioOperationNeedsModel('sound_effect'), true);
});

test('音频处理使用当前音频或所选音频素材且不依赖模型', () => {
  const current = audioNode('trim', { modelId: '', prompt: '', assetRef: { id: 'ignored' } });
  current.data.assetRef = { id: 'asset-current', kind: 'audio', name: '当前音频.mp3' };
  let payload = audioJobRequestForNode(current, { taskId: 'project-1' });
  assert.equal(payload.sourceAssetId, 'asset-current');
  assert.equal(audioOperationNeedsModel('trim'), false);

  const referenced = audioNode('speed', {
    modelId: '', prompt: '',
    referenceAssets: [{ id: 'image-1', kind: 'image' }, { id: 'audio-reference', kind: 'audio' }]
  });
  assert.equal(audioSourceAsset(referenced)?.id, 'audio-reference');
  payload = audioJobRequestForNode(referenced, { taskId: 'project-1' });
  assert.equal(payload.sourceAssetId, 'audio-reference');
});

test('音效旧值会转换成后端接受的 sound_effect', () => {
  const payload = audioJobRequestForNode(audioNode('effect'), { taskId: 'project-1' });
  assert.equal(payload.operation, 'sound_effect');
});

test('音频 API 使用正式项目音频任务入口', async () => {
  const calls = [];
  const result = await createAgentAudioGeneration('project 1', { nodeId: 'audio-1' }, async (url, options) => {
    calls.push({ url, options });
    return { task: { id: 'generation-1' } };
  });
  assert.equal(result.task.id, 'generation-1');
  assert.equal(calls[0].url, '/api/agent/tasks/project%201/audio-generations');
  assert.equal(calls[0].options.method, 'POST');
});

test('画布和编辑器接入音频提交、试听和结果确认', () => {
  assert.match(canvasSource, /createAgentAudioGeneration/);
  assert.match(canvasSource, /submitAudioForNode/);
  assert.match(canvasSource, /confirmAgentGeneration/);
  assert.match(composerSource, /AuthenticatedAgentAudio/);
  assert.match(composerSource, /onConfirmAudio/);
});
