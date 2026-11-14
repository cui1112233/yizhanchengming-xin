const labels = { pending: '等待中', calling: '生成中', received: '已返回', invalid: '格式检查失败', failed: '失败', unknown: '需要核对', cancelled: '已取消', succeeded: '已保存' };

export default function ComicScreenplaySceneProgress({ episodes = [] }) {
  return <div aria-label="分镜场次进度">{episodes.map(episode => <div key={episode.episode_no}>
    <p>第{episode.episode_no}集：{labels[episode.status] || episode.status}{episode.error ? ` · ${episode.error}` : ''}</p>
    {episode.scenes?.length ? <ul>{episode.scenes.map(scene => <li key={scene.index}>第{scene.index}场：{labels[scene.status] || scene.status}{scene.error ? ` · ${scene.error}` : ''}</li>)}</ul> : null}
  </div>)}</div>;
}
