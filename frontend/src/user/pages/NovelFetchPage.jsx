import { useEffect, useRef } from 'react';
import { NovelFetchWorkshopPage } from './NovelFetchWorkshopPage';

// Keep the public V78 workbench as the visual baseline while React parity is
// rebuilt. It is served by the Go embed bundle, not a Node runtime.
export function NovelFetchPage({ theme = 'dark' }) {
  const frameRef = useRef(null);
  const effectiveTheme = theme === 'light' ? 'light' : 'dark';

  useEffect(() => {
    frameRef.current?.contentWindow?.postMessage(
      { type: 'qiantie-theme-sync', theme: effectiveTheme },
      window.location.origin,
    );
  }, [effectiveTheme]);

  return (
    <iframe
      title="批量原文改文系统"
      ref={frameRef}
      src={`/batch-rewrite/v2?theme=${effectiveTheme}`}
      onLoad={event => event.currentTarget.contentWindow?.postMessage(
        { type: 'qiantie-theme-sync', theme: effectiveTheme },
        window.location.origin,
      )}
      style={{ border: 0, display: 'block', height: '100%', width: '100%' }}
    />
  );
}

export { NovelFetchWorkshopPage };
export default NovelFetchPage;
