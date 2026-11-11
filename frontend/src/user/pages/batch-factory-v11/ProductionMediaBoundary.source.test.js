import assert from 'node:assert/strict';
import test from 'node:test';
import { readFileSync } from 'node:fs';

const source = readFileSync(new URL('./ProductionMediaBoundary.jsx', import.meta.url), 'utf8');

test('unavailable external videos show a readable state instead of an empty player', () => {
  assert.match(source, /onPlaybackError/);
  assert.match(source, /视频源地址已失效或无法访问/);
  assert.match(source, /data-bf-media-state="error"/);
});

test('media first-frame overlay uses the shared CM loader instead of LoadingOutlined', () => {
  assert.match(source, /import \{ CmLoader \} from '\.\.\/shuihuo\/CmLoader\.jsx'/);
  assert.match(source, /<CmLoader size=\{40\} title=\{overlayText\} \/>/);
  assert.doesNotMatch(source, /LoadingOutlined/);
});
