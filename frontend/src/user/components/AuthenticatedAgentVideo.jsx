import { useEffect, useState } from 'react';
import { createAuthenticatedMediaSession, nextAuthenticatedMediaRetryAttempt } from '../pages/agent-canvas-media.js';

export function AuthenticatedAgentVideo({ source, className = '', ...videoProps }) {
  const [state, setState] = useState({ url: '', loading: Boolean(source), error: '', reportPlaybackError: null });
  const [retryAttempt, setRetryAttempt] = useState(0);

  useEffect(() => {
    setRetryAttempt(0);
  }, [source]);

  useEffect(() => {
    let session;
    const reportPlaybackError = () => session?.reportPlaybackError();
    session = createAuthenticatedMediaSession(source, {
      onState: mediaState => setState({ ...mediaState, reportPlaybackError })
    });
    return () => session.dispose();
  }, [source, retryAttempt]);

  useEffect(() => {
    const nextAttempt = state.error ? nextAuthenticatedMediaRetryAttempt(source, retryAttempt) : null;
    if (nextAttempt === null) return undefined;
    const timer = window.setTimeout(() => setRetryAttempt(nextAttempt), 750);
    return () => window.clearTimeout(timer);
  }, [source, state.error, retryAttempt]);

  const stateClassName = `${className} agent-canvas-auth-media-state`.trim();
  if (state.loading) return <span className={stateClassName} role="status">正在读取视频…</span>;
  if (state.error) return <span className={`${stateClassName} is-error`} role="alert">视频读取失败：{state.error}<button type="button" onClick={() => setRetryAttempt(value => value + 1)}>重试</button></span>;
  if (!state.url) return null;
  return <video {...videoProps} className={className} src={state.url} onError={event => {
    state.reportPlaybackError?.();
    videoProps.onError?.(event);
  }} />;
}
