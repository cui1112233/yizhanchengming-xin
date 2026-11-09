export const windowBackgrounds = [
  { id: 'none', label: '暂不使用窗口背景', preview: 'none' },
  { id: 'default', label: '默认自定义背景', preview: 'linear-gradient(135deg, #f8fafc 12%, #e9eef8 58%, #dbe7ff)' },
  { id: 'preset2', label: '预设背景 2', preview: 'linear-gradient(145deg, #f7f8fc, #e8e9fb 48%, #c8cdf8)' },
  { id: 'preset3', label: '预设背景 3', preview: 'linear-gradient(135deg, #f5fbff, #dff8f6 50%, #8dddf3)' },
  { id: 'preset4', label: '预设背景 4', preview: 'linear-gradient(145deg, #fffafa, #ffe0df 48%, #9fd8ff)' },
  { id: 'preset5', label: '预设背景 5', preview: 'linear-gradient(140deg, #fbfbff, #e8e4fa 48%, #94b5ff)' }
];

const supported = new Set([...windowBackgrounds.map(item => item.id), 'custom']);

export function normalizeWindowBackgroundId(value, customUrl = '') {
  const id = typeof value === 'string' ? value.trim() : '';
  if (id === 'custom' && !customUrl) return 'none';
  return supported.has(id) ? id : 'none';
}

export function windowBackgroundStyle(id, customUrl = '') {
  const normalized = normalizeWindowBackgroundId(id, customUrl);
  if (normalized === 'none') return {};
  if (normalized === 'custom') return { backgroundImage: `url(${JSON.stringify(customUrl)})` };
  return { backgroundImage: windowBackgrounds.find(item => item.id === normalized)?.preview || 'none' };
}
