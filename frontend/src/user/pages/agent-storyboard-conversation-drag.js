export const storyboardAssetDragType = 'application/x-qiantie-agent-storyboard-asset';

function assetID(asset) {
  return String(asset?.id || asset?.assetId || '').trim();
}

export function hasStoryboardAssetDrop(dataTransfer) {
  return Array.from(dataTransfer?.types || []).includes(storyboardAssetDragType);
}

export function writeStoryboardAssetDrag(dataTransfer, asset) {
  const id = assetID(asset);
  if (!dataTransfer || !id) return false;
  dataTransfer.effectAllowed = 'copy';
  dataTransfer.setData(storyboardAssetDragType, JSON.stringify({ id }));
  return true;
}

export function readStoryboardAssetDrop(dataTransfer, assets) {
  if (!hasStoryboardAssetDrop(dataTransfer)) return null;
  try {
    const { id } = JSON.parse(dataTransfer.getData(storyboardAssetDragType));
    return (Array.isArray(assets) ? assets : []).find(asset => assetID(asset) === String(id || '').trim()) || null;
  } catch {
    return null;
  }
}
