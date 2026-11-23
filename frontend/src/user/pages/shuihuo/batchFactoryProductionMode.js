export function batchFactoryProductionMode(value) {
  return value === 'comic' || value === 'text' ? value : 'video';
}

export function batchFactoryProductionModeVisibility(value) {
  const mode = batchFactoryProductionMode(value);
  return {
    mode,
    video: mode === 'video',
    comic: mode === 'comic',
    text: mode === 'text'
  };
}

export function batchFactoryShouldPrepareAudio(settings = {}) {
  return batchFactoryProductionMode(settings.productionMode) === 'video'
    && settings.audioPlanningEnabled === true;
}

export function batchFactoryShouldLoadPublishEnvironment({ open, activeTab, productionMode } = {}) {
  return open === true && activeTab === 'publish' && batchFactoryProductionMode(productionMode) === 'video';
}

export function batchFactoryShouldRecompileVideoPrompts(settings = {}) {
  return batchFactoryProductionMode(settings.productionMode) === 'video';
}
