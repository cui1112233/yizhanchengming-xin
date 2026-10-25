import { normalizeDirectorProject } from '../../pages/agent-director-desk.js';

function requireDirectorNode(nodeID, getNode) {
  const node = getNode?.(nodeID);
  if (!node || node.data?.type !== 'director') throw new Error('导演台节点不存在或已被移除');
  return node;
}

function imageAsset(value) {
  const asset = value?.asset || value;
  if (!asset?.id || asset.kind !== 'image') throw new Error('导演台截图必须保存为图片素材');
  return asset;
}

export function createDirectorDeskHostAdapter({ nodeID, getNode, saveProject, listImages, uploadCapture, uploadImage, onClose } = {}) {
  const id = String(nodeID || '').trim();
  if (!id) throw new Error('缺少导演台节点');

  return {
    readState() {
      const node = requireDirectorNode(id, getNode);
      return normalizeDirectorProject(node.data?.config?.directorDesk || {});
    },
    async writeState(project) {
      requireDirectorNode(id, getNode);
      if (typeof saveProject !== 'function') throw new Error('导演台保存服务不可用');
      const normalized = normalizeDirectorProject(project);
      await saveProject(id, normalized);
      return normalized;
    },
    async listProjectImages() {
      if (typeof listImages !== 'function') return [];
      const result = await listImages();
      const assets = Array.isArray(result?.assets) ? result.assets : Array.isArray(result) ? result : [];
      return assets.filter(asset => asset?.id && asset.kind === 'image');
    },
    async storeCapture(file) {
      requireDirectorNode(id, getNode);
      if (typeof uploadCapture !== 'function') throw new Error('导演台截图上传服务不可用');
      return imageAsset(await uploadCapture(file));
    },
    async uploadProjectImage(file) {
      requireDirectorNode(id, getNode);
      if (typeof uploadImage !== 'function') throw new Error('导演台图片上传服务不可用');
      return imageAsset(await uploadImage(file));
    },
    close() {
      onClose?.(id);
    }
  };
}
