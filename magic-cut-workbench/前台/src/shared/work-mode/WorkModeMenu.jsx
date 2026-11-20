import { Check, ChevronDown } from 'lucide-react'

import { createWorkModeModel } from './workModeLinks.js'

export function WorkModeMenu() {
  const model = createWorkModeModel(window.location)

  return (
    <details className="work-mode">
      <summary>
        <img className="work-mode__logo" src={model.logoSrc} alt="" />
        <strong>{model.brandLabel}</strong>
        <ChevronDown className="work-mode__chevron" size={15} />
      </summary>
      <nav className="work-mode__menu" aria-label="切换工作模式">
        {model.modes.map((mode) => (
          <a
            key={mode.id}
            href={mode.disabled ? undefined : mode.href}
            aria-disabled={mode.disabled || undefined}
            aria-current={mode.active ? 'page' : undefined}
            className={mode.active ? 'is-active' : mode.disabled ? 'is-disabled' : undefined}
            onClick={mode.disabled ? event => event.preventDefault() : undefined}
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
