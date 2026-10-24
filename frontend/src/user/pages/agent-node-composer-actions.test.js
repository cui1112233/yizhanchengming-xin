import test from 'node:test';
import assert from 'node:assert/strict';
import { compactComposerActionPlan } from './agent-node-composer-actions.js';

test('video keeps references and templates on top', () => {
  assert.deepEqual(
    compactComposerActionPlan({ kind: 'video', source: 'batch_factory', resourceType: 'video' }),
    {
      primary: ['reference-assets', 'prompt-templates'],
      more: ['preview-inputs']
    }
  );
});

test('failed mapped video adds retry only in more', () => {
  assert.deepEqual(
    compactComposerActionPlan({
      kind: 'video',
      source: 'batch_factory',
      resourceType: 'video',
      canRetry: true
    }).more,
    ['preview-inputs', 'retry']
  );
});

test('mapped script keeps real production actions', () => {
  assert.deepEqual(
    compactComposerActionPlan({ kind: 'script', source: 'batch_factory', resourceType: 'book_script' }),
    {
      primary: ['prompt-templates', 'extract-assets', 'generate-storyboard', 'generate-viral'],
      more: ['save-line-count']
    }
  );
});

test('script exposes candidate adoption only after a viral candidate exists', () => {
  assert.deepEqual(
    compactComposerActionPlan({
      kind: 'script',
      source: 'batch_factory',
      resourceType: 'book_script',
      hasViralCandidate: true
    }).more,
    ['save-line-count', 'adopt-viral-candidate']
  );
});

test('image exposes candidate confirmation only when an image candidate exists', () => {
  assert.deepEqual(
    compactComposerActionPlan({
      kind: 'image',
      source: 'batch_factory',
      resourceType: 'book_asset',
      hasImageCandidates: true
    }),
    {
      primary: ['reference-assets', 'prompt-templates'],
      more: ['preview-inputs', 'confirm-image']
    }
  );
});

test('audio exposes candidate confirmation only when an audio candidate exists', () => {
  assert.deepEqual(
    compactComposerActionPlan({
      kind: 'audio',
      source: 'batch_factory',
      resourceType: 'audio',
      hasAudioCandidates: true
    }),
    {
      primary: ['prompt-templates'],
      more: ['preview-inputs', 'confirm-audio']
    }
  );
});

test('merge exposes organization and upload only when it can publish', () => {
  assert.deepEqual(
    compactComposerActionPlan({
      kind: 'video',
      source: 'batch_factory',
      resourceType: 'merge',
      canPublish: true
    }),
    {
      primary: ['publish-organization'],
      more: ['upload-video-management']
    }
  );
});

test('non-Batch-Factory video hides Batch Factory-only template actions', () => {
  assert.deepEqual(
    compactComposerActionPlan({ kind: 'video', source: 'manual', resourceType: 'video' }),
    {
      primary: ['reference-assets'],
      more: ['preview-inputs']
    }
  );
});
