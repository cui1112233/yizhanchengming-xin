/** 右下「下一步」提示卡，文案逐字取自设计稿。 */
export function PrivateStudioNote({ mode }) {
  const copy = mode === 'comic'
    ? '先粘贴正文，再从左侧提取资产。'
    : '确认人物与场景后，点击“继续：生成分镜”。';
  return <aside className="private-studio-note-card" aria-label="下一步">
    <b>下一步</b>
    <p>{copy}</p>
  </aside>;
}

export default PrivateStudioNote;
