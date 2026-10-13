import { Alert, Button, Card, Select, Space, Tag, Typography, message } from 'antd';
import { Clapperboard, ExternalLink } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { generateBatchFactoryVideos } from '../../../shared/api/batchFactory';
import { listModels } from '../../../shared/api/shuihuoProduction';
import { reportClientError } from '../../../shared/error-reporting';

let modelCatalogPromise = null;

export function loadBatchFactoryVideoModels() {
  if (!modelCatalogPromise) {
    modelCatalogPromise = listModels()
      .then(result => (result.models || []).filter(model => model.kind === 'video'))
      .catch(error => {
        modelCatalogPromise = null;
        throw error;
      });
  }
  return modelCatalogPromise;
}

function batchFactoryCompatibleModels(models, batch) {
  const requiredDuration = Number(batch?.settings?.maxVideoDuration || 0);
  return models.filter(model => (
    model.requiresImageInput !== true
    && Number.isInteger(Number(model.maxVideoDuration))
    && Number(model.maxVideoDuration) >= 1
    && (!requiredDuration || Number(model.maxVideoDuration) >= requiredDuration)
  ));
}

function reportProductionIssue(batch, item, messageText) {
  reportClientError({
    kind: 'batch-factory.production-submit-failed',
    message: messageText || '视频提交失败',
    source: '/api/batch-factory/production',
    context: {
      batchId: batch?.id || '',
      itemId: item?.id || '',
      bookTitle: item?.title || '',
      bookId: item?.bookId || '',
      projectId: item?.production?.projectId || '',
      modelName: batch?.settings?.videoModelName || ''
    }
  });
}

function SubmissionErrorAlert({ error }) {
  if (!error?.message) return null;
  const promptFailure = error.stage === 'prompt';
  return <Alert
    type="error"
    showIcon
    message={promptFailure ? 'VIDEO 提示词编译失败' : '上次视频提交未成功'}
    description={error.message}
  />;
}

export function BatchFactoryProductionControls({ batch, item, onRefresh }) {
  const [models, setModels] = useState([]);
  const [legacyModelId, setLegacyModelId] = useState(null);
  const [loadingModels, setLoadingModels] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const boundModelId = Number(batch?.settings?.videoModelId || 0) || null;
  const modelId = boundModelId || legacyModelId;

  useEffect(() => {
    let active = true;
    setLoadingModels(true);
    loadBatchFactoryVideoModels()
      .then(videoModels => {
        if (!active) return;
        setModels(videoModels);
        if (!boundModelId) {
          const compatible = batchFactoryCompatibleModels(videoModels, batch);
          if (compatible.length) setLegacyModelId(current => current || compatible[0].id);
        }
      })
      .catch(error => { if (active) message.error(error.message || '读取视频模型失败'); })
      .finally(() => { if (active) setLoadingModels(false); });
    return () => { active = false; };
  }, [boundModelId, batch?.settings?.maxVideoDuration]);

  const directModels = useMemo(() => batchFactoryCompatibleModels(models, batch), [models, batch]);
  const imageModels = useMemo(() => models.filter(model => model.requiresImageInput === true), [models]);
  const boundModel = useMemo(() => models.find(model => Number(model.id) === boundModelId) || null, [models, boundModelId]);
  const production = item?.production;
  const submissionError = item?.productionSubmissionError;

  async function generate() {
    if (!modelId) return message.warning('当前批次没有可用的视频模型');
    setSubmitting(true);
    try {
      const result = await generateBatchFactoryVideos(batch.id, item.id, modelId);
      await onRefresh?.();
      const queued = result.production?.queued ?? 0;
      const total = result.production?.total ?? 0;
      if (result.production?.failed) {
        reportProductionIssue(batch, item, `${result.production.failed} 个 VIDEO 提交失败，${queued}/${total} 已进入队列`);
        message.warning(`已提交生产：${queued}/${total} 个 VIDEO 进入队列，部分任务需要检查。`);
      } else {
        message.success(`已提交生产：${queued}/${total} 个 VIDEO 已进入视频队列。`);
      }
    } catch (error) {
      reportProductionIssue(batch, item, error.message || '提交视频生产失败');
      await onRefresh?.().catch?.(() => {});
      message.error(error.message || '提交视频生产失败');
    } finally {
      setSubmitting(false);
    }
  }

  if (production?.projectId) {
    return <Card size="small" title="视频生产" style={{ width: '100%' }}>
      <Space direction="vertical" size={10} style={{ width: '100%' }}>
        <Space wrap>
          <Tag color={production.status === 'partial' ? 'gold' : 'processing'}>生产项目 #{production.projectId}</Tag>
          <Tag>{production.modelName || `模型 #${production.modelId}`}</Tag>
          <Typography.Text>{production.queued}/{production.total} 个 VIDEO 已排队</Typography.Text>
          {production.failed ? <Tag color="red">{production.failed} 个提交失败</Tag> : null}
          <Button icon={<ExternalLink size={14} />} onClick={() => { window.location.href = '/shuihuo-production'; }}>打开批量工厂</Button>
        </Space>
        {production.failed ? <Alert type="warning" showIcon message="部分 VIDEO 没有成功进入队列" description="展开上方的视频生成进度，可以看到每个 VIDEO 的实际状态和失败原因。" /> : null}
      </Space>
    </Card>;
  }

  if (boundModelId) {
    const boundName = batch.settings?.videoModelName || boundModel?.name || `模型 #${boundModelId}`;
    const unavailable = !loadingModels && !boundModel;
    return <Card size="small" title="生成视频" style={{ width: '100%' }}>
      <Space direction="vertical" size={10} style={{ width: '100%' }}>
        <Typography.Text type="secondary">本篇会沿用导演开始前锁定的视频模型；模型与时长能力不会在生成阶段再次改变。</Typography.Text>
        <Space wrap>
          <Tag color="blue">{boundName}</Tag>
          <Tag>单次最大 {batch.settings?.maxVideoDuration || '—'}s</Tag>
          <Tag>{batch.settings?.aspectRatio || '9:16'}</Tag>
        </Space>
        <SubmissionErrorAlert error={submissionError} />
        {unavailable ? <Alert type="warning" showIcon message="已绑定模型当前不在可用模型列表中" description="模型可能已被停用或隐藏。恢复该模型后再生成；不要用另一模型直接替换已经完成的导演方案。" /> : null}
        <Button type="primary" icon={<Clapperboard size={15} />} loading={submitting} disabled={unavailable || !modelId} onClick={generate} style={{ alignSelf: 'flex-start' }}>{submissionError ? '重新提交全部 VIDEO' : '生成全部 VIDEO'}</Button>
      </Space>
    </Card>;
  }

  return <Card size="small" title="生成视频 · 历史批次兼容" style={{ width: '100%' }}>
    {directModels.length ? <Space direction="vertical" size={10} style={{ width: '100%' }}>
      <Alert type="info" showIcon message="这是旧批次，创建时没有绑定视频模型" description={`只能选择单次最大时长不小于当前导演上限 ${batch?.settings?.maxVideoDuration || '—'}s 的文生视频模型。新批次会在导演前锁定模型。`} />
      <SubmissionErrorAlert error={submissionError} />
      <Space wrap>
        <Select
          loading={loadingModels}
          value={legacyModelId}
          onChange={setLegacyModelId}
          style={{ minWidth: 320 }}
          options={directModels.map(model => ({ value: model.id, label: `${model.name} · 最大 ${model.maxVideoDuration}s` }))}
          placeholder="选择兼容文生视频模型"
        />
        <Button type="primary" icon={<Clapperboard size={15} />} loading={submitting} onClick={generate}>{submissionError ? '重新提交全部 VIDEO' : '生成全部 VIDEO'}</Button>
      </Space>
      {imageModels.length ? <Typography.Text type="secondary">另有 {imageModels.length} 个图生视频模型未显示；它们不能从批量工厂无图直出。</Typography.Text> : null}
    </Space> : <Alert
      type="warning"
      showIcon
      message="暂无兼容的文生视频模型"
      description="请管理员启用文生视频模型并配置单次最大生成时长；其能力必须覆盖这个历史批次的导演时长上限。"
    />}
  </Card>;
}
