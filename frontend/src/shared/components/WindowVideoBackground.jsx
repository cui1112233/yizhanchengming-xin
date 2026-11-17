import { useEffect, useRef, useState } from 'react';
import { readWindowBackgroundMotion, windowVideoDefinition } from '../styles/window-background.js';
import '../styles/window-background-appearance.js';

// One active decoder, shared posters for static/reduced-motion/error states.
export function WindowVideoBackground({ backgroundId }) {
  const [index, setIndex] = useState(0);
  const [motion, setMotion] = useState(readWindowBackgroundMotion);
  const [reduced, setReduced] = useState(() => window.matchMedia?.('(prefers-reduced-motion: reduce)').matches || false);
  const [visible, setVisible] = useState(() => !document.hidden);
  const [ready, setReady] = useState(false);
  const [failed, setFailed] = useState(false);
  const videoRef = useRef(null);
  const definition = windowVideoDefinition(backgroundId, index);
  const enabled = motion !== 'static' && !reduced && !failed;

  useEffect(() => {
    const observer = new MutationObserver(() => setMotion(document.documentElement.dataset.windowBgMotion || 'soft'));
    observer.observe(document.documentElement, { attributes: true, attributeFilter: ['data-window-bg-motion'] });
    const query = window.matchMedia?.('(prefers-reduced-motion: reduce)');
    const onMotion = event => setReduced(event.matches);
    const onVisibility = () => setVisible(!document.hidden);
    query?.addEventListener('change', onMotion);
    document.addEventListener('visibilitychange', onVisibility);
    return () => { observer.disconnect(); query?.removeEventListener('change', onMotion); document.removeEventListener('visibilitychange', onVisibility); };
  }, []);
  useEffect(() => { setIndex(0); setFailed(false); }, [backgroundId]);
  useEffect(() => { setReady(false); }, [definition?.src]);
  useEffect(() => {
    const video = videoRef.current;
    if (!video) return;
    if (enabled && visible) video.play().catch(() => {});
    else video.pause();
  }, [definition?.src, enabled, visible]);

  if (!definition) return null;
  return <div className="window-video-background" aria-hidden="true">
    <div className="window-video-image">
      <img src={definition.poster} alt="" />
      {enabled && <video key={definition.src} ref={videoRef} className={ready ? 'is-ready' : ''} src={definition.src} poster={definition.poster} autoPlay={visible} muted playsInline loop={backgroundId !== 'cinema-auto'} preload="metadata" onLoadedData={() => setReady(true)} onError={() => setFailed(true)} onEnded={() => { if (backgroundId === 'cinema-auto') setIndex(value => (value + 1) % 4); }} />}
    </div>
  </div>;
}
