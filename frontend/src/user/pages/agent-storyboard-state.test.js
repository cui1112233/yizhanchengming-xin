import test from 'node:test';
import assert from 'node:assert/strict';
import {
  normalizeStoryboardState,
  patchStoryboardState,
  moveStoryboardMaterial,
  reorderStoryboardMaterial,
  storyboardPointerDropTarget,
  storyboardBatchCategoryPatch,
  storyboardBatchAssets,
  storyboardLibraryWidth,
  storyboardSectionHeight,
  toggleStoryboardBatchSelection,
  storyboardCategoryOptions,
  storyboardItemCategory
} from './agent-storyboard-state.js';

test('旧项目会得到稳定的故事板分类、选择和排序状态', () => {
  assert.deepEqual(normalizeStoryboardState(null), {
    directorMode: 'supervised',
    selectedKey: '',
    filters: { text: 'all', audio: 'all', image: 'all', video: 'all' },
    categories: {},
    order: [],
    libraryWidth: 340,
    sectionHeights: { text: 220, audio: 240, image: 320 }
  });
  assert.deepEqual(storyboardCategoryOptions('image').map(item => item.label), ['全部', '角色', '场景', '物品', '其他']);
  assert.deepEqual(storyboardCategoryOptions('video').map(item => item.label), ['全部', '视频片段', '成片']);
});

test('故事板分区高度按类型使用默认值并限制在安全范围', () => {
  assert.equal(storyboardSectionHeight('text'), 220);
  assert.equal(storyboardSectionHeight('audio', 40), 120);
  assert.equal(storyboardSectionHeight('image', 900), 520);
  assert.deepEqual(normalizeStoryboardState({ sectionHeights: { text: 310, audio: 1, image: 999 } }).sectionHeights, {
    text: 310,
    audio: 120,
    image: 520
  });
});

test('批量选择稳定去重并可单独取消', () => {
  assert.deepEqual(toggleStoryboardBatchSelection([], 'asset:a'), ['asset:a']);
  assert.deepEqual(toggleStoryboardBatchSelection(['asset:a'], 'asset:b'), ['asset:a', 'asset:b']);
  assert.deepEqual(toggleStoryboardBatchSelection(['asset:a', 'asset:b'], 'asset:a'), ['asset:b']);
  assert.deepEqual(toggleStoryboardBatchSelection(['asset:a', 'asset:a', ''], 'asset:b'), ['asset:a', 'asset:b']);
});

test('批量分类只处理同类型有效项目且加入对话只返回真实素材', () => {
  const items = [
    { key: 'asset:image-1', kind: 'image', asset: { id: 'image-1', kind: 'image' } },
    { key: 'node:image-2', kind: 'image', asset: null },
    { key: 'asset:video-1', kind: 'video', asset: { id: 'video-1', kind: 'video' } }
  ];
  assert.deepEqual(storyboardBatchCategoryPatch(items, ['asset:image-1', 'node:image-2'], 'character'), {
    'asset:image-1': 'character',
    'node:image-2': 'character'
  });
  assert.equal(storyboardBatchCategoryPatch(items, ['asset:image-1', 'asset:video-1'], 'character'), null);
  assert.deepEqual(storyboardBatchAssets(items, ['asset:image-1', 'node:image-2', 'asset:video-1']).map(asset => asset.id), ['image-1', 'video-1']);
});

test('故事板素材栏宽度有稳定默认值并限制在可用范围内', () => {
  assert.equal(storyboardLibraryWidth(), 340);
  assert.equal(storyboardLibraryWidth(120), 280);
  assert.equal(storyboardLibraryWidth(999), 520);
  assert.equal(storyboardLibraryWidth(460, 410), 410);
  assert.equal(normalizeStoryboardState({ libraryWidth: 416 }).libraryWidth, 416);
});

test('故事板按素材独立保存分类且不会写入原节点或素材', () => {
  const original = normalizeStoryboardState({ selectedKey: 'asset:image-1' });
  const changed = patchStoryboardState(original, {
    selectedKey: 'asset:image-2',
    filters: { image: 'character' },
    categories: { 'asset:image-2': 'character' }
  });

  assert.equal(original.selectedKey, 'asset:image-1');
  assert.equal(changed.selectedKey, 'asset:image-2');
  assert.equal(changed.filters.image, 'character');
  assert.equal(changed.categories['asset:image-2'], 'character');
  assert.equal(storyboardItemCategory({ key: 'asset:image-2', kind: 'image' }, changed), 'character');
});

test('素材排序只移动同一分类中的目标并保持稳定键', () => {
  const state = normalizeStoryboardState({ order: ['asset:a', 'asset:b', 'asset:c'] });
  const moved = moveStoryboardMaterial(state, 'asset:c', -1, ['asset:a', 'asset:b', 'asset:c']);
  assert.deepEqual(moved.order, ['asset:a', 'asset:c', 'asset:b']);
  assert.deepEqual(moveStoryboardMaterial(moved, 'asset:a', -1, ['asset:a', 'asset:c', 'asset:b']).order, moved.order);
});

test('故事板拖拽会把同分区素材放到目标前并保留其他分区顺序', () => {
  const state = normalizeStoryboardState({ order: ['asset:text-1', 'asset:image-1', 'asset:image-2', 'asset:video-1'] });
  const moved = reorderStoryboardMaterial(state, 'asset:image-2', 'asset:image-1', ['asset:image-1', 'asset:image-2']);
  assert.deepEqual(moved.order, ['asset:text-1', 'asset:image-2', 'asset:image-1', 'asset:video-1']);
  assert.deepEqual(reorderStoryboardMaterial(moved, 'asset:image-2', 'asset:image-2', ['asset:image-2', 'asset:image-1']).order, moved.order);
  assert.deepEqual(reorderStoryboardMaterial(moved, 'asset:video-1', 'asset:image-1', ['asset:image-2', 'asset:image-1']).order, moved.order);
});

test('指针拖动只接受落点下同类型的其他素材', () => {
  const elements = [
    { key: 'asset:image-1', kind: 'image' },
    { key: 'asset:video-1', kind: 'video' }
  ];
  assert.equal(storyboardPointerDropTarget('asset:image-2', 'image', elements), 'asset:image-1');
  assert.equal(storyboardPointerDropTarget('asset:image-1', 'image', elements), '');
  assert.equal(storyboardPointerDropTarget('asset:image-2', 'image', elements.slice(1)), '');
  assert.equal(storyboardPointerDropTarget('asset:image-2', 'image', []), '');
});

test('无效分类、重复排序键和未知字段都会被清理', () => {
  assert.deepEqual(normalizeStoryboardState({
    directorMode: 'invalid',
    selectedKey: 42,
    filters: { image: 'unknown', video: 'final_video' },
    categories: { 'asset:image': 'unknown', 'asset:video': 'final_video' },
    order: ['asset:a', '', 'asset:a', 3],
    secret: 'drop'
  }), {
    directorMode: 'supervised',
    selectedKey: '',
    filters: { text: 'all', audio: 'all', image: 'all', video: 'final_video' },
    categories: { 'asset:video': 'final_video' },
    order: ['asset:a'],
    libraryWidth: 340,
    sectionHeights: { text: 220, audio: 240, image: 320 }
  });
});
