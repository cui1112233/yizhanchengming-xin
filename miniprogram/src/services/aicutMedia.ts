import Taro from '@tarojs/taro';
import { apiUrl, requireApiOrigin } from '../config/runtime';
import { getBusinessToken, getSessionCookie } from './session';
import { onSessionInvalidated, sessionRevision } from './sessionBoundary';
import { apiRequest } from './api';
import { cutID, cutPath, type CutRecord } from './aicut';
const outputFiles = new Map<string, string>();
onSessionInvalidated(() => {
  for (const filePath of outputFiles.values()) { try { Taro.getFileSystemManager().unlink({ filePath, fail: () => undefined }); } catch { /* sandbox may already be cleared */ } }
  outputFiles.clear();
});

export async function importCutMedia(project: string): Promise<CutRecord | null> {
  requireApiOrigin();
  const revision = sessionRevision();
  let selected: Awaited<ReturnType<typeof Taro.chooseMedia>>;
  try { selected = await Taro.chooseMedia({ count: 1, mediaType: ['image', 'video'], sourceType: ['album', 'camera'] }); }
  catch (error: any) { if (/cancel/i.test(error?.errMsg || '')) return null; throw error; }
  if (revision !== sessionRevision()) throw new Error('账号已切换，请重新打开工程');
  const filePath = selected.tempFiles[0]?.tempFilePath;
  if (!filePath) return null;
  const token = getBusinessToken(), cookie = getSessionCookie();
  const result = await Taro.uploadFile({ url: apiUrl(cutPath(project, '/assets')), filePath, name: 'file', timeout: 180000,
    header: { ...(token ? { Authorization: `Bearer ${token}` } : {}), ...(cookie ? { Cookie: cookie } : {}) },
  });
  if (revision !== sessionRevision()) throw new Error('账号已切换，已丢弃旧工程上传结果');
  let body: any;
  try { body = JSON.parse(result.data); } catch { throw new Error('上传结果未确认，请刷新素材库核对'); }
  if (result.statusCode < 200 || result.statusCode >= 300 || !body?.asset?.id) throw new Error(typeof body?.error === 'string' ? body.error : '上传结果未确认，请刷新素材库核对');
  return body.asset;
}

export async function downloadCutExport(project: string, timeline: string, job: string, format = 'mp4') {
  if (!['mp4', 'webm', 'mp3', 'mov'].includes(format)) throw new Error('成片格式无效');
  const revision = sessionRevision();
  const key = `${revision}:${project}:${timeline}:${job}:${format}`;
  if (outputFiles.has(key)) return outputFiles.get(key)!;
  const data = await apiRequest<ArrayBuffer>(cutPath(project, `/export-jobs/${cutID(job)}/download`), { responseType: 'arraybuffer', timeout: 120000, header: { 'X-AICut-Timeline-Id': cutID(timeline) } });
  if (revision !== sessionRevision()) throw new Error('账号已切换，已取消旧成片读取');
  if (!(data instanceof ArrayBuffer)) throw new Error('成片未返回有效媒体内容');
  const filePath = `${Taro.env.USER_DATA_PATH}/cut-${cutID(project)}-${cutID(job)}-${Date.now()}.${format}`;
  await new Promise<void>((resolve, reject) => Taro.getFileSystemManager().writeFile({ filePath, data, success: () => resolve(), fail: reject }));
  if (revision !== sessionRevision()) { Taro.getFileSystemManager().unlink({ filePath, fail: () => undefined }); throw new Error('账号已切换，已清理旧成片'); }
  outputFiles.set(key, filePath);
  return filePath;
}
