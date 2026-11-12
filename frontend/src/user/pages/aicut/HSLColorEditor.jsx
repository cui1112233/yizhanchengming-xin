import { Button, InputNumber, Slider } from 'antd';
import { useEffect, useRef, useState } from 'react';
import { HSL_COLOR_BANDS, normalizeHSLAdjustments } from './hslColor.js';
import './hsl-color-editor.css';

const channels = [
  { key: 'hue', label: '色相' },
  { key: 'saturation', label: '饱和度' },
  { key: 'luminance', label: '明度' },
];

export default function HSLColorEditor({ value, onChange, onChangeComplete }) {
  const [selected, setSelected] = useState('red');
  const [draft, setDraft] = useState(() => normalizeHSLAdjustments(value));
  const latest = useRef(draft);
  const signature = JSON.stringify(normalizeHSLAdjustments(value));
  useEffect(() => {
    const next = JSON.parse(signature);
    latest.current = next;
    setDraft(next);
  }, [signature]);

  const update = next => {
    latest.current = next;
    setDraft(next);
    onChange?.(next);
  };
  const setChannel = (key, amount) => {
    if (typeof amount !== 'number' || !Number.isFinite(amount)) return;
    update(normalizeHSLAdjustments({ ...latest.current, [selected]: { ...latest.current[selected], [key]: amount } }));
  };
  const complete = () => onChangeComplete?.(latest.current);
  const reset = all => {
    const next = all ? normalizeHSLAdjustments() : { ...latest.current, [selected]: { hue: 0, saturation: 0, luminance: 0 } };
    update(next);
    onChangeComplete?.(next);
  };
  const bandLabel = HSL_COLOR_BANDS.find(band => band.key === selected).label;

  return (
    <section className="aicut-hsl-editor" aria-label="HSL 分色调节">
      <div className="aicut-hsl-bands" role="group" aria-label="选择颜色范围">
        {HSL_COLOR_BANDS.map(band => (
          <button key={band.key} type="button" className={`aicut-hsl-band${selected === band.key ? ' is-selected' : ''}`} aria-pressed={selected === band.key} aria-label={band.label} title={band.label} style={{ '--hsl-band-color': band.color }} onClick={() => setSelected(band.key)}>
            <span className="aicut-hsl-swatch" aria-hidden="true" />
            <span>{band.label}</span>
          </button>
        ))}
      </div>
      {channels.map(({ key, label }) => (
        <div key={key} className="aicut-hsl-channel">
          <label htmlFor={`aicut-hsl-${selected}-${key}`}>{label}</label>
          <Slider min={-100} max={100} step={1} value={draft[selected][key]} ariaLabelForHandle={`${bandLabel}${label}`} onChange={amount => setChannel(key, amount)} onChangeComplete={complete} />
          <InputNumber id={`aicut-hsl-${selected}-${key}`} aria-label={`${bandLabel}${label}数值`} size="small" min={-100} max={100} step={1} precision={1} value={draft[selected][key]} onChange={amount => setChannel(key, amount)} onBlur={complete} onPressEnter={complete} />
        </div>
      ))}
      <div className="aicut-hsl-actions">
        <Button size="small" type="text" onClick={() => reset(false)}>重置{bandLabel}</Button>
        <Button size="small" type="text" onClick={() => reset(true)}>重置全部</Button>
      </div>
    </section>
  );
}
