import { useEffect, useState } from 'react';
import { createAuthenticatedMediaSession } from '../pages/agent-canvas-media.js';

export function AuthenticatedAgentVideo({ source, className = '', ...videoProps }) {
  const [state, setState] = useState({ url: '', loading: Boolean(source), error: '' });

  useEffect(() => {
    const session = createAuthenticatedMediaSession(source, { onState: setState });
    return () => session.dispose();
  }, [source]);

  const stateClassName = `${className} agent-canvas-auth-media-state`.trim();
  if (state.loading) return <span className={stateClassName} role="status">正在读取视频…</span>;
  if (state.error) return <span className={`${stateClassName} is-error`} role="alert">视频读取失败：{state.error}</span>;
  if (!state.url) return null;
  return <video {...videoProps} className={className} src={state.url} onError={event => {
    videoProps.onError?.(event);
    setState(current => ({ ...current, error: '视频源无法播放' }));
  }} />;
}
