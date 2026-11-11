import assert from 'node:assert/strict';
import test from 'node:test';
import {
  buildStoryboardJson,
  buildStoryboardPreview,
  groupStoryboard,
  parseStoryboardDocument,
  parseStoryboardFiles
} from './storyboardOrganizer.js';

test('分镜解析：竖线表格', () => {
  const content = [
    '镜号 | 时长 | 画面 | 台词',
    '1 | 3秒 | 女主推门进来 | 你终于回来了',
    '2 | 4s | 男主抬头 | 嗯，路上堵车'
  ].join('\n');
  const shots = parseStoryboardDocument(content, { fileName: 'a.txt' });
  assert.equal(shots.length, 2);
  assert.equal(shots[0].duration, 3);
  assert.equal(shots[0].content, '女主推门进来');
  assert.equal(shots[0].dialogue, '你终于回来了');
  assert.equal(shots[1].duration, 4);
});

test('分镜解析：镜头标题块 + 字段', () => {
  const content = [
    '镜头1（3秒）',
    '画面：雨夜街道，霓虹倒影',
    '运镜：缓慢推进',
    '台词：别回头',
    '',
    '镜头2 5秒',
    '画面：主角奔跑'
  ].join('\n');
  const shots = parseStoryboardDocument(content, { fileName: 'b.txt' });
  assert.equal(shots.length, 2);
  assert.equal(shots[0].duration, 3);
  assert.equal(shots[0].content, '雨夜街道，霓虹倒影');
  assert.equal(shots[0].camera, '缓慢推进');
  assert.equal(shots[0].dialogue, '别回头');
  assert.equal(shots[1].duration, 5);
});

test('分镜解析：数字序号条目与默认时长', () => {
  const shots = parseStoryboardDocument('1. 她在门口停下\n2. 回头看向镜头', { defaultShotSeconds: 2.5 });
  assert.equal(shots.length, 2);
  assert.equal(shots[0].duration, 2.5);
  assert.equal(shots[1].content, '回头看向镜头');
});

test('分镜解析：带时间戳的分镜按起止时间算时长', () => {
  const content = '1\n00:00:01,000 --> 00:00:04,500\n开场空镜\n\n2\n00:00:04,500 --> 00:00:08,000\n人物入画\n';
  const shots = parseStoryboardDocument(content, { fileName: 'c.srt' });
  assert.equal(shots.length, 2);
  assert.equal(shots[0].duration, 3.5);
  assert.equal(shots[1].duration, 3.5);
});

test('分镜解析：多文件按导入顺序编号', () => {
  const shots = parseStoryboardFiles([
    { name: 'a.txt', content: '镜头1（2秒）\n画面：A' },
    { name: 'b.txt', content: '镜头1（3秒）\n画面：B' }
  ]);
  assert.equal(shots.length, 2);
  assert.equal(shots[1].index, 2);
  assert.equal(shots[1].source, 'b.txt');
});

test('分镜分组：15 秒上限，超长镜头独占一组', () => {
  const shots = [
    { index: 1, duration: 5, content: 'A' },
    { index: 2, duration: 5, content: 'B' },
    { index: 3, duration: 5, content: 'C' },
    { index: 4, duration: 20, content: 'D' },
    { index: 5, duration: 2, content: 'E' }
  ];
  const groups = groupStoryboard(shots, 15);
  assert.equal(groups.length, 3);
  assert.deepEqual(groups.map(group => group.shots.length), [3, 1, 1]);
  assert.equal(groups[0].duration, 15);
  assert.equal(groups[1].duration, 20);
  assert.equal(groups[2].start, 35);
});

test('分镜分组：30 秒目标把 6 个 5 秒镜头放在一组', () => {
  const shots = Array.from({ length: 6 }, (_, index) => ({ index: index + 1, duration: 5, content: `S${index}` }));
  const groups = groupStoryboard(shots, 30);
  assert.equal(groups.length, 1);
  assert.equal(groups[0].duration, 30);
});

test('分镜 JSON：字段齐全且按组切片', () => {
  const shots = [
    { index: 1, duration: 5, content: 'A', dialogue: '台词A', camera: '推', source: 'a.txt', audio: '', note: '' },
    { index: 2, duration: 5, content: 'B', dialogue: '', camera: '', source: 'a.txt', audio: '', note: '' },
    { index: 3, duration: 5, content: 'C', dialogue: '', camera: '', source: 'a.txt', audio: '', note: '' }
  ];
  const json = buildStoryboardJson(shots, { targetSeconds: 10 });
  assert.equal(json.targetSeconds, 10);
  assert.equal(json.totalShots, 3);
  assert.equal(json.totalSeconds, 15);
  assert.equal(json.groupCount, 2);
  assert.equal(json.groups[0].durationSeconds, 10);
  assert.equal(json.groups[0].shots[0].dialogue, '台词A');
  assert.equal(json.groups[1].shots[0].index, 3);
});

test('分镜预览：包含组号和镜号', () => {
  const preview = buildStoryboardPreview([
    { index: 1, duration: 5, content: '开场', dialogue: '喂' },
    { index: 2, duration: 12, content: '转场' }
  ], { targetSeconds: 15 });
  assert.match(preview, /【第 1 组】/);
  assert.match(preview, /1\. 5秒 开场 台词：喂/);
  assert.match(preview, /【第 2 组】/);
});
