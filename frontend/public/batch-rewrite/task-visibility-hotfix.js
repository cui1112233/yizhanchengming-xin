(() => {
  'use strict';

  if (typeof renderTasks !== 'function' || typeof taskDateKey !== 'function') return;

  const originalRenderTasks = renderTasks;

  function taskTimestamp(task) {
    const value = task?.updated_at || task?.updatedAt || task?.created_at || task?.createdAt || '';
    const date = new Date(value);
    return Number.isNaN(date.getTime()) ? 0 : date.getTime();
  }

  function latestTaskDate(tasks) {
    let latest = null;
    let latestAt = -1;
    for (const task of tasks) {
      const at = taskTimestamp(task);
      if (latest === null || at > latestAt) {
        latest = task;
        latestAt = at;
      }
    }
    return latest ? taskDateKey(latest) : '';
  }

  function appendEmptyState(message) {
    const body = document.getElementById('tasksBody');
    if (!body || body.children.length) return;
    const row = document.createElement('tr');
    row.setAttribute('data-qiantie-task-empty-state', '1');
    const cell = document.createElement('td');
    cell.colSpan = 20;
    cell.style.padding = '24px 12px';
    cell.style.textAlign = 'center';
    cell.style.color = '#64748b';
    cell.textContent = message;
    row.appendChild(cell);
    body.appendChild(row);
  }

  renderTasks = function renderTasksWithVisibleLatest(tasks) {
    const list = Array.isArray(tasks) ? tasks : [];
    const selectedDate = state.taskDate || (typeof todayDateKey === 'function' ? todayDateKey() : '');
    originalRenderTasks(list);

    if (state.tasks?.length) {
      return;
    }

    if (!list.length) {
      appendEmptyState('当前没有可提交任务。请先点击“开始处理”，处理完成后任务会显示在这里。');
      return;
    }

    const currentBatch = state.viewMode === 'current' && state.currentBatchIds?.size;
    appendEmptyState(currentBatch
      ? '当前批次暂无可显示任务。'
      : `当前日期 ${state.taskDate || '-'} 没有任务。已有 ${list.length} 条历史任务，请调整日期筛选。`);
  };
})();
