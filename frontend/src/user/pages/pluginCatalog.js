// 提示词工具在洛水里是整页面板（不是弹窗），href 是面板页地址，供路由按目录派生。
const promptTool = (toolId, promptCategory, progressChunkSize = 0) => ({ kind: 'prompt-tool', toolId, promptCategory, progressChunkSize, href: `/plugins/tool/${toolId}` });
const desktopTool = environment => ({ kind: 'desktop-tool', environment });
// 纯浏览器本地处理，不经过服务端、不调用模型，因此不消耗额度。
const localTool = tool => ({ kind: 'local-tool', tool });
const workspace = name => ({ kind: 'workspace', workspace: name });
const route = href => ({ kind: 'route', href });
// 在 V88 已有能力上真实执行的草稿工具（AI-CUT 项目时间轴）。
const draftTool = tool => ({ kind: 'draft-tool', tool });
// 走服务端火山 AI MediaKit 的超分工作台。
const upscaleTool = () => ({ kind: 'upscale-tool' });
// 洛水封面生成：提示词 + 参考图 + V88 自己的图片生成通道。
const coverTool = () => ({ kind: 'cover-tool' });
// 洛水角色三视图：角色设定 + 参考图 + V88 图片通道，支持本地演练模式。
const turnaroundTool = () => ({ kind: 'turnaround-tool' });
// V88 当前没有可用执行链路的功能，必须在界面上明确标注，不能做成看起来能用的空转。
const unavailable = reason => ({ kind: 'unavailable', reason });

const jianying = desktopTool('剪映草稿与本地媒体执行器');

// 名称、顺序和点击类型来自洛水 v2.9.0 前端目录、处理器映射与工作台代码。
export const PLUGIN_CATALOG = [
  { id: 'asset-reference', name: '资产参考', summary: '文生图/多图参考快速出图', color: '#3b82f6', isNew: true, launch: workspace('asset-reference') },
  { id: 'ai-segment', name: 'AI-分段', summary: '按主轨道 / 文本轨道 / AI对照剧本分集切割', color: '#c084fc', isNew: true, launch: workspace('ai-segment') },
  { id: 'ai-subtitle', name: 'AI-Subtitle', summary: '对照剧本修正剪映语音识别错别字', color: '#22c55e', isNew: true, launch: draftTool('ai-subtitle') },
  { id: 'ai-filler', name: 'AI-Filler', summary: '对照剧本删掉单独成行的语气词', color: '#fb7185', isNew: true, launch: draftTool('ai-filler') },
  { id: 'ai-bilingual-subtitle', name: 'AI-双语字幕', summary: '翻译剪映字幕并新建双语字幕轨', color: '#6366f1', isNew: true, launch: unavailable('缺少字幕翻译模型接入与双语字幕轨写入链路') },
  { id: 'ai-cut-plugin', name: 'AI-CUT插件版', summary: '在线剪辑工作台 · 剪映本地增强', color: '#f43f5e', isNew: true, launch: route('/ai-cut') },
  { id: 'ai-bgm', name: 'AI-BGM', summary: '按剧本和剪映字幕从曲库选乐并写入草稿', color: '#38bdf8', isNew: true, launch: unavailable('缺少本地配乐曲库与草稿音频轨写入链路') },
  { id: 'ai-shot', name: 'AI-Shot', summary: '按剧本和字幕从成片中抽取宣传高光', color: '#f59e0b', isNew: true, launch: unavailable('缺少成片抽帧、高光识别与媒体导出执行器') },
  { id: 'ai-upscale', name: 'AI-Upscale', summary: '视频超分到 1080p / 2K（火山 MediaKit）', color: '#34d399', isNew: true, launch: upscaleTool() },
  { id: 'ai-split', name: 'AI-Split', summary: '按画面变化切开剪映主轨视频', color: '#a855f7', isNew: true, launch: draftTool('ai-split') },
  { id: 'novel-to-drama', name: '小说转剧', summary: '中长篇·纯对话版', color: '#14b8a6', launch: promptTool('novel-to-drama', '中长篇转剧', 2000) },
  { id: 'ai-rewrite', name: 'AI改文', summary: '智能改写文本，优化表达', color: '#60a5fa', launch: route('/plugins/ai-rewrite') },
  { id: 'storyboard-master', name: '分镜大师', summary: '智能生成分镜脚本', color: '#8b5cf6', launch: promptTool('storyboard-master', '分镜', 1500) },
  { id: 'storyboard-organizer', name: '分镜整理', summary: '批量导入分镜文档，按15秒或30秒整理并导出JSON', color: '#8b5cf6', launch: localTool('storyboard-organizer') },
  { id: 'viral-opening', name: '爆款开头', summary: '生成吸引人的开头文案', color: '#f472b6', launch: promptTool('viral-opening', '爆款开头', 0) },
  { id: 'subtitle-alignment', name: '字幕对轴', summary: '字幕时间轴对齐调整', color: '#34d399', launch: desktopTool('音视频与字幕对轴执行器') },
  { id: 'text-format', name: '格式化文本', summary: '文本格式整理与规范化', color: '#fbbf24', launch: localTool('text-format') },
  { id: 'character-edit', name: '角色修改', summary: '批量修改角色名称属性', color: '#fb7185', launch: promptTool('character-edit', '角色修改', 1500) },
  { id: 'voice-clone', name: '声音克隆', summary: 'AI克隆个性化音色', color: '#f97316', launch: desktopTool('声音克隆模型与样本管理工作台') },
  { id: 'subtitle-merge', name: '字幕合并', summary: '多条字幕智能合并处理', color: '#84cc16', launch: localTool('subtitle-merge') },
  { id: 'document-merge', name: '文档合并', summary: '按章节规则批量合并TXT', color: '#14b8a6', launch: localTool('document-merge') },
  { id: 'cover-generation', name: '封面生成', summary: 'AI智能生成精美封面图', color: '#f43f5e', launch: coverTool() },
  { id: 'character-turnaround', name: '角色三视图', summary: '一键生成人物三视图', color: '#6366f1', launch: turnaroundTool() },
  { id: 'voice-design', name: '音色设计', summary: 'AI智能设计角色音色', color: '#06b6d4', launch: desktopTool('音色设计模型与配音工作台') },
  { id: 'draft-merge', name: '草稿合并', summary: '多选剪映草稿按顺序合成一份', color: '#f59e0b', launch: jianying },
  { id: 'jianying-keyframe', name: '剪映关键帧', summary: '为剪映草稿添加关键帧动画', color: '#10b981', launch: jianying }
];

// 4 条提示词工具的整页面板规格，逐项来自洛水 v2.9.0 实测控件表：
// 单栏上下堆叠（输入区 → 主按钮 → 结果区），尺寸、rows、文案按实测值填写。
// 模型和提示词选项不写在这里：必须用 V88 自己的模型库与提示词库。
// formatMode/save 相关字段来自洛水 bundle 静态证据（详见交付说明），不是猜的。
export const PLUGIN_TOOL_PANELS = {
  'novel-to-drama': {
    name: '小说转剧',
    heading: '小说转剧',
    inputPlaceholder: '请输入中长篇小说内容，或上传txt/word文档...',
    inputRows: 6,
    inputHeight: 148,
    emptyInputMessage: '请输入或上传小说内容',
    resultTitle: '生成的剧本',
    resultRows: 10,
    formatLabel: '格式化剧本',
    formatMode: 'script',
    showWordCount: true,
    saveFileName: '中长篇转剧结果',
    saveJson: true,
    iconColor: '#14b8a6',
    // 主按钮配色取自洛水截图实测（小说转剧为青绿 #0ea795）。
    buttonColor: '#0ea795',
    authorized: true,
    concurrent: true
  },
  'storyboard-master': {
    name: '分镜大师',
    heading: 'AI分镜大师',
    inputPlaceholder: '请输入剧本或小说内容，或上传txt/word文档...',
    inputRows: 8,
    inputHeight: 193,
    emptyInputMessage: '请输入或上传内容',
    resultTitle: '生成的分镜',
    resultRows: 8,
    formatLabel: '格式化',
    formatMode: 'lines',
    showWordCount: false,
    saveFileName: 'AI分镜结果',
    saveJson: false,
    iconColor: '#8b5cf6',
    // 主按钮配色取自洛水截图实测（分镜大师为绿 #0eb780）；洛水无「并发生成」。
    buttonColor: '#0eb780',
    authorized: false
  },
  'viral-opening': {
    name: '爆款开头',
    heading: '爆款开头',
    inputPlaceholder: '请输入小说原文，或上传txt/word文档...',
    inputRows: 8,
    inputHeight: 193,
    emptyInputMessage: '请输入或上传内容',
    resultTitle: '生成结果',
    resultRows: 11,
    formatLabel: '格式化',
    formatMode: 'lines',
    showWordCount: false,
    saveFileName: '爆款开头结果',
    saveJson: false,
    iconColor: '#f59430',
    // 主按钮配色取自洛水截图实测（爆款开头为绿 #0eb780）；洛水无「并发生成」。
    buttonColor: '#0eb780',
    authorized: false
  },
  'character-edit': {
    name: '角色修改',
    heading: '角色修改',
    inputPlaceholder: '请输入需要修改角色的内容，或上传txt/word文档...',
    inputRows: 8,
    inputHeight: 193,
    emptyInputMessage: '请输入或上传内容',
    resultTitle: '修改结果',
    resultRows: 8,
    formatLabel: '格式化',
    formatMode: 'lines',
    showWordCount: false,
    saveFileName: '角色修改结果',
    saveJson: false,
    iconColor: '#fb7185',
    // 主按钮配色取自洛水截图实测（角色修改为绿 #0eb780）；洛水无「并发生成」。
    buttonColor: '#0eb780',
    authorized: false
  }
};

export function filterPlugins(catalog, { query = '' } = {}) {
  const normalizedQuery = String(query).trim().toLocaleLowerCase('zh-CN');
  if (!normalizedQuery) return catalog;
  return catalog.filter(plugin => `${plugin.name} ${plugin.summary}`.toLocaleLowerCase('zh-CN').includes(normalizedQuery));
}

export function pluginLaunchTarget(catalog, pluginId) {
  const plugin = catalog.find(item => item.id === pluginId);
  return plugin ? { ...plugin.launch } : null;
}

// 浏览器端当前没有可执行链路的插件，界面上统一标注「未接入」。
export function pluginIsUnavailable(plugin) {
  const kind = plugin?.launch?.kind;
  return kind === 'unavailable' || kind === 'desktop-tool';
}
