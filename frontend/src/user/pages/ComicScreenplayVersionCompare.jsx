import { Button, Empty, Modal, Popconfirm, Spin, Tag, message } from 'antd';
import { Check, Trash2 } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { deleteComicScreenplayDocument, getComicScreenplayDocument, restoreComicScreenplayDocument } from '../../shared/api/comicScreenplay.js';

function countChars(value) {
  return [...String(value || '')].length;
}

function prettyStamp(value) {
  return String(value || '').replace('T', ' ').slice(0, 19);
}

// 版本择优（洛水 wsvp-modal 对齐）：同一剧本的多个版本横排对照，
// 「选用此版」把所选历史版本恢复为新的当前版本；「删除该版」清掉历史版本。
// 差异：洛水按集分组（剧本=每集一份）；V88 剧本是整个项目一份文档，按版本横排对照。
export default function ComicScreenplayVersionCompare({ projectId, documents = [], readOnly = false, open = false, onClose = () => {}, onChanged = async () => {} }) {
  const [contents, setContents] = useState({});
  const [loading, setLoading] = useState(false);
  const [adopting, setAdopting] = useState('');
  const [deleting, setDeleting] = useState('');
  const requestEpoch = useRef(0);

  useEffect(() => {
    if (!open) return undefined;
    const epoch = ++requestEpoch.current;
    const missing = documents.filter(document => contents[document.id] === undefined);
    if (!documents.length || !missing.length) return undefined;
    setLoading(true);
    (async () => {
      const next = {};
      await Promise.all(missing.map(async document => {
        try {
          const full = await getComicScreenplayDocument(projectId, document.id);
          next[document.id] = full?.content || '';
        } catch (error) {
          if (epoch === requestEpoch.current) message.error(error.message || `版本 v${document.version} 内容读取失败`);
          next[document.id] = '';
        }
      }));
      if (epoch === requestEpoch.current) {
        setContents(current => ({ ...current, ...next }));
        setLoading(false);
      }
    })();
    return () => { requestEpoch.current += 1; };
  }, [open, documents]); // eslint-disable-line react-hooks/exhaustive-deps

  async function adopt(document) {
    if (readOnly || !document) return;
    try {
      setAdopting(document.id);
      await restoreComicScreenplayDocument(projectId, document.id);
      await onChanged();
      message.success(`已选用：版本 v${document.version}`);
      onClose();
    } catch (error) {
      message.error(error.message || '选用失败');
    } finally {
      setAdopting('');
    }
  }

  async function remove(document) {
    if (readOnly || !document) return;
    try {
      setDeleting(document.id);
      await deleteComicScreenplayDocument(projectId, document.id);
      await onChanged();
      message.success(`已删除：版本 v${document.version}`);
    } catch (error) {
      message.error(error.message || '删除失败');
    } finally {
      setDeleting('');
    }
  }

  return <Modal className="wsvp-modal" width="min(1100px, 96vw)" title={<span className="wsvp-title">版本择优<span className="wsvp-sub">按版本横排对照，选用更好的一稿</span></span>} open={open} onCancel={onClose} footer={<Button onClick={onClose}>关闭</Button>}>
    {documents.length ? <Spin spinning={loading}>
      <div className="wsvp-columns">
        {documents.map((document, index) => {
          const preferred = document.status === 'active';
          const content = contents[document.id];
          return <article key={document.id} className={`wsvp-col${preferred ? ' preferred' : ''}`}>
            <div className="wsvp-col-head">
              <div className="wsvp-col-title" title={`版本 v${document.version}`}>
                {preferred ? <Tag color="success" bordered={false}>当前版本</Tag> : null}
                <span>{`版本 v${document.version}`}</span>
              </div>
              <div className="wsvp-col-meta">{countChars(content)}字 · {prettyStamp(document.updated_at || document.created_at)}</div>
            </div>
            <div className="wsvp-col-body">{content === undefined ? '（加载中…）' : content || '（空）'}</div>
            <div className="wsvp-col-actions">
              <Button type="primary" size="small" block loading={adopting === document.id} disabled={readOnly || preferred} icon={<Check size={13} />} onClick={() => adopt(document)}>{preferred ? '正在使用' : '选用此版'}</Button>
              <Popconfirm title={`删除版本 v${document.version}？`} description="该版本及其场次、人物、场景会一并删除，不可恢复。" onConfirm={() => remove(document)}>
                <Button size="small" block danger loading={deleting === document.id} disabled={readOnly || preferred} title={preferred ? '当前版本不可删除，请先选用其他版本' : '彻底删除这一版剧本，不可恢复'} icon={<Trash2 size={13} />}>删除该版</Button>
              </Popconfirm>
            </div>
          </article>;
        })}
      </div>
    </Spin> : <Empty description="还没有可对照的剧本版本，先 AI 生成或导入后再来" />}
  </Modal>;
}
