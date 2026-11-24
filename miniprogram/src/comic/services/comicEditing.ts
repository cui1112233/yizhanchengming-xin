import Taro from '@tarojs/taro';
import { apiRequest } from '../../services/api';
import { BATCH_FACTORY_BASE } from '../../services/batchFactoryV11';

function path(batch: string, book: string, suffix: string) {
  if (!batch || !book) throw new Error('请先打开漫画项目');
  return `${BATCH_FACTORY_BASE}/batches/${encodeURIComponent(batch)}/books/${encodeURIComponent(book)}/${suffix}`;
}
export function saveComicAsset(batch: string, book: string, asset: { id: string; revision?: number }, fields: Record<string, unknown>) {
  if (!asset.id || !String(fields.name || '').trim()) throw new Error('请填写资产名称');
  return apiRequest(path(batch, book, `assets/${encodeURIComponent(asset.id)}`), {
    method: 'PATCH', data: { ...fields, expectedRevision: Number(asset.revision || 0) },
  });
}
export function saveComicPagePrompts(batch: string, book: string, video: string, comicPrompt: string, imagePrompt: string) {
  if (!video || !comicPrompt.trim() || !imagePrompt.trim()) throw new Error('漫画提示词与图片提示词都需要填写');
  return apiRequest(path(batch, book, `comic-pages/${encodeURIComponent(video)}`), {
    method: 'PUT', data: { comicPrompt: comicPrompt.trim(), imagePrompt: imagePrompt.trim() },
  });
}
export function setComicAssetPrimary(batch: string, book: string, asset: string, image: string) {
  if (!asset || !image) throw new Error('请先选择图片版本');
  return apiRequest(path(batch, book, `assets/${encodeURIComponent(asset)}/images/${encodeURIComponent(image)}/primary`), { method: 'PUT', data: {} });
}
/** 相册来源通过已有 JSON 上传接口提交，不公开私有媒体。 */
export async function chooseComicImageDataUrl() {
  const picked = await Taro.chooseMedia({ count: 1, mediaType: ['image'], sourceType: ['album', 'camera'] });
  const file = picked.tempFiles[0];
  if (!file?.tempFilePath) throw new Error('没有选择图片');
  if (file.size > 3 * 1024 * 1024) throw new Error('图片超过 3 MB，请压缩后重试');
  const extension = file.tempFilePath.split('.').pop()?.toLowerCase();
  if (!extension || !['jpg', 'jpeg', 'png', 'webp'].includes(extension)) throw new Error('请选择 JPG、PNG 或 WEBP 图片');
  const encoded = await new Promise<string>((resolve, reject) => Taro.getFileSystemManager().readFile({ filePath: file.tempFilePath, encoding: 'base64', success: r => resolve(String(r.data)), fail: reject }));
  return `data:image/${extension === 'jpg' ? 'jpeg' : extension};base64,${encoded}`;
}
export async function replaceComicImage(batch: string, book: string, target: { assetId?: string; videoId?: string }) {
  if (!target.assetId && !target.videoId) throw new Error('请先选择资产或漫画页');
  const dataUrl = await chooseComicImageDataUrl();
  const suffix = target.assetId ? `assets/${encodeURIComponent(target.assetId)}/images/upload` : `comic-pages/${encodeURIComponent(target.videoId!)}/images`;
  return apiRequest(path(batch, book, suffix), { method: 'POST', timeout: 60000, data: { dataUrl } });
}
