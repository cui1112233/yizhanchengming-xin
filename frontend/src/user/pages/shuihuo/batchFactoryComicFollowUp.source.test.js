import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'BatchFactoryNovelList.jsx'), 'utf8');

test('书级重试提交后进入跟随刷新窗口，避免界面看起来没反应', () => {
  assert.match(source, /const \[comicFollowUntil, setComicFollowUntil\] = useState\(0\)/);
  assert.match(source, /setComicFollowUntil\(Date\.now\(\) \+ 180000\)/);
  assert.match(source, /followUntil=\{comicFollowUntil\}/);
  assert.match(source, /const pollDelay = resultPollDelay \|\| \(followActive \? 4000 : 0\)/);
});

test('漫画页行内展示正在生成页数，未开始时提示已提交重试', () => {
  assert.match(source, /正在生成 \{generatingPageCount\} 页/);
  assert.match(source, /已提交重试，等待开始/);
});

test('失败详情不再常驻行内漫画格，交给任务/日志与媒体库明细', () => {
  assert.match(source, /batch-factory-inline-comic-stage/);
  assert.ok(!source.includes('lineMode title="漫画页生成失败"'), '行内漫画格不应再渲染漫画页生成失败错误条');
});

test('生成中有动态反馈：运行标记、忙碌态与 CM 环形指示器', () => {
  assert.match(source, /batch-factory-comic-running/);
  assert.match(source, /generatingPageCount > 0 \? ' is-busy' : \(followActive \? ' is-pending' : ''\)/);
  assert.match(source, /import \{ CmLoader \} from '\.\/CmLoader\.jsx'/);
  assert.match(source, /presentation\.tone === 'busy' \? <CmLoader size=\{30\} title="漫画页生成中" \/>/);
  assert.match(source, /<CmLoader size=\{13\} title="正在生成" \/>/);
  assert.match(source, /<CmLoader size=\{13\} tone="pending" title="等待开始" \/>/);
  // 全项目的加载状态统一为 CM 环形指示器，不再使用 antd LoadingOutlined。
  assert.doesNotMatch(source, /LoadingOutlined/);
});
