import React, { useMemo, useState } from 'react';
import {
  Alert,
  Button,
  Card,
  Col,
  Divider,
  Flex,
  Input,
  InputNumber,
  Row,
  Select,
  Space,
  Tag,
  Typography,
  message,
} from 'antd';
import {
  BATCH_FACTORY_PLATFORMS,
  addBookstoreGroup,
  groupLabel,
  parseBookIds,
  submitIntakeAndStartJob,
} from './batchFactoryIntake.js';

const { Title, Text, Paragraph } = Typography;
const { TextArea } = Input;

const platformOptions = BATCH_FACTORY_PLATFORMS.map(item => ({ value: item.id, label: item.name }));

export default function BatchFactoryPage() {
  const [title, setTitle] = useState('');
  const [platformId, setPlatformId] = useState(BATCH_FACTORY_PLATFORMS[0].id);
  const [rawBookIds, setRawBookIds] = useState('');
  const [manualGender, setManualGender] = useState('');
  const [style, setStyle] = useState('');
  const [maxTxt, setMaxTxt] = useState(4000);
  const [groups, setGroups] = useState([]);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');
  const [result, setResult] = useState(null);
  const [messageApi, contextHolder] = message.useMessage();

  const currentPlatform = useMemo(() => BATCH_FACTORY_PLATFORMS.find(item => item.id === platformId), [platformId]);
  const pendingCount = useMemo(() => parseBookIds(rawBookIds).length, [rawBookIds]);
  const totalBooks = useMemo(() => groups.reduce((sum, group) => sum + (group.books?.length || 0), 0), [groups]);

  function handleAddGroup() {
    setError('');
    setResult(null);
    try {
      const next = addBookstoreGroup(groups, { platformId, platformName: currentPlatform?.name, rawBookIds, manualGender, style, maxTxt });
      setGroups(next);
      setRawBookIds('');
      messageApi.success(`已添加 ${groupLabel(next[next.length - 1])}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : '添加书城失败');
    }
  }

  function handleRemoveGroup(index) {
    setGroups(current => current.filter((_, itemIndex) => itemIndex !== index));
    setResult(null);
    setError('');
  }

  async function handleExecute() {
    setSubmitting(true);
    setError('');
    setResult(null);
    try {
      const nextResult = await submitIntakeAndStartJob({ title, groups });
      setResult(nextResult);
      messageApi.success('已创建批量任务');
    } catch (err) {
      setError(err instanceof Error ? err.message : '立即执行失败');
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="batch-shell">
      {contextHolder}
      <div className="batch-container">
        <div className="batch-hero">
          <div>
            <Text className="eyebrow">BATCH FACTORY</Text>
            <Title level={1}>批量工厂</Title>
            <Paragraph className="hero-copy">按书城分组添加小说。点击立即执行后，系统会创建 Intake、进入 Redis 队列，再由 Go Worker 获取正文并处理男女频与风格。</Paragraph>
          </div>
          <div className="hero-stat"><span>已选书城</span><strong>{groups.length}</strong><span>共 {totalBooks} 本</span></div>
        </div>

        {error ? <Alert className="top-alert" type="error" showIcon message={error} closable onClose={() => setError('')} /> : null}
        {result ? <Alert className="top-alert" type="success" showIcon message="批量任务已进入队列" description={<Space wrap><Tag>Intake: {result.intake?.intake_id}</Tag><Tag>Job: {result.job?.job_id}</Tag><Tag color="processing">{result.job?.status || 'queued'}</Tag><Tag>{result.intake?.book_count ?? totalBooks} 本</Tag></Space>} /> : null}

        <Row gutter={[20, 20]}>
          <Col xs={24} xl={15}>
            <Card className="main-card" bordered={false}>
              <Flex justify="space-between" align="flex-start" gap={16} wrap="wrap">
                <div><Title level={3}>添加书城</Title><Text type="secondary">选择书城后，把这一批书籍 ID 粘贴进输入框。</Text></div>
                <Tag className="count-tag">待添加 {pendingCount} 本</Tag>
              </Flex>
              <Divider />
              <Row gutter={[16, 16]}>
                <Col xs={24} md={12}><label className="field-label">书城</label><Select className="full-width" value={platformId} options={platformOptions} onChange={setPlatformId} showSearch optionFilterProp="label" /></Col>
                <Col xs={24} md={12}><label className="field-label">批次名称（可选）</label><Input value={title} onChange={event => setTitle(event.target.value)} placeholder="例如：10月4日第一批" maxLength={120} /></Col>
                <Col span={24}><label className="field-label">书籍 ID</label><TextArea rows={8} value={rawBookIds} onChange={event => setRawBookIds(event.target.value)} placeholder={'一行一本，也支持逗号或空格分隔\n例如：\n7673480334440139800\n7673480334440139801'} /><Text type="secondary" className="field-help">重复 ID 会自动去重；同一个书城如果重复添加同一本，会直接提示。</Text></Col>
                <Col xs={24} md={8}><label className="field-label">男女频</label><Select className="full-width" value={manualGender} onChange={setManualGender} options={[{ value: '', label: '自动判断' }, { value: 'male', label: '男频' }, { value: 'female', label: '女频' }]} /></Col>
                <Col xs={24} md={8}><label className="field-label">风格（可选）</label><Input value={style} onChange={event => setStyle(event.target.value)} placeholder="例如：现言 / 都市" maxLength={64} /></Col>
                <Col xs={24} md={8}><label className="field-label">获取字数 max_txt</label><InputNumber className="full-width" min={1} max={200000} step={500} value={maxTxt} onChange={value => setMaxTxt(value || 4000)} /></Col>
              </Row>
              <Flex justify="flex-end" className="form-actions"><Button type="primary" size="large" onClick={handleAddGroup}>添加书城</Button></Flex>
            </Card>
          </Col>

          <Col xs={24} xl={9}>
            <Card className="side-card" bordered={false}>
              <Flex justify="space-between" align="center"><div><Title level={3}>所选书城</Title><Text type="secondary">每次“添加书城”都会形成独立标签。</Text></div><Tag>{groups.length} 组</Tag></Flex>
              <div className="group-list">
                {groups.length === 0 ? <div className="empty-state"><span>还没有添加书城</span><small>先在左侧粘贴书籍 ID，再点击“添加书城”。</small></div> : groups.map((group, index) => <div className="group-item" key={`${group.platform_id}-${index}`}><div className="group-head"><Tag color="blue">{groupLabel(group)}</Tag><Button type="text" danger size="small" onClick={() => handleRemoveGroup(index)}>移除</Button></div><div className="group-meta"><span>平台 ID：{group.platform_id}</span><span>max_txt：{group.max_txt}</span><span>男女频：{group.books?.[0]?.manual_gender === 'male' ? '男频' : group.books?.[0]?.manual_gender === 'female' ? '女频' : '自动判断'}</span><span>风格：{group.books?.[0]?.style || '自动判断'}</span></div></div>)}
              </div>
              <Divider />
              <Button type="primary" size="large" block disabled={groups.length === 0} loading={submitting} onClick={handleExecute}>立即执行</Button>
              <Text type="secondary" className="execute-help">立即执行会先保存整批书籍，再创建后台任务；不会从浏览器直接请求 121。</Text>
            </Card>
          </Col>
        </Row>
      </div>
    </div>
  );
}
