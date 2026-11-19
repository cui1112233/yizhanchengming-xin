// 导出本章（洛水作品管理页「导出本章」按钮，WorksApi.export_chapter 取证）：
// 洛水弹出「Text files (*.txt)」另存对话框，UTF-8 写入章节正文，成功提示「已导出：<路径>」；
// 剧本不走本章导出（洛水原文提示「剧本导出请使用复制内容或后续导出功能」）。
// V88 等价：浏览器 Blob 下载 `<章节标题>.txt`，内容为章节原文（source_text）；
// Web 端没有另存为对话框，浏览器下载即等价，剧本仍走手工复制/剧本合集导出。

const ILLEGAL_FILENAME = /[\\/:*?"<>|]+/g;

export function chapterExportFileName(chapter, index) {
  const title = String(chapter?.title || '').trim() || `第${index + 1}章`;
  return `${title.replace(ILLEGAL_FILENAME, '_')}.txt`;
}

// save 注入便于测试；默认走真实浏览器下载。
export function exportChapterText(chapter, index, save = defaultSave) {
  const content = String(chapter?.source_text || '');
  if (!content.trim()) throw Error('这一章还没有原文，没有可导出的内容');
  const fileName = chapterExportFileName(chapter, index);
  save(fileName, content);
  return fileName;
}

function defaultSave(fileName, content) {
  const blob = new Blob([content], { type: 'text/plain;charset=utf-8' });
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement('a');
  anchor.href = url;
  anchor.download = fileName;
  document.body.appendChild(anchor);
  anchor.click();
  anchor.remove();
  URL.revokeObjectURL(url);
}
