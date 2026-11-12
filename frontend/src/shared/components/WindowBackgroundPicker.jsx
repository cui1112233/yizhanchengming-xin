import { Button, message, Segmented, Slider } from 'antd';
import { readWindowBackgroundAppearance, saveWindowBackgroundAppearance } from '../styles/window-background-appearance.js';
import { Check, Upload } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import {
  effectiveWindowBackgroundMotion,
  readWindowBackgroundMotion,
  saveWindowBackgroundMotion,
  windowBackgroundMotions,
  windowBackgroundPreview,
  windowBackgroundChoices
} from '../styles/window-background.js';

export function WindowBackgroundPicker({ theme = '', value = 'none', customUrl = '', onChange, onUpload }) {
  const previewMode = theme === 'light' || theme === 'dark' ? theme : '';
  const inputRef = useRef(null);
  const [uploading, setUploading] = useState(false);
  const [fileName, setFileName] = useState('');
  const [appearance, setAppearance] = useState(readWindowBackgroundAppearance);
  useEffect(() => {
    const sync = () => setAppearance(readWindowBackgroundAppearance());
    window.addEventListener('storage', sync);
    return () => window.removeEventListener('storage', sync);
  }, []);
  function changeAppearance(key, next) {
    setAppearance(previous => saveWindowBackgroundAppearance({ ...previous, [key]: next }));
  }
  const [motion, setMotion] = useState(() => effectiveWindowBackgroundMotion(value, readWindowBackgroundMotion()));

  useEffect(() => {
    const stored = readWindowBackgroundMotion();
    const effective = effectiveWindowBackgroundMotion(value, stored);
    setMotion(effective);
    if (effective !== stored) saveWindowBackgroundMotion(effective);
  }, [value]);

  function changeMotion(next) {
    if (value === 'custom' && next === 'strong') return message.warning('自定义图片只支持柔和动效');
    setMotion(saveWindowBackgroundMotion(effectiveWindowBackgroundMotion(value, next)));
  }

  function choose(item) {
    if (item.id === 'custom' && motion === 'strong') {
      setMotion(saveWindowBackgroundMotion('soft'));
      message.warning('自定义图片只支持柔和动效');
    }
    onChange?.(item.id);
  }

  async function uploadFile(event) {
    const file = event.target.files?.[0];
    event.target.value = '';
    if (!file) return;
    if (!['image/png', 'image/jpeg', 'image/webp'].includes(file.type)) return message.error('仅支持 PNG、JPG 或 WebP 图片');
    if (file.size > 10 * 1024 * 1024) return message.error('背景图片不能超过 10MB');
    setUploading(true);
    try {
      const dataUrl = await new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = () => resolve(reader.result);
        reader.onerror = () => reject(new Error('图片读取失败'));
        reader.readAsDataURL(file);
      });
      await onUpload?.(dataUrl);
      setFileName(file.name);
      message.success('背景图片已保存并启用');
    } catch (error) {
      message.error(error.message || '上传背景图片失败');
    } finally { setUploading(false); }
  }

  const choices = customUrl ? [...windowBackgroundChoices, { id: 'custom', label: '我的上传背景', preview: `url(${JSON.stringify(customUrl)})` }] : windowBackgroundChoices;
  return (
    <section className="window-background-picker" aria-labelledby="window-background-title">
      <div className="window-background-heading">
        <div><h2 id="window-background-title">动态背景</h2><p>选择一段窗外风景，或让四个场景自动轮换。背景随账号保存。</p></div>
      </div>
      <div className="window-background-grid" role="listbox" aria-label="窗口背景">
        {choices.map(item => {
          const selected = item.id === value;
          return <button key={item.id} type="button" role="option" aria-selected={selected} className={`window-background-option${selected ? ' is-selected' : ''}`} onClick={() => choose(item)}>
            <span className={`window-background-preview is-${item.id}`} style={item.id === 'none' ? undefined : { backgroundImage: windowBackgroundPreview(item, previewMode), ...(item.size ? { backgroundSize: item.size } : null) }}>{item.id === 'none' ? '不使用' : null}</span>
            <span className="window-background-label">{item.video ? '▶ ' : ''}{item.label}</span>{selected ? <Check size={15} aria-hidden="true" /> : null}
          </button>;
        })}
      </div>
      <div className="window-background-footer">
        <div className="window-background-motion">
          <span>动效</span>
          <Segmented
            size="small"
            options={windowBackgroundMotions.map(item => ({ label: item.label, value: item.id }))}
            value={motion}
            onChange={changeMotion}
          />
        </div>
        <div className="window-background-actions">
          <Button icon={<Upload size={16} />} loading={uploading} onClick={() => inputRef.current?.click()}>上传图片</Button>
          {fileName ? <span className="window-background-filename" title={fileName}>{fileName}</span> : null}
          <input ref={inputRef} className="window-background-file" type="file" accept="image/png,image/jpeg,image/webp" onChange={uploadFile} hidden tabIndex={-1} />
        </div>
      </div>
      <div className="window-background-adjustments">
        <div className="window-background-adjustment"><div><strong>背景模糊</strong><small>保留光影，淡化画面细节</small></div><Slider ariaLabelForHandle="背景模糊" min={0} max={42} value={appearance.blur} onChange={next => changeAppearance('blur', next)} /><output>{appearance.blur} px</output></div>
        <div className="window-background-adjustment"><div><strong>背景亮度</strong><small>按钮、正文始终保持清晰</small></div><Slider ariaLabelForHandle="背景亮度" min={18} max={85} value={appearance.brightness} onChange={next => changeAppearance('brightness', next)} /><output>{appearance.brightness}%</output></div>
        <p>用于映窗视频壁纸，调整自动保存在当前浏览器。默认模糊 42px、亮度 50%。</p>
      </div>
    </section>
  );
}
