/** 右下「下一步」提示卡，文案逐字取自设计稿。 */
export function PrivateStudioNote({ mode }) {
  const copy = mode === 'comic'
    ? '先提取资产，再生成漫画提示词，确认后生成漫画图片。'
    : '确认人物与场景后，点击右侧“开始生成”。';
  return <aside className="private-studio-note-card" aria-label="下一步">
    <b>下一步</b>
    <p>{copy}</p>
  </aside>;
}

export default PrivateStudioNote;
