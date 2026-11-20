import { Button } from 'antd'
import { Maximize2, Pause, Play, RotateCcw, Volume2, VolumeX } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'

import { formatTimecode, materialFileURL, mediaTimeFromPlayhead, nextVideoClip, playbackInstanceKey, playheadFromMediaTime } from './playerModel.js'

export function Player({ projectID, clip, clips = [], timelineDurationMs = 0, playheadMs = 0, onPlayheadChange, onClipChange }) {
  const canvasRef = useRef(null)
  const videoRef = useRef(null)
  const autoContinueRef = useRef(false)
  const advancingRef = useRef(false)
  const [playing, setPlaying] = useState(false)
  const [muted, setMuted] = useState(false)
  const [error, setError] = useState('')
  const source = materialFileURL(projectID, clip?.materialId)
  const clipEndMs = clip ? clip.startMs + clip.durationMs : 0

  useEffect(() => {
    advancingRef.current = false
    setPlaying(false)
    setError('')
  }, [source])

  useEffect(() => {
    const video = videoRef.current
    if (!video || !clip || !Number.isFinite(video.duration)) return
    const target = mediaTimeFromPlayhead(clip, playheadMs)
    if (Math.abs(video.currentTime - target) > 0.12) video.currentTime = target
  }, [clip, playheadMs])

  const seekToClipStart = () => {
    const video = videoRef.current
    if (!video || !clip) return
    video.pause()
    autoContinueRef.current = false
    advancingRef.current = false
    video.currentTime = mediaTimeFromPlayhead(clip, clip.startMs)
    setPlaying(false)
    onPlayheadChange?.(clip.startMs)
  }

  const togglePlayback = async () => {
    const video = videoRef.current
    if (!video || !clip) return
    if (!video.paused) {
      video.pause()
      autoContinueRef.current = false
      setPlaying(false)
      return
    }
    if (playheadMs >= clipEndMs) seekToClipStart()
    try {
      video.playbackRate = Number(clip.speed) > 0 ? Number(clip.speed) : 1
      await video.play()
      setPlaying(true)
      setError('')
    } catch {
      setPlaying(false)
      setError('浏览器未能开始播放，请再次点击播放按钮')
    }
  }

  const finishClip = video => {
    if (!clip || advancingRef.current) return
    onPlayheadChange?.(clipEndMs)
    const next = nextVideoClip(clips, clip)
    if (!next) {
      video.pause()
      setPlaying(false)
      return
    }
    advancingRef.current = true
    autoContinueRef.current = true
    video.pause()
    setPlaying(false)
    onClipChange?.(next)
    onPlayheadChange?.(next.startMs)
  }

  const updatePlayhead = event => {
    if (!clip) return
    const next = playheadFromMediaTime(clip, event.currentTarget.currentTime)
    onPlayheadChange?.(next)
    if (next >= clipEndMs) {
      finishClip(event.currentTarget)
    }
  }

  const loadSelectedClip = async event => {
    const video = event.currentTarget
    video.currentTime = mediaTimeFromPlayhead(clip, playheadMs)
    video.playbackRate = Number(clip?.speed) || 1
    if (!autoContinueRef.current) return
    autoContinueRef.current = false
    try {
      await video.play()
      setPlaying(true)
      setError('')
    } catch {
      setPlaying(false)
      setError('下一片段已加载，请点击播放继续')
    }
  }

  return (
    <section className="editor-player" aria-label="播放器">
      <div className="editor-player__canvas" ref={canvasRef}>
        {source ? (
          <video
            key={playbackInstanceKey(clip)}
            ref={videoRef}
            src={source}
            preload="metadata"
            playsInline
            muted={muted}
            onLoadedMetadata={loadSelectedClip}
            onTimeUpdate={updatePlayhead}
            onPlay={() => setPlaying(true)}
            onPause={() => setPlaying(false)}
            onEnded={event => finishClip(event.currentTarget)}
            onError={() => setError('素材无法播放，请确认文件仍在素材库中且编码受浏览器支持')}
            onClick={togglePlayback}
          />
        ) : <span>选择时间线中的视频片段开始预览</span>}
        {error ? <div className="editor-player__error" role="alert">{error}</div> : null}
      </div>
      <div className="editor-player__controls">
        <Button type="text" title="回到片段开头" aria-label="回到片段开头" icon={<RotateCcw size={16} />} disabled={!clip} onClick={seekToClipStart} />
        <Button type="text" title={playing ? '暂停' : '播放'} aria-label={playing ? '暂停' : '播放'} icon={playing ? <Pause size={18} /> : <Play size={18} />} disabled={!clip} onClick={togglePlayback} />
        <span>{formatTimecode(playheadMs)} / {formatTimecode(timelineDurationMs)}</span>
        <Button type="text" title={muted ? '打开声音' : '静音'} aria-label={muted ? '打开声音' : '静音'} icon={muted ? <VolumeX size={16} /> : <Volume2 size={16} />} disabled={!clip} onClick={() => setMuted(value => !value)} />
        <Button type="text" title="全屏" aria-label="全屏" icon={<Maximize2 size={16} />} disabled={!clip} onClick={() => canvasRef.current?.requestFullscreen?.()} />
      </div>
    </section>
  )
}
