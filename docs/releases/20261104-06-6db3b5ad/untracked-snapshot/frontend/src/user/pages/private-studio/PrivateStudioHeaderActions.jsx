import { Button } from 'antd';

/** 顶部标题右侧入口：文案按设计稿固定为「历史版本」「更多设置」。 */
export function PrivateStudioHeaderActions({
  onHistory,
  historyDisabled = false,
  historyTitle = '历史版本',
  onSettings,
  settingsDisabled = false
}) {
  return <div className="private-studio-header-actions">
    <Button
      type="text"
      className="private-studio-header-action"
      disabled={historyDisabled}
      title={historyTitle}
      onClick={onHistory}
    >历史版本</Button>
    <Button
      type="text"
      className="private-studio-header-action"
      disabled={settingsDisabled}
      title="更多设置"
      onClick={onSettings}
    >更多设置…</Button>
  </div>;
}

export default PrivateStudioHeaderActions;
