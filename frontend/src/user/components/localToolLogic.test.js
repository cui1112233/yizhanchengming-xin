import assert from 'node:assert/strict';
import test from 'node:test';
import {
  buildSrt,
  detectSubtitleFormat,
  extractChapterNumber,
  formatText,
  mergeDocuments,
  mergeSubtitles,
  parseChineseNumber,
  parseSubtitle,
  parseSubtitleTime
} from './localToolLogic.js';

test('格式化文本：清理空白、中文标点、中文间空格', () => {
  const result = formatText('  他说,"你好,世界!"  \n\n\n  这是   第二段...  \n');
  assert.equal(result.text, '他说，“你好，世界！”\n\n这是第二段……');
  assert.equal(result.stats.linesAfter, 2);
});

test('格式化文本：数字与英文标点不被改坏', () => {
  const result = formatText('版本 3.5 发布于 2024,01,02\nHello, world. It works!');
  assert.equal(result.text, '版本 3.5 发布于 2024,01,02\nHello, world. It works!');
});

test('格式化文本：删除空行与段落缩进可独立开关', () => {
  const result = formatText('第一段\n\n\n第二段', { removeBlank: true, indentParagraphs: true });
  assert.equal(result.text, '　　第一段\n　　第二段');
});

test('字幕解析：SRT 时间与内容', () => {
  const srt = '1\n00:00:01,500 --> 00:00:03,000\n你好\n\n2\n00:00:04,000 --> 00:00:05,000\n再见\n';
  assert.equal(parseSubtitleTime('00:00:01,500'), 1500);
  const parsed = parseSubtitle(srt, { fileName: 'a.srt' });
  assert.equal(parsed.format, 'srt');
  assert.equal(parsed.cues.length, 2);
  assert.equal(parsed.cues[0].end, 3000);
  assert.equal(parsed.cues[0].text, '你好');
});

test('字幕解析：VTT 与 ASS', () => {
  assert.equal(detectSubtitleFormat('a.vtt', ''), 'vtt');
  const vtt = 'WEBVTT\n\n00:00:01.000 --> 00:00:02.000\n<c.yellow>字幕</c>\n';
  assert.deepEqual(parseSubtitle(vtt, { fileName: 'a.vtt' }).cues, [{ start: 1000, end: 2000, text: '字幕' }]);

  const ass = [
    '[Script Info]',
    'Title: demo',
    '[Events]',
    'Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text',
    'Dialogue: 0,0:00:01.00,0:00:02.50,Default,,0,0,0,,{\\an8}第一行\\N第二行',
    'Dialogue: 0,0:00:03.00,0:00:04.00,Default,,0,0,0,,再来, 一句'
  ].join('\n');
  const parsed = parseSubtitle(ass, { fileName: 'a.ass' });
  assert.equal(parsed.format, 'ass');
  assert.equal(parsed.cues.length, 2);
  assert.equal(parsed.cues[0].text, '第一行\n第二行');
  assert.equal(parsed.cues[0].end, 2500);
  assert.equal(parsed.cues[1].text, '再来, 一句');
});

test('字幕合并：按时间重排并重新编号', () => {
  const a = '1\n00:00:05,000 --> 00:00:06,000\n后一句\n';
  const b = '1\n00:00:01,000 --> 00:00:02,000\n前一句\n';
  const result = mergeSubtitles([
    { name: 'a.srt', content: a },
    { name: 'b.srt', content: b }
  ], { mode: 'sort' });
  assert.equal(result.stats.cues, 2);
  assert.match(result.text, /^1\n00:00:01,000 --> 00:00:02,000\n前一句/);
  assert.match(result.text, /2\n00:00:05,000 --> 00:00:06,000\n后一句/);
});

test('字幕合并：偏移、拼接与去重', () => {
  const a = '1\n00:00:01,000 --> 00:00:02,000\n同一句\n';
  const b = '1\n00:00:01,000 --> 00:00:02,000\n同一句\n2\n00:00:03,000 --> 00:00:04,000\n第二句\n';

  const deduped = mergeSubtitles([
    { name: 'a.srt', content: a },
    { name: 'b.srt', content: b }
  ], { mode: 'sort', dedupe: true });
  assert.equal(deduped.stats.duplicates, 1);
  assert.equal(deduped.stats.cues, 2);

  const shifted = mergeSubtitles([{ name: 'a.srt', content: a, offsetMs: 2000 }], { mode: 'sort' });
  assert.equal(shifted.cues[0].start, 3000);

  const concatenated = mergeSubtitles([
    { name: 'a.srt', content: a },
    { name: 'b.srt', content: b }
  ], { mode: 'concat', gapMs: 500 });
  assert.equal(concatenated.cues.length, 3);
  assert.equal(concatenated.cues[1].start, 2500);
  assert.equal(concatenated.cues[2].start, 4500);
});

test('字幕合并：非法片段计入丢弃数', () => {
  const broken = '1\n00:00:05,000 --> 00:00:02,000\n时间反了\n';
  const result = mergeSubtitles([{ name: 'x.srt', content: broken }], { mode: 'sort' });
  assert.equal(result.stats.dropped, 1);
  assert.equal(result.stats.cues, 0);
  assert.equal(result.text, '');
});

test('字幕输出：buildSrt 补齐毫秒位', () => {
  assert.equal(buildSrt([{ start: 1000, end: 61050, text: '内容' }]), '1\n00:00:01,000 --> 00:01:01,050\n内容\n');
});

test('章节号识别：中文数字与文件名', () => {
  assert.equal(parseChineseNumber('二十三'), 23);
  assert.equal(parseChineseNumber('一百零五'), 105);
  assert.equal(parseChineseNumber('12'), 12);
  assert.equal(extractChapterNumber('第12章 归来'), 12);
  assert.equal(extractChapterNumber('第三章 归来'), 3);
  assert.equal(extractChapterNumber('Chapter 7'), 7);
  assert.equal(extractChapterNumber('序章'), null);
});

test('文档合并：按章节号排序并加标题分隔', () => {
  const result = mergeDocuments([
    { name: '第10章.txt', content: '第十章内容' },
    { name: '第2章.txt', content: '第二章内容\n\n\n' },
    { name: '第1章.txt', content: '第一章内容' }
  ], { order: 'chapter', separator: 'title' });
  assert.equal(result.stats.perFile.map(item => item.chapter).join(','), '1,2,10');
  assert.equal(result.text, '=== 第1章 ===\n第一章内容\n\n=== 第2章 ===\n第二章内容\n\n=== 第10章 ===\n第十章内容\n');
});

test('文档合并：自然排序与保持上传顺序', () => {
  const files = [
    { name: 'part10.txt', content: 'C' },
    { name: 'part2.txt', content: 'B' },
    { name: 'Part1.txt', content: 'A' }
  ];
  assert.equal(mergeDocuments(files, { order: 'natural', separator: 'none', collapseBlank: false }).text, 'A\n\nB\n\nC\n');
  assert.equal(mergeDocuments(files, { order: 'upload', separator: 'none', collapseBlank: false }).text, 'C\n\nB\n\nA\n');
});

test('文档合并：重复段落可去掉', () => {
  const result = mergeDocuments([
    { name: 'a.txt', content: '广告\n\n正文一\n\n正文一' }
  ], { order: 'upload', separator: 'none', dedupeParagraphs: true });
  assert.equal(result.text, '广告\n\n正文一\n');
});
