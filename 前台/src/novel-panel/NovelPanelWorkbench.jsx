import React, { useMemo, useState } from 'react'
import { Alert, Button, Card, Col, Drawer, Empty, Input, List, Radio, Row, Select, Space, Switch, Tag, Typography } from 'antd'
import './NovelPanelWorkbench.css'

const { Text, Title } = Typography
const { TextArea } = Input

const defaults = (projectId) => ({
  projectId, revision: 0, mode: 'normal', originalText: '', processedText: '',
  textProcessing: { trimLineWhitespace: true, collapseBlankLines: true, normalizeFullWidthSpaces: true },
  forcedRoster: '', characters: [], relationships: [], contentType: '小说推文', unifiedStyle: '', density: 'standard',
  caseLearning: { enabled: false, material: '' },
  instructions: { generationRules: '', mustCoverDetails: '', shotRhythmRequirements: '', negativeInstructions: '' },
  shots: [],
})

function normalize(projectId, value = {}) {
  const base = defaults(projectId)
  return {
    ...base, ...value, projectId: Number(value.projectId || projectId || 0),
    textProcessing: { ...base.textProcessing, ...(value.textProcessing || {}) },
    caseLearning: { ...base.caseLearning, ...(value.caseLearning || {}) },
    instructions: { ...base.instructions, ...(value.instructions || {}) },
    characters: Array.isArray(value.characters) ? value.characters : [],
    relationships: Array.isArray(value.relationships) ? value.relationships : [],
    shots: Array.isArray(value.shots) ? value.shots : [],
  }
}

function previewText(text, settings) {
  let lines = String(text || '').replace(/\r\n?/g, '\n').split('\n')
  if (settings.normalizeFullWidthSpaces) lines = lines.map((line) => line.replace(/　/g, ' '))
  if (settings.trimLineWhitespace) lines = lines.map((line) => line.trim())
  if (settings.collapseBlankLines) lines = lines.filter((line, index) => line.trim() || (index > 0 && lines[index - 1].trim()))
  return lines.join('\n').trim()
}

const errText = (error) => error?.status === 409 || error?.code === 'NOVEL_PANEL_CONFLICT'
  ? '保存冲突：记录已更新，请重新载入后再保存。'
  : (error?.message || '操作失败')

export default function NovelPanelWorkbench({ projectId, initialWorkspace, api = {}, sharedServices = {} }) {
  const [ws, setWs] = useState(() => normalize(projectId, initialWorkspace))
  const [notice, setNotice] = useState(null)
  const [busy, setBusy] = useState('')
  const [note, setNote] = useState('')
  const [historyOpen, setHistoryOpen] = useState(false)
  const [history, setHistory] = useState([])
  const [relation, setRelation] = useState({ fromId: '', toId: '', type: '' })
  const preview = useMemo(() => previewText(ws.originalText, ws.textProcessing), [ws.originalText, ws.textProcessing])
  const patch = (next) => setWs((current) => ({ ...current, ...next }))
  const patchNested = (key, next) => setWs((current) => ({ ...current, [key]: { ...current[key], ...next } }))

  async function save() {
    if (!api.saveWorkspace) return setNotice({ type: 'warning', message: '保存适配器等待总集成注册。' })
    setBusy('save'); setNotice(null)
    try {
      const result = await api.saveWorkspace(ws.projectId, { workspace: ws, expectedRevision: Number(ws.revision || 0), note })
      setWs(normalize(ws.projectId, result?.workspace || result)); setNote('')
      setNotice({ type: 'success', message: '已保存并生成可恢复记录。' })
    } catch (error) { setNotice({ type: 'error', message: errText(error) }) } finally { setBusy('') }
  }

  async function openHistory() {
    setHistoryOpen(true)
    if (!api.listHistory) return
    setBusy('history')
    try { const result = await api.listHistory(ws.projectId, 50); setHistory(Array.isArray(result) ? result : (result?.items || [])) }
    catch (error) { setNotice({ type: 'error', message: errText(error) }) } finally { setBusy('') }
  }

  async function restore(item) {
    if (!api.restoreHistory) return setNotice({ type: 'warning', message: '恢复适配器等待总集成注册。' })
    setBusy('restore')
    try {
      const result = await api.restoreHistory(ws.projectId, { historyId: item.id, expectedRevision: Number(ws.revision || 0) })
      setWs(normalize(ws.projectId, result?.workspace || result)); setHistoryOpen(false)
      setNotice({ type: 'success', message: '已恢复，并追加恢复记录。' })
    } catch (error) { setNotice({ type: 'error', message: errText(error) }) } finally { setBusy('') }
  }

  async function requestStoryboard() {
    if (!api.requestStoryboard) return setNotice({ type: 'warning', message: '共享 Director 适配器等待总集成注册。' })
    setBusy('storyboard')
    try {
      const result = await api.requestStoryboard(ws.projectId, ws)
      const shots = Array.isArray(result) ? result : result?.shots
      if (!Array.isArray(shots)) throw new Error('共享 Director 未返回 shots')
      patch({ shots }); setNotice({ type: 'success', message: '已载入共享 Director 分镜草稿。' })
    } catch (error) { setNotice({ type: 'error', message: errText(error) }) } finally { setBusy('') }
  }

  const characterOptions = ws.characters.map((item) => ({ value: item.id, label: item.displayName || item.baseName || item.id }))
  const updateCharacter = (index, key, value) => setWs((current) => ({ ...current, characters: current.characters.map((item, i) => i === index ? { ...item, [key]: value } : item) }))
  const updateShot = (index, key, value) => setWs((current) => ({ ...current, shots: current.shots.map((item, i) => i === index ? { ...item, [key]: value } : item) }))
  const shared = [['配音', sharedServices.openAudio], ['人物资产', sharedServices.openAsset, 'characters'], ['场景资产', sharedServices.openAsset, 'scenes'], ['道具资产', sharedServices.openAsset, 'props'], ['图片能力', sharedServices.openImages]]

  return <div className="novel-panel-workbench">
    <div className="novel-panel-toolbar"><div><Title level={3}>小说面板</Title><Text type="secondary">项目 #{ws.projectId || '-'} · 修订 {ws.revision || 0}</Text></div><Space><Button onClick={openHistory}>保存记录与恢复</Button><Button type="primary" loading={busy === 'save'} onClick={save}>保存小说面板</Button></Space></div>
    {notice && <Alert className="novel-panel-notice" showIcon type={notice.type} message={notice.message} />}
    <Row gutter={[16, 16]}>
      <Col xs={24} xl={15}>
        <Card title="模式、原文与画面" className="novel-panel-card"><Space direction="vertical" style={{ width: '100%' }}>
          <Radio.Group aria-label="制作模式" value={ws.mode} onChange={(e) => patch({ mode: e.target.value })} options={[{ label: '普通模式', value: 'normal' }, { label: '精品带图模式', value: 'premium_illustrated' }]} />
          <TextArea aria-label="整段小说原文" value={ws.originalText} onChange={(e) => patch({ originalText: e.target.value })} autoSize={{ minRows: 10, maxRows: 22 }} />
          <Input aria-label="保存说明" value={note} onChange={(e) => setNote(e.target.value)} placeholder="保存说明（可选）" />
          <Space><Button onClick={requestStoryboard} loading={busy === 'storyboard'}>请求共享 Director 分镜</Button><Button onClick={() => patch({ shots: [...ws.shots, { id: `shot_${Date.now()}`, sourceIndex: 1, sourceBasis: '', visual: '', durationSec: 0 }] })}>新增手工分镜</Button></Space>
          {!ws.shots.length ? <Empty description="尚无分镜" /> : ws.shots.map((shot, index) => <Card size="small" key={shot.id || index} title={`分镜 ${index + 1}`}><Space direction="vertical" style={{ width: '100%' }}>
            <Input aria-label={`分镜${index + 1}原文依据`} value={shot.sourceBasis || ''} onChange={(e) => updateShot(index, 'sourceBasis', e.target.value)} placeholder="source_basis：必须来自对应原文" />
            <TextArea aria-label={`分镜${index + 1}画面`} value={shot.visual || ''} onChange={(e) => updateShot(index, 'visual', e.target.value)} placeholder="画面" />
            <Input aria-label={`分镜${index + 1}镜头`} value={shot.camera || ''} onChange={(e) => updateShot(index, 'camera', e.target.value)} placeholder="镜头/机位" />
          </Space></Card>)}
        </Space></Card>
      </Col>
      <Col xs={24} xl={9}>
        <Card title="文本处理" className="novel-panel-card"><Space direction="vertical" style={{ width: '100%' }}>
          {[['trimLineWhitespace', '清理每行首尾空白'], ['collapseBlankLines', '合并连续空行'], ['normalizeFullWidthSpaces', '全角空格转普通空格']].map(([key, label]) => <div className="novel-panel-switch-row" key={key}><Text>{label}</Text><Switch aria-label={label} checked={ws.textProcessing[key]} onChange={(value) => patchNested('textProcessing', { [key]: value })} /></div>)}
          <TextArea aria-label="文本处理预览" value={preview} readOnly autoSize={{ minRows: 4, maxRows: 8 }} />
        </Space></Card>
        <Card title="内容类型与统一风格" className="novel-panel-card"><Space direction="vertical" style={{ width: '100%' }}><Select aria-label="内容类型" value={ws.contentType} onChange={(value) => patch({ contentType: value })} options={['小说推文', '故事视频', '短剧化', '自定义'].map((value) => ({ value }))} /><TextArea aria-label="统一风格" value={ws.unifiedStyle} onChange={(e) => patch({ unifiedStyle: e.target.value })} placeholder="精品带图模式必须填写；真实图片生成走共享服务" /></Space></Card>
        <Card title="人物强制名单与人物卡" className="novel-panel-card"><TextArea aria-label="人物强制名单" value={ws.forcedRoster} onChange={(e) => patch({ forcedRoster: e.target.value })} placeholder="沈清月（青年）、顾川（青年）" />{ws.characters.map((item, index) => <Card size="small" key={item.id || index} title={item.displayName || item.baseName}><Input aria-label={`${item.baseName || '人物'}性别`} value={item.gender || ''} onChange={(e) => updateCharacter(index, 'gender', e.target.value)} /><TextArea aria-label={`${item.baseName || '人物'}外形`} value={item.appearance || ''} onChange={(e) => updateCharacter(index, 'appearance', e.target.value)} /><TextArea aria-label={`${item.baseName || '人物'}人物卡备注`} value={item.cardNote || ''} onChange={(e) => updateCharacter(index, 'cardNote', e.target.value)} />{item.assetRefs?.map((ref) => <Tag key={ref}>{ref}</Tag>)}</Card>)}</Card>
        <Card title="人物关系" className="novel-panel-card"><Space direction="vertical" style={{ width: '100%' }}><Select aria-label="关系起点人物" value={relation.fromId || undefined} onChange={(value) => setRelation({ ...relation, fromId: value })} options={characterOptions} /><Select aria-label="关系终点人物" value={relation.toId || undefined} onChange={(value) => setRelation({ ...relation, toId: value })} options={characterOptions} /><Input aria-label="关系类型" value={relation.type} onChange={(e) => setRelation({ ...relation, type: e.target.value })} /><Button onClick={() => { if (relation.fromId && relation.toId && relation.fromId !== relation.toId && relation.type.trim()) { patch({ relationships: [...ws.relationships, { id: `rel_${Date.now()}`, ...relation, type: relation.type.trim() }] }); setRelation({ fromId: '', toId: '', type: '' }) } }}>添加人物关系</Button><Text type="secondary">已配置 {ws.relationships.length} 条关系</Text></Space></Card>
        <Card title="密度、案例学习与指令设置" className="novel-panel-card"><Space direction="vertical" style={{ width: '100%' }}>
          <Select aria-label="分镜密度" value={ws.density} onChange={(value) => patch({ density: value })} options={[{ value: 'compact', label: '精简' }, { value: 'standard', label: '标准' }, { value: 'detailed', label: '较细' }]} />
          <div className="novel-panel-switch-row"><Text>启用案例学习</Text><Switch aria-label="启用案例学习" checked={ws.caseLearning.enabled} onChange={(enabled) => patchNested('caseLearning', { enabled })} /></div>
          <TextArea aria-label="案例学习素材" disabled={!ws.caseLearning.enabled} value={ws.caseLearning.material} onChange={(e) => patchNested('caseLearning', { material: e.target.value })} />
          {[["generationRules", '生成规则'], ['mustCoverDetails', '必须覆盖细节'], ['shotRhythmRequirements', '镜头节奏要求'], ['negativeInstructions', '负面指令']].map(([key, label]) => <TextArea key={key} aria-label={label} value={ws.instructions[key]} onChange={(e) => patchNested('instructions', { [key]: e.target.value })} />)}
        </Space></Card>
        <Card title="共享能力" className="novel-panel-card"><Space wrap>{shared.map(([label, handler, kind]) => <Button key={label} disabled={!handler} onClick={() => handler?.(kind)}>{label}</Button>)}</Space><div className="novel-panel-shared-note"><Text type="secondary">配音、人物/场景/道具资产、图片能力全部复用共享服务；此处没有 Provider/API Key 管理。</Text></div></Card>
      </Col>
    </Row>
    <Drawer title="保存记录与恢复" open={historyOpen} onClose={() => setHistoryOpen(false)}><List dataSource={history} locale={{ emptyText: '暂无保存记录' }} renderItem={(item) => <List.Item actions={[<Button key="restore" onClick={() => restore(item)} loading={busy === 'restore'}>恢复</Button>]}><List.Item.Meta title={`修订 ${item.revision ?? '-'}${item.note ? ` · ${item.note}` : ''}`} description={item.summary?.sourcePreview || item.createdAt || ''} /></List.Item>} /></Drawer>
  </div>
}
