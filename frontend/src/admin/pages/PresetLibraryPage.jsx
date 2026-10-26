import { Alert, Button, Collapse, Form, Input, Modal, Popconfirm, Segmented, Select, Space, Table, Tag, Typography, message } from 'antd';
import { useEffect, useMemo, useState } from 'react';
import { createPresetDraft, listAdminPresetSlots, listAdminPresets, publishPreset, rollbackPreset, updatePresetDraft } from '../../shared/api/admin';

const modules = [
  { label: '剧本生成', value: 'script' },
  { label: '小说面板', value: 'novel-panel' },
  { label: '小说获取', value: 'novel-fetch' },
  { label: '水货生产', value: 'shuihuo-production' },
  { label: '批量工厂', value: 'batch-factory' }
];

const constraintCategories = [
  { label: '画面前缀词', value: 'prefix' },
  { label: '画质约束', value: 'quality' },
  { label: '画面限制', value: 'restriction' },
  { label: '负面提示词', value: 'negative' }
];

// 剧本页的提示词并非同一用途：按用户在剧本页看到的顺序分区，
// 管理员才能快速定位该改哪一项，而版本仍由每个预设词自己的展开行管理。
const scriptPresetSections = [
  {
    key: 'base-setup',
    title: '基础设定（人物 / 场景）',
    description: '对应剧本生成输入框的「切换指令 → 提取方案」，用于提取人物与场景。',
    matches: preset => preset.protocolLock?.format === 'extract' || ['script-extract', 'script-extract-novel-panel'].includes(preset.id)
  },
  {
    key: 'card-protocol',
    title: '系统级外层协议',
    description: '统一所有剧本模式的外层分镜卡片标题、边界与 10s/15s 时长规则；模式预设只维护卡内格式。',
    matches: preset => preset.id === 'script-card-protocol'
  },
  {
    key: 'constraints',
    title: '约束设置',
    description: '按分镜写入顺序管理：画面前缀词 → 画质约束 / 画面限制 → 分镜正文 → 负面提示词。',
    matches: preset => preset.id === 'script-constraint-wrapper' || preset.protocolLock?.format === 'constraint'
  },
  {
    key: 'formats',
    title: '剧本输出格式',
    description: '控制生成剧本、分镜或短剧文本的输出结构。',
    matches: preset => preset.id.startsWith('script-format-')
  },
  {
    key: 'other',
    title: '其他剧本提示词',
    description: '未归入以上生成环节的剧本模块预设词。',
    matches: () => true
  }
];

const comicSlots = new Set([
  'batch.comic-assets',
  'batch.comic.assets.character',
  'batch.comic.assets.scene',
  'batch.comic.assets.prop',
  'batch.comic.character-image-template',
  'batch.comic-page',
  'batch.comic-image-derivation',
  'batch.comic-style'
]);

const comicAssetSlots = new Set([
  'batch.comic-assets',
  'batch.comic.assets.character',
  'batch.comic.assets.scene',
  'batch.comic.assets.prop',
  'batch.comic.character-image-template'
]);

const comicConstraintCategories = new Set([
  'comic_quality',
  'comic_restriction',
  'comic_negative'
]);

// 批量工厂页与剧本生成页一样按 Collapse 分区展示：资产生成 → 剧情改编 → 视频提示词 → 漫画生成。
// 顺序固定，便于管理员定位；版本管理与归属列保持不变。
const batchFactoryPresetSections = [
  {
    key: 'batch-assets',
    title: '资产生成',
    description: '人物、场景、道具资产提取与一致性提示词。',
    matches: preset => [
      'batch-assets-h3',
      'batch-assets-h3-skill',
      'batch-assets-sd-character-scene',
      'batch-character-h3',
      'batch-character-meta',
      'batch-scene-h3',
      'batch-scene-meta',
      'batch-prop-meta'
    ].includes(preset.id)
  },
  {
    key: 'batch-adaptation',
    title: '剧情改编',
    description: '开头改编与导演拆分规则。',
    matches: preset => [
      'batch-hook-adaptation',
      'batch-original-director',
      'batch-viral-director',
      'batch-opening-meta'
    ].includes(preset.id)
  },
  {
    key: 'batch-video',
    title: '视频提示词',
    description: '视频分镜、画面提示词与视频风格前缀。',
    matches: preset => [
      'batch-video-meta',
      'batch-video-h3-director',
      'batch-video-sd-fixed-shot',
      'batch-video-sd',
      'batch-visual-meta'
    ].includes(preset.id) || preset.id.startsWith('batch-prefix-')
  },
  {
    key: 'batch-comic',
    title: '漫画生成',
    description: '集中管理漫画资产、分镜页面、图片描述词、画风库与漫画专用约束。',
    matches: preset => comicSlots.has(preset.protocolLock?.slot)
      || comicConstraintCategories.has(preset.protocolLock?.category),
    subSections: [
      { title: '资产推导', filter: preset => comicAssetSlots.has(preset.protocolLock?.slot) },
      { title: '漫画分镜与页面', filter: preset => preset.protocolLock?.slot === 'batch.comic-page' },
      { title: '图片描述词', filter: preset => preset.protocolLock?.slot === 'batch.comic-image-derivation' },
      { title: '画风库', filter: preset => preset.protocolLock?.slot === 'batch.comic-style' },
      { title: '漫画约束', filter: preset => comicConstraintCategories.has(preset.protocolLock?.category) }
    ]
  },
  {
    key: 'batch-constraints',
    title: '约束设置',
    description: '基础设定、约束规则、画面前缀词、画质约束、画面限制与负面提示词。',
    matches: preset => preset.id === 'script-constraint-wrapper'
      || (preset.protocolLock?.format === 'constraint' && !comicConstraintCategories.has(preset.protocolLock?.category))
      || [
        'script-extract',
        'script-extract-assets',
        'script-extract-novel-panel',
        'script-card-protocol'
      ].includes(preset.id),
    subSections: [
      { title: '基础设定', ids: ['script-card-protocol', 'script-extract', 'script-extract-assets', 'script-extract-novel-panel'] },
      { title: '约束规则', ids: ['script-constraint-wrapper'] },
      { title: '画面前缀词', filter: preset => preset.protocolLock?.format === 'constraint' && preset.protocolLock?.category === 'prefix' },
      { title: '画质约束', filter: preset => preset.protocolLock?.format === 'constraint' && preset.protocolLock?.category === 'quality' },
      { title: '画面限制', filter: preset => preset.protocolLock?.format === 'constraint' && preset.protocolLock?.category === 'restriction' },
      { title: '负面提示词', filter: preset => preset.protocolLock?.format === 'constraint' && preset.protocolLock?.category === 'negative' }
    ]
  }
];

const formatPriority = [
  'script-card-protocol',
  'script-extract',
  'script-extract-assets',
  'script-extract-novel-panel',
  'script-constraint-wrapper',
  'script-constraint-prefix-smart-unified',
  'script-constraint-prefix-live-action',
  'script-constraint-prefix-3d',
  'script-constraint-prefix-2d',
  'script-constraint-prefix-guoman',
  'script-constraint-quality-4k',
  'script-constraint-quality-h3-shot-narrative',
  'script-constraint-restriction-h3-visual-policy',
  'script-constraint-restriction-no-overlay',
  'script-constraint-negative-general',
  'shuihuo-extract-characters',
  'shuihuo-extract-scenes',
  'shuihuo-smart-segmentation',
  'shuihuo-image-prompt',
  'shuihuo-video-prompt',
  'script-format-q版',
  'script-format-shotlist',
  'script-format-storyboard',
  'script-format-shortdrama',
  'script-format-screenplay'
];

function sortPresets(items) {
  return [...items].sort((left, right) => {
    const leftPriority = formatPriority.indexOf(left.id);
    const rightPriority = formatPriority.indexOf(right.id);
    const leftRank = leftPriority === -1 ? formatPriority.length : leftPriority;
    const rightRank = rightPriority === -1 ? formatPriority.length : rightPriority;
    return leftRank - rightRank || left.name.localeCompare(right.name, 'zh-CN');
  });
}

function errorMessage(error) {
  try {
    return JSON.parse(error.message).error || error.message;
  } catch (_) {
    return error.message || '操作失败';
  }
}

function emptyDraft(module) {
  return {
    id: '',
    module,
    name: '',
    kind: 'base',
    description: '',
    compatibleBaseIds: [],
    body: '',
    protocolLock: {}
  };
}

export function PresetLibraryPage() {
  const [module, setModule] = useState('script');
  const [presets, setPresets] = useState([]);
  const [slots, setSlots] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [editorOpen, setEditorOpen] = useState(false);
  const [editingExisting, setEditingExisting] = useState(false);
  const [editingPreset, setEditingPreset] = useState(null);
  const [saving, setSaving] = useState(false);
  const [form] = Form.useForm();

  async function load() {
    setLoading(true);
    setError('');
    try {
      const requests = [
        listAdminPresets(module),
        listAdminPresetSlots(module)
      ];
      // 批量工厂页需要同时展示约束设置预设（module=script）
      if (module === 'batch-factory') {
        requests.push(listAdminPresets('script'));
      }
      const [presetResult, slotResult, extraPresetResult] = await Promise.all(requests);
      const allPresets = [...(presetResult.presets || [])];
      if (extraPresetResult) {
        allPresets.push(...(extraPresetResult.presets || []));
      }
      setPresets(sortPresets(allPresets));
      setSlots(slotResult.slots || []);
    } catch (requestError) {
      setPresets([]);
      setSlots([]);
      setError(errorMessage(requestError));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => { load(); }, [module]);

  function openCreate() {
    form.setFieldsValue(emptyDraft(module));
    setEditingExisting(false);
    setEditingPreset(null);
    setEditorOpen(true);
  }

  function openEdit(preset) {
    form.setFieldsValue({
      ...preset,
      slot: preset.protocolLock?.slot || undefined,
      constraintCategory: preset.protocolLock?.format === 'constraint' ? preset.protocolLock.category : '',
      protocolLock: JSON.stringify(preset.protocolLock || {}, null, 2)
    });
    setEditingExisting(true);
    setEditingPreset(preset);
    setEditorOpen(true);
  }

  async function saveDraft(values) {
    let protocolLock;
    try {
      protocolLock = JSON.parse(values.protocolLock || '{}');
    } catch (_) {
      form.setFields([{ name: 'protocolLock', errors: ['协议锁必须是合法 JSON'] }]);
      return;
    }
    if (values.constraintCategory) {
      const prefix = `script-constraint-${values.constraintCategory}-`;
      if (module !== 'script' || values.kind !== 'addon' || !values.id.startsWith(prefix)) {
        form.setFields([{ name: 'id', errors: [`${values.constraintCategory} 类约束预设 ID 必须以 ${prefix} 开头`] }]);
        return;
      }
      protocolLock = { ...protocolLock, format: 'constraint', category: values.constraintCategory };
    }
    setSaving(true);
    try {
      const input = {
        ...values,
        module,
        compatibleBaseIds: [],
        protocolLock: { ...protocolLock, slot: values.slot }
      };
      if (editingPreset?.status === 'draft') {
        await updatePresetDraft(editingPreset.id, editingPreset.version, {
          ...input,
          expectedRevision: editingPreset.revision || 1
        });
        message.success(`草稿 V${editingPreset.version} 已更新，发布后才会影响用户生成。`);
      } else {
        await createPresetDraft(input);
        message.success('已另存为新草稿，发布后才会影响用户生成。');
      }
      setEditorOpen(false);
      setEditingPreset(null);
      load();
    } catch (requestError) {
      message.error(errorMessage(requestError));
    } finally {
      setSaving(false);
    }
  }

  async function publish(preset) {
    try {
      await publishPreset(preset.id, preset.version);
      message.success('已发布，后续生成请求将立即使用此版本。');
      load();
    } catch (requestError) {
      message.error(errorMessage(requestError));
    }
  }

  async function rollback(preset) {
    try {
      await rollbackPreset(preset.id, preset.version);
      message.success('已回滚，后续生成请求将立即使用该历史版本。');
      load();
    } catch (requestError) {
      message.error(errorMessage(requestError));
    }
  }

  // 同一预设 ID 的多个版本折叠为一行，展开后仍可逐版本发布或回滚。
  const grouped = useMemo(() => {
    const map = new Map();
    presets.forEach(preset => {
      if (!map.has(preset.id)) map.set(preset.id, []);
      map.get(preset.id).push(preset);
    });
    return [...map.entries()].map(([id, versions]) => {
      const sorted = [...versions].sort((a, b) => b.version - a.version);
      const published = versions.find(version => version.status === 'published');
      const current = published || sorted[0];
      return {
        id,
        name: current.name,
        description: current.description,
        slot: current.protocolLock?.slot || '',
        versions: sorted,
        current
      };
    });
  }, [presets]);

  function statusTag(status) {
    return <Tag color={status === 'published' ? 'green' : status === 'draft' ? 'gold' : 'default'}>{status === 'published' ? '已发布' : status === 'draft' ? '草稿' : '已归档'}</Tag>;
  }

  const columns = [
    { title: '名称', dataIndex: 'name', width: 170 },
    { title: '预设词 ID', dataIndex: 'id', width: 190, ellipsis: true },
    { title: '说明', dataIndex: 'description', ellipsis: true },
    {
      title: '归属', dataIndex: 'slot', width: 150,
      render: slot => slots.find(item => item.id === slot)?.label || '待设置归属'
    },
    { title: '当前版本', width: 100, render: (_, row) => row.current.version },
    {
      title: '状态', width: 94,
      render: (_, row) => statusTag(row.current.status)
    },
    {
      title: '版本数', width: 84,
      render: (_, row) => row.versions.length
    },
    {
      title: '操作', key: 'actions', width: 150,
      render: (_, row) => (
        <Button size="small" onClick={() => openEdit(row.current)}>{row.current.status === 'draft' ? '继续编辑草稿' : '新建草稿'}</Button>
      )
    }
  ];

  const versionColumns = [
    { title: '版本', dataIndex: 'version', width: 76 },
    { title: '状态', dataIndex: 'status', width: 100, render: statusTag },
    { title: '更新时间', dataIndex: 'publishedAt', width: 180, render: value => value ? new Date(value).toLocaleString() : '' },
    {
      title: '操作', key: 'actions', width: 260,
      render: (_, preset) => (
        <Space size="small">
          <Button size="small" onClick={() => openEdit(preset)}>{preset.status === 'draft' ? '编辑草稿' : '另存为新版本'}</Button>
          {preset.status === 'draft' && <Popconfirm title="发布后将立即影响后续生成请求" onConfirm={() => publish(preset)}><Button size="small" type="primary">发布</Button></Popconfirm>}
          {preset.status === 'archived' && <Popconfirm title="确认将此历史版本设为当前线上版本" onConfirm={() => rollback(preset)}><Button size="small">回滚</Button></Popconfirm>}
        </Space>
      )
    }
  ];

  function presetTable(items) {
    return (
      <Table
        rowKey={row => row.id}
        dataSource={items}
        columns={columns}
        loading={loading}
        size="middle"
        scroll={{ x: 'max-content' }}
        expandable={{
          rowExpandable: row => row.versions.length > 1,
          expandedRowRender: row => (
            <Table
              rowKey={version => `${row.id}-${version.version}`}
              size="small"
              pagination={false}
              dataSource={row.versions}
              columns={versionColumns}
            />
          )
        }}
        pagination={false}
      />
    );
  }

  const scriptSections = useMemo(() => {
    const remaining = new Set(grouped.map(item => item.id));
    return scriptPresetSections.map(section => {
      const items = grouped.filter(item => remaining.has(item.id) && section.matches(item.current));
      items.forEach(item => remaining.delete(item.id));
      return { ...section, items };
    }).filter(section => section.items.length > 0);
  }, [grouped]);

  const batchFactorySections = useMemo(() => {
    const remaining = new Set(grouped.map(item => item.id));
    return batchFactoryPresetSections.map(section => {
      const items = grouped.filter(item => remaining.has(item.id) && section.matches(item.current));
      items.forEach(item => remaining.delete(item.id));
      return { ...section, items };
    }).filter(section => section.items.length > 0);
  }, [grouped]);

  return (
    <section className="admin-preset-library">
      <Space direction="vertical" size="middle" style={{ width: '100%' }}>
        <div className="admin-page-heading">
          <div>
            <Typography.Title level={3}>系统预设词</Typography.Title>
            <Typography.Paragraph>保存生成草稿，发布后自动由后端用于后续请求。普通用户无法读取正文。</Typography.Paragraph>
          </div>
          <Button type="primary" onClick={openCreate}>添加预设词</Button>
        </div>
        <Segmented options={modules} value={module} onChange={setModule} />
        {error && <Alert type="error" showIcon message="无法读取此模块预设词" description={error} />}
        {module === 'script' ? (
          <Collapse
            defaultActiveKey={scriptSections.map(section => section.key)}
            items={scriptSections.map(section => ({
              key: section.key,
              label: <Space direction="vertical" size={0}><Typography.Text strong>{section.title}</Typography.Text><Typography.Text type="secondary">{section.description}</Typography.Text></Space>,
              children: presetTable(section.items)
            }))}
          />
        ) : module === 'batch-factory' ? (
          <Collapse
            defaultActiveKey={batchFactorySections.map(section => section.key)}
            items={batchFactorySections.map(section => {
              let children;
              if (section.subSections) {
                children = section.subSections.map(sub => {
                  let subItems;
                  if (sub.ids) {
                    subItems = section.items.filter(item => sub.ids.includes(item.id));
                  } else if (sub.filter) {
                    subItems = section.items.filter(item => sub.filter(item.current));
                  }
                  return subItems && subItems.length > 0 ? (
                    <div key={sub.title} style={{ marginBottom: 16 }}>
                      <Typography.Title level={5} style={{ marginTop: 8, marginBottom: 8 }}>{sub.title}</Typography.Title>
                      {presetTable(subItems)}
                    </div>
                  ) : null;
                }).filter(Boolean);
              } else {
                children = presetTable(section.items);
              }
              return {
                key: section.key,
                label: <Space direction="vertical" size={0}><Typography.Text strong>{section.title}</Typography.Text><Typography.Text type="secondary">{section.description}</Typography.Text></Space>,
                children
              };
            })}
          />
        ) : (
          <Table
            rowKey={row => row.id}
            dataSource={grouped}
            columns={columns}
            loading={loading}
            expandable={{
              rowExpandable: row => row.versions.length > 1,
              expandedRowRender: row => (
                <Table
                  rowKey={version => `${row.id}-${version.version}`}
                  size="small"
                  pagination={false}
                  dataSource={row.versions}
                  columns={versionColumns}
                />
              )
            }}
            pagination={{ pageSize: 10, showSizeChanger: false, showTotal: total => `共 ${total} 个预设词` }}
          />
        )}
      </Space>
      <Modal title={editingPreset?.status === 'draft' ? `编辑草稿 V${editingPreset.version}` : '系统预设词草稿'} open={editorOpen} onCancel={() => setEditorOpen(false)} footer={null} width={760} destroyOnClose>
        <Form form={form} layout="vertical" onFinish={saveDraft} initialValues={emptyDraft(module)}>
          <Space size="middle" style={{ width: '100%' }} align="start">
            <Form.Item label="预设词 ID" name="id" rules={[{ required: true, message: '请输入固定 ID' }, { pattern: /^[\p{L}\p{N}][\p{L}\p{N}._-]{0,63}$/u, message: '只能使用中文、字母、数字、点、下划线或短横线' }]} style={{ flex: 1 }}>
              <Input placeholder="例如 script-custom" disabled={Boolean(editingPreset)} />
            </Form.Item>
            <Form.Item label="名称" name="name" rules={[{ required: true, message: '请输入名称' }]} style={{ flex: 1 }}>
              <Input placeholder="供管理员识别" />
            </Form.Item>
          </Space>
          <Form.Item name="kind" hidden><Input /></Form.Item>
          {!editingExisting && <Alert type="info" showIcon message="选择归属后，发布的提示词会出现在该功能对应的项目下拉选择中。" />}
          {editingPreset?.status === 'draft' && <Alert type="info" showIcon message={`正在编辑草稿 V${editingPreset.version}：保存会原地更新，不会生成新版本。`} />}
          {editingPreset && editingPreset.status !== 'draft' && <Alert type="info" showIcon message={`V${editingPreset.version} 已锁定：保存会另存为新的草稿版本。`} />}
          <Form.Item label="归属" name="slot" rules={[{ required: true, message: '请选择归属' }]}>
            <Select
              placeholder="请选择提示词归属"
              options={slots.map(slot => ({ value: slot.id, label: `${slot.label}（${slot.mode === 'primary' ? '主提示词' : '补充提示词'}）` }))}
              onChange={slotId => {
                const slot = slots.find(item => item.id === slotId);
                if (slot) form.setFieldValue('kind', slot.mode === 'primary' ? 'base' : 'addon');
              }}
            />
          </Form.Item>
          {module === 'script' && <Form.Item label="约束类别（可选）" name="constraintCategory">
            <Select allowClear placeholder="普通模块补充规则" options={constraintCategories} />
          </Form.Item>}
          <Form.Item label="用途说明" name="description"><Input placeholder="说明此预设词会影响的生成环节" /></Form.Item>
          <Form.Item label="系统预设词正文" name="body" rules={[{ required: true, message: '请输入系统预设词正文' }]}>
            <Input.TextArea rows={11} spellCheck={false} />
          </Form.Item>
          <Form.Item label="协议锁 JSON" name="protocolLock" rules={[{ required: true, message: '请输入协议锁 JSON' }]}>
            <Input.TextArea rows={4} spellCheck={false} />
          </Form.Item>
          <Space>
            <Button onClick={() => setEditorOpen(false)}>取消</Button>
            <Button type="primary" htmlType="submit" loading={saving}>{editingPreset?.status === 'draft' ? `保存草稿（V${editingPreset.version}）` : '另存为新版本'}</Button>
          </Space>
        </Form>
      </Modal>
    </section>
  );
}
