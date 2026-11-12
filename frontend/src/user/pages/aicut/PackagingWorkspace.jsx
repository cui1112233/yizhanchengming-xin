import { Button, Empty } from 'antd';
import { projectTemplateCatalog } from './projectTemplates.js';
import './workspace-panels.css';

export default function PackagingWorkspace({ assets = [], onAutoEdit, onTemplate, onCaptions, onSpeech }) {
  const visuals = assets.filter(asset => ['video', 'image'].includes(asset.kind));
  return <div className="aicut-workspace-panel">
    <header><strong>智能包装</strong><p>使用本工程的画面、文字和配音，生成可继续编辑的内容。</p></header>
    {!visuals.length ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="先在素材页导入自己的画面" /> : <p>{visuals.length} 个画面素材可用于包装</p>}
    <Button block type="primary" disabled={!visuals.length} onClick={onAutoEdit}>创建自动剪辑方案</Button>
    <section><h3>工程模板</h3>{projectTemplateCatalog.map(template => <Button block key={template.id} disabled={!visuals.length} onClick={() => onTemplate(template)}>{template.labelZh}</Button>)}</section>
    <section><h3>字幕与声音</h3><Button block onClick={onCaptions}>识别素材字幕</Button><Button block onClick={onSpeech}>文字转语音</Button></section>
  </div>;
}
