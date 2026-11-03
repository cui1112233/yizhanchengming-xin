function percentOf(progress = {}) {
  return Math.max(0, Math.min(100, Number(progress.percent || 0)));
}

export function giantMaterialOcrStatus(job = {}) {
  const state = String(job.state || '').trim().toLowerCase();
  const progress = job.progress || {};
  const percent = percentOf(progress);
  if (state === 'queued' && job.waitingForExecutor === true) {
    return {
      title: '等待 OCR 执行器上线',
      detail: '正文尚未开始读取。请到「设置 → 执行器状态」检查巨量素材执行器是否已安装、配对并在线；执行器上线后会自动领取任务。',
      progressStatus: 'normal'
    };
  }
  if (state === 'queued') {
    return { title: '等待 OCR 执行器领取', detail: '正文尚未开始读取，正在等待执行器领取任务。', progressStatus: 'normal' };
  }
  if (percent <= 0) {
    return { title: '滚屏 OCR', detail: '执行器已领取任务，正在准备读取正文…', progressStatus: 'active' };
  }
  const seconds = progress.total ? ` · 视频 ${progress.completed || 0}/${progress.total} 秒` : '';
  return { title: '滚屏 OCR', detail: `正在读取正文 ${percent}%${seconds}`, progressStatus: 'active' };
}
