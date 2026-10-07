import { cleanup } from '@testing-library/react'
import { message, notification, Modal } from 'antd'
import { afterEach } from 'vitest'

if (typeof window !== 'undefined' && !window.matchMedia) {
  Object.defineProperty(window, 'matchMedia', {
    configurable: true,
    writable: true,
    value: (query) => ({
      matches: false,
      media: query,
      onchange: null,
      addListener: () => {},
      removeListener: () => {},
      addEventListener: () => {},
      removeEventListener: () => {},
      dispatchEvent: () => false,
    }),
  })
}

// jsdom does not support pseudo-element styles and logs a noisy implementation
// error whenever Ant Design probes them. The browser-facing value is irrelevant
// to these component tests, so keep the native element style calculation while
// deliberately ignoring the unsupported pseudo-element argument.
if (typeof window !== 'undefined' && window.getComputedStyle) {
  const getElementStyle = window.getComputedStyle.bind(window)
  window.getComputedStyle = (element) => getElementStyle(element)
}

afterEach(() => {
  if (typeof document === 'undefined') return
  cleanup()
  message.destroy()
  notification.destroy()
  Modal.destroyAll()
})
