import { useEffect, useState } from 'react';
import { createAuthenticatedMediaSession } from '../pages/agent-canvas-media.js';
import { AgentAudioWaveform } from './AgentAudioWaveform';

export function AuthenticatedAgentAudio({ source, className = '', waveform = false, waveformSourceLabel = '', waveformDownstreamLabel = '', ...audioProps }) {
  const [state, setState] = useState({ url: '', loading: Boolean(source), error: '', reportPlaybackError: null });

  useEffect(() => {
    let session;
    const reportPlaybackError = () => session?.reportPlaybackError('音频源无法播放');
    session = createAuthenticatedMediaSession(source, {
      loadErrorMessage: '音频读取失败',
      onState: mediaState => setState({ ...mediaState, reportPlaybackError })
    });
    return () => session.dispose();
  }, [source]);

  const stateClassName = `${className} agent-canvas-auth-media-state`.trim();
  if (state.loading) return <span className={stateClassName} role="status">正在读取音频…</span>;
  if (state.error) return <span className={`${stateClassName} is-error`} role="alert">音频读取失败：{state.error}</span>;
  if (!state.url) return null;
  return <>
    <audio {...audioProps} className={className} src={state.url} onError={event => {
      state.reportPlaybackError?.();
      audioProps.onError?.(event);
    }} />
    {waveform ? <AgentAudioWaveform source={source} sourceLabel={waveformSourceLabel} downstreamLabel={waveformDownstreamLabel} /> : null}
  </>;
}
