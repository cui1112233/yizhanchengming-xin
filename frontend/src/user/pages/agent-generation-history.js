export function generationHistoryRows(tasks = []) {
  const rows = [];
  for (const task of Array.isArray(tasks) ? tasks : []) {
    const taskId = String(task?.id || '').trim();
    const projectId = String(task?.projectId || '').trim();
    const state = String(task?.state || '').trim();
    const modelId = String(task?.modelId || '').trim();
    for (const result of Array.isArray(task?.results) ? task.results : []) {
      const resultId = String(result?.id || '').trim();
      const assetId = String(result?.assetId || result?.asset?.id || '').trim();
      const kind = String(result?.kind || task?.kind || result?.asset?.kind || '').trim();
      if (!taskId || !projectId || !resultId || !assetId || !['image', 'video', 'audio'].includes(kind)) continue;
      rows.push({
        taskId,
        resultId,
        assetId,
        projectId,
        kind,
        state,
        modelId,
        name: String(result?.asset?.name || result?.name || `${kind} 结果`).trim(),
        updatedAt: String(task?.updatedAt || result?.createdAt || '').trim(),
        asset: result?.asset && typeof result.asset === 'object' ? result.asset : {}
      });
    }
  }
  return rows;
}

export const generationHistoryKindOptions = [
  { value: '', label: '全部类型' },
  { value: 'image', label: '图片' },
  { value: 'video', label: '视频' },
  { value: 'audio', label: '音频' }
];
