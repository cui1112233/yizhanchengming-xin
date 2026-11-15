import { useCallback, useEffect, useMemo, useState } from 'react';
import { Button, Empty, Popconfirm, Segmented, Spin, Tag, message } from 'antd';
import { Archive, RotateCcw, Trash2 } from 'lucide-react';
import {
  RETENTION_MODULES,
  getArchivedProjects,
  purgeArchivedProject,
  restoreArchivedProject,
  retentionModuleLabel
} from '../../shared/api/projectRetention.js';

function formatTime(value) {
  if (!value) return '-';
  const date = new Date(value);
  return Number.isFinite(date.getTime()) ? date.toLocaleString('zh-CN', { hour12: false }) : '-';
}

// 全站统一的归档库：批量工厂、漫剧创作、CM 画布、小说创作到期后都会进入这里。
export default function ArchiveLibraryPage() {
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [working, setWorking] = useState('');
  const [filter, setFilter] = useState('all');

  const load = useCallback(async () => {
    setLoading(true);
    try {
      setItems(await getArchivedProjects());
    } catch (error) {
      message.error(error.message || '归档库读取失败');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { void load(); }, [load]);

  const counts = useMemo(() => {
    const result = { all: items.length };
    items.forEach(item => { result[item.module] = (result[item.module] || 0) + 1; });
    return result;
  }, [items]);

  const options = useMemo(() => [
    { value: 'all', label: `全部 ${counts.all || 0}` },
    ...Object.entries(RETENTION_MODULES)
      .filter(([module]) => counts[module])
      .map(([module, label]) => ({ value: module, label: `${label} ${counts[module]}` }))
  ], [counts]);

  const visible = filter === 'all' ? items : items.filter(item => item.module === filter);

  async function restore(item) {
    const key = `restore:${item.module}:${item.projectId}`;
    setWorking(key);
    try {
      await restoreArchivedProject(item.module, item.projectId);
      message.success(`已恢复「${item.title || '未命名作品集'}」，保留期重新开始计时`);
      await load();
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
    } catch (error) {
      message.error(error.message || '彻底删除失败');
    } finally {
      setWorking('');
    }
  }

  return (
    <main className="archive-library-page" style={{ padding: '24px 28px', maxWidth: 1280, margin: '0 auto' }}>
      <header style={{ display: 'flex', flexWrap: 'wrap', gap: 16, alignItems: 'center', justifyContent: 'space-between', marginBottom: 18 }}>
        <div>
          <h1 style={{ display: 'flex', alignItems: 'center', gap: 10, margin: 0 }}><Archive size={22} />归档库</h1>
          <p style={{ margin: '6px 0 0', opacity: 0.72 }}>
            作品集到期后自动进入归档并永久保留。恢复后按设置中的保留时长重新计时；彻底删除才会释放对应数据。
          </p>
        </div>
        <Segmented value={filter} onChange={setFilter} options={options} />
      </header>

      <Spin spinning={loading}>
        {visible.length ? (
          <section style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))', gap: 14 }}>
            {visible.map(item => {
              const restoreKey = `restore:${item.module}:${item.projectId}`;
              const purgeKey = `purge:${item.module}:${item.projectId}`;
              return (
                <article
                  key={`${item.module}:${item.projectId}`}
                  style={{
                    border: '1px solid var(--border-color, rgba(148,163,184,0.28))',
                    borderRadius: 14,
                    padding: 16,
                    background: 'var(--surface-color, rgba(15,23,42,0.55))',
                    display: 'flex',
                    flexDirection: 'column',
                    gap: 10
                  }}
                >
                  <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 8 }}>
                    <Tag color="default">{retentionModuleLabel(item.module)}</Tag>
                    <Tag color="default">已归档</Tag>
                  </div>
                  <strong style={{ fontSize: 16, lineHeight: 1.4 }}>{item.title || '未命名作品集'}</strong>
                  <div style={{ display: 'grid', gap: 4, fontSize: 12, opacity: 0.72 }}>
                    <span>最后活动：{formatTime(item.lastActivityAt)}</span>
                    <span>归档时间：{formatTime(item.archivedAt)}</span>
                  </div>
                  <div style={{ display: 'flex', gap: 8, marginTop: 4 }}>
                    <Button
                      type="primary"
                      icon={<RotateCcw size={15} />}
                      loading={working === restoreKey}
                      disabled={Boolean(working) && working !== restoreKey}
                      onClick={() => restore(item)}
                    >
                      恢复
                    </Button>
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
                      <Button danger icon={<Trash2 size={15} />} loading={working === purgeKey} disabled={Boolean(working) && working !== purgeKey && working !== restoreKey}>
                        彻底删除
                      </Button>
                    </Popconfirm>
                  </div>
                </article>
              );
            })}
          </section>
        ) : (
          <div style={{ padding: '80px 0' }}>
            <Empty description={loading ? '正在读取归档作品' : '归档库还是空的'} />
          </div>
        )}
      </Spin>
    </main>
  );
}
