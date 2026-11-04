// 统一空态：图标 + 一句主文案 + 一句说明 + 一个主操作。
// 文案必须由调用方按真实条件传入，组件本身不声明任何状态。
export function V88EmptyState({ icon = null, title, description = '', action = null }) {
  return (
    <div className="v88-empty" role="status">
      {icon ? <div className="v88-empty-icon" aria-hidden="true">{icon}</div> : null}
      <p className="v88-empty-title">{title}</p>
      {description ? <p className="v88-empty-description">{description}</p> : null}
      {action ? <div className="v88-empty-action">{action}</div> : null}
    </div>
  );
}

export default V88EmptyState;
