import test from 'node:test';
import assert from 'node:assert/strict';
import { analyzeAICutCaptionCorrections, analyzeAICutFillerCandidates, applyAICutCaptionCorrections, applyAICutFillerCandidates, cancelAICutASRJob, cancelAICutExportJob, cleanAICutOralTimeline, cloneAICutVoice, consolidateAICutClips, correctAICutCaptions, createAICutASRJob, createAICutExportJob, createAICutProject, downloadAICutExportJob, exportAICutMedia, extractAICutAudio, extractAICutTrackingFrames, extractAICutVisualFrames, generateAICutProxy, generateAICutSpeech, getAICutAudioAnalysis, getAICutExportJob, getAICutTimelineHistory, importAICutSRT, listAICutASRJobs, listAICutExportJobs, listAICutProjects, loadAICutTimeline, normalizeAICutConsolidateInput, normalizeAICutTTSInput, normalizeAICutVoiceCloneInput, redoAICutTimeline, removeAICutImageBackground, saveAICutTimeline, smartSplitAICutVideo, trimAICutFillers, undoAICutTimeline, updateAICutProject, uploadAICutAsset } from './aicut.js';

test('AI-CUT持久撤销重做使用服务端时间线历史', async () => {
  const calls = [];
  const request = async (path, options = {}) => {
    calls.push({ path, options });
    if (path.endsWith('/undo')) return { timeline: { revision: 4, script: '第一版', tracks: [] }, history: { canUndo: true, canRedo: true } };
    if (path.endsWith('/redo')) return { timeline: { revision: 5, script: '第二版', tracks: [] }, history: { canUndo: true, canRedo: false } };
    return { history: { canUndo: true, canRedo: false } };
  };
  assert.equal((await getAICutTimelineHistory('cut-1', request)).canUndo, true);
  assert.equal((await undoAICutTimeline('cut-1', 3, request)).timeline.script, '第一版');
  assert.equal((await redoAICutTimeline('cut-1', 4, request)).timeline.script, '第二版');
  assert.equal(calls[0].path, '/api/ai-cut/projects/cut-1/timeline/history');
  assert.deepEqual(JSON.parse(calls[1].options.body), { expectedRevision: 3 });
  assert.deepEqual(JSON.parse(calls[2].options.body), { expectedRevision: 4 });
});

test('AI-CUT项目和时间线使用独立登录接口', async () => {
  const calls = [];
  const request = async (path, options = {}) => {
    calls.push({ path, options });
    if (path === '/api/ai-cut/projects' && options.method === 'POST') return { project: { id: 'cut-1', title: '竖屏口播' } };
    if (path.endsWith('/timeline') && options.method === 'PUT') return { timeline: { revision: 2, tracks: [] } };
    if (path.endsWith('/timeline')) return { timeline: { revision: 1, tracks: [] } };
    return { projects: [] };
  };

  assert.deepEqual(await listAICutProjects(request), []);
  assert.equal((await createAICutProject({ title: '竖屏口播', aspectRatio: '9:16' }, request)).id, 'cut-1');
  assert.equal((await loadAICutTimeline('cut-1', request)).revision, 1);
  assert.equal((await saveAICutTimeline('cut-1', { revision: 1, tracks: [] }, request)).revision, 2);
  assert.deepEqual(JSON.parse(calls[3].options.body), { expectedRevision: 1, timeline: { revision: 1, tracks: [] } });
});

test('AI-CUT画布背景设置会持久化到项目', async () => {
  let received;
  const project = await updateAICutProject('cut-1', { backgroundMode: 'transparent', backgroundColor: '#224466' }, async (path, options) => {
    received = { path, options };
    return { project: { id: 'cut-1', backgroundMode: 'transparent', backgroundColor: '#224466' } };
  });
  assert.equal(received.path, '/api/ai-cut/projects/cut-1');
  assert.equal(received.options.method, 'PATCH');
  assert.deepEqual(JSON.parse(received.options.body), { backgroundMode: 'transparent', backgroundColor: '#224466' });
  assert.equal(project.backgroundMode, 'transparent');
  await assert.rejects(() => updateAICutProject('cut-1', { backgroundMode: 'gradient', backgroundColor: '#000000' }, async () => ({})), /背景模式/);
  await assert.rejects(() => updateAICutProject('cut-1', { backgroundMode: 'color', backgroundColor: 'black' }, async () => ({})), /背景颜色/);
});

test('AI-CUT素材上传保留原文件名并使用长任务超时', async () => {
  const video = new Blob(['video'], { type: 'video/mp4' });
  Object.defineProperty(video, 'name', { value: 'source.mp4' });
  let received;
  const asset = await uploadAICutAsset('cut-1', video, async (path, options) => {
    received = { path, options };
    return { asset: { id: 'asset-1', name: 'source.mp4' } };
  });
  assert.equal(received.path, '/api/ai-cut/projects/cut-1/assets');
  assert.ok(received.options.body instanceof FormData);
  assert.equal(received.options.body.get('file').name, 'source.mp4');
  assert.equal(received.options.timeoutMs, 0);
  assert.equal(asset.id, 'asset-1');
});

test('AI-CUT文字转语音生成持久项目素材并传递取消信号', async () => {
  const controller = new AbortController();
  let received;
  const asset = await generateAICutSpeech('cut-1', {
    text: '  这是旁白  ', name: '第一段', voice: 'zh-CN-YunxiNeural', speed: 1.2, pitch: 4, style: 'friendly', signal: controller.signal
  }, async (path, options) => {
    received = { path, options };
    return { asset: { id: 'tts-asset-1', kind: 'audio', durationMs: 1800 } };
  });
  assert.equal(received.path, '/api/ai-cut/projects/cut-1/tts');
  assert.equal(received.options.method, 'POST');
  assert.equal(received.options.signal, controller.signal);
  assert.equal(received.options.timeoutMs, 0);
  assert.deepEqual(JSON.parse(received.options.body), {
    text: '这是旁白', name: '第一段', voice: 'zh-CN-YunxiNeural', speed: 1.2, pitch: 4, style: 'friendly'
  });
  assert.equal(asset.id, 'tts-asset-1');
});

test('AI-CUT文字转语音在请求前拒绝空值和越界参数', async () => {
  assert.throws(() => normalizeAICutTTSInput({ text: ' ' }), /不能为空/);
  assert.throws(() => normalizeAICutTTSInput({ text: '测试', speed: .4 }), /语速/);
  assert.throws(() => normalizeAICutTTSInput({ text: '测试', pitch: 51 }), /音调/);
  assert.equal(normalizeAICutTTSInput({ text: '测试' }).voice, 'zh-CN-XiaoxiaoNeural');
});

test('AI-CUT代理媒体生成使用原素材身份并允许长时间转码', async () => {
  let received;
  const asset = await generateAICutProxy('cut-1', 'video-1', async (path, options) => {
    received = { path, options };
    return { asset: { id: 'video-1', url: '/content', proxyUrl: '/proxy' }, reused: false };
  });
  assert.equal(received.path, '/api/ai-cut/projects/cut-1/assets/video-1/proxy');
  assert.equal(received.options.method, 'POST');
  assert.equal(received.options.timeoutMs, 0);
  assert.equal(asset.id, 'video-1');
  assert.equal(asset.proxyUrl, '/proxy');
  await assert.rejects(() => generateAICutProxy('cut-1', '../video', async () => ({})), /素材无效/);
});

test('AI-CUT渲染合并提交唯一片段、时间线版本和取消信号', async () => {
  const controller = new AbortController();
  let received;
  const asset = await consolidateAICutClips('cut-1', {
    clipIds: [' clip-a ', 'clip-b', 'clip-a'], expectedRevision: 7, signal: controller.signal
  }, async (path, options) => {
    received = { path, options };
    return { asset: { id: 'merged-1', kind: 'video', durationMs: 2200 } };
  });
  assert.equal(received.path, '/api/ai-cut/projects/cut-1/consolidate');
  assert.equal(received.options.method, 'POST');
  assert.equal(received.options.timeoutMs, 0);
  assert.equal(received.options.signal, controller.signal);
  assert.deepEqual(JSON.parse(received.options.body), { clipIds: ['clip-a', 'clip-b'], expectedRevision: 7 });
  assert.equal(asset.id, 'merged-1');
});

test('AI-CUT渲染合并在请求前拒绝空值和非法版本', () => {
  assert.throws(() => normalizeAICutConsolidateInput({ clipIds: ['only'], expectedRevision: 1 }), /2 到 50/);
  assert.throws(() => normalizeAICutConsolidateInput({ clipIds: ['a', 'b'], expectedRevision: -1 }), /版本无效/);
});

test('AI-CUT图片背景移除创建新素材并传递取消信号', async () => {
  const controller = new AbortController();
  let received;
  const asset = await removeAICutImageBackground('cut-1', 'image-1', { name: '人物透明图', signal: controller.signal }, async (path, options) => {
    received = { path, options };
    return { asset: { id: 'transparent-1', kind: 'image', contentType: 'image/png' } };
  });
  assert.equal(received.path, '/api/ai-cut/projects/cut-1/assets/image-1/remove-background');
  assert.equal(received.options.method, 'POST');
  assert.equal(received.options.timeoutMs, 0);
  assert.equal(received.options.signal, controller.signal);
  assert.deepEqual(JSON.parse(received.options.body), { name: '人物透明图' });
  assert.equal(asset.id, 'transparent-1');
  await assert.rejects(() => removeAICutImageBackground('cut-1', '../image', {}, async () => ({})), /素材无效/);
});

test('AI-CUT声音克隆提交参考素材、试听文本和取消信号', async () => {
  let received;
  const controller = new AbortController();
  const asset = await cloneAICutVoice('cut-1', {
    assetId: 'voice-reference', text: '这是克隆音色测试', name: '角色音色', speed: 1.2, pitch: -5, style: 'calm', signal: controller.signal
  }, async (path, options) => {
    received = { path, options };
    return { asset: { id: 'voice-cloned', kind: 'audio' } };
  });
  assert.equal(received.path, '/api/ai-cut/projects/cut-1/voice-clone');
  assert.equal(received.options.method, 'POST');
  assert.equal(received.options.timeoutMs, 0);
  assert.equal(received.options.signal, controller.signal);
  assert.deepEqual(JSON.parse(received.options.body), { assetId: 'voice-reference', text: '这是克隆音色测试', name: '角色音色', speed: 1.2, pitch: -5, style: 'calm' });
  assert.equal(asset.id, 'voice-cloned');
});

test('AI-CUT声音克隆在发请求前拒绝空值和越界参数', () => {
  assert.throws(() => normalizeAICutVoiceCloneInput({ assetId: '', text: '测试' }), /请选择参考音频/);
  assert.throws(() => normalizeAICutVoiceCloneInput({ assetId: 'audio-1', text: ' ' }), /试听文本不能为空/);
  assert.throws(() => normalizeAICutVoiceCloneInput({ assetId: 'audio-1', text: '测试', speed: 3 }), /语速/);
  assert.throws(() => normalizeAICutVoiceCloneInput({ assetId: 'audio-1', text: '测试', pitch: 60 }), /音调/);
});

test('AI-CUT字幕导入发送文件内容和当前版本', async () => {
  let received;
  const result = await importAICutSRT('cut-1', { name: 'dialogue.srt', text: '1\n00:00:00,000 --> 00:00:01,000\n你好' }, 3, async (path, options) => {
    received = { path, options };
    return { timeline: { revision: 4, tracks: [] }, imported: 1, trackId: 'subtitle-1' };
  });
  assert.equal(received.path, '/api/ai-cut/projects/cut-1/subtitles/import');
  assert.deepEqual(JSON.parse(received.options.body), {
    filename: 'dialogue.srt',
    srt: '1\n00:00:00,000 --> 00:00:01,000\n你好',
    expectedRevision: 3
  });
  assert.equal(result.imported, 1);
  assert.equal(result.timeline.revision, 4);
});

test('AI-CUT字幕选择取消时不发请求', async () => {
  let called = false;
  const result = await importAICutSRT('cut-1', null, 0, async () => { called = true; });
  assert.equal(result, null);
  assert.equal(called, false);
});

test('AI-CUT字幕纠错发送剧本、模型、时间线版本和取消信号', async () => {
  let received;
  const controller = new AbortController();
  const result = await correctAICutCaptions('cut-1', {
    script: '故事发生在洛水之畔。', modelId: 'text-main', expectedRevision: 4, signal: controller.signal
  }, async (path, options) => {
    received = { path, options };
    return { timeline: { revision: 5, tracks: [] }, applied: 2, message: '字幕纠错完成' };
  });
  assert.equal(received.path, '/api/ai-cut/projects/cut-1/captions/correct');
  assert.equal(received.options.signal, controller.signal);
  assert.deepEqual(JSON.parse(received.options.body), { script: '故事发生在洛水之畔。', modelId: 'text-main', expectedRevision: 4 });
  assert.equal(result.applied, 2);
});

test('AI-CUT水词删除发送模型、时间线版本和取消信号', async () => {
  let received;
  const controller = new AbortController();
  const result = await trimAICutFillers('cut-1', { modelId: 'text-main', expectedRevision: 5, signal: controller.signal }, async (path, options) => {
    received = { path, options };
    return { timeline: { revision: 6, tracks: [] }, summary: { fillerCount: 1, pauseCount: 0, removedMs: 500 } };
  });
  assert.equal(received.path, '/api/ai-cut/projects/cut-1/filler-trim');
  assert.equal(received.options.signal, controller.signal);
  assert.deepEqual(JSON.parse(received.options.body), { modelId: 'text-main', expectedRevision: 5 });
  assert.equal(result.summary.removedMs, 500);
});

test('AI-CUT Agent去口误停顿使用独立时间线裁剪接口', async () => {
  let received;
  const controller = new AbortController();
  await cleanAICutOralTimeline('cut-1', { modelId: 'text-main', maxPauseMs: 1000, expectedRevision: 6, signal: controller.signal }, async (path, options) => {
    received = { path, options };
    return { timeline: { revision: 7, tracks: [] }, summary: { fillerCount: 1, pauseCount: 1, removedMs: 1500 } };
  });
  assert.equal(received.path, '/api/ai-cut/projects/cut-1/oral-clean');
  assert.equal(received.options.signal, controller.signal);
  assert.deepEqual(JSON.parse(received.options.body), { modelId: 'text-main', maxPauseMs: 1000, expectedRevision: 6 });
});

test('AI-CUT语音识别任务提交、查询和取消使用持久任务接口', async () => {
  const calls = [];
  const request = async (path, options = {}) => {
    calls.push({ path, options });
    if (path.endsWith('/cancel')) return { job: { id: 'asr-1', state: 'canceled' } };
    if (options.method === 'POST') return { job: { id: 'asr-1', state: 'queued' } };
    return { jobs: [{ id: 'asr-1', state: 'running' }] };
  };
  const created = await createAICutASRJob('cut-1', { assetId: 'audio-1', modelId: 'asr-main', language: 'zh', expectedRevision: 2 }, request);
  assert.equal(created.state, 'queued');
  assert.deepEqual(JSON.parse(calls[0].options.body), { assetId: 'audio-1', modelId: 'asr-main', language: 'zh', expectedRevision: 2 });
  assert.equal((await listAICutASRJobs('cut-1', request))[0].state, 'running');
  assert.equal((await cancelAICutASRJob('cut-1', 'asr-1', request)).state, 'canceled');
  assert.equal(calls[2].path, '/api/ai-cut/projects/cut-1/asr-jobs/asr-1/cancel');
});

test('AI-CUT成片导出走服务端渲染并返回文件', async () => {
  let received;
  const expected = new Blob(['rendered'], { type: 'video/mp4' });
  const result = await exportAICutMedia('cut-1', 'mp4', 7, async (path, options) => {
    received = { path, options };
    return expected;
  });
  assert.equal(result, expected);
  assert.equal(received.path, '/api/ai-cut/projects/cut-1/exports/mp4');
  assert.equal(received.options.responseType, 'blob');
  assert.equal(received.options.timeoutMs, 0);
  assert.deepEqual(JSON.parse(received.options.body), { expectedRevision: 7 });
  await assert.rejects(() => exportAICutMedia('cut-1', 'mov', 7, async () => expected), /仅支持/);
});

test('AI-CUT异步导出支持进度、取消和完成后下载', async () => {
  const calls = [];
  const blob = new Blob(['rendered'], { type: 'video/mp4' });
  const request = async (path, options = {}) => {
    calls.push({ path, options });
    if (path.endsWith('/download')) return blob;
    if (path.endsWith('/cancel')) return { job: { id: 'export-1', state: 'cancel_requested', progress: 35 } };
    if (options.method === 'POST') return { job: { id: 'export-1', state: 'queued', progress: 0 } };
    return { job: { id: 'export-1', state: 'running', progress: 35 } };
  };
  assert.equal((await createAICutExportJob('cut-1', 'mp4', 7, request)).state, 'queued');
  assert.equal((await getAICutExportJob('cut-1', 'export-1', request)).progress, 35);
  assert.equal((await cancelAICutExportJob('cut-1', 'export-1', request)).state, 'cancel_requested');
  assert.equal(await downloadAICutExportJob('cut-1', 'export-1', request), blob);
  assert.deepEqual(JSON.parse(calls[0].options.body), { format: 'mp4', expectedRevision: 7 });
  assert.equal(calls[3].options.responseType, 'blob');
});

test('AI-CUT批量导出队列读取当前项目的全部任务', async () => {
  let received;
  const jobs = await listAICutExportJobs('cut-1', async (path, options = {}) => {
    received = { path, options };
    return { jobs: [{ id: 'export-2', state: 'queued' }, { id: 'export-1', state: 'succeeded' }] };
  });
  assert.equal(received.path, '/api/ai-cut/projects/cut-1/export-jobs');
  assert.equal(received.options.suppressGlobalError, true);
  assert.deepEqual(jobs.map(job => job.id), ['export-2', 'export-1']);
});

test('AI-CUT异步导出把专业编码参数发送给服务端并拒绝非法值', async () => {
  const calls = [];
  const request = async (path, options = {}) => {
    calls.push({ path, options });
    return { job: { id: 'export-pro', state: 'queued', progress: 0 } };
  };
  await createAICutExportJob('cut-1', 'mp4', 8, { resolution: '4k', frameRate: 60, quality: 'high', codec: 'hevc', bitDepth: 10, bitrateMbps: 35, hardwareAcceleration: 'software' }, request);
  assert.deepEqual(JSON.parse(calls[0].options.body), {
    format: 'mp4', expectedRevision: 8, resolution: '4k', frameRate: 60, quality: 'high', codec: 'hevc', bitDepth: 10, bitrateMbps: 35, hardwareAcceleration: 'software'
  });
  await assert.rejects(() => createAICutExportJob('cut-1', 'mp4', 8, { resolution: '8k', frameRate: 30, quality: 'high' }, request), /分辨率/);
  await assert.rejects(() => createAICutExportJob('cut-1', 'mp4', 8, { resolution: '1080p', frameRate: 29, quality: 'high' }, request), /帧率/);
  await assert.rejects(() => createAICutExportJob('cut-1', 'mp4', 8, { resolution: '1080p', frameRate: 30, quality: 'lossless' }, request), /质量/);
  await assert.rejects(() => createAICutExportJob('cut-1', 'mp4', 8, { resolution: '1080p', frameRate: 30, quality: 'high', codec: 'h264', bitDepth: 10 }, request), /10-bit/);
  await assert.rejects(() => createAICutExportJob('cut-1', 'mp4', 8, { resolution: '1080p', frameRate: 30, quality: 'high', codec: 'hevc', bitDepth: 10, bitrateMbps: 101 }, request), /码率/);
  await assert.rejects(() => createAICutExportJob('cut-1', 'mp4', 8, { resolution: '1080p', frameRate: 30, quality: 'high', hardwareAcceleration: 'gpu-only' }, request), /硬件加速/);
});

test('AI-CUT透明导出仅允许VP9或ProRes4444', async () => {
  const calls = [];
  const request = async (path, options = {}) => {
    calls.push({ path, options });
    return { job: { id: 'alpha-1', state: 'queued' } };
  };
  await createAICutExportJob('cut-1', 'webm', 2, { codec: 'vp9', alpha: true }, request);
  await createAICutExportJob('cut-1', 'mov', 2, { codec: 'prores4444', bitDepth: 12, alpha: true }, request);
  assert.equal(JSON.parse(calls[0].options.body).alpha, true);
  assert.equal(JSON.parse(calls[1].options.body).codec, 'prores4444');
  await assert.rejects(() => createAICutExportJob('cut-1', 'mp4', 2, { codec: 'h264', alpha: true }, request), /透明/);
});

test('AI-CUT智能切镜和分离音频使用真实媒体接口', async () => {
  const calls = [];
  const request = async (path, options) => {
    calls.push({ path, options });
    if (path.endsWith('/extract-audio')) return { asset: { id: 'audio-1', kind: 'audio' } };
    return { timeline: { revision: 5, tracks: [] }, splitCount: 3 };
  };
  const split = await smartSplitAICutVideo('cut-1', { assetId: 'video-1', threshold: 0.35, minIntervalMs: 700, expectedRevision: 4 }, request);
  assert.equal(split.splitCount, 3);
  assert.deepEqual(JSON.parse(calls[0].options.body), { assetId: 'video-1', threshold: 0.35, minIntervalMs: 700, expectedRevision: 4 });
  assert.equal(calls[0].options.timeoutMs, 0);
  assert.equal((await extractAICutAudio('cut-1', 'video-1', request)).id, 'audio-1');
  assert.equal(calls[1].path, '/api/ai-cut/projects/cut-1/extract-audio');
});

test('AI-CUT视觉理解先调用服务端真实抽帧接口', async () => {
  let received;
  const result = await extractAICutVisualFrames('cut-1', 'video-1', 6, async (path, options) => {
    received = { path, options };
    return { frames: [{ timeMs: 1000, url: 'data:image/jpeg;base64,AA==' }], message: '已提取 1 张代表帧' };
  });
  assert.equal(received.path, '/api/ai-cut/projects/cut-1/visual-frames');
  assert.equal(received.options.timeoutMs, 0);
  assert.deepEqual(JSON.parse(received.options.body), { assetId: 'video-1', count: 6 });
  assert.equal(result.frames[0].timeMs, 1000);
});

test('AI-CUT区域跟踪按片段源区间调用低分辨率连续抽帧接口', async () => {
  let received;
  const controller = new AbortController();
  const result = await extractAICutTrackingFrames('cut-1', {
    assetId: 'video-1', startMs: 1200, durationMs: 3000, count: 13, width: 192, signal: controller.signal
  }, async (path, options) => {
    received = { path, options };
    return { frames: [{ timeMs: 1200, url: 'data:image/jpeg;base64,AA==' }], message: '已提取 1 张跟踪帧' };
  });
  assert.equal(received.path, '/api/ai-cut/projects/cut-1/tracking-frames');
  assert.equal(received.options.timeoutMs, 0);
  assert.equal(received.options.signal, controller.signal);
  assert.deepEqual(JSON.parse(received.options.body), { assetId: 'video-1', startMs: 1200, durationMs: 3000, count: 13, width: 192 });
  assert.equal(result.frames[0].timeMs, 1200);
  await assert.rejects(() => extractAICutTrackingFrames('cut-1', { assetId: '', startMs: 0, durationMs: 1000, count: 5, width: 192 }, async () => ({})), /请选择视频素材/);
  await assert.rejects(() => extractAICutTrackingFrames('cut-1', { assetId: 'video-1', startMs: 0, durationMs: 0, count: 5, width: 192 }, async () => ({})), /跟踪区间/);
  await assert.rejects(() => extractAICutTrackingFrames('cut-1', { assetId: 'video-1', startMs: 0, durationMs: 1000, count: 61, width: 192 }, async () => ({})), /跟踪帧数量/);
});

test('AI-CUT音频波形和节拍分析使用可缓存的只读接口', async () => {
  let received;
  const result = await getAICutAudioAnalysis('cut-1', 'audio-1', 240, async (path, options) => {
    received = { path, options };
    return { analysis: { durationMs: 2000, peaks: [0.1, 0.8], beatsMs: [500, 1000] } };
  });
  assert.equal(received.path, '/api/ai-cut/projects/cut-1/audio-analysis?assetId=audio-1&points=240');
  assert.equal(received.options.timeoutMs, 0);
  assert.equal(result.beatsMs.length, 2);
  await assert.rejects(() => getAICutAudioAnalysis('cut-1', '', 240, async () => ({})), /请选择音频或视频素材/);
  await assert.rejects(() => getAICutAudioAnalysis('cut-1', 'audio-1', 4, async () => ({})), /采样点/);
});

test('AI-CUT字幕纠错先预览分析再写回修改结果', async () => {
  const calls = [];
  const request = async (path, options = {}) => {
    calls.push({ path, options });
    if (JSON.parse(options.body).preview) return { preview: [{ clipId: 'sub-1', oldText: '落水之畔', newText: '洛水之畔', startMs: 1200, endMs: 3000, reason: '剧本原词' }], applied: 1, revision: 4, message: '分析完成' };
    return { timeline: { revision: 5, tracks: [] }, applied: 1, message: '字幕纠错已写回' };
  };
  const controller = new AbortController();

  const analysis = await analyzeAICutCaptionCorrections('cut-1', {
    script: '  故事发生在洛水之畔。 ', modelId: ' text-main ', expectedRevision: 4, signal: controller.signal
  }, request);
  assert.equal(calls[0].path, '/api/ai-cut/projects/cut-1/captions/correct');
  assert.equal(calls[0].options.method, 'POST');
  assert.equal(calls[0].options.signal, controller.signal);
  assert.equal(calls[0].options.timeoutMs, 0);
  assert.deepEqual(JSON.parse(calls[0].options.body), { script: '故事发生在洛水之畔。', modelId: 'text-main', expectedRevision: 4, preview: true });
  assert.equal(analysis.preview[0].newText, '洛水之畔');
  assert.equal(analysis.revision, 4);

  assert.equal((await applyAICutCaptionCorrections('cut-1', {
    expectedRevision: 4, corrections: [{ clipId: 'sub-1', text: '洛水之畔', reason: '剧本原词' }], signal: controller.signal
  }, request)).applied, 1);
  assert.equal(calls[1].path, '/api/ai-cut/projects/cut-1/captions/correct');
  assert.equal(calls[1].options.method, 'POST');
  assert.equal(calls[1].options.signal, controller.signal);
  assert.deepEqual(JSON.parse(calls[1].options.body), { expectedRevision: 4, corrections: [{ clipId: 'sub-1', text: '洛水之畔', reason: '剧本原词' }] });

  await assert.rejects(() => analyzeAICutCaptionCorrections('cut-1', { script: '  ', expectedRevision: 4 }, request), /校对剧本不能为空/);
  assert.equal(calls.length, 2);
});

test('AI-CUT去水词先预览候选再删除字幕', async () => {
  const calls = [];
  const request = async (path, options = {}) => {
    calls.push({ path, options });
    if (JSON.parse(options.body).preview) return { preview: [{ clipId: 'sub-1', startMs: 0, endMs: 500, text: '嗯', reason: '整行只有语气词' }], summary: { fillerCount: 1, pauseCount: 0, removedMs: 500 }, revision: 6, message: '分析完成' };
    return { timeline: { revision: 7, tracks: [] }, summary: { fillerCount: 1, pauseCount: 0, removedMs: 500 }, message: '水词字幕删除完成' };
  };
  const controller = new AbortController();

  const analysis = await analyzeAICutFillerCandidates('cut-1', { modelId: ' text-main ', expectedRevision: 6, signal: controller.signal }, request);
  assert.equal(calls[0].path, '/api/ai-cut/projects/cut-1/filler-trim');
  assert.equal(calls[0].options.method, 'POST');
  assert.equal(calls[0].options.signal, controller.signal);
  assert.equal(calls[0].options.timeoutMs, 0);
  assert.deepEqual(JSON.parse(calls[0].options.body), { modelId: 'text-main', expectedRevision: 6, preview: true });
  assert.equal(analysis.preview[0].text, '嗯');
  assert.equal(analysis.summary.removedMs, 500);

  assert.equal((await applyAICutFillerCandidates('cut-1', {
    expectedRevision: 6, fillers: [{ clipId: 'sub-1', reason: '整行只有语气词' }], signal: controller.signal
  }, request)).summary.fillerCount, 1);
  assert.equal(calls[1].path, '/api/ai-cut/projects/cut-1/filler-trim');
  assert.equal(calls[1].options.method, 'POST');
  assert.equal(calls[1].options.signal, controller.signal);
  assert.deepEqual(JSON.parse(calls[1].options.body), { expectedRevision: 6, fillers: [{ clipId: 'sub-1', reason: '整行只有语气词' }] });
});
