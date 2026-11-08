function count(value, maximum = Number.POSITIVE_INFINITY) {
  return Math.min(maximum, Math.max(0, Number(value || 0)));
}

function stage(key, label, detail, tone, { running = false, canRetry = false } = {}) {
  return { key, label, detail, tone, running, canRetry };
}

export function privateComicGenerationStatus({
  assetCount = 0,
  pagesTotal = 0,
  promptReadyCount = 0,
  imageReadyCount = 0,
  action = '',
  taskFailed = false,
  promptFailed = false,
  imageFailed = false
} = {}) {
  const total = count(pagesTotal);
  const promptsReady = count(promptReadyCount, total);
  const imagesReady = count(imageReadyCount, total);
  const promptsComplete = total > 0 && promptsReady >= total;
  const imagesComplete = promptsComplete && imagesReady >= total;
  const promptDetail = total > 0 ? `已完成 ${promptsReady}/${total} 页` : `${promptsReady} 页`;
  const imageDetail = total > 0 ? `已完成 ${imagesReady}/${total} 页` : `${imagesReady} 页`;
  const promptsRunning = action === 'prompts';
  const imagesRunning = action === 'pages';
  const promptHasFailed = !promptsComplete && (promptFailed || taskFailed);
  const imageHasFailed = promptsComplete && !imagesComplete && (imageFailed || taskFailed);

  let prompts;
  if (promptsComplete) {
    prompts = stage('prompts', '已完成', promptDetail, 'complete');
  } else if (promptsRunning) {
    prompts = stage('prompts', '生成中', promptDetail, 'running', { running: true });
  } else if (promptHasFailed) {
    prompts = stage('prompts', '生成失败', promptDetail, 'failed', { canRetry: true });
  } else if (promptsReady > 0) {
    prompts = stage('prompts', '部分完成', promptDetail, 'partial', { canRetry: true });
  } else {
    prompts = stage('prompts', '未生成', promptDetail, 'idle');
  }

  let images;
  if (imagesComplete) {
    images = stage('images', '已完成', imageDetail, 'complete');
  } else if (!promptsComplete) {
    images = stage('images', '等待提示词', '提示词完成后可生成', 'waiting');
  } else if (imagesRunning) {
    images = stage('images', '生成中', imageDetail, 'running', { running: true });
  } else if (imageHasFailed) {
    images = stage('images', '生成失败', imageDetail, 'failed', { canRetry: true });
  } else if (imagesReady > 0) {
    images = stage('images', '部分完成', imageDetail, 'partial', { canRetry: true });
  } else {
    images = stage('images', '未生成', imageDetail, 'idle');
  }

  return { prompts, images };
}
