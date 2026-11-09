import { Button, Popover } from 'antd';
import { Check, Palette } from 'lucide-react';
import { listThemeDefinitions } from '../styles/theme.js';

const themes = listThemeDefinitions();

function ThemeOptions({ value, onChange }) {
  return (
    <div className="theme-picker-grid" role="listbox" aria-label="外观主题">
      {themes.map(theme => {
        const { tokens } = theme;
        const selected = theme.id === value;
        return (
          <button
            className={`theme-picker-option${selected ? ' is-selected' : ''}`}
            type="button"
            key={theme.id}
            role="option"
            aria-selected={selected}
            aria-pressed={theme.id === value}
            onClick={() => onChange?.(theme.id)}
          >
            <span className="theme-picker-swatches" aria-hidden="true">
              {[tokens.background, tokens.surface, tokens.primary, tokens.textPrimary].map((color, index) => (
                <i key={`${theme.id}-${index}`} style={{ background: color }} />
              ))}
            </span>
            <span>{theme.label}</span>
            {selected ? <Check size={14} aria-hidden="true" /> : null}
          </button>
        );
      })}
    </div>
  );
}

export function ThemePicker({ value = 'system', onChange, compact = false, variant = 'popover' }) {
  const panel = <ThemeOptions value={value} onChange={onChange} />;
  if (variant === 'grid') return <div className="theme-picker-inline">{panel}</div>;
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
