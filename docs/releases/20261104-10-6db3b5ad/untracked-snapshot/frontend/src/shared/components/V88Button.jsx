import { Button } from 'antd';

// 五种变体统一映射到 antd 原生能力，颜色由 ConfigProvider 令牌决定；
// loading 时同时 disabled，保证不能重复提交。
const VARIANT_PROPS = {
  primary: { type: 'primary' },
  secondary: { type: 'default' },
  ghost: { type: 'text' },
  danger: { type: 'default', danger: true },
  icon: { type: 'text' }
};

export function V88Button({
  variant = 'secondary',
  size,
  loading = false,
  disabled = false,
  icon = null,
  className = '',
  children,
  ...rest
}) {
  const props = VARIANT_PROPS[variant] || VARIANT_PROPS.secondary;
  const classes = ['v88-btn', `v88-btn--${variant}`, size === 'small' ? 'v88-btn--sm' : '', className].filter(Boolean).join(' ');
  return (
    <Button {...rest} {...props} size={size} icon={icon} loading={loading} disabled={disabled || loading} className={classes}>
      {children}
    </Button>
  );
}

export default V88Button;
