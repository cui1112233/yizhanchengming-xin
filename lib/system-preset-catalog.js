const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');

const promptsDir = path.join(__dirname, '..', 'prompts');
const SMART_UNIFIED_STYLE_SYSTEM_BODY = fs.readFileSync(path.join(promptsDir, '智能统一视觉分析.md'), 'utf8').trim();
const SMART_UNIFIED_LEGACY_BODY = [
  '【智能统一画面前缀】',
  '',
  '使用当前项目分析得到的统一视觉风格，作为全片所有镜头的画面前缀。',
  '',
  '统一控制：影像媒介、色彩体系、光线质感、明暗关系、对比度、饱和度、颗粒感、镜头质感、氛围和整体美术风格。',
  '',
  '不得改变或新增剧情、人物身份、人物造型、场景内容、道具、动作、台词和事件顺序。',
  '',
  '不要在最终分镜中单独输出“统一风格”标题，只需将该内容作为“画面前缀词”作用于每个镜头。',
  '',
  '如果没有有效的风格分析结果，则根据原文自动判断一种统一视觉风格，并保持全片一致。'
].join('\n');

// The original system "智能统一视觉分析" prompt was published before the
// H3 style.system protocol existed.  Use its exact digest rather than a broad
// text match so a later administrator edit is never silently replaced.
const SMART_UNIFIED_PRE_H3_STYLE_SYSTEM_SHA256 = '0d875416f370e42da9f21e22aa20850973e446292a674943155ad8eabb6da355';
// These are the exact V2/V3 adapted bodies that were already shipped by this
// repository. They can be upgraded to the full supplied source without
// replacing an administrator's later edit.
const SMART_UNIFIED_ADAPTED_V3_SHA256 = '5d28205d6e7d18f387b8bcaf821ae4defed2a90e0ecef5707c6829d7ccaa0cc9';
const H3_ASSET_ADAPTED_V2_SHA256 = 'd8ddaeb802c4755c0de401ee9fca7f0fdd8dec87bbea2393044c5ba7f0b5d4d4';
const H3_DIRECTOR_ADAPTED_V2_SHA256 = 'f6d06c5ec47017b36fdeebbdd685f8084a3cf8f99e5a18aa3d29b1c5402fe046';
const H3_DIRECTOR_ADAPTED_RUNTIME_V2_SHA256 = '9c42d35b88cb4e9987d7bf0475aca704eb2460704a0b5c34eac57e19f8f722d1';
const H3_DIRECTOR_LAYERED_TEMPLATE_V3_SHA256 = '39543b556d360ca3efa94a60b2054ce32c2541af270ea6ed616b37c2db3c6bd0';
const H3_DIRECTOR_WITH_AUDIOVISUAL_POLICY_SHA256 = 'ab1988714f82f6e0345df0c2e1d5126690c31b3f5a91476fe6253b166e30b8bc';
const SCRIPT_EXTRACT_INVALID_JSON_SHA256 = '6487d47e99906292d286e346e249844dd9dd713ffb23d0b7343bbd3787750d51';

function hasBodySha256(body, expected) {
  return crypto.createHash('sha256').update(String(body || '').trim()).digest('hex') === expected;
}

// This is the exact semantic signature of the shipped V3 smart-unified body
// that was published before the H3 JSON protocol.  It is deliberately more
// restrictive than a generic Chinese-label match, so an administrator's later
// custom preset is never silently replaced.
function isShippedChineseLabelSmartUnifiedV3(body) {
  const value = String(body || '').trim();
  return !value.includes('"final_genre"')
    && value.includes('最终将十一项视觉判断融合成一条紧凑、专业、可直接作为后续所有分镜画面提示词共同摄影前缀使用的：')
    && value.includes('`统一风格附加视觉信息`')
    && value.includes('统一风格附加视觉信息：XXXX');
}

function isPreH3SmartUnifiedStyleSystemBody(body) {
  return hasBodySha256(body, SMART_UNIFIED_PRE_H3_STYLE_SYSTEM_SHA256);
}

function isShippedInvalidScriptExtractBody(body) {
  return crypto.createHash('sha256')
    .update(String(body || '').replace(/\r\n/g, '\n'))
    .digest('hex') === SCRIPT_EXTRACT_INVALID_JSON_SHA256;
}

// The former H3 asset preset was the shipped two-request wrapper. Match all
// of its protocol-only signatures so an administrator's edited body is never
// replaced, while the untouched V1 can receive the unified one-request rule.
function isLegacyH3TwoCallAssetBody(body) {
  const value = String(body || '').trim();
  return value.startsWith('# H3 人物场景道具提取')
    && value.includes('本方案独立于普通提取方案')
    && value.includes('## 调用一：H3 人物场景道具事实提取')
    && value.includes('## 调用二：H3 全人物外形编译')
    && value.includes('"character_id":"C001"');
}

// Only this exact shipped placeholder may be replaced during seeding. An
// administrator's edited H3 rule remains their own published preset.
const H3_DIRECTOR_LEGACY_BODY = [
  '## H3 导演分镜 VIDEO 输出规则',
  '',
  '以人物、场景、道具、原文事实和当前镜头连续性为唯一依据。先在内部保持 Scene Memory：场景、空间轴线、人物站位、朝向、视线、手持道具、光线与动作结束状态必须能衔接下一镜。',
  '',
  '每个 VIDEO 必须输出可拍、可见的镜头动作；明确景别、机位、运镜与节奏。人物必须引用已确认的人物资产，禁止临时改写身份、外观或关系。不要用空镜、重复动作或静止画面填充时长。',
  '',
  '最终 VIDEO 由批量工厂 H3 编译器生成：人物定义、视听分层、H3 时间轴和无字幕画面规则均由后端根据结构化分镜输出，浏览器不得拼接或伪造。',
  '',
  '## 最终 Prompt 模板',
  '',
  '管理员可调整本模板的文字、字段顺序与标题；以下占位符由批量工厂以冻结的导演数据和真实时间线确定性填充。',
  '',
  '{{visual_baseline}}',
  '',
  '{{asset_definitions}}',
  '',
  '{{storyboard}}',
  '',
  '{{visual_restriction}}'
].join('\n');
const H3_DIRECTOR_LEGACY_BODY_V0 = [
  '## H3 导演分镜 VIDEO 输出规则',
  '',
  '以人物、场景、道具、原文事实和当前镜头连续性为唯一依据。先在内部保持 Scene Memory：场景、空间轴线、人物站位、朝向、视线、手持道具、光线与动作结束状态必须能衔接下一镜。',
  '',
  '每个 VIDEO 必须输出可拍、可见的镜头动作；明确景别、机位、运镜与节奏。人物必须引用已确认的人物资产，禁止临时改写身份、外观或关系。不要用空镜、重复动作或静止画面填充时长。',
  '',
  '最终 VIDEO 由批量工厂 H3 编译器生成：人物定义、视听分层、H3 时间轴和无字幕画面规则均由后端根据结构化分镜输出，浏览器不得拼接或伪造。'
].join('\n');

const SCRIPT_PRESETS = [
  ['script-extract', '人物场景提取', '人物与场景提取规则', '人物场景提取.md', { format: 'extract' }],
  ['script-extract-assets', '人物场景道具提取', '一次提取人物、场景与关键道具的批量工厂资产规则', '人物场景道具提取.md', { format: 'extract' }],
  ['batch-assets-h3', 'H3 人物场景道具提取', 'H3 一次调用：先提取人物／场景／道具事实与关系，再在同一结果生成全部人物详细外形', '批量工厂-H3人物场景道具提示词.md', { format: 'extract', key: 'h3-assets-full' }],
  ['batch-assets-sd-character-scene', 'SD 人物场景提示词', 'SD V78.3.0.59 人物名单、人物事实、人物关系、人物外形与统一风格一次 AI 调用产出，与现有批量工厂资产库同协议', '批量工厂-SD人物场景提示词.md', { format: 'extract', key: 'sd-character-scene' }],
  ['script-extract-novel-panel', '小说面板提取', 'V77 小说面板人物、关系、场景与统一风格提取规则', '小说面板人物场景提取.md', { format: 'extract' }],
  ['script-hook', '爆款开头', '剧本爆点开头策略', '爆款开头.md'],
  ['script-continuous', '连续开头', '连续剧情开头策略', '连续开头.md'],
  ['script-segmented', '分段开头', 'V77 剧情单元自动分段连续策略', '分段开头.md'],
  ['script-general', '通用规则', '剧本生成通用约束', '通用规则.md'],
  ['script-character-focus', '星标人物聚焦规则', '星标人物作为剧情与镜头聚焦变量的生成规则', '星标人物聚焦规则.md'],
  ['script-audio-match', '匹配音频规则', '普通剧本生成按当前配音实际总时长约束最终分镜', '匹配音频规则.md'],
  ['script-card-protocol', '统一外层分镜卡片协议', '所有剧本输出模式共用的外层分镜卡片边界与时长协议', '统一外层分镜卡片协议.md'],
  ['script-format-screenplay', '剧情模式', '剧情模式输出格式', '剧情模式.md'],
  ['script-format-storyboard', '画布模式', '画布模式输出格式', '画布模式.md'],
  ['script-format-shortdrama', '剧本模式', '短剧模式输出格式', '剧本模式.md'],
  ['script-format-shotlist', '分镜模式', 'V77 小说面板式最终分镜输出格式', '分镜模式.md'],
  ['script-format-q版', 'Q版模式', 'Q版视频镜头输出格式', 'Q版模式.md', { format: 'q版' }],
  ['script-constraint-wrapper', '约束设置规则', '允许格式下的约束输出包裹规则', '约束设置.md'],
  ['script-quick-director-storyboard', '匹配音频', '按当前配音时长和 10s/15s 单元规则生成导演级完整视频分镜', '快速导演分镜.md', { format: 'quick-director' }],
  ['script-quick-director-mode-strict', '导演级完整成品版', '匹配音频的推荐画面描述模式', '快速导演分镜-导演级完整成品版.md', { format: 'quick-director-mode' }],
  ['script-quick-director-mode-concise', '通用小说精简版', '匹配音频的精简画面描述模式', '快速导演分镜-通用小说精简版.md', { format: 'quick-director-mode' }],
  ['script-quick-director-mode-balanced', '通用小说标准版', '匹配音频的标准画面描述模式', '快速导演分镜-通用小说标准版.md', { format: 'quick-director-mode' }],
  ['script-quick-director-mode-detailed', '通用小说较细版', '匹配音频的细节增强画面描述模式', '快速导演分镜-通用小说较细版.md', { format: 'quick-director-mode' }],
  ['script-quick-director-mode-example', '案例学习增强版', '匹配音频的案例学习画面描述模式', '快速导演分镜-案例学习增强版.md', { format: 'quick-director-mode' }],
  ['script-quick-director-mode-reference', '样例对照转换版', '匹配音频的样例对照画面描述模式', '快速导演分镜-样例对照转换版.md', { format: 'quick-director-mode' }],
  ['script-director-storyboard-master', '导演级分镜母版（小说面板同源）', '分镜模式使用的小说面板同源导演规则', '导演级分镜母版.md', { format: 'director-storyboard-master' }]
].map(([id, name, description, source, protocolLock = { format: 'script' }]) => ({ id, module: 'script', name, description, source, protocolLock }));

// 该标记用于把已发布的旧 Q 版预设升级为“只在明确心理内容时出现迷你小人”。
// 已发布的预设通常不会随镜像中的提示词文件自动覆盖，因此采用追加守卫的方式
// 保留用户已有正文，同时保证运行时的硬约束能生效。
const Q_MINI_GUARD_MARKER = 'Q版迷你小人条件守卫 v2（最高优先级）';
const Q_MINI_GUARD = [
  `## ${Q_MINI_GUARD_MARKER}`,
  '先逐段判断镜头是否有原文明确写出的心理活动、内心 OS、内心吐槽、心理独白、脑补画面、弹幕式吐槽、潜台词，或“心里想、暗想、心想、脑海中、内心、幻想/脑补”等非现实内容。',
  '未触发时只按原文现实剧情输出正常比例人物、动作、场景、光影和可确认对白；不得因愤怒、害怕、哭泣、表情或现实对话自行推断心理内容。',
  '未触发时严禁出现 Q 版迷你版主角、迷你小人、缩小版人物、悬浮小人、肩头小人、头顶小人、心理分身或 Q 版画风反差；不得输出【迷你小人细节】或【迷你内心小人】。',
  '只有触发时，才在该心理内容所属的同一镜头增加一个 Q 版迷你版主角；它必须与正常比例现实主角同框，只具象化该段心理内容，不能延伸到前后无关镜头。',
  '每个迷你小人必须能对应原文的明确心理依据；无法指出依据时删除迷你小人，按现实剧情镜头输出。'
].join('\n');

// 已发布格式预设可能仍要求模型无条件输出人物卡或场景卡。该守卫只追加到
// 固定的格式槽位，保留用户已编辑正文，并与服务端/最终卡片层共同保证开关一致。
const BASE_SETUP_GUARD_MARKER = '基础设定开关守卫 v1（最高优先级）';
const BASE_SETUP_GUARD = [
  `## ${BASE_SETUP_GUARD_MARKER}`,
  '仅当本次请求明确提供人物信息和场景信息，且启用了基础设定时，才允许输出独立的【基础设定】或【人物与场景】区块。',
  '未启用基础设定时，不得输出【基础设定】、【人物与场景】、人物卡、场景卡、统一人物、场景环境或任何人物/场景外形档案；只在剧情/镜头正文写原文必要的人名、动作和地点。',
  '不得仅因模板中存在人物或场景示例而编造、补全或单独罗列人物与场景。'
].join('\n');
const BASE_SETUP_GUARD_PRESET_IDS = new Set([
  'script-format-screenplay',
  'script-format-storyboard',
  'script-format-shortdrama',
  'script-format-shotlist',
  'script-format-q版'
]);

// Batch Factory uses the same production semantics as the script workbench,
// but its Director JSON contract is different.  These headings delimit the
// adapted rules in the batch prompt sources so existing personal-center
// presets can receive the rules without replacing an administrator's body.
const BATCH_SCRIPT_ALIGNMENT_MARKERS = Object.freeze({
  'batch-character-meta': '## 剧本生成同源人物规则 v1',
  'batch-scene-meta': '## 剧本生成同源场景规则 v1',
  'batch-video-meta': '## 剧本生成同源 VIDEO 规则 v1',
  'batch-visual-meta': '## 剧本生成同源画面规则 v1'
});

function batchScriptAlignmentGuard(preset) {
  const marker = BATCH_SCRIPT_ALIGNMENT_MARKERS[preset?.id];
  if (!marker) return '';
  const body = String(preset?.body || '');
  const index = body.indexOf(marker);
  return index < 0 ? '' : body.slice(index).trim();
}

const SCRIPT_CONSTRAINT_PRESETS = [
  ['script-constraint-prefix-smart-unified', '智能统一', 'H3 style.system：按视频原文分析内容类型、时代与统一视觉风格', 'prefix', SMART_UNIFIED_STYLE_SYSTEM_BODY],
  ['script-constraint-prefix-live-action', '真人实拍', '真人实拍画面前缀词', 'prefix', '影视级真人实拍短剧，超高清数字电影实拍质感，微弱细腻数字颗粒纹理；色调、对比度、饱和度与光影按当前小说的时代、题材、情绪及场景氛围统一；自然叙事镜头体系，稳定空间关系与克制透视，微弱镜头光晕，动机光优先，保留柔和高光、可读暗部及自然明暗层次；强调人物关系、空间纵深与视线组织的电影级叙事构图，非动画、非卡通、非数字替身。'],
  ['script-constraint-prefix-3d', '3D 国漫', '3D 国漫画面前缀词', 'prefix', '高精度次世代写实国风 3D 角色设计，影视级东方 CG 人物，UE5 虚幻引擎 AAA 级游戏角色渲染，半写实国漫美术风格，精细 3D 建模，非真人摄影，非真人数字替身，非二维动漫，非卡通 Q 版。'],
  ['script-constraint-prefix-2d', '2D 动漫', '2D 动漫画面前缀词', 'prefix', '高质量二维动画，清晰赛璐璐上色，细腻线稿，非真人摄影，非三维建模。'],
  ['script-constraint-prefix-guoman', '国漫', '国漫画面前缀词', 'prefix', '高品质国漫美术风格，东方审美角色设计，细腻线条与电影级光影，非真人摄影。'],
  ['script-constraint-quality-4k', '4K 画质', '4K 超高清画质约束', 'quality', '4K 超高清，浅景深，电影级细节与光影层次；全程不生成内嵌字幕、标题、说明文字、水印或 Logo。'],
  ['script-constraint-quality-sd-detail', 'SD 4K级细节画质', 'SD 最终导出画质约束：4K级细节，人物面部眼睛手部动作清晰可读', 'quality', '以4K级细节为画质目标，实际分辨率与帧率按生成配置设置。主体焦点稳定，景深符合拍摄任务，人物保留真实皮肤、发丝和衣物纹理，关系镜头保留理解站位所需的环境清晰度。环境允许阴暗、夜间或低调，但人物面部、眼睛、手部动作、服装轮廓和关键道具必须保持清晰可读的中间调；使用符合场景来源的主光、柔和补光或轮廓分离光保护主体，并与背景形成明确层次。运动、材质反射与光影变化自然连贯，高光不过曝，黑位保留纹理。无暗角、无人工边缘压黑，不以过度磨皮或锐化代替真实细节。'],
  ['script-constraint-quality-h3-shot-narrative', 'H3 镜头叙事约束', 'H3 场景顺序、镜头边界与可见表演约束', 'quality', '【H3镜头叙事约束】\n严格依次呈现列出的场景与镜头；完成当前场景及其可见结果后，才进入下一场景。\n不得预演、挪用或提前展示后续场景的动作、地点、道具状态、人物揭示或剧情结果。进入后续剧情节点后，不得返回补拍未完成的前一节点。\n场景或镜头边界默认使用干净硬切；只有当前镜头明确是同一物理空间内的连续动作时，才允许连续衔接。\n人物情绪和关系主要通过视线、手势、肢体动作、调度、道具以及环境的可见结果表达。\n画面中人物保持自然闭口状态；只有当前镜头明确要求可见对白口型时，才允许张口说话。'],
  ['script-constraint-restriction-h3-visual-policy', 'H3 画面限制', 'H3 跨镜人物、材质、运动与无文字画面稳定限制', 'restriction', '【H3画面限制】\n人物身份、年龄、脸型、五官、发型、服装和饰品在连续镜头中保持稳定。运动过程中保持五官结构稳定，避免面部融化、身份漂移、重复人物、多余肢体、肢体融合、手指异常、穿模和明显透视错误。\n皮肤保留真实毛孔、细小绒毛和自然肤色变化；眼睛、睫毛、眉毛和发丝保持自然细节，不过度锐化或塑料化磨皮。布料保持真实纤维与褶皱，木材、玻璃、金属、屏幕等保持符合材质属性的反射与粗糙度。\n人物运动和摄影机运动保持清晰稳定，避免不合理运动模糊、局部拖影和背景结构跳变。高光不过曝，黑位不压死，暗部保留有效纹理。保持干净完整的影视画面，不添加水印、品牌标识或与剧情无关的界面元素。'],
  ['script-constraint-restriction-no-overlay', '无关文字限制', '禁止无关文字、横幅与漂浮 UI', 'restriction', '禁止出现任何无关文字、横幅、漂浮 UI 元素、角标、二维码、字幕、水印和 Logo；剧情明确需要的信息载体仅保留必要且最少的可读信息。'],
  ['script-constraint-negative-general', '通用负面提示词', '通用视频负面提示词', 'negative', '禁止字幕、台词字幕、旁白字幕、对白字卡、气泡框文字、水印、Logo及非剧情需要的可读文字；仅当剧情必须呈现手机聊天、系统界面、合同、名单或监控文件时，允许保留必要文字。'],
  ['script-constraint-negative-sd-detail', 'SD 最终导出负面提示词', 'SD 完整负面提示词：失焦、变形、穿模、水印等全部禁止项', 'negative', '主体意外失焦，过度运动模糊，面部变形，身份漂移，人物外形串线，同脸化，人物无故复制，无关人物，多余肢体，手指异常，肢体融合，穿模，服装跳变，道具漂移，空间结构跳变，错误透视，人物瞬移，视线错位，提前出场，动作无故重复，动作状态复位，光线无故闪变，过曝高光，面部欠曝，五官消失，眼睛无细节，手部动作不可辨，主体与背景粘连，纯黑剪影，黑位吞没人物，彩色光污染肤色，塑料皮肤，过度磨皮，过度锐化，暗角，可读字符载体，大面积字符层，可辨识界面文字，水印；人物保持自然闭口。']
].map(([id, name, description, category, body]) => ({
  id,
  module: 'script',
  name,
  kind: 'addon',
  description,
  body,
  compatibleBaseIds: [],
  protocolLock: { format: 'constraint', category }
}));

const NOVEL_PRESETS = [
  {
    id: 'novel-analysis',
    module: 'novel-panel',
    name: '内容分析',
    description: '小说题材、人物与场景分析规则',
    body: [
      '你是小说短剧制作分析服务。只能依据用户提供的小说原文，不能编造人物、关系、关键因果或结局。',
      '只返回合法 JSON，不要 Markdown，不要代码围栏。',
      'JSON 顶层必须包含 genre、trailer_style、camera、characters、scene_options。',
      'characters 是数组；每项至少包含 name、gender、role、appearance、aliases、age_stages。',
      'scene_options 是数组；每项仅描述原文可见场景。appearance 必须是可拍摄的外形描述。'
    ].join('\n'),
    protocolLock: { format: 'json', operation: 'analysis' }
  },
  {
    id: 'novel-character',
    module: 'novel-panel',
    name: '人物卡',
    description: '人物外形卡生成规则',
    body: [
      '你是影视人物外形设计服务。只能保持输入的人物姓名、性别、身份、别名和年龄阶段，不得增删、合并或改名。',
      '只返回合法 JSON，不要 Markdown，不要代码围栏。'
    ].join('\n'),
    protocolLock: { format: 'json', operation: 'character' }
  },
  {
    id: 'novel-outline',
    module: 'novel-panel',
    name: '分镜生成',
    description: '小说分镜和时间轴生成规则',
    body: [
      '你是小说短剧分镜生成服务。只依据用户输入的小说原文和附加规则，不得新增关键剧情、人物、因果或结果。',
      '只返回合法 JSON，不要 Markdown，不要代码围栏。',
      '返回对象必须包含 scenes 和 outline_shots 数组。每个 outline_shot 必须包含 source_index、source_basis、prompt；prompt 必须是具体、可拍摄的中文画面描述。'
    ].join('\n'),
    protocolLock: { format: 'json', operation: 'outline', qualityGate: true }
  }
];

const NOVEL_FETCH_PRESETS = [
  {
    id: 'novel-fetch-induce',
    module: 'novel-fetch',
    name: '诱导排查',
    description: '小说合规检测与柔化优化规则',
    body: [
      '你是网文内容合规审核师与文本优化师。只能依据用户提供的小说原文工作，不得新增、删除或改变人物关系、核心动机、剧情走向和结局。',
      '遵循最小改动原则：优先替换敏感词；仅在必要时将低俗、露骨、血腥或违法内容改为克制的可读表达，并确保前后文衔接自然。',
      '先输出“### 一、分析结果”，下一行只输出 JSON：{"gender":"男或女","style":"一个适合的风格"}。风格只能选：古风虐文、古风甜文、古风通用、年代虐文、年代甜文、年代通用、现代虐文、现代甜文、现代悬疑、现代通用、男频都市、现代女主、玄幻、历史、爆款BGM、家庭奇葩、家庭伤感、职场打脸。',
      '再依次输出“### 二、合规检测报告”“### 三、优化后全文”，必要时输出“### 四、关键修改说明”。优化后全文必须完整、可直接使用。',
      '不要输出代码围栏或额外说明。'
    ].join('\n'),
    protocolLock: { format: 'novel-fetch-process', operation: 'induce' }
  },
  {
    id: 'novel-fetch-hook',
    module: 'novel-fetch',
    name: '爆款优化',
    description: '小说爆款化文本优化规则',
    body: [
      '你是短剧和网文的爆款策划与文本优化师。只能依据用户提供的小说原文工作，不得改变人物关系、核心动机、关键因果、剧情走向或结局。',
      '在保留原文叙事风格的前提下，前置强冲突、身份反差或悬念，压缩无效铺垫，强化可见动作、人物台词和段末卡点；不得凭空添加关键角色、道具或结果。',
      '先输出“### 一、分析结果”，下一行只输出 JSON：{"gender":"男或女","style":"一个适合的风格"}。风格只能选：古风虐文、古风甜文、古风通用、年代虐文、年代甜文、年代通用、现代虐文、现代甜文、现代悬疑、现代通用、男频都市、现代女主、玄幻、历史、爆款BGM、家庭奇葩、家庭伤感、职场打脸。',
      '再依次输出“### 二、优化说明”“### 三、优化后全文”。优化后全文必须完整、可直接用于阅读和后续制作。',
      '不要输出代码围栏或额外说明。'
    ].join('\n'),
    protocolLock: { format: 'novel-fetch-process', operation: 'hook' }
  }
];

const SHUIHUO_PRESETS = [
  {
    id: 'shuihuo-extract-assets',
    module: 'shuihuo-production',
    name: '人物场景、道具提取',
    description: '从项目原文一次性提取人物、场景和关键道具资产',
    body: [
      '你是专业的小说内容分析师，负责从小说原文中一次性提取人物、场景、关键道具，并生成可直接用于 AI 图像/视频生产的视觉描述。',
      '先内部判断题材类型、时代背景、世界观特征和整体视觉风格，但不要输出分析过程。',
      '人物提取：提取所有需要保持一致的出场人物；同一人物在不同阶段、变装、重生前后外观差异显著时拆成“人物名-阶段标签”。外观描述必须是一段式，按“全景，正面拍摄，性别，年龄段，种族/人种，发型发色，面部特征，服装穿着，配饰道具，整体气质。姿态：默认姿态。”组织。',
      '场景提取：提取所有出现的地点；同一地点日夜、破败/繁华等状态差异显著时拆成“地点名-状态”。场景描述必须是一段式，按“空间类型，时间光线，整体环境，建筑/空间结构，地面材质，主要陈设，环境细节，氛围基调，特殊元素”组织，严禁包含人物信息。',
      '道具提取：只提取对剧情推进、身份象征、情感象征、证据、法器、信物等有明确意义的关键道具；不要提取普通桌椅、茶杯、手机、服装和普通陈设。道具描述必须是一段式，按“类别，形态，材质颜色，尺寸，纹样/铭文，特殊视觉效果，质感，陈列状态，整体气质”组织，严禁包含人物信息。',
      '只依据原文，不得编造人物、关系、地点、道具、时代或剧情事件。同名资产不得重复。',
      '只返回合法 JSON 数组，不要 Markdown、说明文字或代码围栏。',
      '数组每项必须是 {"category":"character|scene|prop","name":"资产名称","prompt":"可直接用于画面生成的具体视觉描述"}。',
      '如果原文中没有符合条件的道具，可以不返回 prop 项。',
      '',
      '原文：{{novel_text}}'
    ].join('\n'),
    protocolLock: { format: 'json', operation: 'asset-extraction' }
  },
  {
    id: 'shuihuo-extract-characters',
    module: 'shuihuo-production',
    name: '提取人物',
    description: '从项目原文提取可复用的人物资产与视觉基线',
    body: [
      '你是小说视频生产的人物资产提取服务。只能依据输入原文，不得编造人物、身份、关系、年龄或外形。',
      '只提取原文有明确依据、后续画面需要保持一致的人物；同一人物不得重复。',
      '只返回合法 JSON 数组，不要 Markdown、说明文字或代码围栏。',
      '数组每项必须是 {"category":"character","name":"人物名称","prompt":"可用于画面生成的人物外形与服饰描述"}。',
      '',
      '原文：{{novel_text}}'
    ].join('\n'),
    protocolLock: { format: 'json', operation: 'character-extraction' }
  },
  {
    id: 'shuihuo-extract-scenes',
    module: 'shuihuo-production',
    name: '提取场景',
    description: '从项目原文提取可复用的场景资产与空间基线',
    body: [
      '你是小说视频生产的场景资产提取服务。只能依据输入原文，不得新增地点、时代、天气、陈设或剧情事件。',
      '只提取原文有明确依据、后续画面需要保持一致的场景；同一场景不得重复。',
      '只返回合法 JSON 数组，不要 Markdown、说明文字或代码围栏。',
      '数组每项必须是 {"category":"scene","name":"场景名称","prompt":"可用于画面生成的空间、时间、光线与陈设描述"}。',
      '',
      '原文：{{novel_text}}'
    ].join('\n'),
    protocolLock: { format: 'json', operation: 'scene-extraction' }
  },
  {
    id: 'shuihuo-extract-props',
    module: 'shuihuo-production',
    name: '提取道具',
    description: '从项目原文提取可复用的关键道具资产',
    body: [
      '你是小说视频生产的道具资产提取服务。只能依据输入原文，不得新增道具、用途、材质、时代或剧情事件。',
      '只提取原文有明确依据、后续画面需要保持一致的关键道具；同一道具不得重复。',
      '只返回合法 JSON 数组，不要 Markdown、说明文字或代码围栏。',
      '数组每项必须是 {"category":"prop","name":"道具名称","prompt":"可用于画面生成的材质、形态、颜色、状态与细节描述"}。',
      '',
      '原文：{{novel_text}}'
    ].join('\n'),
    protocolLock: { format: 'json', operation: 'prop-extraction' }
  },
  {
    id: 'shuihuo-asset-binding',
    module: 'shuihuo-production',
    name: '分镜资产绑定',
    description: '将项目人物、场景和道具绑定到对应字幕分镜',
    body: [
      '你是小说视频生产的分镜资产绑定服务。只能依据输入的分镜字幕和项目资产名称、描述做绑定，不得创建新资产、改写字幕或猜测未出现的资产。',
      '每个分镜只能引用本次 assets 数组中存在的 key；角色、场景、道具都可以为空数组。',
      '只返回合法 JSON 数组，不要 Markdown、说明文字或代码围栏。',
      '数组每项必须是 {"segmentId":0,"sceneMode":"start|continue|switch","assetKeys":[]}。',
      '第一段 sceneMode 为 start；没有明确地点变化时为 continue，并沿用上一段 scene 资产 key；只有原文明确切换地点时才为 switch。',
      '',
      '分镜：{{segment_text}}',
      '项目资产：{{project_note}}'
    ].join('\n'),
    protocolLock: { format: 'json', operation: 'asset-binding' }
  },
  {
    id: 'shuihuo-character-color-sheet',
    module: 'shuihuo-production',
    name: '配色图人物设计',
    description: '带三视图和配色板的角色设计参考图',
    body: [
      '你是一位专业的角色设计师，请根据以下角色信息，生成一张完整的角色设计参考图(Character Design Sheet)。',
      '【布局要求】16:9横版，纯白色背景(#FFFFFF)，带中文标注，所有文字使用宋体字体。',
      '- 左上角：显示“我叫{character_name}”大标题，下方标注身份信息。',
      '- 左侧区域：标注“面部特写”，展示角色正面视角的面部特写（与三视图正面保持一致的角度）。',
      '- 左下角：标注“配色板”，展示主要颜色色块（发色、肤色、服装颜色等，带十六进制色值）。',
      '- 中央上方：三视图区域，标注“正面”、“侧面”、“背面”，展示角色全身三个视角。',
      '- 右上角：角色自我介绍文字区域，显示角色台词内容。',
      '- 底部中央：标注“局部细节区”，根据角色描述中的特征，展示3-5个该角色最具代表性的细节特写。',
      '- 右侧：标注“全身照比例照”，展示角色全身立绘配合身高标尺。'
    ].join('\n'),
    protocolLock: { format: 'image-prompt', operation: 'character-sheet' }
  },
  {
    id: 'shuihuo-character-accessory-sheet',
    module: 'shuihuo-production',
    name: '三视图配饰设计',
    description: '带五官与服饰细节的角色三视图',
    body: '纯白色背景，不要带文字，根据角色描述词生成标准概念图(正面、侧面、背面)。右侧细节特写区域包括：面部特写（正面、3/4侧面）、双眼特写、双手平铺展开特写（五指张开、手掌正面）、服饰配件特写（衣领/领带/袖口/衣服下摆/鞋子等）。确保与主图的风格、色彩、特征完全一致。布局格式：三视图在画面左侧，细节特写排列在画面右侧，对半排版，整体比例 16:9。',
    protocolLock: { format: 'image-prompt', operation: 'character-sheet' }
  },
  {
    id: 'shuihuo-character-three-view',
    module: 'shuihuo-production',
    name: '三视图',
    description: '标准全身角色三视图',
    body: '左侧 1/3 区域为超大面部特写，右侧 2/3 区域并排展示正视图、右视图、后视图，整体比例 16:9，构图严谨规整，全身标准三视图，自然标准站立，人物比例精准合理，头身比标准，结构严谨，体态真实自然，真实的皮肤质感、画面干净，纯白色背景，图片不要文字。',
    protocolLock: { format: 'image-prompt', operation: 'character-sheet' }
  },
  {
    id: 'shuihuo-character-expression-sheet',
    module: 'shuihuo-production',
    name: '角色表情多状态',
    description: '角色细节、表情和动作全方位设定图',
    body: [
      '请根据角色设定生成一张角色全方位展示设定图。',
      '整体要求：16:9横版构图，纯白色背景，不要文字、不要水印、不要边框、不要配色板。整张图按三排排版，画面中只保留同一个角色，所有格子里的五官、发型、服饰、配色、材质和关键道具必须完全统一。',
      '第一排放4个细节特写，重点展示发型或发饰、胸肩服饰纹理、腰部或配饰、手部与关键道具；第二排放4个脸部表情特写，展示温和、专注、沉思、冷静等不同日常神态；第三排放4个全身动作展示，表现站姿、行走、指向、坐姿、舞动等自然姿势。',
      '画面简洁，人物清晰，细节准确，超高清，完美构图，画面要求白底、光影清晰、材质细腻、细节丰富。'
    ].join('\n'),
    protocolLock: { format: 'image-prompt', operation: 'character-sheet' }
  },
  {
    id: 'shuihuo-character-single-view',
    module: 'shuihuo-production',
    name: '单视图',
    description: '标准竖版角色单视图',
    body: '9:16竖版构图，纯白色背景，不要文字、不要水印、不要边框、不要配色板。自然标准站立，人物比例精准合理，头身比标准，结构严谨，体态真实自然，真实的皮肤质感、画面干净，纯白色背景，图片不要文字。',
    protocolLock: { format: 'image-prompt', operation: 'character-sheet' }
  },
  {
    id: 'shuihuo-smart-segmentation',
    module: 'shuihuo-production',
    name: '智能识别',
    description: '水货生产导入原文后的智能分段规则',
    body: [
      '你是小说视频生产的智能分段服务。只能依据输入原文，不得编造人物、情节、因果或结局。',
      '按剧情动作、场景变化、人物关系推进、说话人变化和叙事节奏划分连续原文段落。',
      '每个分段必须同时判断发言者。叙述、环境、动作、无法确认说话者的内容填写“旁白”；直接台词必须填写实际说话角色名。',
      '同一角色的连续台词，即使后一句没有重复署名，也沿用最近一次有明确依据的说话角色；第一人称台词在原文能确认是“我”说出时填写“我”。',
      '一段中混有叙述与台词、或混有不同角色台词时，必须按说话人边界拆成多个连续分段；不要把旁白和角色台词放在同一个分段。',
      '不能仅因出现引号就猜测角色；只有结合原文上下文仍无法确定时，才填写“旁白”。',
      '只返回合法 JSON 数组，不要 Markdown、说明文字或代码围栏。',
      '数组每项必须是 {"text":"原文分段","speaker":"旁白或原文角色名"}；每一项 text 必须保留原文内容，不得改写或补写，speaker 不得为空。',
      '',
      '原文：{{novel_text}}'
    ].join('\n'),
    protocolLock: { format: 'json', operation: 'segmentation' }
  },
  {
    id: 'shuihuo-image-prompt',
    module: 'shuihuo-production',
    name: '画面提示词',
    description: '将已确认分镜转为可直接提交图片模型的画面提示词',
    body: [
      '你是专业的 AI 绘画提示词生成器，负责根据当前分镜原文和项目资产，生成即梦等文生图工具可直接使用的图片提示词。',
      '核心原则：只写具象画面，不写抽象概念；只依据当前分镜和项目资产，不补写原文没有的关键人物、事件或结果。',
      '画面提示词必须以“图片设计：”开头，包含景别、角色参考说明、角色位置、动作、表情、背景场景名称、时间/光线。',
      '角色必须使用项目资产名称；多角色时写清“参考图一是A，参考图二是B”，并用“图一中A / 图二中B”描述位置和互动。禁止用“男人、女人、男孩、女孩”等泛称替代角色名。',
      'scene_name 必须从项目资产中的场景名称选择；image_prompt 背景必须包含场景名，例如“背景王府大殿”，禁止只写“背景虚化、背景模糊”。',
      '景别优先使用近景、中景、远景、上半身；面部特写极少使用，仅用于重大转折。相邻分镜避免连续同一景别。',
      '前景/中景/背景只描写环境或物品，不要在前景/中景/背景里写人物。多人互动必须写清谁对谁做什么。',
      '禁止使用血液、染血、乱码、伪文字、无意义符号、压迫感、电影感等容易导致跑偏或抽象的词；情绪用可见表情和动作表达。',
      '只返回合法 JSON，不要 Markdown、说明文字或代码围栏。',
      '优先返回 {"storyboard":[{"image_prompt":"完整中文画面提示词"}]}；也可以返回 [{"segmentId":0,"prompt":"完整中文画面提示词"}]。',
      '',
      '当前分镜：{{segment_text}}',
      '项目资产：{{project_note}}'
    ].join('\n'),
    protocolLock: { format: 'json', operation: 'image-prompt' }
  },
  {
    id: 'shuihuo-video-prompt',
    module: 'shuihuo-production',
    name: '视频提示词',
    description: '将已确认分镜转为可直接提交视频模型的动态镜头提示词',
    body: [
      '你是高分镜连续情绪场景推理执行器，负责把当前分镜原文转成可直接用于 Vidu / Seedance / 即梦视频模型的动态镜头提示词。',
      '只依据当前分镜和项目资产，不遗漏、不改写原文台词或旁白，不补写原文没有的关键剧情、人物、因果或结果。',
      'video_desc 必须独立完整，不能写“同上、延续上段、保持不变”。每条都要写清角色标签、场景人物、地点、空间布局、画面描述、配音脚本、原文摘要。',
      '角色和场景必须使用项目资产名称；同一场景内保持人物位置、服饰、道具、光影和空间结构连续。若人物移动，必须在动作中写明移动过程。',
      '镜头语言要具体：近景/中景/远景/上半身/特写，缓慢推近、横移跟随、轻微跟随、拉远、环绕等必须写清推向哪里、跟随谁、展现什么。',
      '有台词时写嘴型动作，voice_actor 和 line 必须保留原文；旁白/内心独白用动作、光影、环境状态表现。纯画面时 line 为空或“无”。',
      '配音脚本按原文顺序排列，格式可用 [旁白]、[角色名]、[角色名-内心]；内容必须与原文一致，不得改写。',
      '禁止使用与模型不稳定相关的词：血液、染血、泪流满面、背景虚化、压迫感、电影感、同上、保持不变。',
      '只返回合法 JSON，不要 Markdown、说明文字或代码围栏。',
      '优先返回 {"storyboard":[{"characters":"出场角色","scene":"场景名称","voice_actor":"旁白或角色名","emotion":"情绪","line":"原文台词/旁白/无","video_desc":"完整视频提示词"}]}；也可以返回 [{"segmentId":0,"prompt":"完整中文视频提示词"}]。',
      '',
      '当前分镜：{{segment_text}}',
      '项目资产：{{project_note}}'
    ].join('\n'),
    protocolLock: { format: 'json', operation: 'video-prompt' }
  },
  {
    id: 'shuihuo-negative-prompt',
    module: 'shuihuo-production',
    name: '负面提示词',
    description: '将已确认分镜转为图片或视频生成时需要避免的内容约束',
    body: [
      '你是小说视频生产的负面提示词服务。根据当前分镜和已绑定项目资产，输出图片/视频生成时需要避免的内容约束。',
      '重点避免：角色外观不一致、服装突变、场景名称缺失、背景虚化导致白底、前景/中景出现错误人物、无关文字、乱码、伪文字、随机字母、水印、Logo、字幕、二维码、画面接缝、脏乱纹理、噪点、颗粒感、过度锐化、畸形肢体、错误手指、人物数量错误、性别年龄错误、道具消失或凭空出现。',
      '如画面需要文字，要求文字清晰准确可读，中文笔画完整，字距正常，排版规整，不生成错误文字或无意义符号。',
      '不得凭空增加剧情或人物，不要重复正向提示词。',
      '只返回合法 JSON 数组，不要 Markdown、说明文字或代码围栏。',
      '数组每项必须是 {"segmentId":0,"prompt":"完整中文负面提示词"}；segmentId 可留为 0。',
      '',
      '当前分镜：{{segment_text}}',
      '已绑定资产：{{project_note}}'
    ].join('\n'),
    protocolLock: { format: 'json', operation: 'negative-prompt' }
  }
];

const BATCH_PRESETS = [
  ['batch-hook-adaptation', '爆款开头改编元提示词', '批量工厂爆款模式的开头改编规则', '批量工厂-爆款开头改编.md', { format: 'json', operation: 'hook-adaptation' }],
  ['batch-original-director', '原文直转导演元提示词', '批量工厂原文模式的导演拆分规则', '批量工厂-原文直转导演.md', { format: 'json', operation: 'original-director' }],
  ['batch-viral-director', '爆款开头导演元提示词', '批量工厂爆款开头审核后的导演规则', '批量工厂-爆款开头导演.md', { format: 'json', operation: 'viral-director' }],
  ['batch-character-meta', '人物提示词元提示词', '批量工厂人物一致性与人物提示词规则', '批量工厂-人物提示词.md', { format: 'json', operation: 'character-meta' }],
	['batch-character-h3', 'H3 人物提示词', '一次请求返回全部人物的完整 H3 外形，保存到单书资产库', '批量工厂-H3人物提示词.md', { format: 'json', operation: 'character-meta', key: 'h3-character-normal' }],
  ['batch-assets-h3-skill', 'H3 人物资产 Skill', 'H3 V78.3.0.59 人物资产获取 Skill：名单、事实、关系、外形四步产出人物资产，复用现有智能统一结果；分镜规划归现有导演分镜提示词', '批量工厂-H3人物资产Skill-V78.3.0.59.md', { format: 'skill', operation: 'h3-assets-with-smart-unified', version: 'v78.3.0.59', smartUnifiedPresetId: 'script-constraint-prefix-smart-unified', steps: ['character.roster.system__v78.3.0.59', 'character.facts.system__v78.3.0.59', 'character.relationships.system__v78.3.0.59', 'character.appearance.system__v78.3.0.59'] }],
  ['batch-scene-meta', '场景提示词元提示词', '批量工厂场景一致性与场景提示词规则', '批量工厂-场景提示词.md', { format: 'json', operation: 'scene-meta' }],
	['batch-scene-h3', 'H3 场景提示词', 'H3 Scene Memory 驱动的逐场景一致性提示词编译', '批量工厂-H3场景提示词.md', { format: 'json', operation: 'scene-meta', key: 'h3-scene-normal' }],
  ['batch-prop-meta', '道具提示词元提示词', '批量工厂关键道具与道具提示词规则', '批量工厂-道具提示词.md', { format: 'json', operation: 'prop-meta' }],
  ['batch-video-meta', '分镜元提示词（自动组合）', '自动组合剧本生成的分段开头与分镜模式；批量工厂只保留 JSON 分镜输出协议', '批量工厂-分镜元提示词.md', { format: 'json', operation: 'video-meta' }],
  ['batch-video-h3-director', 'H3 导演分镜视频提示词', '使用 H3 Scene Memory、逐镜选角和视听时间轴编译最终 VIDEO 提示词', '批量工厂-H3导演分镜视频提示词.md', { format: 'json', operation: 'video-meta', key: 'h3-video-normal' }],
  ['batch-visual-meta', '画面提示词元提示词', '批量工厂画面图片提示词规则', '批量工厂-画面提示词.md', { format: 'json', operation: 'visual-meta' }],
  ['batch-video-sd-fixed-shot', 'SD 固定镜头提示词', '系统锁定镜头数量与时长，AI 只填画面内容', '批量工厂-SD固定镜头提示词.md', { format: 'json', operation: 'video-meta' }],
  ['batch-video-sd', 'SD 视频提示词', 'SD V78.3.0.59 故事剧本、固定镜头导演、最终 Prompt 模板与分镜规划跟随配音规则一次 AI 调用产出，选择即自动应用全套规则', '批量工厂-SD视频提示词.md', { format: 'json', operation: 'video-meta' }],
  ['batch-opening-meta', '换开头元提示词', '批量工厂换开头：为分镜一生成画面感强且能衔接后续分镜的开场变体', '批量工厂-换开头元提示词.md', { format: 'opening-meta' }]
].map(([id, name, description, source, protocolLock]) => ({ id, module: 'batch-factory', name, description, source, protocolLock }));

const BATCH_PREFIX_PRESETS = [
  ['batch-prefix-general-anime', '通用动漫', '无法明确分类时使用的通用动漫视频前缀', 'general_anime', '高质量动漫短视频，人物比例自然，角色一致性稳定，画面信息清晰，动作与表情可读，电影级构图与光影，避免无意义空镜。'],
  ['batch-prefix-modern-conflict', '现代都市强冲突', '现代都市、豪门、复仇、家庭冲突等强情绪视频', 'modern_conflict', '高情绪现代都市动漫短剧，冲突前置，人物强表演与明显动作反馈，细腻微表情，现实空间质感，电影级构图，高密度剧情推进。'],
  ['batch-prefix-ancient-drama', '古风剧情', '古言、宅斗、权谋、宫廷等古风剧情视频', 'ancient_drama', '高品质国风动漫，东方古典人物设计，服饰与礼制细节准确，宫廷或古宅空间层次清晰，克制而强烈的人物表演，电影级古典光影。'],
  ['batch-prefix-xuanhuan-action', '玄幻战斗', '玄幻、修仙、奇幻战斗与高能动作视频', 'xuanhuan_action', '高品质玄幻国漫，强动作节奏，能量与法术视觉层级清晰，关键动作分解明确，高冲击接触反馈，宏大空间尺度与电影级动态运镜。'],
  ['batch-prefix-suspense', '悬疑惊悚', '悬疑、惊悚、危险发现与反转视频', 'suspense', '悬疑动漫短剧，压迫氛围，信息逐层揭示，克制冷色光影，局部高反差，人物紧张反应真实，关键物件和视线线索清晰，避免无意义惊吓。'],
  ['batch-prefix-era-drama', '年代剧情', '年代文、乡村、知青、家庭生活等视频', 'era_drama', '高质量年代感动漫，时代服装、建筑、家具与生活物件准确，做旧但干净的影像质感，自然人物比例，真实生活动作，情绪表达直接清晰。']
].map(([id, name, description, key, body]) => ({ id, module: 'batch-factory', name, kind: 'base', description, body, compatibleBaseIds: [], protocolLock: { format: 'video-prefix', key } }));

const SYSTEM_PRESET_SLOTS = Object.freeze([
  ['script.extract', 'script', '人物场景提取', 'primary'],
  ['script.asset-extraction', 'script', '人物场景道具提取', 'primary'],
  ['script.novel-panel-extract', 'script', '小说面板提取', 'primary'],
  ['script.hook', 'script', '爆款开头', 'primary'],
  ['script.continuous', 'script', '连续开头', 'primary'],
  ['script.segmented', 'script', '分段开头', 'primary'],
  ['script.general', 'script', '通用规则', 'primary'],
  ['script.character-focus', 'script', '星标人物聚焦规则', 'primary'],
  ['script.audio-match', 'script', '匹配音频规则', 'primary'],
  ['script.card.protocol', 'script', '统一外层分镜卡片协议', 'primary'],
  ['script.format.screenplay', 'script', '剧情模式', 'primary'],
  ['script.format.storyboard', 'script', '画布模式', 'primary'],
  ['script.format.shortdrama', 'script', '剧本模式', 'primary'],
  ['script.format.shotlist', 'script', '分镜模式', 'primary'],
  ['script.format.q版', 'script', 'Q版模式', 'primary'],
  ['script.constraint.wrapper', 'script', '约束设置规则', 'primary'],
  ['script.constraint.prefix', 'script', '画面前缀词', 'addon'],
  ['script.constraint.quality', 'script', '画质约束', 'addon'],
  ['script.constraint.restriction', 'script', '画面限制', 'addon'],
  ['script.constraint.negative', 'script', '负面提示词', 'addon'],
  ['script.quick-director', 'script', '匹配音频', 'primary'],
  ['script.director.master', 'script', '导演级分镜母版', 'primary'],
  ['novel.analysis', 'novel-panel', '内容分析', 'primary'],
  ['novel.character', 'novel-panel', '人物卡', 'primary'],
  ['novel.outline', 'novel-panel', '分镜生成', 'primary'],
  ['novel-fetch.induce', 'novel-fetch', '诱导排查', 'primary'],
  ['novel-fetch.hook', 'novel-fetch', '爆款优化', 'primary'],
  ['shuihuo.asset.extraction', 'shuihuo-production', '人物场景、道具提取', 'primary'],
  ['shuihuo.asset.character-extraction', 'shuihuo-production', '提取人物', 'primary'],
  ['shuihuo.asset.scene-extraction', 'shuihuo-production', '提取场景', 'primary'],
  ['shuihuo.asset.prop-extraction', 'shuihuo-production', '提取道具', 'primary'],
  ['shuihuo.asset.binding', 'shuihuo-production', '分镜资产绑定', 'primary'],
  ['shuihuo.asset.character-sheet', 'shuihuo-production', '人物设定', 'primary'],
  ['shuihuo.segmentation.smart', 'shuihuo-production', '智能识别', 'primary'],
  ['shuihuo.prompt.image', 'shuihuo-production', '画面提示词', 'primary'],
  ['shuihuo.prompt.video', 'shuihuo-production', '视频提示词', 'primary'],
  ['shuihuo.prompt.negative', 'shuihuo-production', '负面提示词', 'primary'],
  ['shuihuo.voice.design', 'shuihuo-production', '音色设计', 'primary'],
  ['shuihuo.voice.dialogue', 'shuihuo-production', '多人配音台词', 'primary'],
  ['batch.hook-adaptation', 'batch-factory', '爆款开头改编', 'primary'],
  ['batch.original-director', 'batch-factory', '原文直转导演', 'primary'],
  ['batch.viral-director', 'batch-factory', '爆款开头导演', 'primary'],
  ['batch.character-meta', 'batch-factory', '人物提示词', 'primary'],
  ['batch.assets-h3-skill', 'batch-factory', 'H3 人物资产 Skill', 'primary'],
  ['batch.scene-meta', 'batch-factory', '场景提示词', 'primary'],
  ['batch.prop-meta', 'batch-factory', '道具提示词', 'primary'],
  ['batch.video-meta', 'batch-factory', '视频提示词', 'primary'],
  ['batch.visual-meta', 'batch-factory', '画面提示词', 'primary'],
  ['batch.prefix', 'batch-factory', '视频风格前缀', 'primary'],
  ['batch.opening-meta', 'batch-factory', '换开头元提示词', 'primary']
].map(([id, module, label, mode]) => Object.freeze({ id, module, label, mode })));

const SLOT_BY_PRESET_ID = Object.freeze({
  'script-extract': 'script.extract',
  'script-extract-assets': 'script.asset-extraction',
  'batch-assets-h3': 'script.asset-extraction',
  'batch-assets-sd-character-scene': 'script.asset-extraction',
  'script-extract-novel-panel': 'script.novel-panel-extract',
  'script-hook': 'script.hook',
  'script-continuous': 'script.continuous',
  'script-segmented': 'script.segmented',
  'script-general': 'script.general',
  'script-character-focus': 'script.character-focus',
  'script-audio-match': 'script.audio-match',
  'script-card-protocol': 'script.card.protocol',
  'script-format-screenplay': 'script.format.screenplay',
  'script-format-storyboard': 'script.format.storyboard',
  'script-format-shortdrama': 'script.format.shortdrama',
  'script-format-shotlist': 'script.format.shotlist',
  'script-format-q版': 'script.format.q版',
  'script-constraint-wrapper': 'script.constraint.wrapper',
  'script-constraint-prefix-smart-unified': 'script.constraint.prefix',
  'script-constraint-prefix-live-action': 'script.constraint.prefix',
  'script-constraint-prefix-3d': 'script.constraint.prefix',
  'script-constraint-prefix-2d': 'script.constraint.prefix',
  'script-constraint-prefix-guoman': 'script.constraint.prefix',
  'script-constraint-quality-4k': 'script.constraint.quality',
  'script-constraint-quality-sd-detail': 'script.constraint.quality',
  'script-constraint-quality-h3-shot-narrative': 'script.constraint.quality',
  'script-constraint-restriction-h3-visual-policy': 'script.constraint.restriction',
  'script-constraint-restriction-no-overlay': 'script.constraint.restriction',
  'script-constraint-negative-general': 'script.constraint.negative',
  'script-constraint-negative-sd-detail': 'script.constraint.negative',
  'script-quick-director-storyboard': 'script.quick-director',
  'script-quick-director-mode-strict': 'script.quick-director',
  'script-quick-director-mode-concise': 'script.quick-director',
  'script-quick-director-mode-balanced': 'script.quick-director',
  'script-quick-director-mode-detailed': 'script.quick-director',
  'script-quick-director-mode-example': 'script.quick-director',
  'script-quick-director-mode-reference': 'script.quick-director',
  'script-director-storyboard-master': 'script.director.master',
  'novel-analysis': 'novel.analysis',
  'novel-character': 'novel.character',
  'novel-outline': 'novel.outline',
  'novel-fetch-induce': 'novel-fetch.induce',
  'novel-fetch-hook': 'novel-fetch.hook',
  'shuihuo-extract-assets': 'shuihuo.asset.extraction',
  'shuihuo-extract-characters': 'shuihuo.asset.character-extraction',
  'shuihuo-extract-scenes': 'shuihuo.asset.scene-extraction',
  'shuihuo-extract-props': 'shuihuo.asset.prop-extraction',
  'shuihuo-asset-binding': 'shuihuo.asset.binding',
  'shuihuo-character-color-sheet': 'shuihuo.asset.character-sheet',
  'shuihuo-character-accessory-sheet': 'shuihuo.asset.character-sheet',
  'shuihuo-character-three-view': 'shuihuo.asset.character-sheet',
  'shuihuo-character-expression-sheet': 'shuihuo.asset.character-sheet',
  'shuihuo-character-single-view': 'shuihuo.asset.character-sheet',
  'shuihuo-smart-segmentation': 'shuihuo.segmentation.smart',
  'shuihuo-image-prompt': 'shuihuo.prompt.image',
  'shuihuo-video-prompt': 'shuihuo.prompt.video',
  'shuihuo-negative-prompt': 'shuihuo.prompt.negative',
  'batch-hook-adaptation': 'batch.hook-adaptation',
  'batch-original-director': 'batch.original-director',
  'batch-viral-director': 'batch.viral-director',
  'batch-character-meta': 'batch.character-meta',
	'batch-character-h3': 'batch.character-meta',
  'batch-assets-h3-skill': 'batch.assets-h3-skill',
  'batch-scene-meta': 'batch.scene-meta',
	'batch-scene-h3': 'batch.scene-meta',
  'batch-prop-meta': 'batch.prop-meta',
  'batch-video-meta': 'batch.video-meta',
	'batch-video-h3-director': 'batch.video-meta',
	'batch-video-sd-fixed-shot': 'batch.video-meta',
	'batch-video-sd': 'batch.video-meta',
  'batch-visual-meta': 'batch.visual-meta',
  'batch-prefix-general-anime': 'batch.prefix',
  'batch-prefix-modern-conflict': 'batch.prefix',
  'batch-prefix-ancient-drama': 'batch.prefix',
  'batch-prefix-xuanhuan-action': 'batch.prefix',
  'batch-prefix-suspense': 'batch.prefix',
  'batch-prefix-era-drama': 'batch.prefix',
  'batch-opening-meta': 'batch.opening-meta'
});

const BATCH_FACTORY_PRESET_REQUIREMENTS = Object.freeze({
  // Batch Factory has one combined extraction contract.  Script's character/
  // scene-only and novel-panel extractors use the same broad JSON format, but
  // they are not interchangeable with this book-level people/scene/prop rule.
  'assets.extraction': Object.freeze({ module: 'script', kind: 'base', slot: 'script.asset-extraction', extractionPreset: true, format: 'extract' }),
	'assets.character': Object.freeze({ module: 'batch-factory', kind: 'base', slot: 'batch.character-meta' }),
	'assets.scene': Object.freeze({ module: 'batch-factory', kind: 'base', slot: 'batch.scene-meta' }),
  'constraints.wrapper': Object.freeze({ module: 'script', kind: 'base', slot: 'script.constraint.wrapper' }),
  constraints: Object.freeze({ module: 'script', kind: 'addon', format: 'constraint' }),
  hook: Object.freeze({ module: 'batch-factory', kind: 'base', slot: 'batch.hook-adaptation' }),
  'director.original': Object.freeze({ module: 'batch-factory', kind: 'base', slot: 'batch.original-director' }),
  'director.viral': Object.freeze({ module: 'batch-factory', kind: 'base', slot: 'batch.viral-director' }),
  video: Object.freeze({ module: 'batch-factory', kind: 'base', slot: 'batch.video-meta' }),
  visual: Object.freeze({ module: 'batch-factory', kind: 'base', slot: 'batch.visual-meta' }),
  prefix: Object.freeze({ module: 'batch-factory', kind: 'base', slot: 'batch.prefix', format: 'video-prefix' })
});

function isPublishedPresetAllowed(preset, requirement) {
  if (!preset || !requirement || preset.module !== requirement.module || preset.kind !== requirement.kind) return false;
  if (requirement.id && preset.id !== requirement.id) return false;
  const format = preset.protocolLock?.format;
  if (requirement.extractionPreset && preset.extractionPreset !== true && format !== 'extract') return false;
  if (requirement.format && format !== requirement.format) return false;
  return !requirement.slot || preset.protocolLock?.slot === requirement.slot;
}

function slotDefinition(slot) {
  return SYSTEM_PRESET_SLOTS.find(item => item.id === slot) || null;
}

function slotsForModule(module) {
  return SYSTEM_PRESET_SLOTS
    .filter(item => item.module === module)
    .map(item => ({ id: item.id, label: item.label, mode: item.mode }));
}

function validatePresetSlot({ module, kind, protocolLock } = {}) {
  const definition = slotDefinition(protocolLock?.slot);
  const expectedMode = kind === 'base' ? 'primary' : kind;
  return Boolean(definition && definition.module === module && definition.mode === expectedMode);
}

function withOwnershipSlot(preset) {
  const slot = SLOT_BY_PRESET_ID[preset.id];
  if (!slot) throw new Error(`Missing system preset slot: ${preset.id}`);
  return { ...preset, protocolLock: { ...preset.protocolLock, slot } };
}

const SYSTEM_PRESETS = Object.freeze([
  ...SCRIPT_PRESETS.map(item => ({
    ...item,
    body: fs.readFileSync(path.join(promptsDir, item.source), 'utf8'),
    protocolLock: { ...item.protocolLock, source: item.source }
  })),
  ...SCRIPT_CONSTRAINT_PRESETS,
  ...NOVEL_PRESETS,
  ...NOVEL_FETCH_PRESETS,
  ...SHUIHUO_PRESETS,
  ...BATCH_PRESETS.map(item => ({ ...item, body: fs.readFileSync(path.join(promptsDir, item.source), 'utf8') })),
  ...BATCH_PREFIX_PRESETS
].map(item => Object.freeze({
  ...withOwnershipSlot(item),
  kind: item.kind || 'base',
  compatibleBaseIds: item.compatibleBaseIds || []
})));

function defaultPreset(id) {
  return SYSTEM_PRESETS.find(item => item.id === id) || null;
}

function defaultBody(id) {
  const preset = defaultPreset(id);
  if (!preset) throw new Error(`Unknown system preset: ${id}`);
  return preset.body;
}

function seedSystemPresets(store, actor) {
  for (const preset of SYSTEM_PRESETS) {
    const current = store.getPublished(preset.id);
    if (!current) {
      const draft = store.createDraft(actor, {
        id: preset.id,
        module: preset.module,
        name: preset.name,
        kind: preset.kind,
        description: preset.description,
        compatibleBaseIds: preset.compatibleBaseIds,
        body: preset.body,
        protocolLock: preset.protocolLock
      });
      store.publish(actor, draft.id, draft.version);
      continue;
    }

    // 对既有已发布版本追加必要守卫，不覆盖用户正文；同一次升级合并写入，
    // 避免 Q 版格式因两个兼容守卫连续产生两个版本。
    const currentBody = String(current.body || '');
    let upgradedBody = currentBody.trim();
    // 只迁移系统原始名称，避免覆盖管理员后来主动改过的自定义名称。
    const renameQuickDirector = preset.id === 'script-quick-director-storyboard'
      && current.name === '快速导演分镜';
    const renameBatchStoryboardMeta = preset.id === 'batch-video-meta'
      && current.name === '视频提示词元提示词';
    // v1-v3 used an independent Batch Factory prompt even after the UI started
    // calling it an automatic composition. Upgrade only the untouched legacy
    // default, never an administrator's custom body.
    const replaceLegacyBatchStoryboardBody = preset.id === 'batch-video-meta'
      && upgradedBody.startsWith('你负责把导演结果组织成可执行的视频分镜描述。');
    if (replaceLegacyBatchStoryboardBody) upgradedBody = String(preset.body || '').trim();
    // Upgrade the exact shipped single-character default, never a customized
    // administrator prompt. Publishing preserves the previous preset version.
    if (preset.id === 'batch-character-h3'
      && upgradedBody === fs.readFileSync(path.join(__dirname, 'fixtures/h3-character-single-legacy.md'), 'utf8').trim()) {
      upgradedBody = String(preset.body || '').trim();
    }
    const replaceAdaptedH3DirectorBody = preset.id === 'batch-video-h3-director'
      && (hasBodySha256(upgradedBody, H3_DIRECTOR_ADAPTED_V2_SHA256)
        || hasBodySha256(upgradedBody, H3_DIRECTOR_ADAPTED_RUNTIME_V2_SHA256)
        || hasBodySha256(upgradedBody, H3_DIRECTOR_LAYERED_TEMPLATE_V3_SHA256)
        || hasBodySha256(upgradedBody, H3_DIRECTOR_WITH_AUDIOVISUAL_POLICY_SHA256));
    if (preset.id === 'batch-video-h3-director'
      && (upgradedBody === H3_DIRECTOR_LEGACY_BODY || upgradedBody === H3_DIRECTOR_LEGACY_BODY_V0 || replaceAdaptedH3DirectorBody)) {
      upgradedBody = String(preset.body || '').trim();
    }
    const replaceLegacyH3AssetBody = preset.id === 'batch-assets-h3'
      && (isLegacyH3TwoCallAssetBody(upgradedBody) || hasBodySha256(upgradedBody, H3_ASSET_ADAPTED_V2_SHA256));
    if (replaceLegacyH3AssetBody) {
      upgradedBody = String(preset.body || '').trim();
    }
    const replaceInvalidScriptExtractBody = preset.id === 'script-extract'
      && isShippedInvalidScriptExtractBody(currentBody);
    if (replaceInvalidScriptExtractBody) {
      upgradedBody = String(preset.body || '').trim();
    }
    const replaceLegacySmartUnifiedBody = preset.id === 'script-constraint-prefix-smart-unified'
      && (upgradedBody === SMART_UNIFIED_LEGACY_BODY
        || isPreH3SmartUnifiedStyleSystemBody(upgradedBody)
        || hasBodySha256(upgradedBody, SMART_UNIFIED_ADAPTED_V3_SHA256)
        || isShippedChineseLabelSmartUnifiedV3(upgradedBody));
    if (replaceLegacySmartUnifiedBody) {
      upgradedBody = String(preset.body || '').trim();
    }
    const nextName = (renameQuickDirector || renameBatchStoryboardMeta) ? preset.name : current.name;
    const nextDescription = (renameQuickDirector || renameBatchStoryboardMeta || replaceLegacySmartUnifiedBody || replaceLegacyH3AssetBody || replaceInvalidScriptExtractBody)
      ? preset.description
      : current.description;
    const metadataChanged = nextName !== current.name || nextDescription !== current.description;
    if (preset.id === 'script-format-q版' && !upgradedBody.includes(Q_MINI_GUARD_MARKER)) {
      upgradedBody = `${upgradedBody}\n\n---\n\n${Q_MINI_GUARD}`;
    }
    if (BASE_SETUP_GUARD_PRESET_IDS.has(preset.id) && !upgradedBody.includes(BASE_SETUP_GUARD_MARKER)) {
      upgradedBody = `${upgradedBody}\n\n---\n\n${BASE_SETUP_GUARD}`;
    }
    const batchAlignmentGuard = batchScriptAlignmentGuard(preset);
    if (batchAlignmentGuard && !upgradedBody.includes(BATCH_SCRIPT_ALIGNMENT_MARKERS[preset.id])) {
      upgradedBody = `${upgradedBody}\n\n---\n\n${batchAlignmentGuard}`;
    }
    if (upgradedBody !== String(current.body || '').trim() || metadataChanged) {
      const draft = store.createDraft(actor, {
        id: current.id,
        module: current.module,
        name: nextName,
        kind: current.kind,
        description: nextDescription,
        compatibleBaseIds: current.compatibleBaseIds,
        body: upgradedBody,
        protocolLock: current.protocolLock
      });
      store.publish(actor, draft.id, draft.version);
      continue;
    }

    if (current.protocolLock?.slot === preset.protocolLock.slot) continue;
    const draft = store.createDraft(actor, {
      id: current.id,
      module: current.module,
      name: current.name,
      kind: current.kind,
      description: current.description,
      compatibleBaseIds: current.compatibleBaseIds,
      body: current.body,
      protocolLock: { ...current.protocolLock, slot: preset.protocolLock.slot }
    });
    store.publish(actor, draft.id, draft.version);
  }
}

function listPublishedForSlot(store, slot) {
  const definition = slotDefinition(slot);
  if (!definition || !store) return [];
  return store.listAll(definition.module)
    .filter(item => item.status === 'published' && item.kind === (definition.mode === 'primary' ? 'base' : definition.mode))
    .filter(item => item.protocolLock?.slot === definition.id)
    .map(item => ({
      id: item.id,
      name: item.name,
      description: item.description || '',
      version: item.version,
      slot: definition.id,
      kind: item.kind,
      constraintCategory: item.constraintCategory || item.protocolLock?.category || ''
    }));
}

function resolveSystemPresetBody(store, id) {
  const base = store?.getPublished(id);
  const preset = base || defaultPreset(id);
  if (!preset) return defaultBody(id);
  const addOns = store?.listAll(preset.module)
    .filter(item => item.status === 'published' && item.kind === 'addon')
    .filter(item => item.protocolLock?.format !== 'constraint')
    .filter(item => item.protocolLock?.slot === preset.protocolLock?.slot)
    .filter(item => item.compatibleBaseIds.length === 0 || item.compatibleBaseIds.includes(preset.id))
    .map(item => item.body) || [];
  return [preset.body, ...addOns].filter(Boolean).join('\n\n---\n\n');
}

module.exports = {
  SYSTEM_PRESETS,
  SYSTEM_PRESET_SLOTS,
  BATCH_FACTORY_PRESET_REQUIREMENTS,
  defaultBody,
  isPublishedPresetAllowed,
  listPublishedForSlot,
  seedSystemPresets,
  resolveSystemPresetBody,
  slotDefinition,
  slotsForModule,
  validatePresetSlot
};
