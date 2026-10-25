import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const dirname = path.dirname(fileURLToPath(import.meta.url));
const styles = fs.readFileSync(path.resolve(dirname, '../../../shared/styles/agent-workspace.css'), 'utf8');
const read = name => fs.readFileSync(path.join(dirname, name), 'utf8');

test('导演台组合真实对象树、视口、检查器和底部工具', () => {
  const source = read('DirectorDeskApp.jsx');
  for (const component of ['DirectorDeskShell', 'DirectorToolRail', 'DirectorSceneTree', 'DirectorAssetPanel', 'DirectorScene', 'DirectorInspector', 'DirectorBottomDock']) {
    assert.match(source, new RegExp(component));
  }
  assert.match(source, /createDirectorSaveQueue/);
  assert.match(source, /viewMode/);
  assert.match(source, /selectedObject/);
});

test('关闭导演台前刷新最后一次保存', () => {
  const source = read('DirectorDeskApp.jsx');
  assert.match(source, /await saveQueue\.flush\(\)/);
  assert.match(source, /adapter\.close\(\)/);
});

test('导演台把选择、视图、变换和变更回调交给真实场景', () => {
  const source = read('DirectorDeskApp.jsx');
  for (const prop of ['selection={selectedObject}', 'viewMode={viewMode}', 'transformMode={transformMode}', 'onSelectionChange={setSelectedObject}', 'onProjectChange={applyProject}', 'onReadyChange={setSceneReady}']) {
    assert.match(source, new RegExp(prop.replace(/[{}]/g, '\\$&')));
  }
  assert.match(source, /<DirectorInspector[\s\S]*?images={images}/);
  assert.match(source, /sceneRef\.current\?\.focusObject\(selection\)/);
  assert.match(source, /sceneRef\.current\?\.resetDirectorView\(\)/);
});

test('连续收起左右面板使用函数式状态更新', () => {
  const source = read('DirectorDeskApp.jsx');
  assert.match(source, /setLeftCollapsed\(current => !current\)/);
  assert.match(source, /setRightCollapsed\(current => !current\)/);
});

test('上传参考图只复用素材上传并刷新图片列表', () => {
  const source = read('DirectorDeskApp.jsx');
  assert.match(source, /await adapter\.uploadProjectImage\(file\)/);
  assert.match(source, /await adapter\.listProjectImages\(\)/);
  assert.match(source, /setActiveTool\('assets'\)/);
  assert.doesNotMatch(source, /createAgent|confirmAgent|generate/i);
});

test('上传成功只清操作错误，不会清除尚未恢复的保存错误', () => {
  const source = read('DirectorDeskApp.jsx');
  assert.match(source, /const \[saveError, setSaveError\] = useState\(''\)/);
  assert.match(source, /const \[operationError, setOperationError\] = useState\(''\)/);
  const upload = source.slice(source.indexOf('const uploadReference'), source.indexOf('const savePrompt'));
  assert.match(upload, /setOperationError\(''\)/);
  assert.doesNotMatch(upload, /setSaveError\(''\)/);
  assert.match(source, /saveError \|\| operationError/);
});

test('画布宿主保存导演台后继续等待外层画布持久化', () => {
  const canvas = fs.readFileSync(path.resolve(dirname, '../AgentCanvas.jsx'), 'utf8');
  assert.match(canvas, /const onFlushCanvasRef = useRef\(onFlushCanvas\)/);
  assert.match(canvas, /onFlushCanvasRef\.current = onFlushCanvas/);
  assert.match(canvas, /flushCanvas:\s*\(\) => \{/);
  assert.match(canvas, /onFlushCanvasRef\.current/);
  assert.match(canvas, /\}, \[directorDeskNodeID\]\)/);
});

test('底部导演工具只暴露当前真实能力', () => {
  const dock = read('DirectorBottomDock.jsx');
  for (const label of ['移动', '重置视角', '截图到画布', '项目图片', '上传参考图', '保存场景描述']) assert.match(dock, new RegExp(label));
  assert.doesNotMatch(dock, /生成场景|动画时间轴|姿势/);
  assert.match(dock, /onSavePrompt/);
  assert.match(dock, /onCapture/);
  assert.match(dock, /input\.value = ''/);
  assert.match(dock, /Math\.min\(.*scrollHeight/s);
  assert.match(styles, /\.agent-director-bottom-dock__composer[^}]*width:\s*min\(620px,\s*calc\(100% - 32px\)\)/s);
});

test('描述输入随宽度和字体变化重新测量并清理监听', () => {
  const dock = read('DirectorBottomDock.jsx');
  assert.match(dock, /ResizeObserver/);
  assert.match(dock, /observer\.observe\(input\)/);
  assert.match(dock, /observer\.disconnect\(\)/);
  assert.match(dock, /document\.fonts/);
  assert.match(dock, /removeEventListener\('loadingdone'/);
});

test('底部工具栏限制在场景中列且不改变场景高度', () => {
  assert.match(styles, /\.agent-director-shell__bottom\s*\{[^}]*grid-column:\s*2\s*;[^}]*position:\s*relative\s*;[^}]*height:\s*0\s*;/s);
  assert.match(styles, /\.agent-director-bottom-dock\s*\{[^}]*position:\s*absolute\s*;[^}]*width:\s*100%\s*;/s);
  assert.doesNotMatch(styles.match(/\.agent-director-bottom-dock\s*\{[^}]*\}/s)[0], /100vw|position:\s*fixed/);
});

test('底部保存状态在浅色主题的深色视口上保持可见', () => {
  const statusRule = styles.match(/\.agent-director-bottom-dock__status\s*\{[^}]*\}/s)?.[0] || '';
  assert.doesNotMatch(statusRule, /color:\s*var\(--director-text\)/);
  assert.match(statusRule, /background:\s*rgba\(10,\s*19,\s*28,\s*\.78\)/);
  assert.match(statusRule, /color:\s*#f3f7f8/);
});

test('导演台外壳提供三栏和可收起面板', () => {
  const shell = read('DirectorDeskShell.jsx');
  const rail = read('DirectorToolRail.jsx');
  for (const slot of ['left', 'viewport', 'right']) assert.match(shell, new RegExp(`agent-director-shell__${slot}`));
  for (const label of ['场景对象', '项目图片', '添加角色', '添加机位', '画幅比例']) assert.match(rail, new RegExp(label));
  assert.doesNotMatch(rail, /姿势|动画时间轴/);
  assert.match(shell, /onToggleLeft/);
  assert.match(shell, /onToggleRight/);
  assert.match(shell, /onClose/);
});

test('场景树及素材面板使用实际对象和操作回调', () => {
  const tree = read('DirectorSceneTree.jsx');
  const assets = read('DirectorAssetPanel.jsx');
  assert.match(tree, /directorSceneObjects\(project\)/);
  for (const callback of ['onSelect', 'onLocate', 'onRemove']) assert.match(tree, new RegExp(`${callback}\\(`));
  for (const callback of ['onSetScene', 'onAddActor']) assert.match(assets, new RegExp(`${callback}\\(asset\\)`));
  assert.doesNotMatch(assets, /shared\/api|fetch\(/);
});

test('全屏 3D 导演台拥有独立场景、素材选择和关闭入口', () => {
  const source = fs.readFileSync(path.join(dirname, 'DirectorDeskApp.jsx'), 'utf8');
  const shell = fs.readFileSync(path.join(dirname, 'DirectorDeskShell.jsx'), 'utf8');
  assert.match(source, /3D导演台/);
  assert.match(`${source}\n${shell}`, /关闭导演台/);
  assert.match(source, /导演视角/);
  assert.match(source, /listProjectImages/);
  assert.match(source, /DirectorScene/);
  assert.match(source, /data-director-desk-open/);
});

test('3D 场景使用背景和演员立面，并将轨道镜头写回项目', () => {
  const source = fs.readFileSync(path.join(dirname, 'DirectorScene.jsx'), 'utf8');
  assert.match(source, /Canvas/);
  assert.match(source, /OrbitControls/);
  assert.match(source, /SceneBackground/);
  assert.match(source, /ImageActorPlane/);
  assert.match(source, /cameraFromControls/);
  assert.match(source, /CameraFrame/);
  assert.match(source, /useThree/);
  assert.match(source, /capture/);
});

test('全屏导演台有稳定的场景视口和窄侧栏布局', () => {
  assert.match(styles, /\.agent-page \.agent-director-desk-fullscreen\s*\{/);
  assert.match(styles, /\.agent-page \.agent-director-desk-fullscreen__viewport\s*\{/);
  assert.match(styles, /\.agent-page \.agent-director-3d-scene\s*\{/);
  assert.match(styles, /@media \(max-width: 900px\)/);
});

test('导演台中列提供真实可用高度供竖屏画幅适配', () => {
  const viewportRule = styles.match(/\.agent-director-shell__viewport\s*\{[^}]*\}/s)?.[0] || '';
  assert.match(viewportRule, /container-type:\s*size/);
  assert.match(viewportRule, /display:\s*flex/);
  assert.match(viewportRule, /align-items:\s*center/);
  assert.match(viewportRule, /justify-content:\s*center/);
});

test('桌面端导演台避开全站左侧导航，素材卡和场景按钮不可被遮挡', () => {
  assert.match(styles, /\.agent-page \.agent-director-desk-fullscreen\s*\{[^}]*inset: 0 0 0 var\(--app-sidebar-width\);/s);
  assert.match(styles, /\.legacy-shell:has\(\.legacy-sidebar\.collapsed\) \.agent-page \.agent-director-desk-fullscreen\s*\{[^}]*inset: 0 0 0 var\(--app-sidebar-collapsed-width\);/s);
  assert.match(styles, /\.legacy-shell:has\(\.agent-director-desk-fullscreen\) > \.cm-penguin-shell\s*\{\s*display: none;/);
  assert.match(styles, /@media \(max-width: 900px\)\s*\{[\s\S]*?\.agent-page \.agent-director-desk-fullscreen\s*\{[^}]*inset: 0;/);
});

test('截图只有在导演台项目保存成功后才回写画布', () => {
  const source = fs.readFileSync(path.join(dirname, 'DirectorDeskApp.jsx'), 'utf8');
  assert.match(source, /createDirectorCaptureTransaction/);
  assert.match(source, /captureTransaction\.start\(\)/);
  assert.match(source, /captureTransaction\.flush\(\)/);
  assert.match(source, /insertCapture:\s*asset\s*=>\s*adapter\.insertCapture\(asset\)/);
  const closeStart = source.indexOf('const close');
  const close = source.slice(closeStart, source.indexOf('const capture', closeStart));
  assert.ok(close.indexOf('await captureTransaction.flush()') < close.indexOf('adapter.close()'));
});
