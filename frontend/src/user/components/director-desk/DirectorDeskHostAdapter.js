import { directorProjectCaptureMetadata, normalizeDirectorProject } from '../../pages/agent-director-desk.js';

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

function imageUpload(value) {
  if (typeof Blob === 'undefined' || !(value instanceof Blob) || !String(value.type || '').toLowerCase().startsWith('image/')) {
    throw new Error('导演台只能上传图片文件');
  }
  return value;
}

export async function queueDirectorCaptureMetadata({ saveQueue, project, asset, updateProject } = {}) {
  if (typeof saveQueue?.schedule !== 'function' || typeof saveQueue?.flush !== 'function') {
    throw new Error('导演台保存队列不可用');
  }
  const next = directorProjectCaptureMetadata(project, asset);
  updateProject?.(next);
  saveQueue.schedule(next);
  await saveQueue.flush();
  return next;
}

export function createDirectorCaptureTransaction({ capture, storeCapture, persistMetadata, insertCapture } = {}) {
  let pending = null;
  let inFlight = null;

  const advance = async create => {
    if (inFlight) return inFlight;
    inFlight = (async () => {
      if (!pending && create) {
        if (typeof capture !== 'function') throw new Error('导演台截图服务不可用');
        pending = { phase: 'upload', file: await capture() };
      }
      if (!pending) return null;
      if (pending.phase === 'upload') {
        if (typeof storeCapture !== 'function') throw new Error('导演台截图上传服务不可用');
        pending = { phase: 'metadata', asset: await storeCapture(pending.file) };
      }
      if (pending.phase === 'metadata') {
        if (typeof persistMetadata !== 'function') throw new Error('导演台截图元数据保存服务不可用');
        await persistMetadata(pending.asset);
        pending = { phase: 'canvas', asset: pending.asset };
      }
      if (pending.phase === 'canvas') {
        if (typeof insertCapture !== 'function') throw new Error('导演台截图画布插入服务不可用');
        const asset = pending.asset;
        await insertCapture(asset);
        pending = null;
        return asset;
      }
      return null;
    })().finally(() => { inFlight = null; });
    return inFlight;
  };

  return {
    start: () => advance(true),
    flush: () => advance(false)
  };
}

export function createDirectorDeskHostAdapter({ nodeID, getNode, saveProject, flushCanvas, listImages, uploadCapture, uploadImage, insertCapture, onClose } = {}) {
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
      if (typeof flushCanvas !== 'function') throw new Error('画布持久化服务不可用');
      const normalized = normalizeDirectorProject(project);
      await saveProject(id, normalized);
      await flushCanvas();
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
      const image = imageUpload(file);
      if (typeof uploadCapture !== 'function') throw new Error('导演台截图上传服务不可用');
      return imageAsset(await uploadCapture(image));
    },
    async uploadProjectImage(file) {
      requireDirectorNode(id, getNode);
      const image = imageUpload(file);
      if (typeof uploadImage !== 'function') throw new Error('导演台图片上传服务不可用');
      return imageAsset(await uploadImage(image));
    },
    async insertCapture(asset) {
      requireDirectorNode(id, getNode);
      if (typeof insertCapture !== 'function') throw new Error('导演台截图画布插入服务不可用');
      if (typeof flushCanvas !== 'function') throw new Error('画布持久化服务不可用');
      const image = imageAsset(asset);
      await insertCapture(image);
      await flushCanvas();
      return image;
    },
    close() {
      onClose?.(id);
    }
  };
}
