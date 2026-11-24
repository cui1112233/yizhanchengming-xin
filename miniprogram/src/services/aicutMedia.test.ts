import { beforeEach, expect, test, vi } from 'vitest';
const h = vi.hoisted(() => ({ revision: 1, invalidate: () => {}, request: vi.fn(), choose: vi.fn(), upload: vi.fn(), write: vi.fn(), unlink: vi.fn() }));
vi.mock('@tarojs/taro', () => ({ default: { env: { USER_DATA_PATH: '/tmp' }, chooseMedia: h.choose, uploadFile: h.upload, getFileSystemManager: () => ({ writeFile: h.write, unlink: h.unlink }) } }));
vi.mock('./api', () => ({ apiRequest: h.request }));
vi.mock('./sessionBoundary', () => ({ sessionRevision: () => h.revision, onSessionInvalidated: (fn: () => void) => { h.invalidate = fn; } }));
vi.mock('./session', () => ({ getBusinessToken: () => 'fake-test-token', getSessionCookie: () => 'fake-test-cookie' }));
vi.mock('../config/runtime', () => ({ apiUrl: (path: string) => `https://test.invalid${path}`, requireApiOrigin: () => {} }));
import { downloadCutExport, importCutMedia } from './aicutMedia';
beforeEach(() => { h.invalidate(); h.revision = 1; h.request.mockReset(); h.upload.mockReset(); h.choose.mockReset(); h.write.mockImplementation(options => options.success()); h.unlink.mockReset(); });
test('downloads named timeline output with matching scope and binary response', async () => {
  h.request.mockResolvedValue(new ArrayBuffer(2));
  expect(await downloadCutExport('p1', 'tl2', 'j1')).toMatch(/\.mp4$/);
  expect(h.request).toHaveBeenCalledWith('/api/ai-cut/projects/p1/export-jobs/j1/download', expect.objectContaining({ responseType: 'arraybuffer', header: { 'X-AICut-Timeline-Id': 'tl2' } }));
});
test('account changes during output read do not save the old account media', async () => {
  h.write.mockClear(); h.request.mockImplementation(async () => { h.revision++; return new ArrayBuffer(2); });
  await expect(downloadCutExport('p1', 'tl2', 'j1')).rejects.toThrow('账号已切换'); expect(h.write).not.toHaveBeenCalled();
});
test('cancelled album selection never uploads', async () => {
  h.choose.mockRejectedValue({ errMsg: 'chooseMedia:fail cancel' });
  expect(await importCutMedia('p1')).toBe(null); expect(h.upload).not.toHaveBeenCalled();
});
test('late album result after account changes cannot upload into a different account', async () => {
  h.choose.mockImplementation(async () => { h.revision++; return { tempFiles: [{ tempFilePath: '/tmp/photo.jpg' }] }; });
  await expect(importCutMedia('p1')).rejects.toThrow('账号已切换'); expect(h.upload).not.toHaveBeenCalled();
});
test('reuses a downloaded output within the account and cleans it on logout', async () => {
  h.request.mockResolvedValue(new ArrayBuffer(2));
  const first = await downloadCutExport('p1', 'tl2', 'j1');
  expect(await downloadCutExport('p1', 'tl2', 'j1')).toBe(first); expect(h.request).toHaveBeenCalledTimes(1);
  h.invalidate(); expect(h.unlink).toHaveBeenCalledWith(expect.objectContaining({ filePath: first }));
});
