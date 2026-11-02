import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const dirname = path.dirname(fileURLToPath(import.meta.url));
const canvasSource = fs.readFileSync(path.resolve(dirname, '../components/AgentCanvas.jsx'), 'utf8');
const composerSource = fs.readFileSync(path.resolve(dirname, '../components/AgentNodeComposer.jsx'), 'utf8');

test('画布加载并轮询当前项目的统一生成任务', () => {
  assert.match(canvasSource, /listAgentGenerations/);
  assert.match(canvasSource, /getAgentGeneration/);
  assert.match(canvasSource, /mergeLatestGenerationTasksByNode/);
  assert.match(canvasSource, /shouldPollGeneration/);
  assert.match(canvasSource, /shouldAcceptGenerationResponse/);
});

test('画布读取并展示服务端的项目任务汇总', () => {
  assert.match(canvasSource, /getAgentGenerationSummary/);
  assert.match(canvasSource, /setGenerationSummary/);
  assert.match(canvasSource, /agent-canvas-generation-summary/);
  assert.match(canvasSource, /generationSummaryText\(generationSummary\)/);
});

test('节点输入框消费统一任务状态并恢复待确认结果', () => {
  assert.match(canvasSource, /mergeGenerationTask/);
  assert.match(canvasSource, /generationPresentation/);
  assert.match(composerSource, /generationState\.presentation/);
  assert.match(composerSource, /generationState\.candidates/);
  assert.match(canvasSource, /result\.generation/);
});

test('图片重新生成会先读取最新失败任务再计算新的幂等编号', () => {
  const submitImage = canvasSource.match(/async function submitImageForNode\(nodeID, selectedModel, draftPrompt\) \{([\s\S]*?)\n  \}/);
  assert.ok(submitImage, '缺少图片生成提交入口');
  assert.match(submitImage[1], /const latestTasks = await refreshGenerationTasks\(requestedTaskId\);/);
  assert.match(submitImage[1], /latestGenerationTaskForNode\(latestTasks, nodeID, 'image'\)/);
  assert.match(submitImage[1], /nextNodeGenerationAttempt/);
});

test('统一任务列表加载失败只显示一次画布级提示，不污染每个节点', () => {
  assert.doesNotMatch(canvasSource, /error:\s*localGeneration\.error[\s\S]{0,180}generationTasksLoadError/);
  assert.match(canvasSource, /className="agent-canvas-global-error"/);
  assert.match(canvasSource, /\{generationTasksLoadError\}/);
});

test('图片和音频使用统一生成任务取消重试，视频仍保持专用任务通道', () => {
  assert.match(canvasSource, /cancelAgentGeneration/);
  assert.match(canvasSource, /retryAgentGeneration/);
  assert.match(canvasSource, /async function cancelGeneration\(nodeID\)/);
  assert.match(canvasSource, /async function retryGeneration\(nodeID\)/);
  assert.match(composerSource, /generationState\.presentation\?\.canCancel/);
  assert.match(composerSource, /generationState\.presentation\?\.canRetry/);
  assert.match(composerSource, /onCancelGeneration/);
  assert.match(composerSource, /onRetryGeneration/);
  assert.match(composerSource, /kind === 'video' && videoState\.canCancel/);
  assert.match(composerSource, /kind === 'video' && videoState\.canRetry/);
});
