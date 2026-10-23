import test from 'node:test';
import assert from 'node:assert/strict';
import { appendUniqueAssetID, classifyComposerAttachment } from './agent-composer-attachments.js';

test('classifyComposerAttachment keeps supported text files in the text context', () => {
  assert.equal(classifyComposerAttachment({ name: 'outline.md', type: 'text/markdown', size: 1024 }).kind, 'text');
});

test('classifyComposerAttachment sends image video and audio through the project asset flow', () => {
  assert.equal(classifyComposerAttachment({ name: 'cover.png', type: 'image/png', size: 1024 }).kind, 'media');
  assert.equal(classifyComposerAttachment({ name: 'shot.mp4', type: 'video/mp4', size: 1024 }).kind, 'media');
  assert.equal(classifyComposerAttachment({ name: 'voice.mp3', type: 'audio/mpeg', size: 1024 }).kind, 'media');
});

test('classifyComposerAttachment rejects unsupported files and oversized attachments', () => {
  assert.equal(classifyComposerAttachment({ name: 'archive.zip', type: 'application/zip', size: 1024 }).kind, 'unsupported');
  assert.equal(classifyComposerAttachment({ name: 'cover.png', type: 'image/png', size: 129 * 1024 * 1024 }).kind, 'too-large');
});

test('appendUniqueAssetID preserves previous conversation assets and avoids duplicates', () => {
  assert.deepEqual(appendUniqueAssetID(['asset-a'], 'asset-b'), ['asset-a', 'asset-b']);
  assert.deepEqual(appendUniqueAssetID(['asset-a'], 'asset-a'), ['asset-a']);
});
