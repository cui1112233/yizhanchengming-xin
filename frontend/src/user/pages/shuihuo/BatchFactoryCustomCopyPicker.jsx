import { Alert, Button, Modal, Radio, Space, Tag } from 'antd';
import { useEffect, useRef, useState } from 'react';
import { apiRequest, getToken } from '../../../shared/api/client.js';
import { apiErrorMessage } from '../../../shared/apiErrors.js';
import { copyInputIDs } from './batchFactoryCustomCopy.js';
import { formatWorkDateTime } from '../../../shared/workDateTime.js';

export function BatchFactoryCustomCopyPicker({ inputText, onAdd, disabled = false }) {
  const [results, setResults] = useState([]);
  const [selections, setSelections] = useState({});
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [preview, setPreview] = useState(null);
  const generation = useRef(0);
  const resultSession = useRef('');
  useEffect(() => {
    const changed = () => {
      if (resultSession.current && resultSession.current !== getToken()) {
        generation.current += 1; setResults([]); setSelections({}); setPreview(null); setBusy(false);
        setError('登录账号已变化，请重新获取文案');
      }
    };
    window.addEventListener('storage', changed);
    return () => window.removeEventListener('storage', changed);
  }, []);
  useEffect(() => {
    generation.current += 1;
    setResults([]); setSelections({}); setError(''); setBusy(false); setPreview(null);
    return () => { generation.current += 1; };
  }, [inputText]);
  async function fetchCopies() {
    let ids;
    try { ids = copyInputIDs(inputText); } catch (err) { setError(err.message); return; }
    const request = ++generation.current;
    const session = getToken();
    resultSession.current = session;
    setBusy(true); setResults([]); setSelections({}); setError('');
    const next = [];
    // Sequential reads avoid flooding the upstream account with dozens of requests.
    for (const bookId of ids) {
      if (request !== generation.current || session !== getToken()) { setBusy(false); return; }
      try {
        const response = await apiRequest(`/api/batch-factory/v12/custom-copies/${encodeURIComponent(bookId)}`);
        next.push({ bookId, copies: response.copies || [] });
      } catch (err) {
        next.push({ bookId, copies: [], error: apiErrorMessage(err, '文案获取失败') });
      }
      if (request === generation.current && session === getToken()) setResults([...next]);
    }
    if (request === generation.current) setBusy(false);
  }
  const ready = results.length > 0 && results.every(result => !result.error && result.copies.some(copy => copy.id === selections[result.bookId]));
  return <div style={{ margin: '12px 0' }} data-testid="custom-copy-picker">
    <p style={{ margin: '0 0 10px', opacity: .8 }}>读取视频管理后台中已有的文案。每本书选择一份，完整正文将直接用于生产。</p>
    <Space><Button onClick={fetchCopies} loading={busy} disabled={disabled || busy || !inputText.trim()}>获取文案</Button>
      <Button type="primary" disabled={disabled || busy || !ready} onClick={() => {
        if (resultSession.current !== getToken()) { setResults([]); setSelections({}); setError('登录账号已变化，请重新获取文案'); return; }
        onAdd(results, selections);
      }}>使用所选文案 · 加入批量</Button></Space>
    {error ? <Alert type="error" showIcon message={error} style={{ marginTop: 10 }} /> : null}
    {results.map(result => <section key={result.bookId} style={{ marginTop: 12, padding: 14, border: '1px solid var(--border-color, rgba(140,150,170,.28))', borderRadius: 12 }}>
      <div style={{ marginBottom: 10, fontWeight: 600 }}>书籍 ID {result.bookId} <Tag>{result.copies.length} 份文案</Tag></div>
      {result.error || !result.copies.length ? <Alert type="warning" showIcon message={result.error || '没有找到该书的文案，请核对书籍 ID'} /> :
        <Radio.Group value={selections[result.bookId]} onChange={event => setSelections(current => ({ ...current, [result.bookId]: event.target.value }))} style={{ width: '100%' }}>
          <Space direction="vertical" style={{ width: '100%' }} size={10}>{result.copies.map((copy, index) => <div key={copy.id} style={{ padding: 12, borderRadius: 10, background: 'rgba(128,140,160,.08)', border: '1px solid rgba(128,140,160,.2)' }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 12 }}>
              <Radio value={copy.id}>{copy.title || result.bookId} · 文案 {index + 1}</Radio>
              <Button size="small" onClick={() => setPreview(copy)}>查看全文</Button>
            </div>
            <div style={{ fontSize: 12, opacity: .75, marginTop: 6 }}>记录 {copy.id} · {Array.from((copy.content || '').replace(/\s/g, '')).length.toLocaleString()} 字{copy.createdAt ? ` · ${formatWorkDateTime(copy.createdAt)}` : ''}</div>
            <div style={{ marginTop: 8, whiteSpace: 'pre-wrap', maxHeight: 72, overflow: 'hidden', lineHeight: 1.7 }}>{copy.content?.trim()}</div>
          </div>)}</Space>
        </Radio.Group>}
    </section>)}
    <Modal title={`${preview?.title || '文案'} · 记录 ${preview?.id || ''}`} open={Boolean(preview)} onCancel={() => setPreview(null)} footer={<Button type="primary" onClick={() => { setSelections(current => ({ ...current, [preview.bookId]: preview.id })); setPreview(null); }}>选择这份文案</Button>} width={760}>
      <div style={{ maxHeight: '60vh', overflowY: 'auto', whiteSpace: 'pre-wrap', lineHeight: 1.9, padding: '12px 4px' }}>{preview?.content?.trim()}</div>
    </Modal>
  </div>;
}
