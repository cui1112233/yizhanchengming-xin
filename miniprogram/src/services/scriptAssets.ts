import { scriptEntityKey, type ScriptConstraints, type ScriptEntity } from './scriptWorkspace';
export type ScriptAssets = { characters: ScriptEntity[]; scenes: ScriptEntity[] };
export type EntityType = keyof ScriptAssets;
export function scriptEntityName(entity: ScriptEntity, fallback = '未命名') {
  const fields = entity.data && typeof entity.data === 'object' ? entity.data as Record<string, unknown> : entity as Record<string, unknown>;
  return String(fields.name || fields.角色名称 || fields.场景名称 || fields.名称 || entity.name || fallback);
}
export function newScriptEntity(type: EntityType, id: string): ScriptEntity {
  return { id, data: type === 'characters' ? { 名称: '', 身份: '', 外形: '', 性格: '' } : { 名称: '', 时段: '', 氛围: '', 描述: '' } };
}
export function validScriptAssets(value: ScriptAssets) {
  return [...value.characters, ...value.scenes].every(entity => scriptEntityName(entity, '').trim());
}
export function retainScriptReferences(value: ScriptAssets, protagonists: string[], constraints: ScriptConstraints = {}) {
  const ids = new Set([...value.characters.map(scriptEntityKey), ...value.scenes.map(scriptEntityKey)]);
  const characterIDs = new Set(value.characters.map(scriptEntityKey));
  return {
    protagonists: protagonists.filter(id => characterIDs.has(id)),
    constraints: { ...constraints, ...(constraints.entityReferences ? { entityReferences: constraints.entityReferences.filter((item: any) => ids.has(String(item?.entityId || ''))) } : {}) },
  };
}
