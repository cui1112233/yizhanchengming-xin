import test from 'node:test';
import assert from 'node:assert/strict';
import { smartUnifiedAnalysisForBook, smartUnifiedDisplayEnabled } from './batchFactorySmartUnified.js';

const analysis = {
  prompt: '现代都市短剧；高级电影感；当代都市。',
  fields: { final_genre: '现代都市短剧', trailer_style: '高级电影感', story_era: '当代都市' },
  preset: { id: 'script-constraint-prefix-smart-unified', version: 7 }
};

test('keeps a persisted style.system result even while smart-unified display is off', () => {
  const batch = { settingsState: { patch: { aiPromptConfig: { constraints: { selections: [] } } } } };
  const book = { directorRevision: { output: { smart_unified_analysis: analysis } }, settingsState: { patch: {} } };
  assert.deepEqual(smartUnifiedAnalysisForBook(book), analysis);
  assert.equal(smartUnifiedDisplayEnabled(batch, book), false);
});

test('enables the saved style.system result only for the smart-unified prefix selection', () => {
  const batch = { settingsState: { patch: { aiPromptConfig: { constraints: { selections: [{ presetId: 'script-constraint-prefix-smart-unified', constraintCategory: 'prefix' }] } } } } };
  const book = { directorRevision: { output: { smart_unified_analysis: analysis } }, settingsState: { patch: {} } };
  assert.equal(smartUnifiedDisplayEnabled(batch, book), true);
});

test('reads the Go-persisted style analysis from the book settings patch', () => {
  const book = { settingsState: { patch: { h3StyleAnalysis: JSON.stringify(analysis) } } };
  assert.deepEqual(smartUnifiedAnalysisForBook(book), analysis);
});
