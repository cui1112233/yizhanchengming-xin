import {ColorPicker as ThemeColorPicker,ConfigProvider,Input,Space} from 'antd';
import {useState,useEffect} from 'react';
import './publish-settings.css';

export default function TextColorPicker({ label, value, onChange }) {
  const [draft, setDraft] = useState(value);
  useEffect(() => setDraft(value), [value]);
  function edit(color) {
    setDraft(color);
    if (/^#[0-9a-f]{6}$/i.test(color)) onChange(color);
  }
  return <Space className="text-theme-color-control">
    <ConfigProvider theme={{token:{colorBgElevated:'var(--sh-panel)',colorBgContainer:'var(--sh-panel-strong)',colorText:'var(--sh-text)',colorTextSecondary:'var(--sh-muted)',colorBorder:'var(--sh-line)',colorPrimary:'var(--sh-accent)'}}}>
      <ThemeColorPicker value={value} disabledAlpha format="hex" rootClassName="batch-factory-theme-color-picker"
        getPopupContainer={trigger => trigger.parentElement} onChange={color => edit(color.toHexString())}>
        <button type="button" className="batch-factory-color-trigger" aria-label={label} style={{'--selected-color':value}}><i /></button>
      </ThemeColorPicker>
    </ConfigProvider>
    <Input aria-label={label+'色值'} value={draft} maxLength={7} style={{width: 100}}
      onChange={event => edit(event.target.value)} onBlur={() => setDraft(value)} />
  </Space>;
}

