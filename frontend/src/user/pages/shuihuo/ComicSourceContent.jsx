import { ExpandOutlined, LeftOutlined, RightOutlined } from '@ant-design/icons';
import { Button, Modal, Select, Space, Tag, Tooltip } from 'antd';
import { useEffect, useMemo, useState } from 'react';
import { comicSourceBoundaryOptions, comicSourceSplitPayload, comicSourceStatusLabel } from './comicSourceContent.js';

export function ComicSourceContent({ segment, ordinal = 0, onSplit, embedded = false }) {
  const options = useMemo(() => comicSourceBoundaryOptions(segment), [segment]);
  const [boundaryOffset, setBoundaryOffset] = useState(options[0]?.value);
  const [expanded, setExpanded] = useState(false);
  const [busy, setBusy] = useState('');
  useEffect(() => setBoundaryOffset(options[0]?.value), [segment?.videoId, options[0]?.value]);
  if (!segment) {
    return <section className="batch-factory-comic-source is-empty"><header><b>画面内容</b><Tag>等待漫画分页</Tag></header><p>提取后，这里显示本页对应的连续小说原文。</p></section>;
  }
  const split = async direction => {
    if (!boundaryOffset || !onSplit || busy) return;
    setBusy(direction);
    try {
      await onSplit(comicSourceSplitPayload(direction, boundaryOffset, segment.revision));
    } finally {
      setBusy('');
    }
  };
  const detail = <>
    <p className="batch-factory-comic-source-text batch-factory-comic-source-full">{segment.sourceText}</p>
    <div className="batch-factory-comic-source-split"><Select size="small" value={boundaryOffset} onChange={setBoundaryOffset} options={options} placeholder={options.length ? '选择安全拆分点' : '本段没有可拆分位置'} disabled={!options.length || Boolean(busy)} /><Space size={4}><Tooltip title="向左拆分"><Button size="small" icon={<LeftOutlined />} loading={busy === 'left'} disabled={!boundaryOffset || Boolean(busy)} onClick={() => split('left')} aria-label="向左拆分" /></Tooltip><Tooltip title="向右拆分"><Button size="small" icon={<RightOutlined />} loading={busy === 'right'} disabled={!boundaryOffset || Boolean(busy)} onClick={() => split('right')} aria-label="向右拆分" /></Tooltip></Space></div>
  </>;
  if (embedded) {
    return <section className="batch-factory-comic-source is-embedded">
      <header><div><b>画面内容</b><small>第 {ordinal + 1} 段 · {segment.characterCount} 字</small></div><Tag color={segment.mediaStale ? 'warning' : segment.promptStale ? 'processing' : 'success'}>{comicSourceStatusLabel(segment)}</Tag></header>
      {detail}
    </section>;
  }
  return <section className="batch-factory-comic-source">
    <header><Button className="batch-factory-comic-source-open" type="text" size="small" icon={<ExpandOutlined />} onClick={() => setExpanded(true)} aria-label="查看画面内容">画面内容</Button><Tag color={segment.mediaStale ? 'warning' : segment.promptStale ? 'processing' : 'success'}>{comicSourceStatusLabel(segment)}</Tag></header>
    <Modal title={`画面内容 · 第 ${ordinal + 1} 段 · ${segment.characterCount} 字`} open={expanded} onCancel={() => setExpanded(false)} footer={<Button onClick={() => setExpanded(false)}>关闭</Button>} width={720}>
      {detail}
    </Modal>
  </section>;
}
