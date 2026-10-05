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

afterEach(() => {
  if (typeof document === 'undefined') return
  cleanup()
  message.destroy()
  notification.destroy()
  Modal.destroyAll()
})
