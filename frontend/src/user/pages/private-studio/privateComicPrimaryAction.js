export function privateComicPrimaryAction({
  sourceReady = false,
  taskReady = false,
  assetCount = 0,
  pagesTotal = 0,
  promptReadyCount = 0,
  imageReadyCount = 0,
  busy = false,
  action = ''
} = {}) {
  const promptsComplete = pagesTotal > 0 && promptReadyCount >= pagesTotal;
  const imagesComplete = promptsComplete && imageReadyCount >= pagesTotal;

  if (imagesComplete) {
    return { key: 'complete', label: '漫画图片已完成', disabled: true };
  }
  if (promptsComplete) {
    return {
      key: 'images',
      label: action === 'pages' ? '生成漫画图片中…' : '生成漫画图片',
      disabled: busy || !taskReady
    };
  }
  return {
    key: 'prompts',
    label: action === 'prompts' ? '生成漫画提示词中…' : promptReadyCount > 0 ? '补齐漫画提示词' : '生成漫画提示词',
    disabled: busy || !sourceReady || !taskReady || assetCount <= 0
  };
}
