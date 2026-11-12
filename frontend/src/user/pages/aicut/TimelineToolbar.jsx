import { Button, Dropdown } from 'antd';
import { Clapperboard, Copy, CopyPlus, DiamondPlus, Flag, MoreHorizontal, Redo2, Scissors, SkipBack, SkipForward, Trash2, Undo2 } from 'lucide-react';

const actionIcons = {
  undo: Undo2, redo: Redo2, trimLeft: SkipBack, trimRight: SkipForward,
  split: Clapperboard, delete: Trash2, cut: Scissors, copy: Copy,
  duplicate: CopyPlus, addKeyframe: DiamondPlus, addMarker: Flag,
};

export default function TimelineToolbar({ primary, primaryActions = [], advancedItems = [], secondary }) {
  return (
    <div className="aicut-timeline-toolbar aicut-timeline-toolbar-v4" role="toolbar" aria-label="时间线工具栏">
      <div className="aicut-timeline-toolbar-primary" role="group" aria-label="剪辑操作">
        {primary}
        {primaryActions.map(action => {
          const Icon = actionIcons[action.key];
          return <Button
            key={action.key}
            size="small"
            aria-label={action.label}
            title={action.title}
            danger={action.danger}
            disabled={action.disabled}
            onClick={action.onClick}
            icon={Icon ? <Icon size={15} aria-hidden="true" /> : undefined}
          >{Icon ? null : action.label}</Button>;
        })}
        {advancedItems.length > 0 && <Dropdown
          trigger={['click']}
          menu={{ items: advancedItems }}
          overlayClassName="aicut-timeline-advanced-menu"
        >
          <Button size="small" icon={<MoreHorizontal size={14} aria-hidden="true" />} aria-label="更多操作" title="更多操作：逐帧、联动、复合片段和轨道设置" aria-haspopup="menu" />
        </Dropdown>}
      </div>
      <div className="aicut-timeline-toolbar-secondary" role="group" aria-label="时间线显示设置">
        {secondary}
      </div>
    </div>
  );
}
