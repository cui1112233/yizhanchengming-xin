import { Button, Checkbox, Modal } from 'antd';
import { X } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { createHydrationGuard } from './hydrationGuard';

// 「自定义预设」弹窗：只负责勾选该分组下要展示的预置模型，不填 Key、不改地址。
// 勾选结果由父组件写入服务端 channelSelection。
export default function ApiPresetSelectionModal({
  open,
  groupTitle = '',
  presets = [],
  selectedIds = [],
  saving = false,
  onCancel,
  onConfirm
}) {
  const [pending, setPending] = useState(selectedIds);
  // 父组件每次渲染都会传入新数组，直接用数组做依赖会把用户刚勾的结果重置掉；
  // 改用内容指纹，并且用户已经动过勾选后不再回灌。
  const guardRef = useRef(null);
  if (!guardRef.current) guardRef.current = createHydrationGuard();
  const guard = guardRef.current;
  const selectedKey = selectedIds.join('\u0000');

  useEffect(() => {
    if (!open) {
      guard.reset();
      return;
    }
    if (!guard.canHydrate()) return;
    setPending(selectedKey ? selectedKey.split('\u0000') : []);
  }, [open, selectedKey, guard]);

  const pendingSet = new Set(pending);
  const selectedPresets = presets.filter(preset => pendingSet.has(preset.id));

  function toggle(id, checked) {
    guard.touch();
    setPending(current => {
      if (checked) return current.includes(id) ? current : [...current, id];
      return current.filter(item => item !== id);
    });
  }

  function restoreDefault() {
    guard.touch();
    setPending(presets.map(preset => preset.id));
  }

  return (
    <Modal
      open={open}
      title={`自定义预设${groupTitle ? ` · ${groupTitle}` : ''}`}
      width={760}
      onCancel={onCancel}
      maskClosable={!saving}
      closable={!saving}
      destroyOnClose
      footer={(
        <div className="api-preset-picker-footer">
          <Button className="api-preset-picker-restore" onClick={restoreDefault} disabled={saving}>初始化/恢复默认</Button>
          <span className="api-preset-picker-footer-right">
            <Button onClick={onCancel} disabled={saving}>取消</Button>
            <Button type="primary" loading={saving} onClick={() => onConfirm(pending)}>确定</Button>
          </span>
        </div>
      )}
    >
      <p className="api-preset-picker-note">
        只选择要展示的预置模型；不改变模型是否可用，填了 Key 的模型仍会进入模型下拉。
      </p>
      <div className="api-preset-picker">
        <div className="api-preset-picker-col">
          <div className="api-preset-picker-head">
            <span>全部预置模型</span>
            <small>{presets.length} 条</small>
          </div>
          <ul className="api-preset-picker-list">
            {presets.map(preset => (
              <li key={preset.id}>
                <Checkbox
                  checked={pendingSet.has(preset.id)}
                  disabled={saving}
                  onChange={event => toggle(preset.id, event.target.checked)}
                >
                  <span className="api-preset-picker-name">{preset.label}</span>
                  <span className="api-preset-picker-meta">
                    {preset.baseUrl || '未填写地址'} · 模型 {preset.model || '未填写'}
                  </span>
                </Checkbox>
              </li>
            ))}
          </ul>
        </div>
        <div className="api-preset-picker-col">
          <div className="api-preset-picker-head">
            <span>已选预置模型</span>
            <small>{selectedPresets.length} 条</small>
          </div>
          {selectedPresets.length ? (
            <ul className="api-preset-picker-list is-selected">
              {selectedPresets.map(preset => (
                <li key={preset.id}>
                  <span className="api-preset-picker-name">{preset.label}</span>
                  <button
                    type="button"
                    className="api-preset-picker-remove"
                    aria-label={`移除预置模型${preset.label}`}
                    disabled={saving}
                    onClick={() => toggle(preset.id, false)}
                  >
                    <X size={14} aria-hidden="true" />
                  </button>
                </li>
              ))}
            </ul>
          ) : (
            <p className="api-preset-picker-empty">还没有勾选预置模型；确定后该分组不显示预置卡片。</p>
          )}
        </div>
      </div>
    </Modal>
  );
}
