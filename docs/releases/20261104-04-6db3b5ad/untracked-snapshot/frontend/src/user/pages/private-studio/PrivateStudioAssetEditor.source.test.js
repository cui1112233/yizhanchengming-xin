import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

const shell = readFileSync(new URL('./PrivateStudioAssetEditor.jsx', import.meta.url), 'utf8');
const shellCss = readFileSync(new URL('./private-studio-asset-editor.css', import.meta.url), 'utf8');
const scriptPage = readFileSync(new URL('../ScriptPage.jsx', import.meta.url), 'utf8');
const shuihuoPage = readFileSync(new URL('../shuihuo/BatchFactoryNovelList.jsx', import.meta.url), 'utf8');
const ratios = readFileSync(new URL('../../../shared/imageAspectRatios.js', import.meta.url), 'utf8');

const scriptEditor = scriptPage.slice(scriptPage.indexOf('function EntityEditor('));
const shuihuoEditor = shuihuoPage.slice(shuihuoPage.indexOf('<PrivateStudioAssetEditor'), shuihuoPage.indexOf('</section>;', shuihuoPage.indexOf('<PrivateStudioAssetEditor')));

test('两个页面共用同一套编辑资产外壳与同一套分区', () => {
  assert.match(shell, /export function PrivateStudioAssetEditor\(/);
  assert.match(shell, /PrivateStudioAssetEditorSection title="基础"/);
  assert.match(shell, /PrivateStudioAssetEditorSection title="设定"/);
  assert.match(shell, /PrivateStudioAssetEditorSection title="提示词"/);
  assert.match(shell, /className="private-studio-asset-editor-aside"/);
  assert.match(shell, /className="private-studio-asset-editor-footer"/);
  assert.match(scriptPage, /import \{ PrivateStudioAssetEditor \} from '\.\/private-studio\/PrivateStudioAssetEditor';/);
  assert.match(shuihuoPage, /import \{ PrivateStudioAssetEditor \} from '\.\.\/private-studio\/PrivateStudioAssetEditor';/);
  assert.match(scriptEditor, /<PrivateStudioAssetEditor[\s\S]{0,400}编辑资产 · \$\{entityName\(fields\)/);
  assert.match(shuihuoEditor, /<PrivateStudioAssetEditor[\s\S]{0,400}编辑资产 · \$\{assetName\(editing\)\}/);
});

test('剧本人物四个字段全部渲染在统一弹窗里，外形按剧本侧命名', () => {
  assert.match(shell, /key: '名称', label: '名称', section: 'basic', source: '名称'/);
  assert.match(shell, /key: '身份', label: '身份', section: 'settings', source: '身份'/);
  assert.match(shell, /key: '外形', label: '外形（剧本）', section: 'settings', source: '外形'/);
  assert.match(shell, /key: '性格', label: '性格', section: 'settings', source: '性格'/);
  assert.match(scriptEditor, /const basicFields = fieldsToRender\.filter/);
  assert.match(scriptEditor, /const settingFields = fieldsToRender\.filter/);
  assert.match(scriptEditor, /'外形（剧本）'/);
});

test('漫画人物三组设定全部独立输入，不与剧本外形合并', () => {
  assert.match(shuihuoEditor, /人物事实<Input\.TextArea/);
  assert.match(shuihuoEditor, /人物关系<Input\.TextArea/);
  assert.match(shuihuoEditor, /人物外形<Input\.TextArea/);
  assert.match(shell, /注意：剧本的「外形」与漫画的「人物外形」语义不同，绝不合并成一个输入框。/);
});

test('三视图开关只在人物类型出现', () => {
  assert.match(shuihuoEditor, /editing\?\.kind === 'character' \? <Tooltip title="开启后人物资产图按/);
  assert.match(shuihuoEditor, /人物三视图/);
});

test('画幅下拉为七档共用常量，没有写死的 16:9 兜底', () => {
  assert.match(ratios, /export const IMAGE_ASPECT_RATIOS = \['16:9', '9:16', '1:1', '4:3', '3:4', '3:2', '2:3'\];/);
  assert.match(shuihuoPage, /const imageAspectRatio = normalizeImageAspectRatio\(''\);/);
  assert.doesNotMatch(shuihuoPage, /imageAspectRatio \|\| '16:9'/);
  assert.match(shuihuoEditor, /options=\{IMAGE_ASPECT_RATIO_OPTIONS\}/);
  assert.match(shuihuoEditor, /placeholder=\{`默认 \$\{imageAspectRatio\}`\}/);
});

test('胶囊点击仍打开统一弹窗，添加入口一个不少', () => {
  assert.match(scriptPage, /onClick: \(\) => openEntityEditor\('characters', item\.id\)/);
  assert.match(scriptPage, /onClick: \(\) => openEntityEditor\('scenes', item\.id\)/);
  assert.match(scriptPage, /function openEntityEditor\(type, id\)/);
  assert.match(shuihuoPage, /setEditing\(\{ kind, name: '', prompt: '' \}\)/);
});

test('统一外壳只写 private-studio 作用域，且没有删除建议', () => {
  assert.doesNotMatch(shellCss, /\.batch-factory-/);
  assert.match(shellCss, /\.private-studio-asset-editor/);
  assert.doesNotMatch(shell + scriptEditor + shuihuoEditor, /建议删除/);
});
