export function batchFactoryProductionMode(value) {
  return value === 'comic' ? 'comic' : 'video';
}

export function batchFactoryProductionModeVisibility(value) {
  const mode = batchFactoryProductionMode(value);
  return {
    mode,
    video: mode === 'video',
    comic: mode === 'comic'
  };
}
