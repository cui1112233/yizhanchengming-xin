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
    event.preventDefault();
    event.stopImmediatePropagation();
    if (!ids.length) {
      window.qiantiePublishNovelFetchStatus?.('请先在任务列表勾选需要提交的任务', 'error');
      return;
    }
    if (typeof window.qiantieSubmitSelectedTasks !== 'function') {
      window.qiantiePublishNovelFetchStatus?.('提交网络按钮未准备好，请刷新页面后重试', 'error');
      return;
    }
    window.qiantieSubmitSelectedTasks(ids);
  }, true);
})();
