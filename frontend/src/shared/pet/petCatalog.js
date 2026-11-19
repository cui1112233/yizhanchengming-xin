const littleCompanionPet = Object.freeze({
  id: 'little-companion',
  displayName: '小家伙',
  description: '小家伙，陪你创作的原创桌面宠物。',
  renderer: 'multi-sheet',
  assetBasePath: '/pets/little-companion',
  previewPose: 'idle',
  behaviorProfile: 'cm-v1',
  speechProfile: 'cm-v1',
  renderMode: 'smooth'
});

// 前贴宠物只保留小家伙一只；历史上保存过 CM（stacky）/貔貅（pixiu）的账号
// 会在 getPetDefinition 里回落到默认，不会出现缺资源的渲染。
export const DEFAULT_PET_ID = 'little-companion';
export const PET_DEFINITIONS = Object.freeze([littleCompanionPet]);
export const PET_SELECTION_EVENT = 'qiantie:pet-selection';

export function findPetDefinition(value) {
  const id = typeof value === 'string' ? value : value?.id;
  return PET_DEFINITIONS.find(pet => pet.id === id) || null;
}

export function getPetDefinition(value) {
  // 只有“从未选择”或“未知标识”才回落到默认宠物。
  return findPetDefinition(value) || findPetDefinition(DEFAULT_PET_ID);
}

export function getPetOptions() {
  return PET_DEFINITIONS.map(pet => ({ label: pet.displayName, value: pet.id }));
}

export function petAnimationDelay(pet, state) {
  return state === 'working' ? 120 : 180;
}

export function dispatchPetSelection(value) {
  if (typeof window === 'undefined') return;
  window.dispatchEvent(new CustomEvent(PET_SELECTION_EVENT, { detail: { pet: getPetDefinition(value) } }));
}

export function previewPetSelection(value) {
  const pet = getPetDefinition(value);
  dispatchPetSelection(pet.id);
  return pet;
}
