// @vitest-environment jsdom
import React from 'react'
import { cleanup, render } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import WorkspaceIcon from './WorkspaceIcon.jsx'

afterEach(cleanup)

describe('WorkspaceIcon', () => {
  it.each(['script', 'novel', 'batch', 'shuihuo', 'agent', 'tts', 'arrow'])('renders %s as a reusable SVG rather than a character glyph', (name) => {
    const { container } = render(<WorkspaceIcon name={name} />)
    expect(container.querySelector('svg')).toBeTruthy()
    expect(container.textContent).toBe('')
  })
})
