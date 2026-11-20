import { Button, Slider } from 'antd'
import { Copy, Redo2, Scissors, Trash2, Undo2, ZoomIn } from 'lucide-react'

export function Timeline() {
  return (
    <section className="editor-timeline" aria-label="时间线">
      <div className="timeline-toolbar" aria-label="工具栏">
        <Button type="text" icon={<Undo2 size={16} />}>撤销</Button>
        <Button type="text" icon={<Redo2 size={16} />}>重做</Button>
        <Button type="text" icon={<Scissors size={16} />}>分割</Button>
        <Button type="text" icon={<Copy size={16} />}>复制</Button>
        <Button type="text" danger icon={<Trash2 size={16} />}>删除</Button>
        <span className="timeline-toolbar__spacer" />
        <ZoomIn size={15} /><Slider min={1} max={10} defaultValue={4} />
      </div>
      <div className="timeline-ruler"><span>00:00</span><span>00:05</span><span>00:10</span><span>00:15</span><span>00:20</span></div>
      <div className="timeline-track"><b>视频 1</b><div><span>把素材拖到这里开始剪辑</span></div></div>
      <div className="timeline-track"><b>音频 1</b><div /></div>
    </section>
  )
}
