(() => {
  function byId(id) { return document.getElementById(id); }
  function todayDateKey() {
    const date = new Date();
    return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`;
  }
  function shiftTaskDateKey(value, offset) {
    const match = String(value || '').match(/^(\d{4})-(\d{2})-(\d{2})$/);
    if (!match) return todayDateKey();
    const date = new Date(Date.UTC(Number(match[1]), Number(match[2]) - 1, Number(match[3])));
    date.setUTCDate(date.getUTCDate() + Number(offset || 0));
    return date.toISOString().slice(0, 10);
  }
  function setTaskDate(date) {
    state.viewMode = 'date';
    state.taskDate = date;
    const input = byId('taskDateFilter');
    if (input) input.value = date;
  }
  async function loadDateFilteredTasks(originalLoadTasks, options = {}) {
    return originalLoadTasks(options);
  }
  function v78TaskToday() {
    setTaskDate(todayDateKey());
    void loadDateFilteredTasks(loadTasks);
  }
  function bindLegacyTaskDateFilter(originalLoadTasks) {
    const input = byId('taskDateFilter');
    if (input && input.dataset.v78DateBound !== '1') {
      input.dataset.v78DateBound = '1';
      input.addEventListener('change', event => {
        event.stopPropagation();
        setTaskDate(input.value || todayDateKey());
        void loadDateFilteredTasks(originalLoadTasks);
      }, true);
    }
    const today = byId('taskTodayBtn');
    if (today) today.onclick = v78TaskToday;
    const previous = byId('taskPrevDayBtn');
    if (previous) previous.onclick = () => {
      setTaskDate(shiftTaskDateKey(state.taskDate || todayDateKey(), -1));
      void loadDateFilteredTasks(originalLoadTasks);
    };
    const next = byId('taskNextDayBtn');
    if (next) next.onclick = () => {
      setTaskDate(shiftTaskDateKey(state.taskDate || todayDateKey(), 1));
      void loadDateFilteredTasks(originalLoadTasks);
    };
  }
  function install() {
    if (window.__qiantieNovelFetchDateBooted) return true;
    if (typeof loadTasks !== 'function' || typeof state !== 'object') return false;
    const originalLoadTasks = loadTasks;
    window.__qiantieNovelFetchDateBooted = true;
    bindLegacyTaskDateFilter(originalLoadTasks);
    return true;
  }

  if (!install()) {
    let attempts = 0;
    const timer = window.setInterval(() => {
      if (install() || ++attempts >= 20) window.clearInterval(timer);
    }, 250);
  }
})();
