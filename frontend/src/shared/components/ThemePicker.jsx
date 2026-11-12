import { Button, Popover } from 'antd';
import { Palette } from 'lucide-react';
import { composeThemeId, listThemeModes, listThemeSchemes, parseThemeId } from '../styles/theme.js';

const schemes = listThemeSchemes();
const modes = listThemeModes();

function ThemeOptions({ value, onChange }) {
  const { mode: activeMode } = parseThemeId(value);
  const activeScheme = 'glass';
  // 「跟随系统」时没有确定深浅，缩略卡按深色预览（四套方向都以深色为主场景）。
  const previewMode = activeMode === 'light' ? 'light' : 'dark';

  return (
    <div className="theme-picker-panel">
      <p className="theme-picker-group-label">风格</p>
      <div className="theme-scheme-grid" role="radiogroup" aria-label="配色风格">
        {schemes.map(scheme => {
          const selected = scheme.id === activeScheme;
          const preview = scheme.preview[previewMode];
          return (
            <button
              className={`theme-scheme-card${selected ? ' is-selected' : ''}`}
              type="button"
              key={scheme.id}
              data-scheme-preview={scheme.id}
              role="radio"
              aria-checked={selected}
              onClick={() => onChange?.(composeThemeId(scheme.id, activeMode))}
              style={{
                '--swatch-page': preview.page,
                '--swatch-surface': preview.surface,
                '--swatch-primary': preview.primary
              }}
            >
              <span className="theme-scheme-swatch" aria-hidden="true">
                <span className="theme-scheme-swatch-surface" />
                <span className="theme-scheme-swatch-lines">
                  <i />
                  <i />
                </span>
                <span className="theme-scheme-swatch-accent" />
              </span>
              <span className="theme-scheme-name">{scheme.label}</span>
            </button>
          );
        })}
      </div>
      <p className="theme-picker-group-label">深浅</p>
      <div className="theme-segmented" role="radiogroup" aria-label="深浅模式">
        {modes.map(mode => {
          const selected = mode.id === activeMode;
          return (
            <button
              className={`theme-segmented-option${selected ? ' is-selected' : ''}`}
              type="button"
              key={mode.id}
              role="radio"
              aria-checked={selected}
              onClick={() => onChange?.(composeThemeId(activeScheme, mode.id))}
            >
              {mode.label}
            </button>
          );
        })}
      </div>
    </div>
  );
}

export function ThemePicker({ value = 'coral-system', onChange, variant = 'popover' }) {
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
        主题
      </Button>
    </Popover>
  );
}
