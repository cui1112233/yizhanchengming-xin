// 异步回灌守卫：用户已经改过的本地输入，不允许被异步返回的旧值覆盖。
// 统一写法（API 设置 / 插件配置页共用）：
//   1. 用户 onChange 时 guard.touch()；
//   2. 异步回包写入本地状态前 if (!guard.canHydrate()) return；
//   3. 弹窗关闭或重新打开时 guard.reset()，只回灌一次。
// 请求飞行中可能又发生本地修改时，用 createRevisionTracker()：只认最新版本的回包。
export function createHydrationGuard() {
  let touched = false;
  return {
    touch() { touched = true; },
    canHydrate() { return touched === false; },
    reset() { touched = false; },
    get touched() { return touched; }
  };
}

export function createRevisionTracker() {
  let revision = 0;
  return {
    bump() { revision += 1; return revision; },
    current() { return revision; },
    isCurrent(captured) { return revision === captured; }
  };
}

// 异步保存回包只清理“刚提交的那份草稿”，保存期间用户新输入的内容保留。
export function sameText(left, right) {
  return String(left ?? '') === String(right ?? '');
}
