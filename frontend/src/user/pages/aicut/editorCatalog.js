const numberParameter = (key, labelZh, defaultValue, min, max, step) => ({
  key, labelZh, type: 'number', defaultValue, min, max, step,
});

const durationParameter = () => numberParameter('durationMs', '持续时间', 500, 100, 3000, 50);
const intensityParameter = () => numberParameter('intensity', '强度', 0.5, 0, 1, 0.05);

function resource(category, id, labelZh, description, options = {}) {
  return Object.freeze({
    id: `${category}-${id}`,
    labelZh,
    category,
    description,
    parameters: options.parameters || [],
    supportedKinds: options.supportedKinds || ['video', 'image'],
    previewAdapter: options.previewAdapter || `${category}:${id}`,
    renderAdapter: options.renderAdapter || `${category}:${id}`,
  });
}

const filterDefinitions = [
  ['original', '原片', '不改变素材原有色彩'],
  ['bright', '明亮', '提升整体亮度和通透感'],
  ['clear', '清晰', '增强细节和局部对比'],
  ['natural', '自然', '轻微校色并保持真实肤色'],
  ['warm-sun', '暖阳', '增加柔和的暖黄色调'],
  ['cool-blue', '冷蓝', '增加清爽的冷蓝色调'],
  ['cinema', '电影', '压低高光并增强电影感'],
  ['retro', '复古', '模拟褪色胶片质感'],
  ['monochrome', '黑白', '转换为黑白影像'],
  ['high-contrast', '高对比', '强化明暗反差'],
  ['low-saturation', '低饱和', '降低颜色浓度'],
  ['teal-orange', '青橙', '形成青橙电影配色'],
  ['night', '夜景', '提亮暗部并控制噪点'],
  ['portrait', '人像', '柔化肤色并突出人物'],
  ['vivid', '鲜艳', '提升色彩鲜艳程度'],
  ['soft', '柔和', '降低反差并柔化画面'],
  ['golden-hour', '黄金时刻', '模拟日落前后的温暖金色光线'],
  ['noir', '黑色电影', '高反差黑白电影质感'],
  ['vintage', '怀旧', '褪色、暖调并带轻微年代感'],
  ['cyberpunk', '赛博朋克', '强化蓝紫霓虹与高反差色彩'],
  ['vaporwave', '蒸汽波', '粉蓝高饱和复古电子色调'],
  ['mint-haze', '薄荷雾', '低反差的清凉薄荷色雾感'],
  ['bleach-bypass', '漂白旁路', '低饱和并强化银盐般的反差'],
  ['winter-cool', '冬日冷调', '降低饱和并增加冷蓝空气感'],
  ['rosy', '玫瑰', '增加柔和玫瑰红与肤色光泽'],
  ['film', '胶片', '添加胶片反差、颗粒与暗角'],
  ['super-8', 'Super 8 胶片', '模拟 Super 8 的暖色颗粒与暗角'],
  ['instant-print', '拍立得', '低反差、浅色和暖白相纸质感'],
];

const effectDefinitions = [
  ['blur', '模糊', '对画面添加可调节模糊', [intensityParameter(), numberParameter('radius', '半径', 8, 0, 40, 1)]],
  ['vignette', '暗角', '压暗画面四周以突出主体'],
  ['glow', '发光', '为高光区域添加柔光'],
  ['grain', '颗粒', '添加可控胶片颗粒'],
  ['shake', '抖动', '模拟手持镜头抖动'],
  ['breathe-zoom', '呼吸缩放', '缓慢循环缩放画面'],
  ['flash', '闪烁', '按速度改变画面亮度'],
  ['glitch', '故障', '添加数字信号故障效果'],
  ['chromatic', '色差', '错开红蓝通道形成色差'],
  ['mosaic', '马赛克', '对画面进行像素化处理'],
  ['fisheye', '鱼眼', '模拟鱼眼镜头畸变'],
  ['sharpen', '锐化', '增强画面边缘细节'],
];

const transitionDefinitions = [
  ['cross-dissolve', '交叉溶解', '前后画面自然叠化'],
  ['fade', '淡入淡出', '通过透明度平滑切换'],
  ['flash-black', '闪黑', '经由短暂黑场切换'],
  ['flash-white', '闪白', '经由短暂白场切换'],
  ['push-left', '向左推', '下一画面从右向左推入'],
  ['push-right', '向右推', '下一画面从左向右推入'],
  ['push-up', '向上推', '下一画面从下向上推入'],
  ['push-down', '向下推', '下一画面从上向下推入'],
  ['wipe-left', '向左擦除', '从右向左擦除当前画面'],
  ['wipe-right', '向右擦除', '从左向右擦除当前画面'],
  ['wipe-up', '向上擦除', '从下向上擦除当前画面'],
  ['wipe-down', '向下擦除', '从上向下擦除当前画面'],
  ['zoom', '缩放', '通过快速缩放连接镜头'],
  ['blur', '模糊', '模糊当前画面后切换'],
  ['glitch', '故障', '以数字故障连接镜头'],
  ['light-leak', '漏光', '通过胶片漏光连接镜头'],
];

const textDefinitions = [
  ['basic-caption', '基础字幕', '简洁易读的底部字幕'],
  ['bold-title', '粗体标题', '适合片头的大号粗体标题'],
  ['outline-caption', '描边字幕', '在复杂背景上保持可读'],
  ['shadow-caption', '阴影字幕', '使用投影增强层次'],
  ['panel-caption', '底板字幕', '带半透明底板的字幕'],
  ['word-highlight', '逐字高亮', '随播放进度逐字强调'],
  ['news-bar', '新闻条', '新闻资讯风格信息条'],
  ['lower-third', '下三分之一', '人物和地点信息条'],
  ['comic', '漫画字', '适合漫剧的醒目文字'],
  ['handwriting', '手写字', '模拟自然手写标题'],
  ['vertical', '竖排字', '从上到下排列的文字'],
  ['gradient-title', '渐变标题', '带渐变填充的大标题'],
];

const filterParameters = () => [intensityParameter()];
const effectParameters = (parameters) => parameters || [intensityParameter(), numberParameter('speed', '速度', 1, 0.1, 5, 0.1)];
const textParameters = () => [
  numberParameter('fontSize', '字号', 48, 12, 240, 1),
  numberParameter('opacity', '不透明度', 1, 0, 1, 0.05),
];

export const editorCatalog = Object.freeze([
  ...filterDefinitions.map(([id, label, description]) => resource('filter', id, label, description, { parameters: filterParameters() })),
  ...effectDefinitions.map(([id, label, description, parameters]) => resource('effect', id, label, description, { parameters: effectParameters(parameters) })),
  ...transitionDefinitions.map(([id, label, description]) => resource('transition', id, label, description, { parameters: [durationParameter()], supportedKinds: ['video', 'image', 'gap'] })),
  ...textDefinitions.map(([id, label, description]) => resource('text', id, label, description, { parameters: textParameters(), supportedKinds: ['text', 'caption'] })),
]);

const byId = new Map(editorCatalog.map((item) => [item.id, item]));

export function catalogById(id) {
  return byId.get(String(id || '')) || null;
}

export function catalogForTab(tab) {
  return editorCatalog.filter((item) => item.category === tab);
}

export function normalizeCatalogParameters(id, values = {}) {
  const item = catalogById(id);
  if (!item) return {};
  return Object.fromEntries(item.parameters.map((parameter) => {
    const rawValue = values?.[parameter.key];
    if (parameter.type !== 'number') return [parameter.key, rawValue ?? parameter.defaultValue];
    const parsed = Number(rawValue);
    const value = Number.isFinite(parsed) ? parsed : parameter.defaultValue;
    return [parameter.key, Math.min(parameter.max, Math.max(parameter.min, value))];
  }));
}
