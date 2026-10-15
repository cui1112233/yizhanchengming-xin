(() => {
  const original = window.qiantieSubmitSelectedTasks;
  if (typeof original !== 'function' || original.__qiantieTaskListFallbackWrapped) return;

  const wrapped = ids => {
    const explicitIds = Array.isArray(ids)
      ? ids.map(id => String(id || '').trim()).filter(Boolean)
      : [];
    return explicitIds.length ? original(explicitIds) : original();
  };

  wrapped.__qiantieTaskListFallbackWrapped = true;
  window.qiantieSubmitSelectedTasks = wrapped;
})();
