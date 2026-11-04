import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const unifiedSettings = fs.readFileSync(path.join(here, 'BatchFactoryUnifiedSettingsModal.jsx'), 'utf8');
const publishHandler = fs.readFileSync(
  path.resolve(here, '../../../../../backend/internal/httpapi/batch_factory_n8_comic_publish.go'),
  'utf8'
);

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
  assert.match(unifiedSettings, /message="登录已失效，请重新登录"/);
  // 目录读失败且有真实原因时，直接展示上游 code/message。
  assert.match(unifiedSettings, /message=\{directoryFailureMessage \|\| directoryWarning\}/);
  // 只有确认 directoryAvailable=false 且没有具体原因时，才提示未配置。
  assert.match(unifiedSettings, /directoryFailureKind !== 'auth' && !\(directoryFailureMessage \|\| directoryWarning\) && directoryAvailable === false/);
  assert.match(unifiedSettings, /N8 目录服务未配置（服务器缺少 N8 后台登录令牌）/);
  // 读取失败时不能回写“未配置”的假状态。
  assert.doesNotMatch(unifiedSettings, /setDirectoryConfigured\(false\)/);
});

test('指定音频读取失败为显式错误态并带重试，只有真空数据才显示没有可用', () => {
  assert.match(unifiedSettings, /const \[audiosError, setAudiosError\] = useState\(''\)/);
  assert.match(unifiedSettings, /const failure = describeN8DirectoryFailure\(error\)/);
  assert.match(unifiedSettings, /setAudiosError\(failure\.message\)/);
  assert.match(unifiedSettings, /placeholder="请选择指定音频（可选）"/);
  assert.match(unifiedSettings, /{audiosError \? <small>{audiosError}<Button type="link" size="small" onClick=\{\(\) => setAudiosReloadKey\(key => key \+ 1\)\}>重试<\/Button><\/small> : null}/);
  assert.match(unifiedSettings, /{!audiosError && !audios\.length && comicPublishSettings\.organizationId && comicPublishSettings\.audioCategoryId \? <small>当前分类和组织下没有可用指定音频<\/small> : null}/);
});
