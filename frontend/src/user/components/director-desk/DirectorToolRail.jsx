import { Images, Layers3, Ratio, UserRoundPlus, Video } from 'lucide-react';

export const DIRECTOR_TOOLS = Object.freeze([
  { id: 'scene', label: '场景对象', icon: Layers3 },
  { id: 'assets', label: '项目图片', icon: Images },
  { id: 'actors', label: '添加角色', icon: UserRoundPlus },
  { id: 'cameras', label: '添加机位', icon: Video },
  { id: 'ratio', label: '画幅比例', icon: Ratio }
]);

export default function DirectorToolRail({ activeTool, onToolChange }) {
  return <nav className="agent-director-tool-rail" aria-label="导演台工具">
    {DIRECTOR_TOOLS.map(({ id, label, icon: Icon }) => <button key={id} type="button" aria-label={label} title={label} aria-pressed={activeTool === id} onClick={() => onToolChange(id)}><Icon size={18} /></button>)}
  </nav>;
}
