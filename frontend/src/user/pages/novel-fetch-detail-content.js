export function taskDetailNeedsContent(data, tab) {
  return data?.content_loaded === false && ['original', 'ai', 'logs'].includes(tab);
}
