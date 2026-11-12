import { Button, Dropdown } from 'antd';
import { MoreHorizontal } from 'lucide-react';

export default function TimelineToolbar({ primary, advancedItems = [], secondary }) {
  return (
    <div className="aicut-timeline-toolbar aicut-timeline-toolbar-v4" role="toolbar" aria-label="时间线工具栏">
      <div className="aicut-timeline-toolbar-primary" role="group" aria-label="剪辑操作">
        {primary}
        <Dropdown
          trigger={['click']}
          menu={{ items: advancedItems }}
          overlayClassName="aicut-timeline-advanced-menu"
        >
          <Button size="small" icon={<MoreHorizontal size={14} aria-hidden="true" />} aria-label="更多操作" title="更多操作：剪贴板、关键帧、标记和复合片段" aria-haspopup="menu" />
        </Dropdown>
      </div>
      <div className="aicut-timeline-toolbar-secondary" role="group" aria-label="时间线显示设置">
        {secondary}
      </div>
    </div>
  );
}
