import { useEffect, useState } from 'react';
import { createAuthenticatedMediaSession } from '../pages/agent-canvas-media.js';

export function AuthenticatedAgentVideo({ source, className = '', ...videoProps }) {
  const [state, setState] = useState({ url: '', loading: Boolean(source), error: '', reportPlaybackError: null });

  useEffect(() => {
    let session;
    const reportPlaybackError = () => session?.reportPlaybackError();
    session = createAuthenticatedMediaSession(source, {
      onState: mediaState => setState({ ...mediaState, reportPlaybackError })
    });
    return () => session.dispose();
  }, [source]);

  const stateClassName = `${className} agent-canvas-auth-media-state`.trim();
  if (state.loading) return <span className={stateClassName} role="status">正在读取视频…</span>;
  if (state.error) return <span className={`${stateClassName} is-error`} role="alert">视频读取失败：{state.error}</span>;
  if (!state.url) return null;
  return <video {...videoProps} className={className} src={state.url} onError={event => {
    state.reportPlaybackError?.();
    videoProps.onError?.(event);
  }} />;
}
