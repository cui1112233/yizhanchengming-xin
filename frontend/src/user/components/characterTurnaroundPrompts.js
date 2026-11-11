// 角色三视图的提示词模板。正文取自 V88 提示词库中已迁移的三个洛水同名预设：
// shuihuo-character-three-view / shuihuo-character-accessory-sheet / shuihuo-character-color-sheet。
// 洛水导出的 181 条语料文件当前不在仓库里，第 212/214 条原文无法在本机核对；
// 这里不冒充语料原文，只使用 V88 已有的同名模板，补齐时替换 body 即可。

export const TURNAROUND_TEMPLATES = [
  {
    id: 'shuihuo-character-three-view',
    name: '三视图（标准全身）',
    ratio: '16:9',
    body: '左侧 1/3 区域为超大面部特写，右侧 2/3 区域并排展示正视图、右视图、后视图，整体比例 16:9，构图严谨规整，全身标准三视图，自然标准站立，人物比例精准合理，头身比标准，结构严谨，体态真实自然，真实的皮肤质感、画面干净，纯白色背景，图片不要文字。'
  },
  {
    id: 'shuihuo-character-accessory-sheet',
    name: '三视图 + 配饰细节',
    ratio: '16:9',
    body: '纯白色背景，不要带文字，根据角色描述词生成标准概念图(正面、侧面、背面)。右侧细节特写区域包括：面部特写（正面、3/4侧面）、双眼特写、双手平铺展开特写（五指张开、手掌正面）、服饰配件特写（衣领/领带/袖口/衣服下摆/鞋子等）。确保与主图的风格、色彩、特征完全一致。布局格式：三视图在画面左侧，细节特写排列在画面右侧，对半排版，整体比例 16:9。'
  },
  {
    id: 'shuihuo-character-color-sheet',
    name: '配色图 + 三视图',
    ratio: '16:9',
    body: '你是一位专业的角色设计师，请根据以下角色信息，生成一张完整的角色设计参考图(Character Design Sheet)。\n【布局要求】16:9横版，纯白色背景(#FFFFFF)，带中文标注，所有文字使用宋体字体。\n- 左上角：显示“我叫{character_name}”大标题，下方标注身份信息。\n- 左侧区域：标注“面部特写”，展示角色正面视角的面部特写（与三视图正面保持一致的角度）。\n- 左下角：标注“配色板”，展示主要颜色色块（发色、肤色、服装颜色等，带十六进制色值）。\n- 中央上方：三视图区域，标注“正面”、“侧面”、“背面”，展示角色全身三个视角。\n- 右上角：角色自我介绍文字区域，显示角色台词内容。\n- 底部中央：标注“局部细节区”，根据角色描述中的特征，展示3-5个该角色最具代表性的细节特写。\n- 右侧：标注“全身照比例照”，展示角色全身立绘配合身高标尺。'
  }
];

export function buildTurnaroundPrompt(template, { name = '', description = '' } = {}) {
  const base = String(template?.body || TURNAROUND_TEMPLATES[0].body)
    .replace(/\{character_name\}/g, name.trim() || '角色');
  const lines = [base];
  if (name.trim()) lines.push(`角色名：${name.trim()}`);
  if (description.trim()) lines.push(`角色设定：${description.trim()}`);
  return lines.join('\n\n');
}
