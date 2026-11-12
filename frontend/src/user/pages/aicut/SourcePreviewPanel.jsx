export default function SourcePreviewPanel({ asset, onClose }) {
  return <section className="aicut-preview aicut-reference-stage aicut-source-preview">
    <div className="aicut-panel-title"><div><strong>素材预览</strong><small title={asset.name}>{asset.name}</small></div><button type="button" onClick={onClose}>返回时间线</button></div>
    <div className="aicut-source-media">
      {asset.kind === 'image' ? <img src={asset.url} alt={asset.name} /> : asset.kind === 'audio' ? <div><span>{asset.name}</span><audio key={asset.id} src={asset.url} controls preload="metadata" /></div> : <video key={asset.id} src={asset.proxyUrl || asset.url} controls preload="metadata" playsInline />}
    </div>
    <div className="aicut-source-footer">预览原素材 · 双击素材或拖入轨道后开始剪辑</div>
  </section>;
}
