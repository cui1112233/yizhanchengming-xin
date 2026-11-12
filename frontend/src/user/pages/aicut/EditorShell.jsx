import { forwardRef } from 'react';

const EditorShell = forwardRef(function EditorShell({ children, className = '', inspectorDrawerOpen = false, hasSelection = true, layout }, ref) {
  const style = {
    '--aicut-resource-width': `${layout.resourceCollapsed ? 48 : layout.resourceWidth}px`,
    '--aicut-inspector-width': `${layout.inspectorCollapsed || layout.inspectorMode === 'drawer' ? 0 : layout.inspectorWidth}px`,
    '--aicut-inspector-panel-width': `${layout.inspectorWidth}px`,
    '--aicut-timeline-height': `${layout.timelineCollapsed ? 40 : layout.timelineHeight}px`,
  };
  const classes = [
    'aicut-editor-shell', 'aicut-jianying-shell', 'aicut-reference-editor', 'aicut-editor-v3', 'aicut-editor-v4', className,
    layout.resourceCollapsed ? 'media-collapsed' : '',
    layout.inspectorCollapsed ? 'inspector-collapsed' : '',
    layout.timelineCollapsed ? 'timeline-collapsed' : '',
    layout.inspectorMode === 'drawer' ? 'inspector-drawer-mode' : '',
    !hasSelection ? 'inspector-empty' : '',
    inspectorDrawerOpen ? 'inspector-drawer-open' : '',
  ].filter(Boolean).join(' ');
  return <section ref={ref} style={style} className={classes} data-inspector-mode={layout.inspectorMode}>{children}</section>;
});

export default EditorShell;
