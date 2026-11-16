import { Button, Input, InputNumber, Modal } from 'antd';

export default function BlueprintEditor({ value, onChange, onCancel, onConfirm }) {
  function update(field, next) { onChange({ ...value, [field]: next }); }
  function updateRow(field, index, key, next) { update(field, value[field].map((row, i) => i === index ? { ...row, [key]: next } : row)); }
  return <Modal rootClassName="novel-modal" open={value !== null} title="确认或修改创作蓝图" width={720} okText="确认蓝图" cancelText="取消" onCancel={onCancel} onOk={() => onConfirm(value)}>
    {value && <div className="novel-panel"><p>核对作品设定、人物和卷章安排，确认后才开始正文。</p>
      {[
        ['title', '作品名称'], ['genre', '题材'], ['world', '世界设定'], ['coreConflict', '核心冲突']
      ].map(([key, label]) => <label key={key}>{label}<Input.TextArea aria-label={'蓝图' + label} rows={key === 'world' || key === 'coreConflict' ? 3 : 1} value={value[key] || ''} onChange={e => update(key, e.target.value)} /></label>)}
      <h3>人物</h3>{(value.characters || []).map((row, i) => <section className="novel-item" key={i}>{[['name', '姓名'], ['role', '身份'], ['goal', '目标'], ['conflict', '冲突']].map(([key, label]) => <label key={key}>{label}<Input value={row[key] || ''} onChange={e => updateRow('characters', i, key, e.target.value)} /></label>)}<Button onClick={() => update('characters', value.characters.filter((_, j) => j !== i))}>移除此人物</Button></section>)}
      <Button onClick={() => update('characters', [...(value.characters || []), { name: '', role: '', goal: '', conflict: '' }])}>添加人物</Button>
      <h3>卷章安排</h3>{(value.volumes || []).map((row, i) => <section className="novel-item" key={i}><label>卷序<InputNumber min={1} value={row.number} onChange={n => updateRow('volumes', i, 'number', n)} /></label><label>卷名<Input value={row.title || ''} onChange={e => updateRow('volumes', i, 'title', e.target.value)} /></label><label>本卷目标<Input.TextArea value={row.goal || ''} onChange={e => updateRow('volumes', i, 'goal', e.target.value)} /></label><div className="novel-actions"><label>起始章<InputNumber min={1} value={row.chapterRange?.[0]} onChange={n => updateRow('volumes', i, 'chapterRange', [n, row.chapterRange?.[1] || n])} /></label><label>结束章<InputNumber min={1} value={row.chapterRange?.[1]} onChange={n => updateRow('volumes', i, 'chapterRange', [row.chapterRange?.[0] || 1, n])} /></label></div><Button onClick={() => update('volumes', value.volumes.filter((_, j) => j !== i))}>移除此卷</Button></section>)}
      <Button onClick={() => update('volumes', [...(value.volumes || []), { number: (value.volumes?.length || 0) + 1, title: '', goal: '', chapterRange: [1, 10] }])}>添加卷章安排</Button>
    </div>}
  </Modal>;
}
