import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const source = readFileSync(new URL('./BatchFactoryNovelList.jsx', import.meta.url), 'utf8');

test('book asset catalog keeps protected URLs without eagerly downloading every original image', () => {
  const hook = source.slice(
    source.indexOf('function useAuthenticatedBookAssetThumbs'),
    source.indexOf('function renderStoryboardPromptTextPart')
  );
  assert.doesNotMatch(hook, /loadBookAssetDisplayURL/);
  assert.match(hook, /primary\?\.(?:thumbnailUrl|url)/);
});

test('offscreen prompt rows are deferred before they can request prompts or asset images', () => {
  assert.match(source, /function DeferredInlineBookPrompts/);
  assert.match(source, /<DeferredInlineBookPrompts[\s\S]*?batch=\{batch\}/);
});
