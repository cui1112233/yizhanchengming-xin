function imageAspectRatio(settings = {}) {
  return String(settings.imageAspectRatio || settings.aspectRatio || '').trim() || '9:16';
}

export function comicPromptForImageAspect(settings = {}, prompt = '') {
  const aspectRatio = imageAspectRatio(settings);
  return String(prompt || '')
    .replaceAll('1440×1920 或等比例 3:4', `与 ${aspectRatio} 画幅匹配的高清分辨率`)
    .replaceAll('3:4 竖版', `${aspectRatio} 画幅`)
    .replaceAll('3:4竖版', `${aspectRatio}画幅`)
    .replaceAll('{{image_aspect_ratio}}', aspectRatio)
    .replaceAll('3:4', aspectRatio)
    .trim();
}

export function comicPageGenerationPayload(settings = {}, prompt = '') {
  const normalizedPrompt = comicPromptForImageAspect(settings, prompt);
  return {
    comicPrompt: normalizedPrompt,
    imagePrompt: normalizedPrompt,
    provider: String(settings.imageProvider || '').trim(),
    modelId: String(settings.imageModelId || '').trim(),
    aspectRatio: imageAspectRatio(settings)
  };
}
