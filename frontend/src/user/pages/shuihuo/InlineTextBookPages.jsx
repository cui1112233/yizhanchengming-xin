import { LeftOutlined, PictureOutlined, RightOutlined } from '@ant-design/icons';
import { useEffect, useState } from 'react';
import { listTextPages } from '../../../shared/api/batchFactoryText.js';
import AuthenticatedTextPageImage from './AuthenticatedTextPageImage.jsx';

function PagePager({ selectedIndex, total, onMove, kind }) {
  const image = kind === 'image';
  return <div className="batch-factory-cell-pager">
    <button type="button" aria-label={image ? '上一文字页' : '上一页正文'} disabled={selectedIndex <= 0} onClick={() => onMove(selectedIndex - 1)}><LeftOutlined /></button>
    <span>{total ? `第 ${selectedIndex + 1} 页` : '尚未分页'}</span>
    <button type="button" aria-label={image ? '下一文字页' : '下一页正文'} disabled={!total || selectedIndex >= total - 1} onClick={() => onMove(selectedIndex + 1)}><RightOutlined /></button>
  </div>;
}

export default function InlineTextBookPages({ batchId, book, pageTarget = 150, automationActive = false, onOpen }) {
  const [snapshot, setSnapshot] = useState({ pages: [], artifacts: [] });
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    const controller = new AbortController();
    let timer;
    async function read() {
      try {
        const result = await listTextPages(batchId, book.id, { signal: controller.signal, suppressGlobalError: true });
        if (controller.signal.aborted) return;
        setSnapshot({ pages: Array.isArray(result?.pages) ? result.pages : [], artifacts: Array.isArray(result?.artifacts) ? result.artifacts : [] });
        setError('');
      } catch (readError) {
        if (!controller.signal.aborted) setError(readError?.message || '文字页读取失败');
      } finally {
        if (!controller.signal.aborted) {
          setLoading(false);
          if (automationActive) timer = setTimeout(read, 2000);
        }
      }
    }
    setLoading(true);
    void read();
    return () => { controller.abort(); clearTimeout(timer); };
  }, [batchId, book.id, automationActive]);

  const pages = snapshot.pages;
  const artifacts = snapshot.artifacts;
  useEffect(() => {
    setSelectedIndex(current => Math.max(0, Math.min(current, Math.max(0, pages.length - 1))));
  }, [pages.length]);

  const move = index => setSelectedIndex(Math.max(0, Math.min(index, Math.max(0, pages.length - 1))));
  const page = pages[selectedIndex];
  const activeArtifact = artifacts.find(item => item.id === page?.activeArtifactId && item.kind === 'page');
  const emptyText = loading ? '正在读取分配内容…' : error || '尚未分配页面内容';
  const imageState = loading ? '正在读取文字页…' : error || (page ? '尚未生成本页' : '尚未分配文字页');

  return <>
    <div className="shuihuo-workbench-cell batch-factory-text-page-cell is-copy">
      <div className="batch-factory-cell-shell">
        <header className="batch-factory-text-page-head"><div><b>页面内容</b><small>{page ? `第 ${page.ordinal} 页 · ${String(page.text || '').length} 字` : `单页约 ${pageTarget} 字`}</small></div><span>{pages.length ? `${selectedIndex + 1}/${pages.length}` : '未分页'}</span></header>
        <div className="batch-factory-text-page-content"><p>{page?.text || emptyText}</p></div>
        <PagePager selectedIndex={selectedIndex} total={pages.length} onMove={move} kind="copy" />
      </div>
    </div>
    <div className="shuihuo-workbench-cell shuihuo-library-cell batch-factory-book-library-cell batch-factory-text-rendered-cell">
      <div className="batch-factory-cell-shell">
        <header className="batch-factory-text-page-head"><div><b>文字页</b><small>{page ? `第 ${page.ordinal} 页` : `目标 ${pageTarget} 字/页`}</small></div><button type="button" className="batch-factory-text-page-open" disabled={!page} onClick={() => page && onOpen?.(page.id)}>查看 / 调整</button></header>
        <button type="button" className="batch-factory-text-page-image" disabled={!page} onClick={() => page && onOpen?.(page.id)}>
          {activeArtifact?.contentUrl
            ? <AuthenticatedTextPageImage passive contentUrl={activeArtifact.contentUrl} alt={`文字页 ${page.ordinal}`} />
            : <span className="batch-factory-text-page-empty"><PictureOutlined />{imageState}</span>}
        </button>
        <PagePager selectedIndex={selectedIndex} total={pages.length} onMove={move} kind="image" />
      </div>
    </div>
  </>;
}
