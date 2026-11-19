import { Button, Checkbox } from 'antd';

// Catalog text is supplied by the backend prompt module; only selected IDs are editable.
export default function ReviewDimensions({ catalog, value = [], onChange, disabled }) {
  const selected = new Set(value);
  const update = (id, checked) => onChange?.(catalog.filter(d => d.id === id ? checked : selected.has(d.id)).map(d => d.id));
  return <section className="novel-agent-section novel-review-dimensions">
    <div className="novel-agent-section-heading">
      <div><h3>审稿维度（37 项）</h3><p>勾选后即参与审稿；维度说明为内置文案，不可改字，但可自由开关。当前已选 {value.length} / {catalog.length} 项。</p></div>
      <div className="novel-review-dimension-actions"><Button type="text" size="small" disabled={disabled || !catalog.length} onClick={() => onChange?.(catalog.map(d => d.id))}>全选</Button><Button type="text" size="small" disabled={disabled || !catalog.length} onClick={() => onChange?.([])}>全不选</Button></div>
    </div>
    {!catalog.length ? <p role="status">审稿维度目录加载失败，请重新打开 Agent 设置。</p> : <div className="novel-review-dimension-list">
      {catalog.map(d => <label className="novel-review-dimension-row" key={d.id}>
        <Checkbox aria-label={`审稿维度${d.id} ${d.title}`} checked={selected.has(d.id)} disabled={disabled} onChange={e => update(d.id, e.target.checked)} />
        <span><span className="novel-review-dimension-title"><span>{d.id}.</span><strong>{d.title}</strong></span><span className="novel-review-dimension-description">{d.content}</span></span>
      </label>)}
    </div>}
  </section>;
}
