const PRESET_CATALOG_MODULES = ['script', 'batch-factory'];

export async function loadBatchFactoryPresetCatalog(loadModule) {
  const results = await Promise.allSettled(PRESET_CATALOG_MODULES.map(module => loadModule(module)));
  const catalog = [];
  const failedModules = [];
  results.forEach((result, index) => {
    if (result.status === 'fulfilled') {
      if (Array.isArray(result.value?.catalog)) catalog.push(...result.value.catalog);
      return;
    }
    failedModules.push(PRESET_CATALOG_MODULES[index]);
  });
  return { catalog, failedModules };
}
