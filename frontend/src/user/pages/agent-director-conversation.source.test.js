import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

const source = readFileSync(new URL('../components/AgentDirectorConversation.jsx', import.meta.url), 'utf8');
const page = readFileSync(new URL('./AgentPageV2.jsx', import.meta.url), 'utf8');

test('storyboard director conversation exposes modes and review actions', () => {
  assert.match(source, /逐步确认/);
  assert.match(source, /自动执行/);
  assert.match(source, /确认执行/);
  assert.match(source, /反馈调整/);
  assert.match(source, /directorPlanPresentation/);
});

test('storyboard uses a compact director composer instead of the full general chat toolbar', () => {
  assert.match(page, /is-director/);
  assert.match(page, /告诉导演你想完成什么/);
  assert.match(page, /viewState\.activeView !== 'storyboard'/);
});

test('director feedback is sent back as a new planned request instead of a local notification', () => {
  assert.match(page, /feedbackToPlanId/);
  assert.match(page, /重新整理一个待确认的画布方案/);
  assert.match(page, /detectDirectorMode/);
});
