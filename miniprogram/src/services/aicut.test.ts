import { describe, expect, it, vi } from 'vitest';
import { createAICutClient } from './aicut';
import { addAsset, editClip, removeClip } from './aicutEditing';
import { moduleRoute } from '../config/modules';

describe('mobile AI-CUT contract', () => {
  it('has its own entry instead of opening CM canvas', () => expect(moduleRoute('aicut')).toBe('/comic/pages/aicut/index'));
  it('saves the whole timeline with revision and named scope; never retries conflict', async () => {
    const request = vi.fn().mockRejectedValue(Object.assign(new Error('conflict'), { statusCode: 409 }));
    const timeline = { revision: 12, tracks: [], script: 'keep', markers: [{ id: 'm', timeMs: 2 }] };
    await expect(createAICutClient(request).save('p1', 'tl-2', timeline)).rejects.toThrow('conflict');
    expect(request).toHaveBeenCalledExactlyOnceWith('/api/ai-cut/projects/p1/timelines/tl-2/snapshot', {
      method: 'PUT', header: { 'X-AICut-Timeline-Id': 'tl-2' }, data: { expectedRevision: 12, timeline },
    });
  });
  it('creates exports using the saved revision and selected timeline', async () => {
    const request = vi.fn().mockResolvedValue({ job: { id: 'j1', state: 'queued', progress: 0 } });
    expect(await createAICutClient(request).export('p1', 'default', 5)).toMatchObject({ state: 'queued' });
    expect(request).toHaveBeenCalledWith('/api/ai-cut/projects/p1/export-jobs', expect.objectContaining({ data: {
      format: 'mp4', expectedRevision: 5, resolution: '1080p', frameRate: 30, quality: 'standard', codec: 'h264', bitDepth: 8,
    } }));
  });
  it('does not turn malformed reads into an empty project library', async () => {
    await expect(createAICutClient(vi.fn().mockResolvedValue({})).list()).rejects.toThrow('工程列表');
  });
  it('cancels within the selected timeline instead of default', async () => {
    const request = vi.fn().mockResolvedValue({ job: { id: 'j1' } });
    await createAICutClient(request).cancel('p1', 'tl-2', 'j1');
    expect(request).toHaveBeenCalledWith('/api/ai-cut/projects/p1/export-jobs/j1/cancel', expect.objectContaining({ header: { 'X-AICut-Timeline-Id': 'tl-2' } }));
  });
});

const clip = { id: 'c', kind: 'video', assetId: 'a', name: '原片', startMs: 0, durationMs: 4000, sourceStartMs: 0, sourceDurationMs: 6000, volume: 1, extra: { keep: true } };
const timeline = () => ({ revision: 9, tracks: [{ id: 't', kind: 'video', name: '视频', order: 0, locked: false, muted: false, clips: [{ ...clip }] }], script: '原剧本' });
describe('mobile timeline editing', () => {
  it('preserves unknown data without mutating the original', () => {
    const value = timeline(); const next = editClip(value, 't', 'c', { name: '新名', sourceStartMs: 1000, durationMs: 3000 });
    expect(next.tracks[0].clips[0]).toMatchObject({ extra: { keep: true }, sourceStartMs: 1000, durationMs: 3000 });
    expect(next.script).toBe('原剧本'); expect(value.tracks[0].clips[0].name).toBe('原片');
  });
  it('rejects source overflow and locked track edits', () => {
    expect(() => editClip(timeline(), 't', 'c', { sourceStartMs: 4000, durationMs: 3000 })).toThrow('素材时长');
    const locked = timeline(); locked.tracks[0].locked = true;
    expect(() => removeClip(locked, 't', 'c')).toThrow('锁定');
  });
  it('does not trim complex clips through a basic numeric form', () => {
    const value: any = timeline(); value.tracks[0].clips[0].speedCurve = [{ position: 0, rate: 1 }];
    expect(() => editClip(value, 't', 'c', { durationMs: 1000 })).toThrow('复杂');
  });
  it('appends assets to an unlocked matching track and keeps other tracks', () => {
    const value = timeline(); const next = addAsset(value, { id: 'im', name: '插图', kind: 'image', durationMs: 0 }, 'new');
    expect(next.tracks[0].clips[1]).toMatchObject({ assetId: 'im', startMs: 4000, durationMs: 3000, kind: 'image' });
    expect(value.tracks[0].clips).toHaveLength(1);
  });
});
