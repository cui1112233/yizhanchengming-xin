import React from 'react'

const paths = {
  script: <><path d="M5 4h10l4 4v12H5z" /><path d="M15 4v5h5M8 13h8M8 17h6" /></>,
  novel: <><path d="M4 5.5A3.5 3.5 0 0 1 7.5 2H12v18H7.5A3.5 3.5 0 0 0 4 23z" /><path d="M20 5.5A3.5 3.5 0 0 0 16.5 2H12v18h4.5A3.5 3.5 0 0 1 20 23z" /></>,
  batch: <><rect x="3" y="4" width="18" height="5" rx="1.5" /><rect x="3" y="11" width="18" height="9" rx="1.5" /><path d="M8 14v3M12 14v3M16 14v3" /></>,
  shuihuo: <><path d="m7 3 10 6-10 6z" /><path d="M4 20h16" /></>,
  agent: <><path d="M12 3 4.5 7.5v9L12 21l7.5-4.5v-9z" /><circle cx="12" cy="12" r="2.5" /></>,
  tts: <><path d="M5 9v6h4l5 4V5L9 9z" /><path d="M17 9a4 4 0 0 1 0 6M19.5 6.5a8 8 0 0 1 0 11" /></>,
  arrow: <path d="M5 12h14M14 7l5 5-5 5" />,
}

export default function WorkspaceIcon({ name, className = '' }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true" focusable="false">
      {paths[name] || paths.batch}
    </svg>
  )
}
