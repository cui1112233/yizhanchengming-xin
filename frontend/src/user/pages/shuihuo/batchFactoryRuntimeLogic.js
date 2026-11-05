import { batchFactoryCleanSourceText, batchFactoryPreviewText, batchFactorySourceText } from './batchFactoryContentRange.js';

function normalizedText(value) {
  return String(value ?? '').trim();
}

export function resolveBookProductionText(book) {
  const working = batchFactorySourceText(book?.workingFrontContent);
  if (working) return working;
  return normalizedText(batchFactoryPreviewText(book?.sourceText, book));
}

export function resolveBookEditableContent(book, productionMode) {
  const working = batchFactorySourceText(book?.workingFrontContent);
  if (working) return working;
  if (productionMode === 'comic') return batchFactoryCleanSourceText(book?.sourceText);
  return resolveBookProductionText(book);
}

function stableHash(input) {
  let hash = 0x811c9dc5;
  for (let index = 0; index < input.length; index += 1) {
    hash ^= input.charCodeAt(index);
    hash = Math.imul(hash, 0x01000193);
  }
  return (hash >>> 0).toString(16).padStart(8, '0');
}

export function audioDurationFingerprint(book, tts = {}) {
  return stableHash(JSON.stringify({
    text: resolveBookProductionText(book),
    voice: normalizedText(tts.voice),
    style: normalizedText(tts.style),
    speed: Number(tts.speed ?? 1),
    pitch: Number(tts.pitch ?? 0)
  }));
}

export function batchMediaCounts(books = [], productionStatus = {}, mergeStatus = {}) {
  const bookList = Array.isArray(books) ? books : [];
  const storyboards = bookList.reduce(
    (sum, book) => sum + (Array.isArray(book?.videos) ? book.videos.length : 0),
    0
  );

  // 「可播放」按分镜口径统计：同一个 ordinal 下，active 记录或历史（orphaned）
  // 记录里有任何一个生成成功的任务就算这个分镜可播放。这样既不会把历史轮次的
  // 多条任务累加成「6 可播放」，也不会让仍然存在的上一版成功视频被漏掉。
  const slotByVideoId = new Map();
  const activeSlots = new Set();
  for (const book of bookList) {
    const bookKey = String(book?.id || book?.bookId || '');
    (Array.isArray(book?.videos) ? book.videos : []).forEach((video, index) => {
      const ordinal = Number.isFinite(Number(video?.ordinal)) ? Number(video.ordinal) : index;
      const slot = `${bookKey}:${ordinal}`;
      activeSlots.add(slot);
      if (video?.id) slotByVideoId.set(String(video.id), slot);
    });
    for (const video of Array.isArray(book?.videoHistory) ? book.videoHistory : []) {
      const ordinal = Number.isFinite(Number(video?.ordinal)) ? Number(video.ordinal) : 0;
      if (!video?.id) continue;
      slotByVideoId.set(String(video.id), `${bookKey}:${ordinal}`);
    }
  }

  const playable = new Set();
  for (const job of productionStatus?.jobs || []) {
    for (const task of job?.tasks || []) {
      if (String(task?.status || '').toLowerCase() !== 'succeeded') continue;
      const mediaUrl = normalizedText(task?.mediaUrl);
      if (!mediaUrl) continue;
      const videoId = String(task?.videoId || '');
      if (slotByVideoId.size > 0) {
        const slot = slotByVideoId.get(videoId);
        if (!slot || !activeSlots.has(slot)) continue;
        playable.add(slot);
        continue;
      }
      playable.add(String(task?.id || `${videoId}:${mediaUrl}`));
    }
  }

  const merges = new Set();
  for (const job of mergeStatus?.jobs || []) {
    if (String(job?.status || '').toLowerCase() !== 'succeeded') continue;
    const outputUrl = normalizedText(job?.outputUrl);
    if (!outputUrl) continue;
    merges.add(String(job?.id || `${job?.bookId || ''}:${outputUrl}`));
  }

  return { storyboards, playable: playable.size, merges: merges.size };
}

// batchFactoryNovelCopyText 生成“复制全部小说”的制表符文本：书名 / 书城ID /
// 书城，每本一行，顺序沿用列表展示顺序（即序号升序）。缺失字段留空，
// 始终保留三列，方便直接粘贴到 Excel。bookId 取页面显示的 source_book_id。
export function batchFactoryNovelCopyText(books = [], platformNames = {}) {
  const list = Array.isArray(books) ? books : [];
  return list.map((book) => {
    const metadata = book?.sourceMetadata || {};
    const platformCode = String(metadata.platformCode || '').trim().toUpperCase();
    const platform = String(book?.platform || '').trim().toUpperCase();
    let platformName = String(metadata.platformName || metadata.platformLabel || platformNames[String(book?.platform || '')] || '').trim();
    if (platformCode === 'YG' || platform === 'YG' || platform === '阳光') platformName = '点众付费';
    return [String(book?.title || '').trim(), String(book?.bookId || '').trim(), platformName].join('\t');
  }).join('\n');
}

// bookVideoVersions 把被轮换成 orphaned 的历史视频记录重新挂回同一个分镜。
// 已经生成成功的旧媒体不能因为 active 记录被重建就消失：它仍出现在版本列表里，
// 标记 isHistory，并且可以被选为该分镜的主版本。
export function bookVideoVersions(book, status) {
  const entries = new Map();
  for (const job of status?.jobs || []) {
    for (const task of job?.tasks || []) {
      if (String(task?.status || '').toLowerCase() !== 'succeeded' || !normalizedText(task?.mediaUrl)) continue;
      const versions = entries.get(String(task.videoId || '')) || [];
      versions.push({ ...task, productionJobId: job.id });
      entries.set(String(task.videoId || ''), versions);
    }
  }
  const historyByOrdinal = new Map();
  const historyVideoById = new Map();
  for (const video of Array.isArray(book?.videoHistory) ? book.videoHistory : []) {
    const ordinal = Number.isFinite(Number(video?.ordinal)) ? Number(video.ordinal) : 0;
    if (!video?.id) continue;
    historyVideoById.set(String(video.id), video);
    const ids = historyByOrdinal.get(ordinal) || [];
    ids.push(String(video.id));
    historyByOrdinal.set(ordinal, ids);
  }
  const result = new Map();
  const videos = Array.isArray(book?.videos) ? book.videos : [];
  videos.forEach((video, index) => {
    const ordinal = Number.isFinite(Number(video?.ordinal)) ? Number(video.ordinal) : index;
    const merged = (entries.get(String(video?.id || '')) || []).map(item => ({ ...item, isHistory: false }));
    const seen = new Set(merged.map(item => item.id));
    for (const historyVideoId of historyByOrdinal.get(ordinal) || []) {
      for (const item of entries.get(historyVideoId) || []) {
        if (seen.has(item.id)) continue;
        seen.add(item.id);
        merged.push({ ...item, isHistory: true, historyVideoId });
      }
      // 生产任务列表可能未包含旧槽位的任务；videoHistory 已下发的
      // mediaUrl/productionTaskId 仍应合成一条可播放、可设主版本的历史版本。
      const historyVideo = historyVideoById.get(historyVideoId);
      const taskId = String(historyVideo?.productionTaskId || '');
      const mediaUrl = normalizedText(historyVideo?.mediaUrl);
      if (!taskId || !mediaUrl || seen.has(taskId)) continue;
      seen.add(taskId);
      merged.push({
        id: taskId,
        videoId: historyVideoId,
        mediaUrl,
        provider: historyVideo?.provider || '',
        status: 'succeeded',
        createdAt: historyVideo?.generatedAt || '',
        updatedAt: historyVideo?.generatedAt || '',
        isHistory: true,
        historyVideoId,
        productionTaskId: taskId,
      });
    }
    // 所有成功版本按生成时间倒序展示（最新在前），不折叠、不截断。
    merged.sort((left, right) => String(right?.updatedAt || right?.createdAt || '').localeCompare(String(left?.updatedAt || left?.createdAt || '')));
    result.set(String(video?.id || ''), merged);
  });
  return result;
}
