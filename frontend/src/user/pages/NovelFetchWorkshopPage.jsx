import { Alert, Button, Checkbox, Form, Input, InputNumber, Modal, Select, Space, Switch, Table, Tabs, Tag, Typography, message } from 'antd';
import { ArrowLeft, Download, Eye, RefreshCw, RotateCcw, Save, Trash2, UploadCloud, Wand2 } from 'lucide-react';
import { useEffect, useState } from 'react';
import {
  deleteWorkshopTasks, fetchWorkshopOriginal, generateWorkshopAi, getWorkshopConfig, getWorkshopTask,
  listWorkshopTasks, processBatch, saveWorkshopConfig, testWorkshopAi
} from '../../shared/api/novelFetchWorkshop';
import { createNovelFetchIntake } from '../../shared/api/batchFactoryV11';
import { batchFactoryNovelFetchHandoffPath } from './shuihuo/batchFactoryNovelFetchHandoff';

// 输入格式（parse 模块 PARSE_MODES）
const PARSE_MODE_OPTIONS = [
  { value: 'smart', label: '智能（smart）' },
  { value: 'header', label: '表头（header）' },
  { value: 'multi_header', label: '多表头（multi_header）' },
  { value: 'fixed_full_11', label: '固定11列（fixed_full_11）' },
  { value: 'fixed_from_b', label: '固定从B列（fixed_from_b）' },
  { value: 'fixed_paid_basic', label: '固定付费基础（fixed_paid_basic）' },
  { value: 'custom', label: '自定义列顺序（custom）' }
];

// 列顺序预设（parse 模块 COLUMN_PRESETS）
const COLUMN_PRESET_OPTIONS = [
  { value: 'paid_name_reason', label: '付费ID/书名/推荐理由' },
  { value: 'paid_name_gender_reason', label: '付费ID/书名/男女频/推荐理由' },
  { value: 'free_paid_name_gender_reason', label: '免费ID/付费ID/书名/男女频/推荐理由' },
  { value: 'sample_input', label: '书籍ID/书名/推荐理由/男女频/标签/评级' },
  { value: 'full_11', label: '完整11列' }
];

// 改文方案（rewrite.strategy）
const REWRITE_STRATEGY_OPTIONS = [
  { value: 'instruction', label: '指令（instruction）' },
  { value: 'opening_instruction', label: '开头指令（opening_instruction）' },
  { value: 'high_imitation', label: '高仿（high_imitation）' }
];

// 状态色：failed 红 / waiting 黄 / done 绿
function statusColor(status) {
  const s = String(status || '').toLowerCase();
  if (s.includes('failed')) return 'red';
  if (s.includes('done')) return 'green';
  return 'gold';
}

function taskStatusText(status) {
  const map = {
    created: '已创建',
    original_done: '原文已抓取',
    original_failed: '原文失败',
    original_restored: '原文已恢复'
  };
  return map[status] || status || '已创建';
}

function originalStatusText(status) {
  if (status === 'done') return '成功';
  if (status === 'failed') return '失败';
  return '未抓取';
}

function aiStatusText(task) {
  const count = Number(task.aiGeneratedCount) || 0;
  if (count > 0) return `${count} 版`;
  if (task.aiStatus === 'failed') return '失败';
  if (task.aiStatus === 'partial') return '部分';
  if (task.aiStatus) return task.aiStatus;
  return '未生成';
}

function downloadText(filename, content) {
  const blob = new Blob([content], { type: 'text/plain;charset=utf-8' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
}

export function NovelFetchWorkshopPage() {
  const [processForm] = Form.useForm();
  const [configForm] = Form.useForm();
  const [aiConfigForm] = Form.useForm();

  // 配置数据（GET /config）
  const [configData, setConfigData] = useState({ platforms: [], styles: [], aiConfig: null });

  // 任务列表（处理 / 任务 两个 Tab 共享）
  const [tasks, setTasks] = useState([]);
  const [tasksLoading, setTasksLoading] = useState(false);
  // 任务表格勾选（仅任务 Tab 使用，批量删除）
  const [selectedRowKeys, setSelectedRowKeys] = useState([]);

  // 处理 Tab
  const [processing, setProcessing] = useState(false);
  const [processResult, setProcessResult] = useState(null);

  // 任务详情（弹窗）
  const [detail, setDetail] = useState(null);     // { bookId, data, loading }

  // 单行操作状态
  const [fetchingBook, setFetchingBook] = useState(null);   // 正在重新抓原文的 bookId
  const [genModal, setGenModal] = useState(null);           // { bookId, count }
  const [generating, setGenerating] = useState(false);

  // 配置保存 / AI 测试
  const [savingConfig, setSavingConfig] = useState(false);
  const [testingAi, setTestingAi] = useState(false);
  const [handoffBook, setHandoffBook] = useState(null);

  async function loadConfig() {
    try {
      const data = await getWorkshopConfig();
      setConfigData({
        platforms: (data && data.platforms) || [],
        styles: (data && data.styles) || [],
        aiConfig: (data && data.aiConfig) || null
      });
      const app = (data && data.appConfig) || {};
      const workflow = app.workflow || {};
      const fetchCfg = app.fetch || {};
      const rewrite = app.rewrite || {};
      configForm.setFieldsValue({
        auto_classify_missing: workflow.auto_classify_missing,
        auto_fetch_original: workflow.auto_fetch_original,
        auto_rewrite_after_fetch: workflow.auto_rewrite_after_fetch,
        default_max_txt: fetchCfg.default_max_txt,
        default_ai_count: rewrite.default_ai_count,
        max_ai_count: rewrite.max_ai_count,
        process_line_count: rewrite.process_line_count,
        anchor_line_count: rewrite.anchor_line_count,
        temperature: rewrite.temperature,
        strategy: rewrite.strategy,
        method_sequence: Array.isArray(rewrite.method_sequence) ? rewrite.method_sequence.join(',') : '',
        prompt: rewrite.prompt
      });
      const ai = (data && data.aiConfig && data.aiConfig.ai) || app.ai || {};
      aiConfigForm.setFieldsValue({
        base_url: ai.base_url,
        api_key: ai.api_key,
        model: ai.model,
        timeout_seconds: ai.timeout_seconds,
        max_concurrency: ai.max_concurrency,
        retry_times: ai.retry_times,
        max_tokens: ai.max_tokens,
        temperature: ai.temperature,
        top_p: ai.top_p,
        presence_penalty: ai.presence_penalty,
        frequency_penalty: ai.frequency_penalty,
        stream: ai.stream,
        json_mode: ai.json_mode,
        enable_thinking: ai.enable_thinking,
        force_serial_batch: ai.force_serial_batch,
        extra_body_json: ai.extra_body_json
      });
    } catch (error) {
      message.error(error.message || '读取配置失败');
    }
  }

  async function loadTasks() {
    setTasksLoading(true);
    try {
      const data = await listWorkshopTasks();
      setTasks((data && data.tasks) || []);
    } catch (error) {
      message.error(error.message || '获取任务列表失败');
    } finally {
      setTasksLoading(false);
    }
  }

  useEffect(() => {
    loadConfig();
    loadTasks();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // ===== 处理 Tab =====
  async function handleProcess() {
    const values = await processForm.validateFields().catch(() => null);
    if (!values) return;
    if (!values.inputText || !values.inputText.trim()) { message.warning('请填写待处理清单'); return; }
    setProcessing(true);
    setProcessResult(null);
    try {
      const data = await processBatch({
        platformId: values.platformId,
        inputText: values.inputText,
        parseMode: values.parseMode,
        columnPresetId: values.columnPresetId,
        columnOrder: values.columnOrder || '',
        maxTxt: values.maxTxt,
        aiCount: values.aiCount
      });
      setProcessResult(data || {});
      setTasks((data && data.tasks) || []);
      message.success('处理完成');
    } catch (error) {
      message.error(error.message || '处理失败');
    } finally {
      setProcessing(false);
    }
  }

  function renderProcessSummary(result) {
    if (!result) return null;
    const lines = [
      `解析行数：${result.parsed ?? '-'}`,
      `有效任务：${result.uniqueTasks ?? '-'}`,
      `重复：${result.duplicateCount ?? '-'}`,
      `空ID：${result.emptyIdCount ?? '-'}`,
      `抓取成功：${result.fetched ?? '-'}`,
      `抓取/处理失败：${result.fetchFailed ?? '-'}`,
      `AI版本数：${result.generatedAiFiles ?? '-'}`
    ];
    if (Array.isArray(result.classifyErrors) && result.classifyErrors.length) {
      lines.push(`分类错误：${result.classifyErrors.length}`);
    }
    return (
      <Alert
        type={Number(result.fetchFailed) > 0 ? 'warning' : 'success'}
        showIcon
        message="处理结果摘要"
        description={
          <Space direction="vertical" size={0}>
            <Typography.Text>{lines.join('　')}</Typography.Text>
            {Array.isArray(result.classifyErrors) && result.classifyErrors.length ? (
              <Typography.Paragraph type="secondary" style={{ margin: '4px 0 0', whiteSpace: 'pre-wrap' }}>
                {result.classifyErrors.join('\n')}
              </Typography.Paragraph>
            ) : null}
          </Space>
        }
        style={{ maxWidth: 720 }}
      />
    );
  }

  // ===== 任务详情 =====
  async function openDetail(row) {
    setDetail({ bookId: row.bookId, data: null, loading: true });
    try {
      const data = await getWorkshopTask(row.bookId);
      setDetail({ bookId: row.bookId, data, loading: false });
    } catch (error) {
      setDetail({ bookId: row.bookId, data: null, loading: false });
      message.error(error.message || '获取任务详情失败');
    }
  }

  async function reloadDetail(bookId) {
    if (!detail || detail.bookId !== bookId) return;
    try {
      const data = await getWorkshopTask(bookId);
      setDetail({ bookId, data, loading: false });
    } catch (_) {
      setDetail(current => current ? { ...current, loading: false } : current);
    }
  }

  async function handleFetchOriginal(row) {
    setFetchingBook(row.bookId);
    try {
      await fetchWorkshopOriginal(row.bookId, row.maxTxt);
      message.success('重新抓取原文成功');
      loadTasks();
      reloadDetail(row.bookId);
    } catch (error) {
      message.error(error.message || '重新抓取原文失败');
    } finally {
      setFetchingBook(null);
    }
  }

  async function handleGenerateAi() {
    if (!genModal) return;
    const count = Math.floor(Number(genModal.count) || 1);
    if (count < 1 || count > 20) { message.warning('生成数量需为 1~20 的整数'); return; }
    setGenerating(true);
    try {
      await generateWorkshopAi(genModal.bookId, count);
      message.success('AI 生成完成');
      setGenModal(null);
      loadTasks();
      reloadDetail(genModal.bookId);
    } catch (error) {
      message.error(error.message || 'AI 生成失败');
    } finally {
      setGenerating(false);
    }
  }

  async function handleDownloadOriginal(row) {
    try {
      const data = await getWorkshopTask(row.bookId);
      const text = (data && data.original) || '';
      if (!text) { message.warning('暂无原文内容，请先抓取原文'); return; }
      downloadText(`${row.bookId}.txt`, text);
    } catch (error) {
      message.error(error.message || '下载失败');
    }
  }

  async function handleAddToBatchFactory(row) {
    setHandoffBook(row.bookId);
    try {
      const data = await getWorkshopTask(row.bookId);
      const original = String(data?.original || data?.processedText || '').trim();
      if (!original) {
        message.warning('请先成功抓取原文，再交给批量工厂');
        return;
      }
      const meta = data?.meta || {};
      const intake = await createNovelFetchIntake({
        books: [{
          bookId: String(row.bookId),
          sourceTaskId: String(row.bookId),
          title: String(meta.bookName || row.bookName || row.bookId),
          platform: String(meta.platformName || row.platformName || row.platform || ''),
          sourceText: original,
          txtText: original,
          sourceMetadata: {
            source: 'novel-fetch-workshop',
            gender: meta.gender || row.gender || '',
            style: meta.style || row.style || '',
            bookInfo: data?.bookInfo || {}
          }
        }],
        metadata: { source: 'novel-fetch-workshop', version: 'go-native' }
      });
      const intakeId = intake?.intake?.id || intake?.id;
      if (!intakeId) throw new Error('批量工厂交接未返回编号');
      message.success('已交给批量工厂');
      window.location.assign(batchFactoryNovelFetchHandoffPath(intakeId));
    } catch (error) {
      message.error(error.message || '交给批量工厂失败');
    } finally {
      setHandoffBook(null);
    }
  }

  // ===== 配置 Tab =====
  async function handleSaveConfig() {
    const values = await configForm.validateFields().catch(() => null);
    if (!values) return;
    setSavingConfig(true);
    try {
      const methodSequence = String(values.method_sequence || '')
        .split(/[,，\s]+/)
        .map(item => item.trim())
        .filter(Boolean);
      await saveWorkshopConfig({
        appConfig: {
          workflow: {
            auto_classify_missing: Boolean(values.auto_classify_missing),
            auto_fetch_original: Boolean(values.auto_fetch_original),
            auto_rewrite_after_fetch: Boolean(values.auto_rewrite_after_fetch)
          },
          fetch: { default_max_txt: values.default_max_txt },
          rewrite: {
            default_ai_count: values.default_ai_count,
            max_ai_count: values.max_ai_count,
            process_line_count: values.process_line_count,
            anchor_line_count: values.anchor_line_count,
            temperature: values.temperature,
            strategy: values.strategy,
            method_sequence: methodSequence,
            prompt: values.prompt
          }
        }
      });
      message.success('配置已保存');
    } catch (error) {
      message.error(error.message || '保存配置失败');
    } finally {
      setSavingConfig(false);
    }
  }

  async function handleSaveAiConfig() {
    const values = await aiConfigForm.validateFields().catch(() => null);
    if (!values) return;
    setSavingConfig(true);
    try {
      await saveWorkshopConfig({
        appConfig: {
          ai: {
            base_url: values.base_url,
            api_key: values.api_key,
            model: values.model,
            timeout_seconds: values.timeout_seconds,
            max_concurrency: values.max_concurrency,
            retry_times: values.retry_times,
            max_tokens: values.max_tokens,
            temperature: values.temperature,
            top_p: values.top_p,
            presence_penalty: values.presence_penalty,
            frequency_penalty: values.frequency_penalty,
            stream: Boolean(values.stream),
            json_mode: Boolean(values.json_mode),
            enable_thinking: Boolean(values.enable_thinking),
            disable_thinking: !values.enable_thinking,
            force_serial_batch: Boolean(values.force_serial_batch),
            extra_body_json: values.extra_body_json || ''
          }
        }
      });
      message.success('AI 配置已保存');
    } catch (error) {
      message.error(error.message || '保存 AI 配置失败');
    } finally {
      setSavingConfig(false);
    }
  }

  async function handleTestAi() {
    setTestingAi(true);
    try {
      const data = await testWorkshopAi('rewrite');
      Modal.info({
        title: 'AI 接口测试结果',
        content: data && data.ok
          ? `连通成功，返回内容：${(data.content || '').slice(0, 200)}`
          : `测试失败：${(data && data.error) || '未知错误'}`
      });
    } catch (error) {
      message.error(error.message || 'AI 接口测试失败');
    } finally {
      setTestingAi(false);
    }
  }

  // ===== 任务表格（处理 / 任务 Tab 共用）=====
  // 批量删除选中任务：确认后调 DELETE /tasks，成功后刷新列表并清空勾选
  function handleDeleteSelected() {
    const ids = selectedRowKeys;
    if (!ids.length) return;
    Modal.confirm({
      title: '删除选中任务',
      content: `确定删除选中的 ${ids.length} 个任务？将同时删除其原文、AI 版本与日志文件，且不可恢复。`,
      okText: '删除',
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          const data = await deleteWorkshopTasks(ids);
          message.success(`已删除 ${(data && data.deleted) || 0} 个任务`);
          setSelectedRowKeys([]);
          await loadTasks();
        } catch (error) {
          message.error(error.message || '删除失败');
        }
      }
    });
  }

  const actionColumn = {
    title: '操作',
    key: 'actions',
    width: 380,
    render: (_, row) => (
      <Space size={4} wrap>
        <Button size="small" icon={<Eye size={14} aria-hidden="true" />} onClick={() => openDetail(row)}>查看</Button>
        <Button
          size="small"
          icon={<RotateCcw size={14} aria-hidden="true" />}
          loading={fetchingBook === row.bookId}
          disabled={fetchingBook !== null && fetchingBook !== row.bookId}
          onClick={() => handleFetchOriginal(row)}
        >重新抓原文</Button>
        <Button size="small" icon={<Wand2 size={14} aria-hidden="true" />} onClick={() => setGenModal({ bookId: row.bookId, count: Number(row.aiCount) || 1 })}>生成AI</Button>
        <Button size="small" icon={<Download size={14} aria-hidden="true" />} onClick={() => handleDownloadOriginal(row)}>下载</Button>
        <Button size="small" icon={<UploadCloud size={14} aria-hidden="true" />} loading={handoffBook === row.bookId} onClick={() => handleAddToBatchFactory(row)}>交给批量工厂</Button>
      </Space>
    )
  };

  const taskTableColumns = [
    { title: 'ID', dataIndex: 'bookId', width: 180, ellipsis: true },
    { title: '书名', dataIndex: 'bookName', ellipsis: true },
    { title: '平台', dataIndex: 'platformName', width: 110 },
    { title: '风格', dataIndex: 'style', width: 110 },
    { title: '男女频', dataIndex: 'gender', width: 80 },
    {
      title: '原文状态',
      dataIndex: 'originalStatus',
      width: 100,
      render: (v) => <Tag color={statusColor(v)}>{originalStatusText(v)}</Tag>
    },
    {
      title: 'AI状态',
      dataIndex: 'aiStatus',
      width: 100,
      render: (_, row) => <Tag color={statusColor(Number(row.aiGeneratedCount) > 0 ? 'done' : (row.aiStatus || ''))}>{aiStatusText(row)}</Tag>
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 110,
      render: (v) => <Tag color={statusColor(v)}>{taskStatusText(v)}</Tag>
    },
    actionColumn
  ];

  function TaskTable({ withSelection, selectedRowKeys: keys, onSelectionChange }) {
    return (
      <>
        {withSelection ? (
          <Space style={{ marginBottom: 8 }} wrap>
            <Typography.Text type="secondary">已选 {keys.length} 项</Typography.Text>
            <Button
              danger
              size="small"
              icon={<Trash2 size={14} aria-hidden="true" />}
              disabled={!keys.length}
              onClick={handleDeleteSelected}
            >删除选中</Button>
          </Space>
        ) : null}
        <Table
          size="small"
          rowKey="bookId"
          loading={tasksLoading}
          dataSource={tasks}
          columns={taskTableColumns}
          pagination={{ pageSize: 20, showSizeChanger: false }}
          rowSelection={withSelection ? { type: 'checkbox', selectedRowKeys: keys, onChange: onSelectionChange } : undefined}
        />
      </>
    );
  }

  // ===== 任务详情弹窗 =====
  const meta = (detail && detail.data && detail.data.meta) || {};
  const aiCount = Number(meta.aiGeneratedCount) || 0;

  const metaItems = [
    ['书籍 ID', meta.bookId],
    ['书名', meta.bookName],
    ['平台', meta.platformName],
    ['风格', meta.style],
    ['男女频', meta.gender],
    ['状态', taskStatusText(meta.status)],
    ['原文状态', originalStatusText(meta.originalStatus)],
    ['AI 状态', aiStatusText(meta)],
    ['字符数', meta.originalChars != null ? String(meta.originalChars) : '-'],
    ['最大抓取字数', meta.maxTxt != null ? String(meta.maxTxt) : '-'],
    ['创建时间', meta.createdAt || '-'],
    ['更新时间', meta.updatedAt || '-']
  ];

  return (
    <div style={{ height: '100%', overflowY: 'auto', boxSizing: 'border-box' }}>
      <Space direction="vertical" size={16} style={{ width: '100%', maxWidth: 1180, padding: 24, boxSizing: 'border-box' }}>
        <Space align="center">
          <Button icon={<ArrowLeft size={16} aria-hidden="true" />} onClick={() => window.history.back()}>返回</Button>
          <Typography.Title level={3} style={{ margin: 0 }}>改文工作台</Typography.Title>
        </Space>

      <Tabs
        items={[
          {
            key: 'process',
            label: '处理',
            children: (
              <Space direction="vertical" size={16} style={{ width: '100%' }}>
                <Form
                  form={processForm}
                  layout="vertical"
                  initialValues={{ platformId: '2', parseMode: 'smart', columnPresetId: 'sample_input', maxTxt: 4000, aiCount: 1 }}
                  style={{ maxWidth: 840 }}
                >
                  <Space size={12} wrap>
                    <Form.Item name="platformId" label="平台">
                      <Select
                        style={{ width: 150 }}
                        options={(configData.platforms || [])
                          .filter(p => p && p.visible !== false)
                          .map(p => ({ value: String(p.id), label: p.name }))}
                      />
                    </Form.Item>
                    <Form.Item name="parseMode" label="输入格式">
                      <Select style={{ width: 200 }} options={PARSE_MODE_OPTIONS} />
                    </Form.Item>
                    <Form.Item name="columnPresetId" label="列顺序预设">
                      <Select style={{ width: 220 }} options={COLUMN_PRESET_OPTIONS} />
                    </Form.Item>
                    <Form.Item name="columnOrder" label="自定义列顺序">
                      <Input style={{ width: 220 }} placeholder="例：书籍ID,书名,推荐理由,男女频,标签,评级" />
                    </Form.Item>
                    <Form.Item name="maxTxt" label="截取字数">
                      <InputNumber min={100} max={100000} style={{ width: 110 }} />
                    </Form.Item>
                    <Form.Item name="aiCount" label="AI文案数量">
                      <InputNumber min={1} max={20} style={{ width: 90 }} />
                    </Form.Item>
                  </Space>
                  <Form.Item name="inputText" label="待处理清单（一行一本书）" rules={[{ required: true, message: '请填写待处理清单' }]}>
                    <Input.TextArea
                      rows={6}
                      placeholder={'每行一本书，示例：\n1\t书A\t推荐理由\t女频\t标签\t评级'}
                    />
                  </Form.Item>
                  <Space>
                    <Button type="primary" icon={<Wand2 size={14} aria-hidden="true" />} loading={processing} onClick={handleProcess}>开始处理</Button>
                    <Button icon={<RefreshCw size={14} aria-hidden="true" />} loading={tasksLoading} onClick={loadTasks}>刷新任务</Button>
                  </Space>
                </Form>
                {renderProcessSummary(processResult)}
                <Typography.Title level={5} style={{ margin: 0 }}>任务列表</Typography.Title>
                <TaskTable withSelection={false} />
              </Space>
            )
          },
          {
            key: 'tasks',
            label: '任务',
            children: (
              <Space direction="vertical" size={12} style={{ width: '100%' }}>
                <Typography.Paragraph type="secondary" style={{ margin: 0 }}>
                  支持勾选批量删除选中任务；批量重试一期未提供，单本操作：查看、重新抓原文、生成AI、下载。
                </Typography.Paragraph>
                <TaskTable withSelection selectedRowKeys={selectedRowKeys} onSelectionChange={setSelectedRowKeys} />
              </Space>
            )
          },
          {
            key: 'config',
            label: '配置',
            children: (
              <Space direction="vertical" size={20} style={{ width: '100%', maxWidth: 760 }}>
                <div className="legacy-panel-card" style={{ padding: 16 }}>
                  <Typography.Title level={5} style={{ margin: '0 0 12px' }}>自动处理区</Typography.Title>
                  <Form form={configForm} layout="vertical">
                    <Space size={16} wrap>
                      <Form.Item name="default_max_txt" label="截取字数">
                        <InputNumber min={100} max={100000} style={{ width: 120 }} />
                      </Form.Item>
                      <Form.Item name="default_ai_count" label="默认AI文案数量">
                        <InputNumber min={1} max={20} style={{ width: 110 }} />
                      </Form.Item>
                      <Form.Item name="max_ai_count" label="最大AI文案数量">
                        <InputNumber min={1} max={20} style={{ width: 110 }} />
                      </Form.Item>
                      <Form.Item name="process_line_count" label="处理行数">
                        <InputNumber min={1} max={100} style={{ width: 100 }} />
                      </Form.Item>
                      <Form.Item name="anchor_line_count" label="锚点行数">
                        <InputNumber min={0} max={100} style={{ width: 100 }} />
                      </Form.Item>
                      <Form.Item name="temperature" label="改文温度">
                        <InputNumber min={0} max={2} step={0.05} style={{ width: 100 }} />
                      </Form.Item>
                      <Form.Item name="strategy" label="改文方案">
                        <Select style={{ width: 220 }} options={REWRITE_STRATEGY_OPTIONS} />
                      </Form.Item>
                      <Form.Item name="method_sequence" label="出文轮换顺序（逗号分隔）">
                        <Input style={{ width: 300 }} placeholder="high_imitation,opening_instruction,instruction" />
                      </Form.Item>
                    </Space>
                    <Form.Item name="prompt" label="改文提示词">
                      <Input.TextArea rows={4} />
                    </Form.Item>
                    <Space size={24} wrap>
                      <Form.Item name="auto_classify_missing" label="自动补齐风格/男女频" valuePropName="checked">
                        <Switch />
                      </Form.Item>
                      <Form.Item name="auto_fetch_original" label="自动抓原文" valuePropName="checked">
                        <Switch />
                      </Form.Item>
                      <Form.Item name="auto_rewrite_after_fetch" label="抓取后自动生成AI文案" valuePropName="checked">
                        <Switch />
                      </Form.Item>
                    </Space>
                    <Button type="primary" icon={<Save size={14} aria-hidden="true" />} loading={savingConfig} onClick={handleSaveConfig}>保存配置</Button>
                  </Form>
                </div>

                <div className="legacy-panel-card" style={{ padding: 16 }}>
                  <Typography.Title level={5} style={{ margin: '0 0 12px' }}>AI 接口区</Typography.Title>
                  <Form form={aiConfigForm} layout="vertical">
                    <Space size={16} wrap>
                      <Form.Item name="base_url" label="API地址" style={{ width: 340 }}>
                        <Input placeholder="https://api.example.com/v1/chat/completions" />
                      </Form.Item>
                      <Form.Item name="api_key" label="API Key">
                        <Input.Password style={{ width: 220 }} autoComplete="new-password" />
                      </Form.Item>
                      <Form.Item name="model" label="模型名">
                        <Input style={{ width: 200 }} />
                      </Form.Item>
                      <Form.Item name="timeout_seconds" label="超时(秒)">
                        <InputNumber min={1} max={600} style={{ width: 110 }} />
                      </Form.Item>
                      <Form.Item name="max_concurrency" label="并发">
                        <InputNumber min={1} max={50} style={{ width: 90 }} />
                      </Form.Item>
                      <Form.Item name="retry_times" label="重试">
                        <InputNumber min={0} max={10} style={{ width: 90 }} />
                      </Form.Item>
                      <Form.Item name="max_tokens" label="max_tokens">
                        <InputNumber min={1} max={32000} style={{ width: 110 }} />
                      </Form.Item>
                      <Form.Item name="temperature" label="temperature">
                        <InputNumber min={0} max={2} step={0.05} style={{ width: 110 }} />
                      </Form.Item>
                      <Form.Item name="top_p" label="top_p">
                        <InputNumber min={0} max={1} step={0.05} style={{ width: 100 }} />
                      </Form.Item>
                      <Form.Item name="presence_penalty" label="presence">
                        <InputNumber min={-2} max={2} step={0.1} style={{ width: 110 }} />
                      </Form.Item>
                      <Form.Item name="frequency_penalty" label="frequency">
                        <InputNumber min={-2} max={2} step={0.1} style={{ width: 110 }} />
                      </Form.Item>
                    </Space>
                    <Form.Item label="选项" style={{ marginBottom: 12 }}>
                      <Space size={16} wrap>
                        <Form.Item name="stream" valuePropName="checked" style={{ marginBottom: 0 }}>
                          <Checkbox>流式</Checkbox>
                        </Form.Item>
                        <Form.Item name="json_mode" valuePropName="checked" style={{ marginBottom: 0 }}>
                          <Checkbox>JSON</Checkbox>
                        </Form.Item>
                        <Form.Item name="enable_thinking" valuePropName="checked" style={{ marginBottom: 0 }}>
                          <Checkbox>思考</Checkbox>
                        </Form.Item>
                        <Form.Item name="force_serial_batch" valuePropName="checked" style={{ marginBottom: 0 }}>
                          <Checkbox>强制串行</Checkbox>
                        </Form.Item>
                      </Space>
                    </Form.Item>
                    <Form.Item name="extra_body_json" label="额外JSON">
                      <Input.TextArea rows={3} placeholder='{"extra_param": "value"}' />
                    </Form.Item>
                    <Space>
                      <Button type="primary" icon={<Save size={14} aria-hidden="true" />} loading={savingConfig} onClick={handleSaveAiConfig}>保存AI配置</Button>
                      <Button icon={<RotateCcw size={14} aria-hidden="true" />} loading={testingAi} onClick={handleTestAi}>测试接口</Button>
                    </Space>
                  </Form>
                </div>

                <div className="legacy-panel-card" style={{ padding: 16 }}>
                  <Typography.Title level={5} style={{ margin: '0 0 12px' }}>平台表 / 风格表（一期只读）</Typography.Title>
                  <Space size={24} align="start" wrap>
                    <div>
                      <Typography.Text strong>平台表</Typography.Text>
                      <Table
                        size="small"
                        rowKey="id"
                        dataSource={(configData.platforms || []).map((p, i) => ({ ...p, index: i + 1 }))}
                        pagination={false}
                        columns={[
                          { title: 'ID', dataIndex: 'id', width: 60 },
                          { title: '名称', dataIndex: 'name' }
                        ]}
                        style={{ marginTop: 8 }}
                      />
                    </div>
                    <div>
                      <Typography.Text strong>风格表</Typography.Text>
                      <Typography.Paragraph type="secondary" style={{ margin: '8px 0 0', maxWidth: 400 }}>
                        {(configData.styles || []).join('、') || '（空）'}
                      </Typography.Paragraph>
                    </div>
                  </Space>
                </div>
              </Space>
            )
          }
        ]}
      />

      {/* 任务详情弹窗 */}
      <Modal
        title={detail ? `任务详情 — ${detail.bookId}` : ''}
        open={Boolean(detail)}
        width={1000}
        onCancel={() => setDetail(null)}
        footer={[
          <Button key="dl" icon={<Download size={14} aria-hidden="true" />} onClick={() => detail && detail.data && detail.data.original ? downloadText(`${detail.bookId}.txt`, detail.data.original) : message.warning('暂无原文内容')}>下载原文</Button>,
          <Button key="fetch" icon={<RotateCcw size={14} aria-hidden="true" />} loading={fetchingBook === (detail && detail.bookId)} disabled={fetchingBook !== null && fetchingBook !== (detail && detail.bookId)} onClick={() => detail && handleFetchOriginal({ bookId: detail.bookId, maxTxt: meta.maxTxt })}>重新抓原文</Button>,
          <Button key="gen" type="primary" icon={<Wand2 size={14} aria-hidden="true" />} onClick={() => detail && setGenModal({ bookId: detail.bookId, count: aiCount || 1 })}>生成AI</Button>,
          <Button key="close" onClick={() => setDetail(null)}>关闭</Button>
        ]}
      >
        {detail ? (
          <Tabs
            items={[
              {
                key: 'meta',
                label: '元信息',
                children: (
                  <Space direction="vertical" size={8}>
                    {metaItems.map(([label, value]) => (
                      <div key={label} style={{ display: 'flex', gap: 8 }}>
                        <Typography.Text type="secondary" style={{ width: 90, flexShrink: 0 }}>{label}：</Typography.Text>
                        <Typography.Text>{value || '-'}</Typography.Text>
                      </div>
                    ))}
                  </Space>
                )
              },
              {
                key: 'original',
                label: '处理后原文',
                children: (
                  <Input.TextArea value={(detail.data && detail.data.original) || ''} rows={18} readOnly className="novel-fetch-preview" />
                )
              },
              {
                key: 'ai',
                label: `AI版本${aiCount > 0 ? `（${aiCount}）` : ''}`,
                children: aiCount > 0 ? (
                  <Space direction="vertical" size={8}>
                    <Space wrap>
                      {Array.from({ length: aiCount }, (_, i) => i + 1).map(n => <Tag key={n} color="blue">ai{n}</Tag>)}
                    </Space>
                    <Typography.Paragraph type="secondary" style={{ margin: 0 }}>
                      一期接口未提供 AI 版本文本与下载，改文版本文件保存在服务器 ai/ai{n}/ 目录。
                    </Typography.Paragraph>
                  </Space>
                ) : (
                  <Typography.Paragraph type="secondary" style={{ margin: 0 }}>尚未生成 AI 版本。</Typography.Paragraph>
                )
              },
              {
                key: 'logs',
                label: '日志',
                children: (
                  <Space direction="vertical" size={4} style={{ width: '100%' }}>
                    {(detail.data && Array.isArray(detail.data.logs) && detail.data.logs.length
                      ? detail.data.logs
                      : []
                    ).map((log, i) => (
                      <Typography.Paragraph key={i} style={{ margin: 0, fontSize: 12 }}>
                        <Typography.Text type="secondary">{log.time || ''}</Typography.Text>　{log.event || ''}
                        {log.data && Object.keys(log.data).length ? <Typography.Text type="secondary">　{JSON.stringify(log.data)}</Typography.Text> : null}
                      </Typography.Paragraph>
                    ))}
                    {!(detail.data && Array.isArray(detail.data.logs) && detail.data.logs.length) ? (
                      <Typography.Paragraph type="secondary" style={{ margin: 0 }}>暂无日志。</Typography.Paragraph>
                    ) : null}
                  </Space>
                )
              }
            ]}
          />
        ) : null}
      </Modal>

      {/* 生成 AI 弹窗 */}
      <Modal
        title="生成 AI 改文版本"
        open={Boolean(genModal)}
        onCancel={() => setGenModal(null)}
        footer={[
          <Button key="cancel" onClick={() => setGenModal(null)}>取消</Button>,
          <Button key="ok" type="primary" icon={<Wand2 size={14} aria-hidden="true" />} loading={generating} onClick={handleGenerateAi}>生成</Button>
        ]}
      >
        <Space direction="vertical" size={8}>
          <Typography.Text>书籍 ID：{genModal ? genModal.bookId : ''}</Typography.Text>
          <Typography.Text>生成数量（1~20）：</Typography.Text>
          <InputNumber
            min={1}
            max={20}
            value={genModal ? genModal.count : 1}
            onChange={v => setGenModal(current => current ? { ...current, count: v } : current)}
          />
        </Space>
      </Modal>
      </Space>
    </div>
  );
}

export default NovelFetchWorkshopPage;
