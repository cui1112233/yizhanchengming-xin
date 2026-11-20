import { Button } from 'antd'
import { Maximize2, Pause, Play, RotateCcw, Volume2 } from 'lucide-react'
import { useState } from 'react'

export function Player() {
  const [playing, setPlaying] = useState(false)
  return (
    <section className="editor-player" aria-label="播放器">
      <div className="editor-player__canvas"><span>选择时间线中的素材开始预览</span></div>
      <div className="editor-player__controls">
        <Button type="text" icon={<RotateCcw size={16} />} />
        <Button type="text" icon={playing ? <Pause size={18} /> : <Play size={18} />} onClick={() => setPlaying(value => !value)} />
        <span>00:00:00 / 00:00:00</span>
        <Button type="text" icon={<Volume2 size={16} />} />
        <Button type="text" icon={<Maximize2 size={16} />} />
      </div>
    </section>
  )
}
