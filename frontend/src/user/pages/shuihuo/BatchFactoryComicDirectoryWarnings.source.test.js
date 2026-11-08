import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const unifiedSettings = fs.readFileSync(path.join(here, 'BatchFactoryUnifiedSettingsModal.jsx'), 'utf8');
const shuihuoStyles = fs.readFileSync(path.resolve(here, '../shuihuo-production.css'), 'utf8');
const publishHandler = fs.readFileSync(
  path.resolve(here, '../../../../../backend/internal/httpapi/batch_factory_n8_comic_publish.go'),
  'utf8'
);

test('漫画发布先展示账号本地目录，并在打开时自动后台刷新', () => {
  const start = unifiedSettings.indexOf('export function BatchFactoryComicPublishSettingsForm');
  const end = unifiedSettings.indexOf('export function BatchFactoryAiReasoningForm', start);
  const form = unifiedSettings.slice(start, end);
  assert.match(form, /readComicPublishDirectoryCache\(\)/);
  assert.match(form, /writeComicPublishDirectoryCache\(/);
  assert.match(form, /useEffect\(\(\) => \{[\s\S]*?getN8ComicPublishStatus\(/);
  assert.doesNotMatch(form, /comicPublishSettings\.directorySnapshot/);
  assert.doesNotMatch(form, /同步发布目录/);
  assert.doesNotMatch(form, /保存统一配置后下次会直接显示/);
});

test('status 把“目录可用”和“发布密钥已配置”拆成独立字段', () => {
  assert.match(publishHandler, /organizationsWarning\s*string/);
  assert.match(publishHandler, /audioCategoriesWarning\s*string/);
  assert.match(publishHandler, /"organizationsWarning":\s*organizationsWarning/);
  assert.match(publishHandler, /"audioCategoriesWarning":\s*audioCategoriesWarning/);
  assert.match(publishHandler, /"directoryAvailable":\s*directoryAvailable/);
  assert.match(publishHandler, /"publishKeyConfigured":\s*publishKeyConfigured/);
  assert.match(publishHandler, /"publishKeyHint":\s*"QIANTIE_N8_COMIC_OPEN_API_KEY"/);
  // configured 保留旧含义且不再参与目录判定。
  assert.match(publishHandler, /"configured":\s*publishKeyConfigured/);
  assert.match(publishHandler, /func n8DirectoryWarning\(label string, err error\) string/);
  assert.match(publishHandler, /n8DirectoryWarning\("组织目录", err\)/);
  assert.match(publishHandler, /n8DirectoryWarning\("音频分类目录", err\)/);
});

test('目录告警只看 directoryAvailable，目录可读时不再出现未配置/令牌失效红字', () => {
  assert.match(unifiedSettings, /setDirectoryAvailable\(typeof response\?\.directoryAvailable === 'boolean' \? response\.directoryAvailable : null\)/);
  assert.match(unifiedSettings, /setPublishKeyConfigured\(typeof response\?\.publishKeyConfigured === 'boolean' \? response\.publishKeyConfigured : null\)/);
  assert.doesNotMatch(unifiedSettings, /response\?\.configured === 'boolean'/);
  assert.doesNotMatch(unifiedSettings, /N8 目录服务未配置或令牌已失效/);
  // directoryAvailable=true 且 publishKeyConfigured=false：只显示中性提示。
  assert.match(unifiedSettings, /发布密钥未配置：目录可以读取，但发布\/上传不可用/);
  assert.match(unifiedSettings, /QIANTIE_N8_COMIC_OPEN_API_KEY/);
});

test('三种失败各自使用独立文案，401 不再被说成 N8 令牌失效', () => {
  // 401/403 只归到登录态失效。
  assert.match(unifiedSettings, /status === 401 \|\| status === 403\) return \{ kind: 'auth', message: '登录已失效，请重新登录' \}/);
  assert.match(unifiedSettings, /directoryFailureKind === 'auth'/);
  assert.match(unifiedSettings, /directoryFailureKind === 'auth' \? '登录已失效，请重新登录' : ''/);
  // 目录读失败且有真实原因时，直接展示上游 code/message。
  assert.match(unifiedSettings, /producerAccountsWarning\s*\|\|\s*directoryFailureMessage/);
  // 只有确认 directoryAvailable=false 且没有具体原因时，才提示未配置。
  assert.match(unifiedSettings, /producerAccountsWarning[\s\S]{0,200}?directoryFailureKind === 'auth'[\s\S]{0,200}?directoryAvailable === false/);
  assert.match(unifiedSettings, /N8 目录服务未配置（服务器缺少 N8 后台登录令牌）/);
  // 读取失败时不能回写“未配置”的假状态。
  assert.doesNotMatch(unifiedSettings, /setDirectoryConfigured\(false\)/);
});

test('指定音频读取失败为显式错误态并带重试，只有真空数据才显示没有可用', () => {
  assert.match(unifiedSettings, /const \[audiosError, setAudiosError\] = useState\(''\)/);
  assert.match(unifiedSettings, /const failure = describeN8DirectoryFailure\(error\)/);
  assert.match(unifiedSettings, /setAudiosError\(describeN8DirectoryFailure\(error\)\.message\)/);
  assert.match(unifiedSettings, /manualAudioReady \? '选择分类后随机带出，可手动更换' : '请先选择组织和音频分类'/);
  assert.match(unifiedSettings, /{audiosError \? <small className="batch-factory-engine-note">{audiosError}<\/small> : null}/);
  assert.match(unifiedSettings, /{!audiosError && !audios\.length && comicPublishSettings\.organizationId && comicPublishSettings\.audioCategoryId \? <small>当前分类和组织下没有可用指定音频<\/small> : null}/);
});

test('统一配置不展示指定音频，单本配置选齐组织和分类后才允许指定', () => {
  assert.match(unifiedSettings, /const manualAudioReady = Boolean\(singleBook && manualAudioMode && comicPublishSettings\.organizationId && comicPublishSettings\.audioCategoryId\)/);
  assert.match(unifiedSettings, /\{singleBook \? <div className="batch-factory-engine-field"><span><b>指定音频<\/b><\/span>/);
  assert.match(unifiedSettings, /disabled=\{!manualAudioReady\}/);
  assert.match(unifiedSettings, /manualAudioReady \? '选择分类后随机带出，可手动更换' : '请先选择组织和音频分类'/);
  assert.match(unifiedSettings, /if \(singleBook\) loadComicPublishAudios\(/);
});

test('音频分类只保留一个名称加 ID 的选择框', () => {
  const labels = unifiedSettings.match(/<span><b>启用中的音频分类<\/b>/g) || [];
  assert.equal(labels.length, 1);
  assert.doesNotMatch(unifiedSettings, /音频分类 ID（必填）/);
  assert.match(unifiedSettings, /label: `\$\{item\.name\}（ID \$\{item\.id\}）`/);
});

test('文案来源三个选项保留足够宽度，不被选中图标挤成省略号', () => {
  assert.match(unifiedSettings, /className="batch-factory-copy-source-segmented"/);
  assert.match(shuihuoStyles, /\.batch-factory-unified-settings-modal \.batch-factory-copy-source-segmented\s*\{[^}]*min-width:\s*240px/);
  assert.match(shuihuoStyles, /\.batch-factory-unified-settings-modal \.batch-factory-copy-source-segmented \.ant-segmented-item\s*\{[^}]*min-width:\s*76px/);
});

test('漫画发布说明收进标题问号，分类与作品匹配左右对齐', () => {
  assert.match(unifiedSettings, /className="batch-factory-field-title-inline"><b>音频匹配方式<\/b><HelpTip title="音频匹配方式">发布时按每本书的男女频和风格自动匹配并随机选择音频；重试会继续使用第一次选中的音频。<\/HelpTip>/);
  assert.doesNotMatch(unifiedSettings, /!manualAudioMode \? <small>发布时按每本书的男女频和风格自动匹配/);
  assert.match(unifiedSettings, /<label className="batch-factory-engine-field">\s*<span><b>作品匹配方式<\/b><\/span>/);
  assert.doesNotMatch(unifiedSettings, /<label className="batch-factory-engine-field" style=\{\{ gridColumn: '1 \/ -1' \}\}>\s*<span><b>作品匹配方式<\/b><\/span>/);
  assert.match(shuihuoStyles, /\.batch-factory-engine-field\s*>\s*\.batch-factory-field-title-inline\s*\{[^}]*display:\s*inline-flex/);
});
