import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const dirname = path.dirname(fileURLToPath(import.meta.url));
const styles = fs.readFileSync(path.resolve(dirname, '../../../shared/styles/agent-workspace.css'), 'utf8');

test('全屏 3D 导演台拥有独立场景、素材选择和关闭入口', () => {
  const source = fs.readFileSync(path.join(dirname, 'DirectorDeskApp.jsx'), 'utf8');
  assert.match(source, /3D导演台/);
  assert.match(source, /关闭导演台/);
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

test('桌面端导演台避开全站左侧导航，素材卡和场景按钮不可被遮挡', () => {
  assert.match(styles, /\.agent-page \.agent-director-desk-fullscreen\s*\{[^}]*inset: 0 0 0 var\(--app-sidebar-width\);/s);
  assert.match(styles, /\.legacy-shell:has\(\.legacy-sidebar\.collapsed\) \.agent-page \.agent-director-desk-fullscreen\s*\{[^}]*inset: 0 0 0 var\(--app-sidebar-collapsed-width\);/s);
  assert.match(styles, /\.legacy-shell:has\(\.agent-director-desk-fullscreen\) > \.cm-penguin-shell\s*\{\s*display: none;/);
  assert.match(styles, /@media \(max-width: 900px\)\s*\{[\s\S]*?\.agent-page \.agent-director-desk-fullscreen\s*\{[^}]*inset: 0;/);
});

test('截图只有在导演台项目保存成功后才回写画布', () => {
  const source = fs.readFileSync(path.join(dirname, 'DirectorDeskApp.jsx'), 'utf8');
  assert.match(source, /const saved = await saveProject\(next\);/);
  assert.match(source, /if \(!saved\) throw new Error\('3D 构图状态保存失败'\);/);
});
