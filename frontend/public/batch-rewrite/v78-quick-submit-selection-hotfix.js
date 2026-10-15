(() => {
  function checkedTaskIds() {
    return [...document.querySelectorAll('.task-check:checked')]
      .map(input => String(input.dataset.id || '').trim())
      .filter(Boolean);
  }

  document.addEventListener('click', event => {
    const button = event.target.closest('#v78QuickSubmit');
    if (!button) return;

    const ids = checkedTaskIds();
    if (!ids.length || typeof window.qiantieSubmitSelectedTasks !== 'function') return;

    event.preventDefault();
    event.stopImmediatePropagation();
    window.qiantieSubmitSelectedTasks(ids);
  }, true);
})();
