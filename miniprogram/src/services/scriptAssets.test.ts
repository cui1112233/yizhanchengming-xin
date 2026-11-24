import { expect, test } from 'vitest';
import { newScriptEntity, retainScriptReferences, scriptEntityName, validScriptAssets } from './scriptAssets';
test('new manual entities use the same script fields as the web editor', () => {
  expect(newScriptEntity('characters', 'c1')).toEqual({ id: 'c1', data: { 名称: '', 身份: '', 外形: '', 性格: '' } });
  expect(newScriptEntity('scenes', 's1')).toEqual({ id: 's1', data: { 名称: '', 时段: '', 氛围: '', 描述: '' } });
});
test('blank manual names cannot be applied', () => {
  expect(validScriptAssets({ characters: [newScriptEntity('characters', 'c1')], scenes: [] })).toBe(false);
  expect(scriptEntityName({ data: { 名称: '林夏' } })).toBe('林夏');
});
test('removing a person prunes focus and constraint references without losing other layers', () => {
  const constraints = { enabled: true, quality: { enabled: true, presetId: 'q1' }, entityReferences: [{ entityId: 'removed', layer: 'old' }, { entityId: 's1', image: 'keep' }] };
  const next = retainScriptReferences({ characters: [{ id: 'c1', name: '林夏' }], scenes: [{ id: 's1', name: '店铺' }] }, ['c1', 'removed'], constraints);
  expect(next).toEqual({ protagonists: ['c1'], constraints: { enabled: true, quality: constraints.quality, entityReferences: [{ entityId: 's1', image: 'keep' }] } });
  expect(constraints.entityReferences).toHaveLength(2);
});
