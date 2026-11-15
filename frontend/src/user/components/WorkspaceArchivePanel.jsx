import { useCallback, useEffect, useMemo, useState } from 'react';
import { Button, Empty, Modal, Popconfirm, Spin, Tag, message } from 'antd';
import { Archive, RotateCcw, Trash2 } from 'lucide-react';
import {
  getArchivedProjects,
  purgeArchivedProject,
  restoreArchivedProject,
  retentionModuleLabel
} from '../../shared/api/projectRetention.js';

const ARCHIVE_GRACE_MS = 24 * 60 * 60 * 1000;

function formatTime(value) {
  if (!value) return '-';
  const date = new Date(value);
  return Number.isFinite(date.getTime()) ? date.toLocaleString('zh-CN', { hour12: false }) : '-';
}

function remainingLabel(archivedAt) {
  if (!archivedAt) return '等待清理';
  const left = new Date(archivedAt).getTime() + ARCHIVE_GRACE_MS - Date.now();
  if (!Number.isFinite(left)) return '等待清理';
  if (left <= 0) return '即将清理';
  const hours = Math.ceil(left / (60 * 60 * 1000));
  return `可恢复 ${hours} 小时`;
}

// 每个工作区内的归档库：到期作品先进入这里，保留 24 小时供恢复，之后自动彻底删除。
export function WorkspaceArchivePanel({ open, modules = [], onClose, onChanged }) {
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(false);
  const [working, setWorking] = useState('');

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const all = await getArchivedProjects();
      const wanted = new Set(modules);
      setItems(wanted.size ? all.filter(item => wanted.has(item.module)) : all);
    } catch (error) {
      message.error(error.message || '归档库读取失败');
    } finally {
      setLoading(false);
    }
  }, [modules]);

  useEffect(() => {
    if (open) void load();
  }, [open, load]);

  const options = useMemo(
    () => (modules.length ? modules : Array.from(new Set(items.map(item => item.module)))),
    [items, modules]
  );

  async function restore(item) {
    const key = `restore:${item.module}:${item.projectId}`;
    setWorking(key);
    try {
      await restoreArchivedProject(item.module, item.projectId);
      message.success(`已恢复「${item.title || '未命名作品集'}」，保留期重新开始计时`);
      await load();
      await onChanged?.();
    } catch (error) {
      message.error(error.message || '恢复失败');
    } finally {
      setWorking('');
    }
  }

  async function purge(item) {
    const key = `purge:${item.module}:${item.projectId}`;
    setWorking(key);
    try {
      await purgeArchivedProject(item.module, item.projectId);
      message.success(`已彻底删除「${item.title || '未命名作品集'}」`);
      await load();
      await onChanged?.();
    } catch (error) {
      message.error(error.message || '彻底删除失败');
    } finally {
      setWorking('');
    }
  }

  return (
    <Modal
      open={open}
      onCancel={onClose}
      footer={null}
      width={980}
      title={<span style={{ display: 'inline-flex', alignItems: 'center', gap: 8 }}><Archive size={18} />归档库</span>}
      destroyOnClose
    >
      <p style={{ margin: '0 0 14px', opacity: 0.7, fontSize: 13 }}>
        到期作品会先进入归档库，可在这里保留 <strong>24 小时</strong> 供恢复；超过 24 小时将自动彻底删除。恢复后按设置中的保留天数重新计时。
      </p>
      <Spin spinning={loading}>
        {items.length ? (
          <section style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: 12, maxHeight: '58vh', overflow: 'auto', paddingRight: 4 }}>
            {items.map(item => {
              const restoreKey = `restore:${item.module}:${item.projectId}`;
              const purgeKey = `purge:${item.module}:${item.projectId}`;
              return (
                <article key={`${item.module}:${item.projectId}`} style={{ border: '1px solid rgba(148,163,184,0.28)', borderRadius: 12, padding: 14, display: 'flex', flexDirection: 'column', gap: 8 }}>
                  <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 8 }}>
                    <Tag>{retentionModuleLabel(item.module)}</Tag>
                    <Tag color={remainingLabel(item.archivedAt) === '即将清理' ? 'red' : 'orange'}>{remainingLabel(item.archivedAt)}</Tag>
                  </div>
                  <strong style={{ fontSize: 15, lineHeight: 1.4 }}>{item.title || '未命名作品集'}</strong>
                  <div style={{ fontSize: 12, opacity: 0.68, display: 'grid', gap: 3 }}>
                    <span>最后活动：{formatTime(item.lastActivityAt)}</span>
                    <span>归档时间：{formatTime(item.archivedAt)}</span>
                  </div>
                  <div style={{ display: 'flex', gap: 8, marginTop: 2 }}>
                    <Button type="primary" size="small" icon={<RotateCcw size={14} />} loading={working === restoreKey} onClick={() => restore(item)}>恢复</Button>
                    <Popconfirm
                      title="彻底删除这个作品集？"
                      description={item.module === 'batch_factory'
                        ? '将删除数据库记录、TOS 云文件与本地缓存，且不可恢复。'
                        : item.module === 'cm_canvas'
                          ? '将删除数据库记录与画布素材对象，且不可恢复。'
                          : '将删除数据库记录，且不可恢复。'}
                      okText="彻底删除"
                      okButtonProps={{ danger: true }}
                      cancelText="取消"
                      onConfirm={() => purge(item)}
                    >
                      <Button danger size="small" icon={<Trash2 size={14} />} loading={working === purgeKey}>彻底删除</Button>
                    </Popconfirm>
                  </div>
                </article>
              );
            })}
          </section>
        ) : (
          <div style={{ padding: '48px 0' }}><Empty description={loading ? '正在读取归档作品' : `归档库为空（${options.map(retentionModuleLabel).join('、')}）`} /></div>
        )}
      </Spin>
    </Modal>
  );
}
