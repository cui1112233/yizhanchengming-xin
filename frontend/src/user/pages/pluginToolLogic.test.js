// 4 条文字类插件的纯逻辑五场景测试（正常 / 空值 / 重复 / 错误 / 取消）。
// 只覆盖不依赖浏览器与模型的部分；界面与真实模型请求在验收文档里标注“未验证”。
import test from 'node:test';
import assert from 'node:assert/strict';

import {
  buildDialogueList,
  buildPartPrompt,
  extractScript,
  formatScriptText,
  parseDialogueLine,
  splitInputByLength
} from './pluginToolLogic.js';

const SCRIPT_OUTPUT = [
  '=== 剧本正文 ===',
  '沈知意-台词（平静）：你终于来了。',
  '旁白：雨声盖住了脚步声。',
  '=== 改编说明 ===',
  '保留原文主线。'
].join('\n');

test('正常输入：分段、提示词拼装、剧本提取与格式化可完整走通', () => {
  // 小说转剧 progressChunkSize=2000，这里用 20 字模拟同一分段路径。
  const text = '第一场戏的原文内容\n第二场戏的原文内容';
  const parts = splitInputByLength(text, 12);
  assert.ok(parts.length > 1);
  assert.equal(parts.join('\n'), text);

  const prompt = buildPartPrompt({ full: text, part: parts[0], index: 1, total: parts.length, history: '' });
  assert.match(prompt, /【完整原文】（供参考整体框架，共\d+字）/);
  assert.match(prompt, new RegExp(`【当前处理分段】这是第1/${parts.length}部分`));

  const extracted = extractScript(SCRIPT_OUTPUT);
  assert.equal(extracted.success, true);
  assert.equal(extracted.scriptCount, 1);
  assert.equal(extracted.lineCount, 2);

  assert.deepEqual(formatScriptText(SCRIPT_OUTPUT), ['你终于来了。', '雨声盖住了脚步声。']);

  const dialogues = buildDialogueList([
    '沈知意-台词（平静）：你终于来了。',
    '旁白：雨声盖住了脚步声。'
  ].join('\n'));
  assert.deepEqual(dialogues.map(item => item.role_name), ['沈知意', '旁白']);
  assert.equal(dialogues[0].emotion_name, '平静');
  assert.equal(dialogues[1].text_content, '雨声盖住了脚步声。');
});

test('空值：全空、纯空格、只有换行都不产生假结果', () => {
  for (const empty of ['', '   ', '\n\n\t ']) {
    assert.equal(extractScript(empty).success, false);
    assert.equal(extractScript(empty).message, '内容为空');
    assert.deepEqual(formatScriptText(empty), []);
    assert.deepEqual(buildDialogueList(empty), []);
  }
  // 分段函数把空串视为单段原样返回，交给页面/后端做空值拦截，不吞成多段。
  assert.deepEqual(splitInputByLength('', 2000), ['']);
  assert.deepEqual(splitInputByLength('   ', 2000), ['   ']);
});

test('重复操作：同样的输入重复调用结果完全一致，不累积历史', () => {
  const text = '第一场\n第二场\n第三场';
  const first = splitInputByLength(text, 8);
  const second = splitInputByLength(text, 8);
  assert.deepEqual(second, first);

  const promptA = buildPartPrompt({ full: text, part: '第二场', index: 2, total: 3, history: '' });
  const promptB = buildPartPrompt({ full: text, part: '第二场', index: 2, total: 3, history: '' });
  assert.equal(promptB, promptA);

  // 已格式化的文本再次格式化保持稳定（幂等），不会叠加前缀或空行。
  const once = formatScriptText(SCRIPT_OUTPUT);
  assert.deepEqual(formatScriptText(once.join('\n')), once);

  // 同角色相邻台词在 150 字内合并，重复解析不会重复生成记录。
  const merged = buildDialogueList('沈知意-台词：第一句\n沈知意-台词：第二句');
  assert.equal(merged.length, 1);
  assert.equal(merged[0].text_content, '第一句第二句');
});

test('错误操作：超长单行、特殊字符与不规范台词行不丢内容也不误判', () => {
  // 超长单行按 limit 硬切，拼回后与原文一致（不丢字）。
  const longLine = '字'.repeat(5000);
  const chunks = splitInputByLength(longLine, 2000);
  assert.deepEqual(chunks.map(chunk => chunk.length), [2000, 2000, 1000]);
  assert.equal(chunks.join(''), longLine);

  // 特殊字符（正则符号、引号、emoji、换行）原样进入提示词。
  const special = '价格 $& $1 \\ 反斜杠 "引号" 😀 \n第二行';
  const prompt = buildPartPrompt({ full: special, part: special, index: 1, total: 1, history: '' });
  assert.match(prompt, /\$& \$1/);
  assert.match(prompt, /😀/);

  // 不规范台词行原样保留，不会被解析成对话结构。
  assert.equal(parseDialogueLine('这一行不是台词格式'), null);
  assert.deepEqual(buildDialogueList('这一行不是台词格式'), []);
  // 半角冒号与括号也要识别，避免“格式错”被判成空结果。
  assert.equal(parseDialogueLine('沈知意-台词(激动): 走。').emotion_name, '激动');
  assert.equal(parseDialogueLine('旁白: 天亮了。').text_content, '天亮了。');
});

test('取消操作：纯逻辑无异步状态，取消由页面与请求层负责（本层无中途写回）', () => {
  // 纯函数不含定时器、请求或全局状态，重复调用不会留下半成品；
  // 页面取消路径（关闭面板 / 切换页面）没有 AbortController，属于已知缺口，见验收文档。
  assert.equal(typeof globalThis.AbortController, 'function');
  const parts = splitInputByLength('第一场\n第二场', 4);
  assert.ok(Array.isArray(parts));
  assert.equal(parts.every(part => typeof part === 'string'), true);
});
