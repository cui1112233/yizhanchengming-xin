/** 「已识别素材 / 可编辑」标签区；点击标签复用既有编辑入口。 */
export function PrivateStudioChips({ title, actionLabel, items = [], onAdd, addLabel = '+ 添加' }) {
  return <div className="private-studio-assets">
    {title || actionLabel ? <div className="private-studio-assets-head">
      <span>{title}</span>
      {actionLabel ? <span className="private-studio-assets-action">{actionLabel}</span> : null}
    </div> : null}
    <div className="private-studio-chips">
      {items.map(item => item.onClick
        ? <button
          key={item.key}
          type="button"
          className="private-studio-chip"
          title={item.title || item.label}
          onClick={item.onClick}
        >{item.label}{item.suffix ? <span> · {item.suffix}</span> : null}</button>
        : <span key={item.key} className="private-studio-chip" title={item.title || item.label}>{item.label}{item.suffix ? <span> · {item.suffix}</span> : null}</span>)}
      {onAdd ? <button type="button" className="private-studio-chip private-studio-chip-add" onClick={onAdd}>{addLabel}</button> : null}
    </div>
  </div>;
}

export default PrivateStudioChips;
