const PLATFORM_CODE_LABELS = {
  HY: '黑岩付费',
  YG: '点众付费'
};

const PLATFORM_NAME_LABELS = {
  黑岩: '黑岩付费',
  黑岩付费书城: '黑岩付费',
  阳光: '点众付费'
};

export function bookPlatformName(book, platformNames = {}) {
  const metadata = book?.sourceMetadata || {};
  const platformCode = String(metadata.platformCode || '').trim().toUpperCase();
  if (PLATFORM_CODE_LABELS[platformCode]) return PLATFORM_CODE_LABELS[platformCode];
  const platform = String(book?.platform || '').trim();
  const normalizedPlatform = platform.toUpperCase();
  if (PLATFORM_CODE_LABELS[normalizedPlatform]) return PLATFORM_CODE_LABELS[normalizedPlatform];
  if (PLATFORM_NAME_LABELS[platform]) return PLATFORM_NAME_LABELS[platform];
  return String(metadata.platformName || metadata.platformLabel || platformNames[platform] || '').trim() || '未命名书城';
}

export function sensitiveReviewPresentation(metadata = {}, productionMode = 'video') {
  const count = Number(metadata.sensitiveReviewCount) || 0;
  return {
    title: `有 ${count} 处敏感词已自动打码`,
    description: String(productionMode || '').trim().toLowerCase() === 'comic'
      ? '漫画自动化会继续执行，不需要人工确认。'
      : '自动化会继续执行，不需要人工确认。',
  };
}
