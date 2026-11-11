// 4 条文字类插件的纯逻辑：分段、提示词拼装、剧本提取与格式化。
// 从 PluginToolPage.jsx 抽出，便于用 node --test 直接验证五种场景，不依赖浏览器构建。

// 洛水实测的结果空态文案，4 条共用。
export const RESULT_PLACEHOLDER = '选择模型和提示词，输入内容后点击AI生成';

// 以下 3 个函数照搬洛水 bundle 里的 fc/pu/bc 行为，用于小说转剧的复制、格式化和保存 JSON。
export function extractScript(text) {
  if (!text || !text.trim()) return { success: false, message: '内容为空' };
  const pattern = /={3,}\s*剧本正文\s*={3,}\s*\n([\s\S]*?)(?=\n\s*={3,}\s*改编说明\s*={3,})/g;
  const blocks = [];
  let match;
  while ((match = pattern.exec(text)) !== null) blocks.push(match[1]);
  if (blocks.length === 0) return { success: false, message: '未找到剧本正文标记，将复制原内容' };
  const lines = [];
  for (const block of blocks) lines.push(...block.trim().split('\n').filter(line => line.trim()));
  const content = lines.join('\n');
  if (!content.trim()) return { success: false, message: '提取到的剧本内容为空' };
  return { success: true, content, scriptCount: blocks.length, lineCount: lines.length };
}

export function formatScriptText(text) {
  const extracted = extractScript(text);
  const base = extracted.success ? extracted.content : text;
  const lines = [];
  for (const raw of base.split('\n')) {
    const line = raw.trim();
    if (!line || line.startsWith('---') || (line.startsWith('（') && line.endsWith('）'))) continue;
    const matched = line.match(/^(?:.+?-(?:台词|内心)(?:（[^）]*）)?|旁白)[：:]\s*(.+)$/);
    lines.push(matched ? matched[1] : line);
  }
  return lines;
}

export function parseDialogueLine(line) {
  if (line.startsWith('旁白：') || line.startsWith('旁白:')) {
    const content = line.includes('：') ? line.split('：').slice(1).join('：') : line.split(':').slice(1).join(':');
    return { role_name: '旁白', text_content: content.trim(), emotion_name: '平静', strength_name: '中等' };
  }
  let matched = line.match(/^(.+?)-台词[（(]([^）)]+)[）)][：:]\s*(.+)$/);
  if (matched) return { role_name: matched[1].trim(), text_content: matched[3].trim(), emotion_name: matched[2].trim(), strength_name: '中等' };
  matched = line.match(/^(.+?)-台词[：:]\s*[（(]([^）)]+)[）)]\s*(.+)$/);
  if (matched) return { role_name: matched[1].trim(), text_content: matched[3].trim(), emotion_name: matched[2].trim(), strength_name: '中等' };
  matched = line.match(/^(.+?)-台词[：:]\s*(.+)$/);
  if (matched) return { role_name: matched[1].trim(), text_content: matched[2].trim(), emotion_name: '平静', strength_name: '中等' };
  matched = line.match(/^(.+?)-内心[（(]([^）)]+)[）)][：:]\s*(.+)$/);
  if (matched) return { role_name: matched[1].trim(), text_content: matched[3].trim(), emotion_name: matched[2].trim(), strength_name: '中等' };
  matched = line.match(/^(.+?)-内心[：:]\s*(.+)$/);
  if (matched) return { role_name: matched[1].trim(), text_content: matched[2].trim(), emotion_name: '平静', strength_name: '中等' };
  return null;
}

export function buildDialogueList(text) {
  const parsed = [];
  for (const raw of text.trim().split('\n')) {
    const line = raw.trim();
    if (!line || line.startsWith('---') || (line.startsWith('（') && line.endsWith('）'))) continue;
    const item = parseDialogueLine(line);
    if (item) parsed.push(item);
  }
  const merged = [];
  let current = null;
  for (const item of parsed) {
    if (current === null) {
      current = { ...item };
    } else if (item.role_name === current.role_name) {
      const combined = current.text_content + item.text_content;
      if (combined.length <= 150) {
        current.text_content = combined;
      } else {
        merged.push(current);
        current = { ...item };
      }
    } else {
      merged.push(current);
      current = { ...item };
    }
  }
  if (current) merged.push(current);
  return merged;
}

// 洛水 Python 侧 smart_split_text 的分段语义：按行累积到 segment_length，
// 超长单行再硬切；segment_length<=0 时整段一次发送。
export function splitInputByLength(text, limit) {
  const source = String(text || '');
  const size = Number(limit) || 0;
  if (!source.trim() || size <= 0 || source.length <= size) return [source];
  const parts = [];
  let current = '';
  const flush = () => {
    if (current) {
      parts.push(current);
      current = '';
    }
  };
  for (const line of source.split('\n')) {
    const candidate = current ? `${current}\n${line}` : line;
    if (candidate.length <= size) {
      current = candidate;
      continue;
    }
    if (line.length > size) {
      flush();
      for (let start = 0; start < line.length; start += size) parts.push(line.slice(start, start + size));
      continue;
    }
    flush();
    current = line;
  }
  flush();
  return parts.length ? parts : [source];
}

// 对应洛水 build_part_prompt：单段直接发送；多段时每段都带完整原文作为整体框架，
// 顺序模式再附加上一段之前的历史记录，并提示保持连贯。
export function buildPartPrompt({ full, part, index, total, history }) {
  if (total <= 1) return `请按照要求转换以下内容：\n\n${part}`;
  const original = `【完整原文】（供参考整体框架，共${full.length}字）：\n\n${full}`;
  if (history && history.trim()) {
    return `【重要提示】这是上次对话的历史记录，你可以参考每个场景的内容，确保上下场景连贯。\n\n${history}\n\n\n${original}\n\n\n【当前处理分段】这是第${index}部分，请继续按照之前的格式转换以下内容，注意保持与前面内容的连贯性：\n\n${part}`;
  }
  return `${original}\n\n\n【当前处理分段】这是第${index}/${total}部分，请按照要求转换以下内容：\n\n${part}`;
}
