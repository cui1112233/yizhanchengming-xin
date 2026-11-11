import { Button, Popover } from 'antd';
import { Palette } from 'lucide-react';
import { listThemeDefinitions } from '../styles/theme.js';

const themes = listThemeDefinitions();
// 分段控件固定按 浅色 → 深色 → 跟随系统 排列，与目录里的定义先后无关。
const segmentOrder = ['light', 'dark', 'system'];
const orderedThemes = [...themes].sort(
  (a, b) => segmentOrder.indexOf(a.id) - segmentOrder.indexOf(b.id)
);

function ThemeOptions({ value, onChange }) {
  return (
    <div className="theme-segmented" role="radiogroup" aria-label="外观主题">
      {orderedThemes.map(theme => {
        const selected = theme.id === value;
        return (
          <button
            className={`theme-segmented-option${selected ? ' is-selected' : ''}`}
            type="button"
            key={theme.id}
            role="radio"
            aria-checked={selected}
            onClick={() => onChange?.(theme.id)}
          >
            {theme.label}
          </button>
        );
      })}
    </div>
  );
}

export function ThemePicker({ value = 'system', onChange, compact = false, variant = 'popover' }) {
  const panel = <ThemeOptions value={value} onChange={onChange} />;
  if (variant === 'grid') return <div className="theme-picker-inline">{panel}</div>;
  // theme-picker-trigger 只作为按用途限定的样式钩子（当前仅管理端侧栏使用）。
  return (
    <Popover
      content={panel}
      title="外观主题"
      trigger="click"
      placement="rightBottom"
      getPopupContainer={() => document.body}
      overlayClassName="theme-picker-popover"
    >
      <Button className="theme-picker-trigger" type="text" icon={<Palette size={18} />} aria-label="选择外观主题">
        {compact ? null : '主题'}
      </Button>
    </Popover>
  );
}
