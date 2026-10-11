const CAPABILITY_LABELS = Object.freeze({
  'production.submit': '正式生产',
  'merge.run': '批量合成',
  'publish.121': '121 视频发布',
  'publish.yadi': 'Yadi 发布'
});

const REASON_LABELS = Object.freeze({
  'Production slice not released': '正式生产模块尚未发布',
  'Merge slice not released': '合成模块尚未发布',
  'Yadi is not enabled': 'Yadi 发布尚未启用'
});

export function unavailableCapabilitySummary(capabilities = {}) {
  return Object.entries(CAPABILITY_LABELS)
    .map(([key, label]) => {
      const record = capabilities?.[key];
      if (!record || record.available === true) return null;
      return {
        key,
        label,
        reason: REASON_LABELS[record.reason] || record.reason || '当前不可用'
      };
    })
    .filter(Boolean);
}
