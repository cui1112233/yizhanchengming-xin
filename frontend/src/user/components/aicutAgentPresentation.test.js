import test from 'node:test';
import assert from 'node:assert/strict';
import { agentFollowupFromMessages, agentStepRows, formatAgentToolMessage } from './aicutAgentPresentation.js';

test('followup survives reload and is superseded by a later answer', () => {
 const question = { question: '保留哪一段？', fields: [{ id: 'topic', label: '主题', required: true }] };
 const messages = [{ role: 'assistant', content: `请回答\n${JSON.stringify(question)}` }];
 assert.equal(agentFollowupFromMessages(messages).question, question.question);
 assert.equal(agentFollowupFromMessages([...messages, { role: 'user', content: '保留产品介绍' }]), null);
});
test('steps show one row per call with success and readable arguments', () => {
 const events = [{ sequence: 1, type: 'tool_start', toolCallId: 'a', toolName: 'move_clip', message: '开始' }, { sequence: 2, type: 'tool_done', toolCallId: 'a', toolName: 'move_clip', message: '已移动', revision: 2 }];
 const rows = agentStepRows(events);
 assert.equal(rows.length, 1); assert.equal(rows[0].state, 'done'); assert.equal(rows[0].revision, 2);
 assert.equal(formatAgentToolMessage(JSON.stringify({ name: 'move_clip', ok: false, message: '轨道已锁定' })), '移动片段：轨道已锁定');
});
