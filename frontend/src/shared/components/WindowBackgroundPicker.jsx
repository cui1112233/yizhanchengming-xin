import { Button, message } from 'antd';
import { Check, Upload } from 'lucide-react';
import { useRef, useState } from 'react';
import { windowBackgrounds } from '../styles/window-background.js';

export function WindowBackgroundPicker({ value = 'none', customUrl = '', onChange, onUpload }) {
  const inputRef = useRef(null);
  const [uploading, setUploading] = useState(false);

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
      message.success('背景图片已保存并启用');
    } catch (error) {
      message.error(error.message || '上传背景图片失败');
    } finally { setUploading(false); }
  }

  const choices = customUrl ? [...windowBackgrounds, { id: 'custom', label: '我的上传背景', preview: `url(${JSON.stringify(customUrl)})` }] : windowBackgrounds;
  return (
    <section className="window-background-picker" aria-labelledby="window-background-title">
      <div className="window-background-heading">
        <div><h2 id="window-background-title">窗口背景</h2><p>预设图片随账号同步，自定义图片保存到服务器用户资源。</p></div>
        <Button icon={<Upload size={16} />} loading={uploading} onClick={() => inputRef.current?.click()}>上传图片</Button>
        <input ref={inputRef} className="window-background-file" type="file" accept="image/png,image/jpeg,image/webp" onChange={uploadFile} />
      </div>
      <div className="window-background-grid" role="listbox" aria-label="窗口背景">
        {choices.map(item => {
          const selected = item.id === value;
          return <button key={item.id} type="button" role="option" aria-selected={selected} className={`window-background-option${selected ? ' is-selected' : ''}`} onClick={() => onChange?.(item.id)}>
            <span className={`window-background-preview is-${item.id}`} style={item.id === 'none' ? undefined : { backgroundImage: item.preview }}>{item.id === 'none' ? '不使用' : null}</span>
            <span className="window-background-label">{item.label}</span>{selected ? <Check size={15} aria-hidden="true" /> : null}
          </button>;
        })}
      </div>
    </section>
  );
}
