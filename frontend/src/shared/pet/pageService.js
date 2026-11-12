// Current-page services are installed by application pages, never model responses.
let currentService = null;
let currentOwner = null;
const listeners = new Set();

function notify() {
  for (const listener of [...listeners]) listener();
}

export function getPetPageService() {
  return currentService;
}

export function subscribePetPageService(listener) {
  listeners.add(listener);
  return () => listeners.delete(listener);
}

export function registerPetPageService(service) {
  if (!service || !['id', 'scope', 'label'].every(key => typeof service[key] === 'string' && service[key].trim()) || typeof service.render !== 'function') {
    throw new TypeError('页面宠物服务需要 id、scope、label 和真实面板 render。');
  }
  const sameScope = currentService?.id === service.id && currentService?.scope === service.scope;
  const registered = Object.freeze({
    id: service.id,
    scope: service.scope,
    label: service.label,
    render: service.render,
    openRequest: sameScope ? currentService.openRequest : 0
  });
  // Keep ownership separate from snapshots, because opening replaces the snapshot.
  const owner = {};
  currentOwner = owner;
  currentService = registered;
  notify();
  return () => {
    if (currentOwner !== owner) return;
    currentOwner = null;
    currentService = null;
    notify();
  };
}

export function openPetPageService(id) {
  if (!currentService || currentService.id !== id) return false;
  currentService = Object.freeze({ ...currentService, openRequest: currentService.openRequest + 1 });
  notify();
  return true;
}
