import { Button, Segmented } from 'antd'
import { Download, Settings2 } from 'lucide-react'
import { useEffect, useState } from 'react'

import { ExportDialog } from '../export/ExportDialog.jsx'
import { MaterialPanel } from '../materials/MaterialPanel.jsx'
import { Player } from '../player/Player.jsx'
import { videoClipsFromTimeline } from '../player/playerModel.js'
import { Timeline } from '../timeline/Timeline.jsx'
import { commandForShortcut } from './editorModel.js'
import { AudioInspector } from './AudioInspector.jsx'

export function EditorPage({ projectID }) {
  const [queuedMaterial, setQueuedMaterial] = useState(null)
  const [exportOpen, setExportOpen] = useState(false)
  const [preview, setPreview] = useState(null)
  const [playheadMs, setPlayheadMs] = useState(0)
  const [timelineSnapshot, setTimelineSnapshot] = useState(null)
  const [selection, setSelection] = useState(null)
  const [queuedClipUpdate, setQueuedClipUpdate] = useState(null)
  const videoClips = videoClipsFromTimeline(timelineSnapshot)
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
        <div><Button icon={<Settings2 size={16} />}>工程设置</Button><Button type="primary" icon={<Download size={16} />} onClick={() => setExportOpen(true)}>导出</Button></div>
      </header>
      <div className="editor-workspace">
        <MaterialPanel projectID={projectID} onInsert={setQueuedMaterial} />
        <Player
          projectID={projectID}
          clip={preview?.clip}
          clips={videoClips}
          timelineDurationMs={timelineSnapshot?.durationMs || preview?.timelineDurationMs || 0}
          playheadMs={playheadMs}
          onPlayheadChange={setPlayheadMs}
          onClipChange={clip => setPreview({ clip, timelineDurationMs: timelineSnapshot?.durationMs || 0 })}
        />
        {selection?.trackKind === 'audio' ? <AudioInspector clip={selection.clip} onUpdate={patch => {
          setSelection(current => current ? { ...current, clip: { ...current.clip, ...patch } } : current)
          setQueuedClipUpdate({ clipId: selection.clip.id, patch, nonce: crypto.randomUUID() })
        }} /> : <aside className="editor-panel inspector-panel" aria-label="检查器">
          <header><b>属性</b></header>
          <div className="inspector-empty"><Settings2 size={28} /><p>{selection ? '视频片段已选择' : '选择一个素材或片段'}</p><small>{selection ? '音频片段会在这里显示音量、倍速和时长' : '这里会显示画面、音频和变速参数'}</small></div>
        </aside>}
      </div>
      <Timeline
        projectID={projectID}
        materialToInsert={queuedMaterial}
        onMaterialInserted={() => setQueuedMaterial(null)}
        clipUpdate={queuedClipUpdate}
        onClipUpdated={() => setQueuedClipUpdate(null)}
        playheadMs={playheadMs}
        onSeek={setPlayheadMs}
        onTimelineChange={setTimelineSnapshot}
        onSelectClip={(clip, timelineDurationMs, trackKind) => {
          setSelection({ clip, trackKind })
          if (trackKind === 'video') setPreview({ clip, timelineDurationMs })
          setPlayheadMs(clip.startMs)
        }}
      />
      <ExportDialog open={exportOpen} projectID={projectID} onClose={() => setExportOpen(false)} />
    </main>
  )
}
