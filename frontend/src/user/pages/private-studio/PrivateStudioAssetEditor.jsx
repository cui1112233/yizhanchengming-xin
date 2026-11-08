import { Button, Modal, Typography } from 'antd';
import { X } from 'lucide-react';

/**
 * 剧本创作（私人定制）与水货生产「编辑资产」共用的一套资产编辑外壳。
 * 只负责结构、分区与按钮位；字段控件与保存接口仍由调用方传入，避免混淆两边的 payload。
 *
 * 字段映射表（来源字段 → 统一后的分区）：
 * - script.characters: 名称 → 基础；身份 / 外形 / 性格 → 设定（各自独立输入）
 * - script.scenes: 名称 → 基础；其余场景字段（若存在）→ 设定
 * - comic.character: 名称 → 基础；人物事实 / 人物关系 / 人物外形 → 设定（各自独立输入）
 * 注意：剧本的「外形」与漫画的「人物外形」语义不同，绝不合并成一个输入框。
 */
export const ASSET_EDITOR_FIELD_MAP = {
  script: {
    characters: [
      { key: '名称', label: '名称', section: 'basic', source: '名称' },
      { key: '身份', label: '身份', section: 'settings', source: '身份' },
      { key: '外形', label: '外形（剧本）', section: 'settings', source: '外形' },
      { key: '性格', label: '性格', section: 'settings', source: '性格' }
    ],
    scenes: [
      { key: '名称', label: '名称', section: 'basic', source: '名称' },
      { key: '描述', label: '设定 / 描述', section: 'settings', source: '描述' }
    ]
  },
  comic: {
    character: [
      { key: 'name', label: '名称', section: 'basic', source: 'name' },
      { key: 'facts', label: '人物事实', section: 'settings', source: 'facts' },
      { key: 'relationships', label: '人物关系', section: 'settings', source: 'relationships' },
      { key: 'appearance', label: '人物外形', section: 'settings', source: 'appearance' },
      { key: 'prompt', label: '合并资产提示词 / AI 生图提示词', section: 'prompts', source: 'prompt' }
    ],
    scene: [
      { key: 'name', label: '名称', section: 'basic', source: 'name' },
      { key: 'prompt', label: '合并资产提示词 / AI 生图提示词', section: 'prompts', source: 'prompt' }
    ],
    prop: [
      { key: 'name', label: '名称', section: 'basic', source: 'name' },
      { key: 'prompt', label: '合并资产提示词 / AI 生图提示词', section: 'prompts', source: 'prompt' }
    ]
  }
};

export function PrivateStudioAssetEditorSection({ title, description, children }) {
  return <section className="private-studio-asset-editor-section">
    <div className="private-studio-asset-editor-section-head">
      <Typography.Text strong>{title}</Typography.Text>
      {description ? <Typography.Text type="secondary" className="private-studio-asset-editor-section-desc">{description}</Typography.Text> : null}
    </div>
    <div className="private-studio-asset-editor-section-body">{children}</div>
  </section>;
}

export function PrivateStudioAssetEditor({ open, title, onClose, width = 920, fullscreen = false, basic, settings, prompts, ruleNote, aside, footer }) {
  return <Modal
    open={open}
    title={null}
    footer={null}
    closable={false}
    onCancel={onClose}
    maskClosable
    keyboard
    width={fullscreen ? '100vw' : width}
    style={fullscreen ? { top: 0, paddingBottom: 0 } : undefined}
    rootClassName="private-studio-asset-editor-modal"
    destroyOnClose
  >
    <div className="private-studio-asset-editor">
      <div className="private-studio-asset-editor-head">
        <Typography.Text strong className="private-studio-asset-editor-heading">{title}</Typography.Text>
        <Button className="private-studio-asset-editor-close" type="text" aria-label="关闭" title="关闭" icon={<X size={16} strokeWidth={1.8} aria-hidden="true" />} onClick={onClose} />
      </div>
      <div className="private-studio-asset-editor-body">
        <div className="private-studio-asset-editor-main">
          <PrivateStudioAssetEditorSection title="基础">{basic}</PrivateStudioAssetEditorSection>
          <PrivateStudioAssetEditorSection title="设定">{settings}</PrivateStudioAssetEditorSection>
          <PrivateStudioAssetEditorSection title="提示词">{prompts}</PrivateStudioAssetEditorSection>
          {ruleNote ? <p className="private-studio-asset-editor-note">{ruleNote}</p> : null}
        </div>
        <aside className="private-studio-asset-editor-aside">{aside}</aside>
      </div>
      <div className="private-studio-asset-editor-footer">{footer}</div>
    </div>
  </Modal>;
}

export default PrivateStudioAssetEditor;
