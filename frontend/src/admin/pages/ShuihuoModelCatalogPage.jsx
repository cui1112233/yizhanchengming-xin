import { Button, Drawer, Form, Input, Popconfirm, Select, Switch, Table, Tag, Typography, message } from 'antd';
import { useEffect, useState } from 'react';
import { createAdminModel, deleteAdminModel, listAdminModels, setAdminModelEnabled, testAdminModel, updateAdminModel } from '../../shared/api/shuihuoProduction';

const adapters = [
  { value: 'text_completion', label: '文本分析', kind: 'text' },
  { value: 'jimeng_image', label: '即梦生图', kind: 'image' },
  { value: 'vidu_image_to_video', label: 'Vidu 图生视频', kind: 'video' },
  { value: 'yd_video', label: 'YD2.0 Mini 图生视频', kind: 'video' },
  { value: 'local_executor_video', label: '本地豆包执行器', kind: 'video' },
  { value: 'generic_http', label: '通用 HTTP 配音（仅管理员）', kind: 'audio' }
];

export function ShuihuoModelCatalogPage() {
  const [models, setModels] = useState([]);
  const [open, setOpen] = useState(false);
  const [editingModel, setEditingModel] = useState(null);
  const [saving, setSaving] = useState(false);
  const [testingId, setTestingId] = useState('');
  const [form] = Form.useForm();
  const adapterKind = Form.useWatch('adapterKind', form);
  const isGenericAdapter = adapterKind === 'generic_http';
  const isYDVideoAdapter = adapterKind === 'yd_video';
  const isLocalExecutorAdapter = adapterKind === 'local_executor_video';

  const refresh = async () => {
    try {
      const result = await listAdminModels();
      setModels(result.models || []);
    } catch (error) {
      message.error(error.message || '读取模型目录失败');
    }
  };

  useEffect(() => { refresh(); }, []);

  function openCreate() {
    setEditingModel(null);
    form.resetFields();
    setOpen(true);
  }

  function openEdit(model) {
    setEditingModel(model);
    form.setFieldsValue({ ...model, modelId: model.modelId || model.id, endpoint: model.endpoint || model.baseUrl || '', credentialRef: '' });
    setOpen(true);
  }

  async function submit() {
    try {
      const values = await form.validateFields();
      setSaving(true);
      const payload = {
        ...values,
        parameterSchema: values.parameterSchema || '{}',
        credentialRef: values.credentialRef || '',
        endpoint: values.endpoint || '',
        requestTemplate: values.requestTemplate || '',
        responseMapping: values.responseMapping || ''
      };
      if (editingModel) {
        if (!payload.credentialRef) delete payload.credentialRef;
        await updateAdminModel(editingModel.modelId || editingModel.id, payload);
      } else {
        await createAdminModel(payload);
      }
      setOpen(false);
      form.resetFields();
      setEditingModel(null);
      await refresh();
      message.success('模型已保存');
    } catch (error) {
      if (!error?.errorFields) message.error(error.message || '保存模型失败');
    } finally {
      setSaving(false);
    }
  }

  async function remove(model) {
    try {
      await deleteAdminModel(model.modelId || model.id);
      await refresh();
      message.success('模型已停用并移除');
    } catch (error) {
      message.error(error.message || '移除模型失败');
    }
  }

  async function setEnabled(model, enabled) {
    const id = model.modelId || model.id;
    try {
      await setAdminModelEnabled(id, enabled);
      await refresh();
      message.success(enabled ? '模型已启用' : '模型已停用');
    } catch (error) {
      message.error(error.message || (enabled ? '启用模型失败' : '停用模型失败'));
    }
  }

  async function test(model) {
    const id = model.modelId || model.id;
    setTestingId(id);
    try {
      const result = await testAdminModel(id);
      message.success(result.message || '模型连接成功');
    } catch (error) {
      message.error(error.message || '模型连接测试失败');
    } finally {
      setTestingId('');
    }
  }

  return <>
    <Typography.Title level={3}>水货生产模型</Typography.Title>
    <Typography.Paragraph>管理员在此配置文本分析、即梦生图、图生视频和配音。密钥只保留服务器端的引用；运行地址、请求模板和返回映射不会在模型列表或用户工作台返回。</Typography.Paragraph>
    <Button type="primary" onClick={openCreate}>新增模型</Button>
    <Table rowKey="id" style={{ marginTop: 16 }} dataSource={models} pagination={false} columns={[
      { title: '名称', dataIndex: 'name' },
      { title: '能力', dataIndex: 'kind' },
      { title: '适配器', dataIndex: 'adapterKind' },
      { title: '状态', dataIndex: 'enabled', render: value => <Tag color={value ? 'green' : 'default'}>{value ? '已启用' : '已停用'}</Tag> },
      { title: '凭据引用', dataIndex: 'credentialConfigured', render: value => <Tag color={value ? 'green' : 'gold'}>{value ? '已登记' : '未登记'}</Tag> },
      { title: '运行配置', dataIndex: 'providerConfigured', render: value => <Tag color={value ? 'green' : 'gold'}>{value ? '模型记录完整' : '待完善'}</Tag> },
      { title: '操作', key: 'actions', render: (_, model) => <><Button type="link" loading={testingId === (model.modelId || model.id)} onClick={() => test(model)}>测试连接</Button><Button type="link" onClick={() => openEdit(model)}>编辑</Button><Popconfirm title={model.enabled ? '确认停用这个模型？' : '确认启用这个模型？'} onConfirm={() => setEnabled(model, !model.enabled)}><Button type="link">{model.enabled ? '停用' : '启用'}</Button></Popconfirm><Popconfirm title="确认移除这个模型？" onConfirm={() => remove(model)}><Button type="link" danger>移除</Button></Popconfirm></> }
    ]} />
    <Drawer title={editingModel ? '编辑模型' : '新增模型'} open={open} onClose={() => { setOpen(false); setEditingModel(null); }} width={560} extra={<Button type="primary" loading={saving} onClick={submit}>保存</Button>}>
      <Form form={form} layout="vertical" initialValues={{ kind: 'image', adapterKind: 'jimeng_image', enabled: false, parameterSchema: '{}' }}>
        <Form.Item label="模型标识" name="modelId" rules={[{ required: true, message: '请填写模型标识' }, { pattern: /^[a-z0-9]+(?:-[a-z0-9]+)*$/, message: '请使用小写 kebab-case，例如 video-vidu-admin' }]}><Input placeholder="例如 video-vidu-admin" /></Form.Item>
        <Form.Item label="名称" name="name" rules={[{ required: true, message: '请填写模型名称' }]}><Input /></Form.Item>
        <Form.Item label="适配器" name="adapterKind"><Select options={adapters} onChange={nextAdapterKind => form.setFieldValue('kind', adapters.find(item => item.value === nextAdapterKind)?.kind)} /></Form.Item>
        <Form.Item name="kind" hidden><Input /></Form.Item>
        <Form.Item label="启用" name="enabled" valuePropName="checked"><Switch /></Form.Item>
        {adapterKind !== 'local_executor_video' ? <Form.Item label="服务地址" name="endpoint" rules={[{ required: true, message: '请填写服务地址' }]} extra="文本和图片模型支持 OpenAI 兼容地址；视频/配音按适配器协议填写"><Input placeholder="例如 https://provider.example.com/v1" /></Form.Item> : null}
        <Form.Item label="公开参数 Schema" name="parameterSchema"><Input.TextArea rows={3} /></Form.Item>
        <Form.Item label="密钥引用" name="credentialRef"><Input.Password placeholder="例如 IMAGE_PROVIDER_TOKEN，仅写入不回显" /></Form.Item>
        {isYDVideoAdapter ? <>
          <Typography.Paragraph type="secondary">固定服务商：YD2.0 Mini 图生视频。服务端按适配器内置协议调用，固定 1 秒、720p，仅支持 9:16 或 16:9。</Typography.Paragraph>
        </> : null}
        {isLocalExecutorAdapter ? <Typography.Paragraph type="secondary">本地执行器会从平台领取任务；豆包账号登录态仅保存在用户自己的电脑。本模型不需要填写服务器密钥或请求地址。</Typography.Paragraph> : null}
        {isGenericAdapter ? <>
          <Form.Item label="服务端请求地址" name="endpoint" rules={[{ required: true, message: '请填写 HTTPS 配音接口地址' }]}><Input placeholder="https://provider.example.com/v1/audio/speech" /></Form.Item>
          <Form.Item label="请求模板" name="requestTemplate" rules={[{ required: true, message: '请填写请求模板 JSON' }]} extra="可用占位符：{{prompt}}、{{voice}}、{{speech_rate}}、{{pitch}}、{{credential}}"><Input.TextArea rows={7} placeholder={'{"method":"POST","headers":{"Authorization":"Bearer {{credential}}"},"body":{"text":"{{prompt}}","voice":"{{voice}}","rate":"{{speech_rate}}","pitch":"{{pitch}}"}}'} /></Form.Item>
          <Form.Item label="返回映射" name="responseMapping" rules={[{ required: true, message: '请填写返回映射 JSON' }]} extra={'填写音频 URL 的 JSON 路径，例如 {"resultUrl":"data.url"}'}><Input.TextArea rows={3} placeholder={'{"resultUrl":"data.url"}'} /></Form.Item>
        </> : null}
      </Form>
    </Drawer>
  </>;
}
