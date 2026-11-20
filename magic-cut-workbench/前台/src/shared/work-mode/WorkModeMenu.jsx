import { Check, ChevronDown } from 'lucide-react'

const modes = [
  { id: 'studio', label: '一战景铭', href: 'http://127.0.0.1:4001/' },
  { id: 'agent', label: 'Agent 代理', href: 'http://127.0.0.1:4001/agent' },
  { id: 'magic-cut', label: '魔剪', href: '/', active: true },
  { id: 'duet', label: 'Duet', disabled: true },
]

export function WorkModeMenu() {
  return (
    <details className="work-mode">
      <summary>
        <span className="work-mode__mark">M</span>
        <span>魔剪</span>
        <ChevronDown size={15} />
      </summary>
      <nav className="work-mode__menu" aria-label="切换工作模式">
        {modes.map((mode) => (
          <a
            key={mode.id}
            href={mode.disabled ? undefined : mode.href}
            aria-disabled={mode.disabled || undefined}
            className={mode.active ? 'is-active' : mode.disabled ? 'is-disabled' : undefined}
          >
            <span>{mode.label}</span>
            {mode.active ? <Check size={15} /> : null}
            {mode.disabled ? <small>未实现</small> : null}
          </a>
        ))}
      </nav>
    </details>
  )
}
