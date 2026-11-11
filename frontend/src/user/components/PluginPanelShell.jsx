import { ArrowLeft, X } from 'lucide-react';
import './plugin-panel.css';

// 插件面板公共骨架：面包屑 + 标题 + 内容区 + 结果区。
// 布局由各条自己决定（data-layout），不强行套同一个模子。
export function openPluginMarket() {
  window.history.pushState({}, '', '/plugins');
  window.dispatchEvent(new PopStateEvent('popstate'));
}

export function openPluginRoute(path) {
  window.history.pushState({}, '', path);
  window.dispatchEvent(new PopStateEvent('popstate'));
}

export default function PluginPanelShell({ name, heading, subtitle, layout = 'stack', authorized = false, icon = null, iconColor, primaryColor, onClose, children, result }) {
  // 主按钮跟随各插件自己的洛水主题色，不再是统一珊瑚红。
  const themeStyle = primaryColor
    ? { '--pp-primary': primaryColor, '--pp-primary-hover': `color-mix(in srgb, ${primaryColor} 86%, #000)` }
    : undefined;
  return (
    <main className="plugin-panel-page" style={themeStyle}>
      <div className="plugin-panel-inner">
        <nav className="plugin-panel-crumb" aria-label="面包屑">
          <button type="button" className="plugin-panel-back" onClick={openPluginMarket}>
            <ArrowLeft size={15} aria-hidden="true" />
            返回
          </button>
          <span aria-hidden="true">/</span>
          <button type="button" onClick={openPluginMarket}>插件市场</button>
          <span aria-hidden="true">&gt;</span>
          <span className="plugin-panel-current" aria-current="page">{name}</span>
        </nav>

        <div className="plugin-panel-titlebar">
          {icon ? (
            <span
              className="plugin-panel-title-icon"
              style={iconColor ? { background: `${iconColor}20` } : undefined}
            >
              {icon}
            </span>
          ) : null}
          <h1 className="plugin-panel-title">{heading || name}</h1>
          {authorized ? <span className="plugin-panel-authorized">已授权</span> : null}
          {onClose ? (
            <button type="button" className="plugin-panel-close" onClick={onClose} aria-label="关闭">
              <X size={16} aria-hidden="true" />
            </button>
          ) : null}
        </div>
        {subtitle ? <p className="plugin-panel-subtitle">{subtitle}</p> : null}

        <div className="plugin-panel-body" data-layout={layout}>{children}</div>

        {result ? <section className="plugin-panel-result" aria-label="结果">{result}</section> : null}
      </div>
    </main>
  );
}
