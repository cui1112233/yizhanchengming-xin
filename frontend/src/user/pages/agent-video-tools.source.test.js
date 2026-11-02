import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';

const root = path.resolve(import.meta.dirname, '..');
const composer = fs.readFileSync(path.join(root, 'components', 'AgentNodeComposer.jsx'), 'utf8');
const canvas = fs.readFileSync(path.join(root, 'components', 'AgentCanvas.jsx'), 'utf8');
const api = fs.readFileSync(path.join(root, '..', 'shared', 'api', 'agent.js'), 'utf8');

test('video tools create a linked result node and submit only after the canvas is durable', () => {
  assert.match(composer, /AgentVideoTools/);
  assert.match(canvas, /function submitVideoTool\(nodeID, input\)/);
  assert.match(canvas, /createNodeAt\(/);
  assert.match(canvas, /runAfterDurableCanvasSave\(onFlushCanvas/);
  assert.match(canvas, /createAgentVideoTool\(requestedTaskId/);
  assert.match(canvas, /relation: 'reference'/);
  assert.match(api, /video-tools/);
});

test('video tool results use the common generation confirmation path', () => {
  assert.match(composer, /isVideoToolTask/);
  assert.match(composer, /onConfirmImage\?\.\(node\.id, selectedVideoToolID\)/);
  assert.match(canvas, /String\(task\.mode \|\| ''\)\.startsWith\('video_'\)/);
});

test('video generation handler is available to every rendered canvas node', () => {
  assert.match(canvas, /submitTextForNode, submitImageForNode, submitImageTool, submitVideoTool, submitVideoForNode, submitAudioForNode/);
});
