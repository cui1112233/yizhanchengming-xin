import test from 'node:test';
import assert from 'node:assert/strict';
import { readStoryboardPromptEditableText, storyboardPromptTextForSave } from './storyboardPromptEditor.js';

test('reads a rich storyboard editor without losing its line boundaries', () => {
  const editor = { innerText: '基础设定：人物：林晚\n视频提示词：\n镜头缓慢推进\u00a0' };

  assert.equal(
    readStoryboardPromptEditableText(editor),
    '基础设定：人物：林晚\n视频提示词：\n镜头缓慢推进 '
  );
});

test('uses the current rich editor text when saving a storyboard prompt', () => {
  const editor = { innerText: '镜头一\n用户刚输入的修改' };

  assert.equal(storyboardPromptTextForSave(editor, '过期的状态值'), '镜头一\n用户刚输入的修改');
});
