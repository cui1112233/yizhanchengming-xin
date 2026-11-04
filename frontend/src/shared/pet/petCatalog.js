const stackyPet = Object.freeze({
  id: 'stacky',
  displayName: 'CM',
  description: 'CM，前贴的桌面宠物。',
  spriteVersionNumber: 2,
  spritesheetPath: '/pets/stacky/spritesheet.webp',
  atlasProfile: 'stacky-v2',
  behaviorProfile: 'cm-v1',
  speechProfile: 'cm-v1',
  renderMode: 'pixelated'
});

const pixiuPet = Object.freeze({
  id: 'pixiu',
  displayName: '貔貅',
  description: '貔貅，前贴的招财守护宠物。',
  spriteVersionNumber: 2,
  spritesheetPath: '/pets/pixiu/spritesheet.webp',
  atlasProfile: 'stacky-v2',
  animationProfile: 'pixiu-v1',
  behaviorProfile: 'cm-v1',
  speechProfile: 'cm-v1',
  renderMode: 'smooth'
});

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

// Keep the primary companion first so the settings selector and all new
// workspaces agree on the same default without a separate ordering rule.
export const DEFAULT_PET_ID = 'little-companion';
export const DEFAULT_PET_LABEL_SUFFIX = '（默认）';
export const PET_DEFINITIONS = Object.freeze([littleCompanionPet, stackyPet, pixiuPet]);
export const PET_SELECTION_EVENT = 'qiantie:pet-selection';

export function findPetDefinition(value) {
  const id = typeof value === 'string' ? value : value?.id;
  return PET_DEFINITIONS.find(pet => pet.id === id) || null;
}

export function getPetDefinition(value) {
  // 只有“从未选择”或“未知标识”才回落到默认宠物；用户已保存的 stacky/pixiu 不受影响。
  return findPetDefinition(value) || findPetDefinition(DEFAULT_PET_ID);
}

export function getPetOptions() {
  return PET_DEFINITIONS.map(pet => ({
    label: pet.id === DEFAULT_PET_ID ? `${pet.displayName}${DEFAULT_PET_LABEL_SUFFIX}` : pet.displayName,
    value: pet.id
  }));
}

export function petAnimationDelay(pet, state) {
  const resolved = getPetDefinition(pet);
  if (resolved.id !== 'pixiu') return state === 'working' ? 120 : 180;
  if (state === 'working') return 220;
  if (state === 'success' || state === 'error') return 260;
  return 320;
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
