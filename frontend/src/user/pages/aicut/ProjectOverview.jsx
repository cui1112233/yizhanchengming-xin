export default function ProjectOverview({ project, assetCount, trackCount, durationLabel }) {
  return <div className="aicut-project-overview">
    <section><h3>工程信息</h3><dl>
      <dt>工程名称</dt><dd>{project?.title || '未命名工程'}</dd>
      <dt>保存方式</dt><dd>云端工程</dd>
      <dt>素材数量</dt><dd>{assetCount} 个素材</dd>
    </dl></section>
    <section><h3>时间线</h3><dl>
      <dt>画面比例</dt><dd>{project?.aspectRatio || '9:16'}</dd>
      <dt>成片时长</dt><dd>{durationLabel}</dd>
      <dt>轨道数量</dt><dd>{trackCount} 条轨道</dd>
      <dt>画布背景</dt><dd>{project?.backgroundMode === 'transparent' ? '透明' : project?.backgroundColor || '#000000'}</dd>
    </dl></section>
    <p>选择时间线片段后，在这里调整画面、声音和文字。</p>
  </div>;
}
