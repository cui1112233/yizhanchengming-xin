import { useEffect, useMemo, useState } from 'react';
import { getAgentAudioWaveform } from '../../shared/api/agent';

function sourceIdentity(source) {
  try {
    const pathname = new URL(String(source || ''), window.location.origin).pathname;
    const match = pathname.match(/^\/api\/agent\/tasks\/([^/]+)\/assets\/([^/]+)\/content$/);
    if (!match) return null;
    return { taskId: decodeURIComponent(match[1]), assetId: decodeURIComponent(match[2]) };
  } catch {
    return null;
  }
}

function durationLabel(durationMs) {
  const totalSeconds = Math.max(0, Math.round(Number(durationMs) || 0) / 1000);
  return `${Math.floor(totalSeconds / 60)}:${String(totalSeconds % 60).padStart(2, '0')}`;
}

export function AgentAudioWaveform({ source, sourceLabel = '', downstreamLabel = '' }) {
  const identity = useMemo(() => sourceIdentity(source), [source]);
  const [state, setState] = useState({ loading: Boolean(identity), summary: null, error: '' });

  useEffect(() => {
    let active = true;
    if (!identity) {
      setState({ loading: false, summary: null, error: '' });
      return () => { active = false; };
    }
    setState({ loading: true, summary: null, error: '' });
    getAgentAudioWaveform(identity.taskId, identity.assetId)
      .then(summary => {
        if (active) setState({ loading: false, summary, error: '' });
      })
      .catch(error => {
        if (active) setState({ loading: false, summary: null, error: error?.message || '波形读取失败' });
      });
    return () => { active = false; };
  }, [identity]);

  if (state.loading) return <span className="agent-audio-waveform" role="status">正在读取波形…</span>;
  if (state.error || !state.summary) return null;
  const peaks = Array.isArray(state.summary.peaks) ? state.summary.peaks.slice(0, 96) : [];
  if (!peaks.length) return null;
  const duration = durationLabel(state.summary.durationMs);
  return <div className="agent-audio-waveform" aria-label={`音频波形，时长 ${duration}`}>
    <div className="agent-audio-waveform__bars" aria-hidden="true">
      {peaks.map((peak, index) => <i key={index} className="agent-audio-waveform__bar" style={{ height: `${Math.max(8, Math.min(100, Math.round(Number(peak) * 100)))}%` }} />)}
    </div>
    <span className="agent-audio-waveform__meta">{duration}{sourceLabel ? ` · ${sourceLabel}` : ''}</span>
    {downstreamLabel ? <span className="agent-audio-waveform__downstream">{downstreamLabel}</span> : null}
  </div>;
}
