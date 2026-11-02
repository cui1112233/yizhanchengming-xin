import assert from 'node:assert/strict';
import test from 'node:test';
import { buildBatchFactoryH3Constraints } from './batchFactoryH3Constraints.js';

test('maps enabled script constraint layers into the H3 compile contract', () => {
  const result = buildBatchFactoryH3Constraints({
    enabled: true,
    baseSetup: { enabled: true },
    enabledCategories: ['prefix', 'quality', 'restriction', 'negative'],
    selections: [
      { constraintCategory: 'prefix', presetId: 'script-constraint-prefix-live-action', body: 'PREFIX' },
      { constraintCategory: 'quality', presetId: 'script-constraint-quality-4k', body: 'QUALITY' },
      { constraintCategory: 'restriction', presetId: 'script-constraint-restriction-no-overlay', body: 'RESTRICTION' },
      { constraintCategory: 'negative', presetId: 'script-constraint-negative-general', body: 'NEGATIVE' }
    ]
  });

  assert.deepEqual(result, {
    prefix_text: 'PREFIX',
    quality_text: 'QUALITY',
    visual_restriction_text: 'RESTRICTION',
    negative_text: 'NEGATIVE',
    switches: {
      smart_unified: false,
      base_setup: true,
      prefix: true,
      quality: true,
      visual_restriction: true,
      negative: true
    }
  });
});

test('uses only the saved H3 visual baseline for smart unified and ignores the retired master switch', () => {
  const selected = {
    enabled: true,
    baseSetup: { enabled: true },
    enabledCategories: ['prefix'],
    selections: [{ constraintCategory: 'prefix', presetId: 'script-constraint-prefix-smart-unified', body: 'MUST-NOT-LEAK' }]
  };
  assert.deepEqual(buildBatchFactoryH3Constraints(selected), {
    prefix_text: '', quality_text: '', visual_restriction_text: '', negative_text: '',
    switches: { smart_unified: true, base_setup: true, prefix: false, quality: false, visual_restriction: false, negative: false }
  });

  assert.deepEqual(buildBatchFactoryH3Constraints({ ...selected, enabled: false }), {
    prefix_text: '', quality_text: '', visual_restriction_text: '', negative_text: '',
    switches: { smart_unified: true, base_setup: true, prefix: false, quality: false, visual_restriction: false, negative: false }
  });
});

test('keeps legacy batch selections active when their old settings omit the master flag', () => {
  const result = buildBatchFactoryH3Constraints({
    baseSetup: { enabled: true },
    enabledCategories: ['quality'],
    selections: [{ constraintCategory: 'quality', presetId: 'script-constraint-quality-4k', body: 'QUALITY' }]
  });
  assert.equal(result.switches.quality, true);
  assert.equal(result.switches.base_setup, true);
});

test('uses script-compatible defaults for legacy batches without category or base-setting fields', () => {
  const result = buildBatchFactoryH3Constraints({
    selections: [{ constraintCategory: 'negative', presetId: 'legacy-negative', body: 'NO-WATERMARK' }]
  });
  assert.equal(result.negative_text, 'NO-WATERMARK');
  assert.equal(result.switches.negative, true);
  assert.equal(result.switches.base_setup, true);
});
