import test from 'node:test';
import assert from 'node:assert/strict';
import { buildRuleConfigPatch } from './novel-fetch-rule-config.js';

test('chapter and pair rule edits preserve sibling rules in canonical library', () => {
  const data = { knowledgeLibraries: { chapter_rules: { chapter_exact_rules: [{ find: '^目录$' }] }, symbol_rules: { symbol_rules: [{ find: '，' }] } } };
  assert.deepEqual(buildRuleConfigPatch('chapter_inline_rules', { chapter_inline_rules: [{ find: '^章节' }], chapter_remove_inline: true }, data).knowledge.chapter_rules, { chapter_exact_rules: [{ find: '^目录$' }], chapter_inline_rules: [{ find: '^章节' }], chapter_remove_inline: true });
  assert.deepEqual(buildRuleConfigPatch('pair_fill_rules', { pair_fill_rules: [{ find: '“', replace: '”' }] }, data).knowledge.symbol_rules.symbol_rules, [{ find: '，' }]);
});

test('layout rule switches target the runtime layout while keeping spacing in its library', () => {
  const patch = buildRuleConfigPatch('layout_rules', { indent_spaces: 2, apply_sensitive: false }, { appConfig: { layout: { apply_pair_fill: false } } });
  assert.deepEqual(patch.layout, { apply_pair_fill: false, apply_sensitive: false });
  assert.equal(patch.knowledge.layout_rules.indent_spaces, 2);
});
