import { InputNumber, Slider } from 'antd'
import { Volume2 } from 'lucide-react'

import { audioInspectorValues, patchAudioClip } from './audioInspectorModel.js'

export function AudioInspector({ clip, onUpdate }) {
  const values = audioInspectorValues(clip)
  const update = change => onUpdate?.(patchAudioClip(clip, change))

  return (
    <aside className="editor-panel inspector-panel audio-inspector" aria-label="音频属性">
      <header><b>音频属性</b><small>{clip.name || '音频片段'}</small></header>
      <div className="audio-inspector__body">
        <div className="audio-inspector__title"><Volume2 size={18} /><span>基础设置</span></div>
        <div className="audio-property">
          <label htmlFor="audio-volume">音量</label>
          <div><Slider id="audio-volume" min={0} max={200} step={1} marks={{ 100: '100%' }} value={values.volumePercent} onChange={value => update({ volumePercent: value })} /><InputNumber min={0} max={200} step={1} value={values.volumePercent} suffix="%" onChange={value => value != null && update({ volumePercent: value })} /></div>
          <small>调到 0% 即静音</small>
        </div>
        <div className="audio-property">
          <label htmlFor="audio-speed">倍速</label>
          <div><Slider id="audio-speed" min={0.1} max={4} step={0.1} marks={{ 1: '1x', 2: '2x', 3: '3x' }} value={values.speed} onChange={value => update({ speed: value })} /><InputNumber min={0.1} max={4} step={0.1} precision={1} value={values.speed} suffix="x" onChange={value => value != null && update({ speed: value })} /></div>
        </div>
        <div className="audio-property">
          <label htmlFor="audio-duration">时长</label>
          <div><Slider id="audio-duration" min={values.minDurationSeconds} max={values.maxDurationSeconds} step={0.01} value={values.durationSeconds} onChange={value => update({ durationSeconds: value })} /><InputNumber min={values.minDurationSeconds} max={values.maxDurationSeconds} step={0.1} precision={2} value={values.durationSeconds} suffix="s" onChange={value => value != null && update({ durationSeconds: value })} /></div>
        </div>
      </div>
    </aside>
  )
}
