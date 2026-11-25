import test from 'node:test';
import assert from 'node:assert/strict';
import {comicScriptEpisodes, comicEpisodesDocument} from './comic-screenplay-stages.js';
import {buildStoryboardInput} from './comic-screenplay-storyboard-input.js';
import {splitDocumentEpisodes, chapterHeadingMatches, splitSourceChapters} from './comic-screenplay-episode-headings.js';
import {matchStoryboardChapter} from './comic-screenplay-storyboard-settings.js';

test('普通与Markdown中文集标题保留前言、空集位置，正文提及和章回不切集',()=>{
 const text='故事前言\r\n第1集：相遇\r\n对白：第2集会再见。\r\n第2集会再见。\r\n第十二章 章节文字\r\n第九回 回目\r\n## 第 二 集\r\n### 第十二集 转折\r\n完整尾句。';
 const rows=comicScriptEpisodes(text);
 assert.deepEqual(rows.map(e=>e.no),[1,2,12]);
 assert.equal(rows[0].prefix,'故事前言');assert.equal(rows[0].title,'相遇');
 assert.equal(rows[1].content,'');assert.equal(rows[2].content,'完整尾句。');
 assert.ok(rows[0].content.includes('第2集会再见。'));assert.ok(rows[0].content.includes('第十二章 章节文字'));
 assert.ok(comicEpisodesDocument(rows).startsWith('故事前言\n\n# 第1集 相遇'));
 assert.deepEqual(comicScriptEpisodes(comicEpisodesDocument(rows)).map(e=>e.no),[1,2,12]);
});
test('中文百十与逐位数字保持真实集号',()=>{
 assert.deepEqual(comicScriptEpisodes('第十一集\na\n第两百零三集：标题\nb\n第 二〇四 集\nc').map(e=>e.no),[11,203,204]);
});
test('章节文本匹配使用同一集标题规则，章回和正文不被删掉',()=>{
 assert.equal(matchStoryboardChapter([{id:'a',script_text:'## 第十二集：标题\r\n正文。'}],{content:'正文。'}),'a');
 assert.equal(matchStoryboardChapter([{id:'b',script_text:'第十二章\n正文。'}],{content:'正文。'}),'');
 assert.equal(matchStoryboardChapter([{id:'c',script_text:'第2集会再见。\n正文。'}],{content:'正文。'}),'');
});
test('生成输入保留前言和空集号，重复集号在转换成映射之前拒绝',()=>{
 const w={project:{model_id:'text'},document:{content:'前言\n第1集\n正文\n第二集\n第三集\n尾句'}};
 assert.deepEqual(buildStoryboardInput(w,0,{},'e').episodes,{'1':'前言\n\n正文','2':'','3':'尾句'});
 assert.throws(()=>buildStoryboardInput({...w,document:{content:'第1集\n甲\n# 第一集\n乙'}},0,{},'e'),/重复.*1|1.*重复/);
 assert.throws(()=>buildStoryboardInput({...w,document:{content:'第1集\n甲\n# 第一集\n乙'}},1,{},'e'),/重复/);
});

test('页面分集使用完整标题和前言，空集仍占原位置，无标题正文仍为第一集',()=>{
 assert.deepEqual(splitDocumentEpisodes('前言\r\n第1集：标题\r\n正文\r\n## 第二集\r\n### 第十二集\r\n尾句'),[
  {no:1,content:'前言\r\n第1集：标题\r\n正文'},
  {no:2,content:'## 第二集'},
  {no:12,content:'### 第十二集\r\n尾句'}
 ]);
 assert.deepEqual(splitDocumentEpisodes('无标题正文。'),[{no:1,content:'无标题正文。'}]);
 assert.deepEqual(splitDocumentEpisodes('  '),[]);
});

test('章节拆分识别章回节篇卷部集与Chapter/Episode，前言单独成章',()=>{
 const text='前言文字\n第一章 开场\n甲\n第十二回 回目\n乙\n第三节 一节\n丙\n第四篇 一篇\n丁\n第五卷 一卷\n戊\n第六部 一部\n己\n第7集 集标题\n庚\nChapter 8 英文\n辛\nEpisode 9\n壬';
 const rows=splitSourceChapters(text);
 assert.deepEqual(rows.map(row=>row.title),['前言','第一章 开场','第十二回 回目','第三节 一节','第四篇 一篇','第五卷 一卷','第六部 一部','第7集 集标题','Chapter 8 英文','Episode 9']);
 assert.equal(rows[0].sourceText,'前言文字');
 assert.deepEqual(rows.slice(1).map(row=>row.sourceText),['甲','乙','丙','丁','戊','己','庚','辛','壬']);
 assert.equal(chapterHeadingMatches(text).length,9);
});

test('正文里的章回提及不算标题，无标题原文整篇一章',()=>{
 assert.deepEqual(splitSourceChapters('这是前言，提到第一章和第二章。\n只有正文。'),[{title:'',sourceText:'这是前言，提到第一章和第二章。\n只有正文。'}]);
 assert.deepEqual(splitSourceChapters('无标题正文。'),[{title:'',sourceText:'无标题正文。'}]);
 assert.deepEqual(splitSourceChapters('  '),[]);
});

test('Episode 作为分集标题切分，Chapter 留在正文',()=>{
 const rows=comicScriptEpisodes('前言\nEpisode 1: Start\nA\nEpisode 2\nB\nChapter 3\nC');
 assert.deepEqual(rows.map(row=>row.no),[1,2]);
 assert.ok(rows[0].content.includes('A'));
 assert.ok(!rows[0].content.includes('Chapter 3'));
 assert.ok(rows[1].content.includes('Chapter 3'));
});
