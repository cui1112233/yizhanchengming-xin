// 左侧 180px 导航:账号 / 任务 / 日志 / 设置 + 底部状态栏。
import { route, selectedTaskId } from '../state';
import type { Route } from '../state';
import { StatusBar } from './status-bar';

const NAV: { key: Route; label: string }[] = [
  { key: 'accounts', label: '账号' },
  { key: 'tasks', label: '任务' },
  { key: 'logs', label: '日志' },
  { key: 'settings', label: '设置' },
];

export function Sidebar() {
  return (
    <aside class="sidebar">
      <nav class="nav" aria-label="主导航">
        {NAV.map((item) => (
          <button
            key={item.key}
            type="button"
            class={`nav-item${route.value === item.key ? ' active' : ''}`}
            aria-current={route.value === item.key ? 'page' : undefined}
            onClick={() => {
              // 已在任务页时再点"任务"回到列表
              if (item.key === 'tasks' && route.value === 'tasks') selectedTaskId.value = null;
              route.value = item.key;
            }}
          >
            {item.label}
          </button>
        ))}
      </nav>
      <StatusBar />
    </aside>
  );
}
