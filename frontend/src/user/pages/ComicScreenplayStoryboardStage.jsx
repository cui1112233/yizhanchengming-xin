import { Alert, Button, Dropdown, Empty, Space } from 'antd';
import { useMemo, useState } from 'react';
import OriginalIcon from './ComicScreenplayIcon.jsx';
import { comicScriptEpisodes } from './comic-screenplay-stages.js';

export default function ComicScreenplayStoryboardStage({ workspace, canGenerate, working, onGenerate, onEditor, onTools, onView, promptName, engine, modelId }) {
  const [columns,setColumns]=useState(1);
  const episodes = useMemo(() => comicScriptEpisodes(workspace?.document?.content), [workspace?.document?.content]);
  const boards = (workspace?.storyboards || []).filter(board => !engine || board.engine === engine);
  return <div className="comic-storyboard-stage"><Alert type="info" showIcon icon={<OriginalIcon name="InformationCircleOutline" size={16}/>} closable message="按集制作分镜，先核对剧本与资产。图片、音色和剪辑工具可在当前阶段打开；切换页面不会自动生成。" />
    <div className="comic-stage-title-row"><div><h1>分镜</h1><p>共 {episodes.length} 集 · 已生成 {new Set(boards.filter(item=>item.shots?.length).map(item => item.episode_no)).size} 集 · {boards.reduce((sum, item) => sum + (item.shots?.length || 0), 0)} 个分镜</p></div>
      <div className="comic-storyboard-original-toolbar"><Button type="text" aria-label="分镜生成日志" title="查看制作任务与已保存的执行日志" icon={<OriginalIcon name="DocumentTextOutline" size={18}/>} onClick={()=>onTools('production')}/><Button type="text" aria-label={columns===2?'切换为单列显示':'切换为双列显示'} title={columns===2?'切换为单列显示':'切换为双列显示'} icon={<OriginalIcon name={columns===2?'ListOutline':'GridOutline'} size={18}/>} onClick={()=>setColumns(value=>value===2?1:2)}/><Button type="text" className="comic-storyboard-prompt-trigger" title="分镜词：在模型配置中选择并保存" onClick={()=>onTools('storyboard-prompts')}><span>{promptName||'分镜词'}</span><OriginalIcon name="ChevronDownOutline" size={12} style={{opacity:.5}}/></Button><Button type="text" title="导入本工作台导出的分镜文件" aria-label="导入分镜文件" icon={<OriginalIcon name="CloudUploadOutline" size={14}/>} onClick={()=>onTools('files')}>导入</Button><Dropdown trigger={['click']} menu={{items:[{key:'media',label:'媒体与剪辑'},{key:'production',label:'自动制作任务'},{key:'files',label:'导入 / 导出文件'},{key:'settings',label:'模型与创作配置'}],onClick:({key})=>onTools(key)}}><Button type="text" aria-label="更多分镜工具" title="打开本系统已接通的媒体、文件与任务工具" icon={<OriginalIcon name="EllipsisHorizontalOutline" size={14}/>}/></Dropdown></div>
    </div>
    {!modelId?<Alert type="info" showIcon message="尚未选择文本模型，请先在页头选择并保存；仍可查看或导入已有分镜。"/>:null}
    <div className={`comic-episode-production-list${columns===2?' is-double-column':''}`}>{episodes.map(episode => {
      const episodeBoards = boards.filter(item => item.episode_no === episode.no), shots = episodeBoards.flatMap(item => item.shots || []), scenes = (workspace?.shots || []).filter(item => item.episode_no === episode.no);
      return <article key={episode.no} className="comic-episode-production-card"><header><div><span>第 {episode.no} 集</span><h2>{episode.title}</h2><p>{scenes.length} 个场次 · {shots.length} 个分镜{shots.length ? ` · ${shots.reduce((sum, item) => sum + Number(item.duration_sec || 0), 0).toFixed(1)} 秒` : ' · 还未生成分镜脚本'}</p></div></header>
        <p className="comic-episode-script-preview" title="点击查看完整剧本" onClick={()=>onView(episode)}><b>文字梳理：</b>{episode.content.slice(0, 240)}{episode.content.length > 240 ? '…' : ''}</p>
        <Space wrap className="comic-storyboard-episode-actions"><Button size="small" shape="round" aria-label={shots.length?'重新生成分镜':'生成分镜脚本'} icon={<OriginalIcon name="FilmOutline" size={14}/>} disabled={!canGenerate || !!working || !promptName || !modelId} title={!modelId?'请先在页头选择并保存文本模型':!promptName?'请先选择并保存分镜提示词':undefined} loading={working === 'storyboard'} type={shots.length?'default':'primary'} onClick={() => onGenerate(episode.no)}>{shots.length ? '重新生成' : '分镜脚本'}</Button><Button size="small" shape="round" aria-label="进入分镜编辑器" icon={<OriginalIcon name="CreateOutline" size={14}/>} disabled={!shots.length} title={!shots.length?'请先生成或导入本集分镜，再进入编辑器':undefined} onClick={() => onEditor(episode.no)}>编辑</Button><Button size="small" shape="round" aria-label="导入分镜" icon={<OriginalIcon name="CloudUploadOutline" size={14}/>} onClick={() => onTools('files')}>导入</Button><Button size="small" shape="round" icon={<OriginalIcon name="BookOutline" size={14}/>} onClick={() => onView(episode)}>查看剧本</Button>{shots.length ? <Button size="small" shape="round" aria-label="查看已生成脚本" icon={<OriginalIcon name="ListOutline" size={14}/>} onClick={() => onEditor(episode.no)}>查看脚本</Button> : null}</Space>
      </article>;
    })}</div>
    {!episodes.length ? <Empty description="还没有可制作的剧本；可返回剧本阶段生成或导入并保存剧本。" /> : null}
  </div>;
}
