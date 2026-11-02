import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import assert from 'node:assert/strict';

const source = fs.readFileSync(path.resolve('src/user/pages/shuihuo/BatchFactoryEngineSettingsDrawer.jsx'), 'utf8');

test('图片模型配置 exposes a selectable generation provider and local executor status', () => {
  assert.match(source, /图片生成通道/);
  assert.match(source, /imageProvider/);
  assert.match(source, /from '.\/providerModelBinding'/);
  assert.match(source, /providerOptionsForModels\(models, 'image'\)/);
  assert.match(source, /providerOptionsForModels\(models, 'video'\)/);
});

test('legacy engine settings reconcile the model when its provider changes', () => {
  assert.match(source, /providerModelSettingsPatch\(\{ models, kind: 'image'/);
  assert.match(source, /providerModelSettingsPatch\(\{ models, kind: 'video'/);
});
