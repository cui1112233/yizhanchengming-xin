import { Box, ImagePlus, UserRoundPlus } from 'lucide-react';

export default function DirectorAssetPanel({ images = [], onSetScene, onAddActor, onAddProp }) {
  return <section className="agent-director-assets" aria-label="项目图片">
    <h2>项目图片</h2>
    {!images.length ? <p>项目中还没有图片素材</p> : images.map(asset => <article className="agent-director-assets__item" key={asset.id}>
      {asset.url && <img src={asset.url} alt="" />}
      <strong>{asset.name || '未命名图片'}</strong>
       <div><button type="button" onClick={() => onSetScene(asset)}><ImagePlus size={14} />设为场景</button><button type="button" onClick={() => onAddActor(asset)}><UserRoundPlus size={14} />添加角色</button><button type="button" onClick={() => onAddProp?.(asset)}><Box size={14} />添加道具</button></div>
    </article>)}
  </section>;
}
