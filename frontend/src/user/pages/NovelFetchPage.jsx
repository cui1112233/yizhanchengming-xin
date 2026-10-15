import { NovelFetchWorkshopPage } from './NovelFetchWorkshopPage';

// Keep the public V78 workbench as the visual baseline while React parity is
// rebuilt. It is served by the Go embed bundle, not a Node runtime.
export function NovelFetchPage() {
  return (
    <iframe
      title="批量原文改文系统"
      src="/batch-rewrite/v2?theme=dark"
      style={{ border: 0, display: 'block', height: '100%', width: '100%' }}
    />
  );
}

export { NovelFetchWorkshopPage };
export default NovelFetchPage;
