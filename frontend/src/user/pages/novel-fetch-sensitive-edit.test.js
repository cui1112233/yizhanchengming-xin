import test from 'node:test';
import assert from 'node:assert/strict';
import { updateSensitiveRule, parseSensitiveBulkLines } from './novel-fetch-sensitive-edit.js';

test('editing replacement to empty records explicit deletion and preserves other fields', () => {
 const old = { find: '代孕', replace: '旧值', enabled: false, note: '保留' };
 const edited = updateSensitiveRule(old, 'replace', '');
 assert.deepEqual(edited, { ...old, replace: '', replace_explicit: true });
 assert.equal(old.replace, '旧值');
 assert.equal(updateSensitiveRule(old, 'enabled', true).replace_explicit, undefined);
});
test('existing arrow format distinguishes missing replacement from explicit empty and keeps remaining arrows', () => {
 const rows = parseSensitiveBulkLines('代孕=>\n原词\n原词二=>甲=>乙');
 assert.equal(rows[0].replace_explicit, true);
 assert.equal(rows[0].replace, '');
 assert.equal(rows[1].replace_explicit, undefined);
 assert.equal(rows[2].replace, '甲=>乙');
});
