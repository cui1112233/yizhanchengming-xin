// 列表行漫画页的展示语义。后端已经返回 promptStatus/imageStatus/generationState，
// 前端不再只用“有没有主图”判断，否则「生成中」「生成失败」「还没提取提示词」
// 都会被渲染成同一句“暂无漫画页”，无法区分“没生成”和“生成失败/取不到”。
export function comicInlinePresentation({ page, selectedVideo, loading, error, hasPrimary }) {
  if (hasPrimary) return { statusText: '当前主图', stateText: '已选主图', placeholder: '', tone: 'ready', retry: false };
  if (error) return { statusText: '读取失败', stateText: '读取失败', placeholder: '漫画页读取失败，点击重试', tone: 'error', retry: true };
  if (!page) {
    if (loading) return { statusText: '读取中', stateText: '读取中', placeholder: '正在读取漫画页…', tone: '', retry: false };
    const hasPrompt = String(selectedVideo?.comicPrompt || '').trim().length > 0;
    return hasPrompt
      ? { statusText: '等待生成图片', stateText: '待生成', placeholder: '漫画页尚未生成，点击打开媒体库', tone: '', retry: false }
      : { statusText: '等待提取漫画提示词', stateText: '待分析', placeholder: '尚未提取漫画提示词', tone: '', retry: false };
  }
  if (page.generationState === 'generating') return { statusText: '生成中', stateText: '生成中', placeholder: '漫画页生成中…', tone: 'busy', retry: false };
  if (page.generationState === 'failed') return { statusText: '生成失败', stateText: '失败', placeholder: '漫画页生成失败，点击查看', tone: 'error', retry: false };
  if (page.promptStatus === 'missing') return { statusText: '提示词缺失', stateText: '待分析', placeholder: '尚未提取漫画提示词', tone: '', retry: false };
  if (page.imageStatus === 'candidates') return { statusText: '已有候选图', stateText: '待选主图', placeholder: '已有候选图，打开媒体库选主图', tone: '', retry: false };
  if (page.imageStatus === 'none' && page.status === 'generation_failed') return { statusText: '生成失败', stateText: '失败', placeholder: '漫画页生成失败，点击查看', tone: 'error', retry: false };
  return { statusText: '等待生成图片', stateText: '待生成', placeholder: '等待生成漫画页图片', tone: '', retry: false };
}

// 列表行/缩略图格优先取服务端缩略图；没有 thumbnailUrl 时回落到原图。
// 详情、放大预览必须传 original，保证看到的仍是原图。
export function resolveComicImageSources(image) {
  const original = String(image?.url || '');
  const thumbnail = String(image?.thumbnailUrl || '') || original;
  return { thumbnail, original };
}
