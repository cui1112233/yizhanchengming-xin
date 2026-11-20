import { Check, ChevronDown } from 'lucide-react'

import { createWorkModeLinks } from './workModeLinks.js'

export function WorkModeMenu() {
  const links = createWorkModeLinks(window.location)
  const modes = [
    { id: 'studio', label: '一战晟铭', href: links.studio },
    { id: 'agent', label: 'Agent 代理', href: links.agent },
    { id: 'magic-cut', label: '魔剪', href: links.magicCut, active: true },
    { id: 'duet', label: 'Duet', disabled: true },
  ]

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
