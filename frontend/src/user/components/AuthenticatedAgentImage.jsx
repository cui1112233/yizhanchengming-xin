import { useEffect, useState } from 'react';
import { createAuthenticatedMediaSession } from '../pages/agent-canvas-media.js';

export function AuthenticatedAgentImage({ source, className = '', alt = '', ...imageProps }) {
  const [state, setState] = useState({ url: '', loading: Boolean(source), error: '', reportPlaybackError: null });

  useEffect(() => {
    let session;
    const reportPlaybackError = () => session?.reportPlaybackError('图片源无法显示');
    session = createAuthenticatedMediaSession(source, {
      loadErrorMessage: '图片读取失败',
      onState: mediaState => setState({ ...mediaState, reportPlaybackError })
    });
    return () => session.dispose();
  }, [source]);

  const stateClassName = `${className} agent-canvas-auth-media-state`.trim();
  if (state.loading) return <span className={stateClassName} role="status">正在读取图片…</span>;
  if (state.error) return <span className={`${stateClassName} is-error`} role="alert">图片读取失败：{state.error}</span>;
  if (!state.url) return null;
  return <img {...imageProps} className={className} src={state.url} alt={alt} onError={event => {
    state.reportPlaybackError?.();
    imageProps.onError?.(event);
  }} />;
}
