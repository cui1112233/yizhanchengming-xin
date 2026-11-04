import { Modal, Typography } from 'antd';

/**
 * 剧本创作与漫画生成共用的「约束条件」弹窗。
 * 两个页签共用同一个壳和同一套分区；字段差异只体现在传入的分区列表里：
 * - 剧本：画面前缀词 / 画质约束 / 画面限制 / 负面提示词（支持系统预设与我的提示词）
 * - 漫画：画质约束 / 画面限制 / 负面提示词（系统预设）＋漫画提示词 / 图片描述词推导 / 画风
 */
export const PRIVATE_STUDIO_CONSTRAINT_FIELDS = {
  script: [
    { key: 'prefix', label: '画面前缀词' },
    { key: 'quality', label: '画质约束' },
    { key: 'restriction', label: '画面限制' },
    { key: 'negative', label: '负面提示词' }
  ],
  comic: [
    { key: 'comic_quality', label: '画质约束' },
    { key: 'comic_restriction', label: '画面限制' },
    { key: 'comic_negative', label: '负面提示词' }
  ]
};

export function PrivateStudioConstraintSection({ title, description, control, children }) {
  return <div className="private-studio-constraint-section">
    <div className="private-studio-constraint-section-head">
      <Typography.Text strong>{title}</Typography.Text>
      {control ? <div className="private-studio-constraint-section-control">{control}</div> : null}
    </div>
    {description ? <Typography.Paragraph type="secondary" className="private-studio-constraint-description">{description}</Typography.Paragraph> : null}
    {children}
  </div>;
}

export function PrivateStudioConstraintsModal({ open, onCancel, onOk, okText = '保存本次设置', title = '约束条件', footerNote = '', children }) {
  return <Modal
    title={title}
    open={open}
    onCancel={onCancel}
    onOk={onOk}
    okText={okText}
    width={720}
    destroyOnClose
    rootClassName="private-studio-constraints-modal"
  >
    <div className="private-studio-constraint-body">{children}</div>
    {footerNote ? <Typography.Paragraph className="script-constraint-help" type="secondary">{footerNote}</Typography.Paragraph> : null}
  </Modal>;
}

export default PrivateStudioConstraintsModal;
