import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

const sources = [
  'BatchFactoryAiReasoningModal.jsx',
  'BatchFactoryBookSettingsModal.jsx',
  'BatchFactoryUnifiedSettingsModal.jsx'
].map(name => ({ name, source: readFileSync(new URL(`./${name}`, import.meta.url), 'utf8') }));

test('user asset selectors use only the unified comic asset slot', () => {
  for (const { name, source } of sources) {
    assert.match(source, /batch\.comic-assets/, `${name} must expose the unified comic asset slot`);
    assert.doesNotMatch(source, /startsWith\(['"]batch\.comic\.assets/, `${name} must not prefix-match internal asset slots`);
    assert.doesNotMatch(source, /batch\.comic\.assets\.(character|scene|prop)/, `${name} must not expose internal derivation slots`);
    assert.doesNotMatch(source, /batch\.comic\.character-image-template/, `${name} must not expose the three-view template slot`);
  }
});
