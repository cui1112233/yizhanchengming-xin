import { useId } from 'react';

// CM 环形指示器：来自 Uiverse 的 dash/spin 描边动效，用于所有“正在执行 / 等待开始”的状态。
export function CmLoader({ size = 22, label = 'CM', tone = 'active', className = '', title = '加载中' }) {
  const gradientId = useId().replace(/:/g, '');
  const stroke = tone === 'pending' ? 'var(--sh-muted, #8391a4)' : 'var(--sh-accent, #63dfc0)';
  const textFill = tone === 'pending' ? 'var(--sh-muted, #8391a4)' : 'var(--sh-text, #f5f7fb)';
  return (
    <svg
      className={`cm-loader${className ? ` ${className}` : ''}`}
      width={size}
      height={size}
      viewBox="0 0 120 120"
      role="img"
      aria-label={title}
    >
      <title>{title}</title>
      <defs>
        <linearGradient id={`cm-loader-${gradientId}`} x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stopColor={stroke} stopOpacity="0.35" />
          <stop offset="100%" stopColor={stroke} stopOpacity="1" />
        </linearGradient>
      </defs>
      <g className="cm-loader-spin" style={{ transformOrigin: '60px 60px' }}>
        <circle
          className="cm-loader-dash"
          cx="60"
          cy="60"
          r="52"
          fill="none"
          stroke={`url(#cm-loader-${gradientId})`}
          strokeWidth="7"
          strokeLinecap="round"
        />
      </g>
      <text
        x="60"
        y="62"
        textAnchor="middle"
        dominantBaseline="central"
        fontSize="38"
        fontWeight="700"
        fill={textFill}
        className="cm-loader-text"
      >
        {label}
      </text>
    </svg>
  );
}

export default CmLoader;
