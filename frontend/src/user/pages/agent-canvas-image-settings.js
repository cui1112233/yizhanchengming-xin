export const IMAGE_NODE_SETTINGS_DEFAULTS = Object.freeze({
  ratio: '16:9',
  quality: 'standard',
  resolution: '2K',
  count: 1
});

const RATIOS = new Set(['16:9', '9:16', '1:1']);
const QUALITIES = new Set(['standard', 'high']);
const RESOLUTIONS = new Set(['1K', '2K', '4K']);

const SIZE_BY_RESOLUTION = {
  '1K': { '16:9': '1024x576', '9:16': '576x1024', '1:1': '1024x1024' },
  '2K': { '16:9': '2048x1152', '9:16': '1152x2048', '1:1': '2048x2048' },
  '4K': { '16:9': '4096x2304', '9:16': '2304x4096', '1:1': '4096x4096' }
};

export function normalizeImageNodeSettings(config = {}) {
  const ratio = RATIOS.has(config.ratio) ? config.ratio : IMAGE_NODE_SETTINGS_DEFAULTS.ratio;
  const quality = QUALITIES.has(config.quality) ? config.quality : IMAGE_NODE_SETTINGS_DEFAULTS.quality;
  const resolution = RESOLUTIONS.has(config.resolution) ? config.resolution : IMAGE_NODE_SETTINGS_DEFAULTS.resolution;
  const countValue = Number(config.count);
  const count = Number.isInteger(countValue) && countValue >= 1 && countValue <= 4
    ? countValue
    : IMAGE_NODE_SETTINGS_DEFAULTS.count;
  return { ratio, quality, resolution, count };
}

export function imageNodeSettingsSummary(config = {}) {
  const settings = normalizeImageNodeSettings(config);
  const qualityLabel = settings.quality === 'high' ? '高清画质' : '标准画质';
  return `${settings.ratio} · ${qualityLabel} · ${settings.resolution} · ${settings.count}张`;
}

export function imageGenerationRequestSettings(config = {}) {
  const settings = normalizeImageNodeSettings(config);
  return {
    aspectRatio: settings.ratio,
    quality: settings.quality,
    resolution: settings.resolution,
    count: settings.count,
    size: SIZE_BY_RESOLUTION[settings.resolution][settings.ratio]
  };
}
