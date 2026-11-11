const definitions = [
  ['web-upload', '媒体', '网页上传', '视频/图片/音频上传到项目', 'browser', 'uploadAsset'],
  ['project-assets', '媒体', '项目素材', '读取当前项目已上传的素材', 'cloud', 'openAssets'],
  ['extract-audio', '音频', '提取音频', '从选中的视频分离声音', 'render', 'extractAudio'],
  ['default-text', '文本', '默认文本', '添加和编辑普通字幕文字', 'browser', 'openText'],
  ['batch-text', '文本', '批量编辑', '批量查找并替换字幕文字', 'browser', 'openBatchText'],
  ['visual-effects', '特效', '画面特效', '可预览并进入成片的画面参数', 'render', 'openVisualEffects'],
  ['basic-transition', '转场', '基础转场', '闪黑和闪白转场', 'render', 'openTransitions'],
  ['transcribe', '字幕', '智能识别字幕', '识别语音并写入字幕轨', 'ai', 'openTranscription'],
  ['caption-correction', '字幕', '字幕纠错', '结合剧本检查并修正错字', 'ai', 'openCaptionCorrection'],
  ['subtitle-upload', '字幕', '字幕文件上传', '导入 SRT 字幕文件', 'browser', 'openSubtitleUpload'],
  ['subtitle-download', '字幕', '字幕文件下载', '导出当前字幕为 SRT', 'browser', 'downloadSubtitles'],
  ['clean-speech', 'AI工具', '智能剪口播', '删除口误和长停顿', 'ai', 'openSpeechCleanup'],
  ['smart-split', 'AI工具', '智能切镜', '按画面变化切分所选视频', 'render', 'smartSplit'],
];

export const aicutCapabilityCategories = ['媒体', '音频', '文本', '特效', '转场', '字幕', 'AI工具'];

export function createAICutCapabilities(handlers = {}) {
  return definitions.flatMap(([id, category, title, detail, source, handler]) => {
    const action = handlers[handler];
    return typeof action === 'function' ? [{ id, category, title, detail, source, action }] : [];
  });
}

export function filterAICutCapabilities(items, { category = '', filter = '全部', favorites = [], recent = [], search = '' } = {}) {
  const query = String(search || '').trim().toLowerCase();
  let result = items.filter(item => (!category || item.category === category) && (!query || `${item.title}${item.detail}`.toLowerCase().includes(query)));
  if (filter === '收藏') result = result.filter(item => favorites.includes(item.id));
  if (filter === '最近使用') {
    const positions = new Map(recent.map((id, index) => [id, index]));
    result = result.filter(item => positions.has(item.id)).sort((a, b) => positions.get(a.id) - positions.get(b.id));
  }
  return result;
}

export function toggleFavoriteAICutFeature(items, id) {
  return items.includes(id) ? items.filter(item => item !== id) : [id, ...items];
}

export function recordRecentAICutFeature(items, id, limit = 20) {
  return [id, ...items.filter(item => item !== id)].slice(0, limit);
}
