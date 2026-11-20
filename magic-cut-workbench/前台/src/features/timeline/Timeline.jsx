import { Alert, Button, Slider, Spin } from 'antd'
import { Copy, Redo2, Scissors, Trash2, Undo2, ZoomIn } from 'lucide-react'
import { useCallback, useEffect, useState } from 'react'

import { apiRequest } from '../../shared/api/client.js'
import { executeCommand, needsAutosave } from './commands.js'

function clipID() {
  return `clip_${crypto.randomUUID().replaceAll('-', '')}`
}

export function Timeline({ projectID, materialToInsert, onMaterialInserted, playheadMs = 0, onSeek, onSelectClip }) {
  const [timeline, setTimeline] = useState(null)
  const [selectedClip, setSelectedClip] = useState('')
  const [undoStack, setUndoStack] = useState([])
  const [redoStack, setRedoStack] = useState([])
  const [dirtyVersion, setDirtyVersion] = useState(0)
  const [savedDirtyVersion, setSavedDirtyVersion] = useState(0)
  const [saveState, setSaveState] = useState('读取中')
  const [conflict, setConflict] = useState('')

  useEffect(() => {
    apiRequest(`/api/magic-cut/projects/${projectID}/timeline`)
      .then(payload => { setTimeline(payload.timeline); setSaveState('已保存') })
      .catch(error => setConflict(error.message || '时间线读取失败'))
  }, [projectID])

  const run = useCallback(command => {
    setTimeline(current => {
      if (!current) return current
      const result = executeCommand(current, command)
      setUndoStack(stack => [...stack, result.inverse])
      setRedoStack([])
      setDirtyVersion(value => value + 1)
      return result.state
    })
  }, [])

  const undo = useCallback(() => {
    setUndoStack(stack => {
      if (!stack.length) return stack
      const inverse = stack.at(-1)
      setTimeline(current => {
        const result = executeCommand(current, inverse)
        setRedoStack(items => [...items, result.inverse])
        setDirtyVersion(value => value + 1)
        return result.state
      })
      return stack.slice(0, -1)
    })
  }, [])

  const redo = useCallback(() => {
    setRedoStack(stack => {
      if (!stack.length) return stack
      const inverse = stack.at(-1)
      setTimeline(current => {
        const result = executeCommand(current, inverse)
        setUndoStack(items => [...items, result.inverse])
        setDirtyVersion(value => value + 1)
        return result.state
      })
      return stack.slice(0, -1)
    })
  }, [])

  useEffect(() => {
    if (!materialToInsert || !timeline) return
    const kind = materialToInsert.mediaType === 'audio' ? 'audio' : 'video'
    const track = timeline.tracks.find(item => item.kind === kind) || timeline.tracks[0]
    const startMs = Math.max(0, ...track.clips.map(clip => clip.startMs + clip.durationMs))
    run({ type: 'insert', trackId: track.id, clip: { id: clipID(), materialId: materialToInsert.id, name: materialToInsert.originalName, startMs, durationMs: materialToInsert.metadata?.durationMs || 5000, sourceStartMs: 0, speed: 1 } })
    onMaterialInserted?.()
  }, [materialToInsert, timeline, run, onMaterialInserted])

  useEffect(() => {
    if (!needsAutosave(dirtyVersion, savedDirtyVersion) || !timeline || conflict) return
    setSaveState('保存中')
    const snapshot = timeline
    const dirtyAtSchedule = dirtyVersion
    const timer = window.setTimeout(() => {
      apiRequest(`/api/magic-cut/projects/${projectID}/timeline`, { method: 'PUT', body: JSON.stringify({ expectedRevision: snapshot.revision, timeline: snapshot }) })
        .then(payload => { setTimeline(current => ({ ...current, revision: payload.timeline.revision })); setSavedDirtyVersion(dirtyAtSchedule); setSaveState('已保存') })
        .catch(error => {
          if (error.code === 'TIMELINE_REVISION_CONFLICT') setConflict('工程已在另一个页面修改，自动保存已暂停')
          else setConflict(error.message || '时间线保存失败')
          setSaveState('保存失败')
        })
    }, 700)
    return () => window.clearTimeout(timer)
  }, [dirtyVersion, savedDirtyVersion, timeline, projectID, conflict])

  useEffect(() => {
    const handleCommand = event => {
      const command = event.detail
      if (command === 'undo') undo()
      if (command === 'redo') redo()
      if (command === 'delete-selection' && selectedClip) run({ type: 'delete', clipId: selectedClip })
      if (command === 'copy' && selectedClip) window.sessionStorage.setItem('magic-cut-copied-clip', selectedClip)
      if (command === 'paste' && selectedClip) run({ type: 'duplicate', clipId: selectedClip, newClipId: clipID(), offsetMs: 250 })
    }
    window.addEventListener('magic-cut-command', handleCommand)
    return () => window.removeEventListener('magic-cut-command', handleCommand)
  }, [redo, run, selectedClip, undo])

  if (!timeline) return <section className="editor-timeline timeline-loading"><Spin /><span>{conflict || '正在读取时间线'}</span></section>

  const selected = timeline.tracks.flatMap(track => track.clips).find(clip => clip.id === selectedClip)
  const seekFromPointer = event => {
    const bounds = event.currentTarget.getBoundingClientRect()
    const next = Math.max(0, Math.min(timeline.durationMs, Math.round((event.clientX - bounds.left) * 100)))
    const videoClip = timeline.tracks.find(track => track.kind === 'video')?.clips.find(clip => next >= clip.startMs && next <= clip.startMs + clip.durationMs)
    if (videoClip) {
      setSelectedClip(videoClip.id)
      onSelectClip?.(videoClip, timeline.durationMs)
    }
    onSeek?.(next)
  }
  return (
    <section className="editor-timeline" aria-label="时间线">
      <div className="timeline-toolbar" aria-label="工具栏">
        <Button type="text" icon={<Undo2 size={16} />} disabled={!undoStack.length} onClick={undo}>撤销</Button>
        <Button type="text" icon={<Redo2 size={16} />} disabled={!redoStack.length} onClick={redo}>重做</Button>
        <Button type="text" icon={<Scissors size={16} />} disabled={!selected} onClick={() => selected && run({ type: 'split', clipId: selected.id, atMs: selected.startMs + Math.round(selected.durationMs / 2), newClipId: clipID() })}>分割</Button>
        <Button type="text" icon={<Copy size={16} />} disabled={!selected} onClick={() => selected && run({ type: 'duplicate', clipId: selected.id, newClipId: clipID(), offsetMs: 250 })}>复制</Button>
        <Button type="text" danger icon={<Trash2 size={16} />} disabled={!selected} onClick={() => selected && run({ type: 'delete', clipId: selected.id })}>删除</Button>
        <span className="timeline-toolbar__spacer" />
        <small className={`timeline-save is-${saveState}`}>{saveState}</small>
        <ZoomIn size={15} /><Slider min={1} max={10} defaultValue={4} />
      </div>
      {conflict ? <Alert className="timeline-conflict" type="warning" showIcon message={conflict} /> : null}
      <div className="timeline-ruler"><span>00:00</span><span>00:05</span><span>00:10</span><span>00:15</span><span>00:20</span></div>
      {timeline.tracks.map(track => (
        <div className="timeline-track" key={track.id}>
          <b>{track.kind === 'audio' ? '音频' : '视频'} 1</b>
          <div className="timeline-lane" onClick={seekFromPointer}>
            <i className="timeline-playhead" style={{ left: `${Math.max(0, playheadMs / 100)}px` }} aria-hidden="true" />
            {track.clips.length ? track.clips.map(clip => (
              <button type="button" key={clip.id} className={`timeline-clip ${selectedClip === clip.id ? 'is-selected' : ''}`} style={{ left: `${Math.max(0, clip.startMs / 100)}px`, width: `${Math.max(72, clip.durationMs / 100)}px` }} onClick={event => { event.stopPropagation(); setSelectedClip(clip.id); onSelectClip?.(clip, timeline.durationMs); onSeek?.(clip.startMs) }}>
                {clip.name || clip.materialId}
              </button>
            )) : <span>双击左侧素材添加到这里</span>}
          </div>
        </div>
      ))}
    </section>
  )
}
