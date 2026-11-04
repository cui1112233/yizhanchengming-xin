const FANKE_VIDEO_PREFIX = 'fanke-video-api::';

const PROVIDER_BY_MODEL = {
  'minimax-h3-video': 'autodl_comfyui',
  'yd2-mini-video': 'personal_api',
  'yd2.0-mini': 'personal_api',
  // 官方直连 Seedance 统一归 YFAI 分组；执行时后端仍走同一个 YFAI Seedance 适配器。
  'seedance-2-0-official': 'yfai',
  // 梵客聚合站的裸记录与具体型号都归梵客分组。
  'fanke-video-api': 'yfai_seedance',
  'local-doubao-executor-video': 'doubao_local_executor',
  'doubao-seedance': 'doubao_local_executor'
};

export function videoProviderForModel(modelId, fallback = 'personal_api') {
  const normalized = String(modelId || '').trim().toLowerCase();
  if (normalized.startsWith(FANKE_VIDEO_PREFIX)) return 'yfai_seedance';
  return PROVIDER_BY_MODEL[normalized] || fallback || 'personal_api';
}
