import { Button, Segmented } from 'antd'
import { Download, Settings2 } from 'lucide-react'
import { useEffect } from 'react'

import { MaterialPanel } from '../materials/MaterialPanel.jsx'
import { Player } from '../player/Player.jsx'
import { Timeline } from '../timeline/Timeline.jsx'
import { commandForShortcut } from './editorModel.js'

export function EditorPage({ projectID }) {
  useEffect(() => {
    const onKeyDown = event => {
      const command = commandForShortcut(event)
      if (!command || /INPUT|TEXTAREA|SELECT/.test(event.target?.tagName || '')) return
      event.preventDefault()
      window.dispatchEvent(new CustomEvent('magic-cut-command', { detail: command }))
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [])

  return (
    <main className="editor-page">
      <header className="editor-topbar">
        <div><b>未命名工程</b><small>{projectID}</small></div>
        <Segmented size="small" options={['剪辑', '字幕', '配音']} defaultValue="剪辑" />
        <div><Button icon={<Settings2 size={16} />}>工程设置</Button><Button type="primary" icon={<Download size={16} />}>导出</Button></div>
      </header>
      <div className="editor-workspace">
        <MaterialPanel projectID={projectID} />
        <Player />
        <aside className="editor-panel inspector-panel" aria-label="检查器">
          <header><b>属性</b></header>
          <div className="inspector-empty"><Settings2 size={28} /><p>选择一个素材或片段</p><small>这里会显示画面、音频和变速参数</small></div>
        </aside>
      </div>
      <Timeline />
    </main>
  )
}
