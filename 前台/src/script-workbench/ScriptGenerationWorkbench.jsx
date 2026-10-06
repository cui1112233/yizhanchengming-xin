import React, { useEffect, useMemo, useRef, useState } from 'react'
import {
  Alert,
  Button,
  Card,
  Checkbox,
  Col,
  Divider,
  Drawer,
  Form,
  Input,
  Modal,
  Radio,
  Row,
  Select,
  Space,
  Tabs,
  Tag,
  Typography,
} from 'antd'
import * as api from '../api.js'
import {
  joinShotDocument,
  loadWorkbenchRecord,
  replaceInText,
  saveWorkbenchDraft,
  saveWorkbenchVersion,
  splitShotDocument,
  workbenchExportArtifact,
} from './scriptWorkbenchState.js'

const { TextArea } = Input

const OPENING_OPTIONS = [
  { label: '连续开头', value: 'continuous' },
  { label: '爆款开头', value: 'hook' },
  { label: '分段开头', value: 'segmented' },
]

const OUTPUT_OPTIONS = [
  { label: '画布模式', value: 'canvas' },
  { label: '剧本模式', value: 'script' },
  { label: '剧情模式', value: 'story' },
  { label: '分镜模式', value: 'shotlist' },
  { label: 'Q版模式', value: 'q' },
]

const STAGES = ['SCRIPT', 'HOOK', 'DIRECTOR', 'FINAL_PROMPT']

const defaultApiClient = {
  getBookGeneration: (...args) => api.getBookGeneration(...args),
  runBookGeneration: (...args) => api.runBookGeneration(...args),
}

function emptyEntity(kind, index) {
  return {
    id: kind + '-' + String(Date.now()) + '-' + String(index),
    kind,
    name: '',
    description: '',
    protagonist: false,
    referenceImages: [],
  }
}

function normalizeEntity(value, kind, index) {
  return {
    id: value?.id || kind + '-' + String(index + 1),
    kind,
    name: value?.name || '',
    description: value?.description || '',
    protagonist: Boolean(value?.protagonist),
    referenceImages: Array.isArray(value?.referenceImages) ? value.referenceImages : [],
  }
}

function safeStage(summary, stage) {
  return summary?.latest?.[stage] || null
}

function outputFromSummary(summary) {
  return summary?.editableOutput
    || safeStage(summary, 'SCRIPT')?.outputText
    || ''
}

function compiledPromptFromSummary(summary) {
  return summary?.compiledPrompt
    || safeStage(summary, 'FINAL_PROMPT')?.outputText
    || ''
}

function historyOutput(entry) {
  return entry?.editableOutput
    || entry?.latest?.SCRIPT?.outputText
    || ''
}

export default function ScriptGenerationWorkbench({
  projectId,
  bookId,
  bookTitle,
  apiClient = defaultApiClient,
  onGenerated,
  onRequestReferenceImage,
  onExport,
}) {
  const [loading, setLoading] = useState(true)
  const [working, setWorking] = useState(false)
  const [notice, setNotice] = useState(null)
  const [sourceText, setSourceText] = useState('')
  const [openingMode, setOpeningMode] = useState('continuous')
  const [outputMode, setOutputMode] = useState('canvas')
  const [directorMode, setDirectorMode] = useState('normal')
  const [characters, setCharacters] = useState([])
  const [scenes, setScenes] = useState([])
  const [constraints, setConstraints] = useState({
    visualPrefix: '',
    quality: '',
    pictureLimit: '',
    negativePrompt: '',
  })
  const [summary, setSummary] = useState(null)
  const [editorOutput, setEditorOutput] = useState('')
  const [undoStack, setUndoStack] = useState([])
  const [historyOpen, setHistoryOpen] = useState(false)
  const [replaceOpen, setReplaceOpen] = useState(false)
  const [findText, setFindText] = useState('')
  const [replacement, setReplacement] = useState('')
  const [localVersions, setLocalVersions] = useState([])
  const loadedRef = useRef(false)

  const draftSnapshot = useMemo(() => ({
    sourceText,
    openingMode,
    outputMode,
    directorMode,
    characters,
    scenes,
    constraints,
    editorOutput,
  }), [sourceText, openingMode, outputMode, directorMode, characters, scenes, constraints, editorOutput])

  const restoreDraft = (draft) => {
    if (!draft || typeof draft !== 'object') return
    setSourceText(draft.sourceText || '')
    setOpeningMode(draft.openingMode || 'continuous')
    setOutputMode(draft.outputMode || 'canvas')
    setDirectorMode(draft.directorMode || 'normal')
    setCharacters(Array.isArray(draft.characters) ? draft.characters.map((item, index) => normalizeEntity(item, 'character', index)) : [])
    setScenes(Array.isArray(draft.scenes) ? draft.scenes.map((item, index) => normalizeEntity(item, 'scene', index)) : [])
    setConstraints({
      visualPrefix: draft.constraints?.visualPrefix || '',
      quality: draft.constraints?.quality || '',
      pictureLimit: draft.constraints?.pictureLimit || '',
      negativePrompt: draft.constraints?.negativePrompt || '',
    })
    setEditorOutput(draft.editorOutput || '')
    setUndoStack([])
  }

  useEffect(() => {
    let active = true
    loadedRef.current = false
    setLoading(true)
    setNotice(null)
    const record = loadWorkbenchRecord(window.localStorage, projectId, bookId)
    setLocalVersions(record.versions)
    apiClient.getBookGeneration(projectId, bookId)
      .then((payload) => {
        if (!active) return
        setSummary(payload || null)
        if (record.draft) {
          restoreDraft(record.draft)
          setNotice({ type: 'info', message: '已恢复本机保存的剧本工作台草稿；服务端生成历史仍可在“历史恢复”中查看。' })
        } else {
          setSourceText(payload?.sourceText || '')
          setEditorOutput(outputFromSummary(payload))
        }
      })
      .catch((error) => {
        if (!active) return
        if (record.draft) restoreDraft(record.draft)
        setNotice({ type: 'error', message: error?.message || '读取剧本工作台失败' })
      })
      .finally(() => {
        if (active) {
          loadedRef.current = true
          setLoading(false)
        }
      })
    return () => { active = false }
  }, [projectId, bookId, apiClient])

  useEffect(() => {
    if (!loadedRef.current || loading) return
    saveWorkbenchDraft(window.localStorage, projectId, bookId, draftSnapshot)
  }, [draftSnapshot, loading, projectId, bookId])

  const updateEntity = (setter, index, patch) => {
    setter((current) => current.map((item, itemIndex) => itemIndex === index ? { ...item, ...patch } : item))
  }

  const removeEntity = (setter, index) => {
    setter((current) => current.filter((_, itemIndex) => itemIndex !== index))
  }

  const addReferenceImage = (index, raw) => {
    const refs = String(raw || '').split(/\n+/).map((item) => item.trim()).filter(Boolean)
    updateEntity(setCharacters, index, { referenceImages: refs.slice(0, 8) })
  }

  const requestReferenceImage = async (index) => {
    if (typeof onRequestReferenceImage !== 'function') {
      setNotice({ type: 'warning', message: '共享图片服务尚未注入。当前参考图 URL 只会作为文本元数据写入提示词，不会进行图片理解，也不会触发图片生成。' })
      return
    }
    try {
      const result = await onRequestReferenceImage({ projectId, bookId, character: characters[index] })
      if (result?.url) {
        const current = characters[index]?.referenceImages || []
        updateEntity(setCharacters, index, { referenceImages: [...current, result.url].slice(0, 8) })
      }
    } catch (error) {
      setNotice({ type: 'error', message: error?.message || '共享图片服务调用失败' })
    }
  }

  const addTextFile = async (event) => {
    const file = event.target.files?.[0]
    event.target.value = ''
    if (!file) return
    if (!file.name.toLowerCase().endsWith('.txt') && file.type !== 'text/plain') {
      setNotice({ type: 'error', message: '只支持 TXT 文本文件。' })
      return
    }
    try {
      const text = await file.text()
      setSourceText(text)
      setNotice({ type: 'success', message: 'TXT 原文已载入，可继续提取人物与场景。' })
    } catch {
      setNotice({ type: 'error', message: 'TXT 文件读取失败。' })
    }
  }

  const extractEntities = async () => {
    if (!sourceText.trim()) {
      setNotice({ type: 'warning', message: '请先输入原文或载入 TXT。' })
      return
    }
    setWorking(true)
    try {
      const payload = await apiClient.runBookGeneration(projectId, bookId, {
        workbench: true,
        action: 'extract',
        sourceText,
        openingMode,
        outputMode,
        directorMode,
        requestId: 'extract-' + String(Date.now()),
      })
      setCharacters((payload?.extraction?.characters || []).map((item, index) => normalizeEntity(item, 'character', index)))
      setScenes((payload?.extraction?.scenes || []).map((item, index) => normalizeEntity(item, 'scene', index)))
      setNotice({ type: 'success', message: '人物与场景已提取。请先检查、编辑并确认主角/参考图，再生成剧本。' })
    } catch (error) {
      setNotice({ type: 'error', message: error?.message || '人物与场景提取失败' })
    } finally {
      setWorking(false)
    }
  }

  const generate = async () => {
    if (!sourceText.trim()) {
      setNotice({ type: 'warning', message: '请先输入原文或载入 TXT。' })
      return
    }
    setWorking(true)
    try {
      const generated = await apiClient.runBookGeneration(projectId, bookId, {
        workbench: true,
        action: 'generate',
        force: true,
        sourceText,
        openingMode,
        outputMode,
        directorMode,
        hookEnabled: openingMode === 'hook',
        plotMode: outputMode === 'story',
        characters,
        scenes,
        constraints,
        matchAudio: false,
        requestId: 'script-workbench-' + String(Date.now()),
      })
      const refreshed = await apiClient.getBookGeneration(projectId, bookId)
      const nextOutput = outputFromSummary(refreshed) || outputFromSummary(generated)
      setSummary(refreshed)
      setUndoStack((current) => editorOutput ? [...current, editorOutput].slice(-30) : current)
      setEditorOutput(nextOutput)
      setNotice({ type: 'success', message: '生成链已返回并回读服务端历史。可编辑成品来自 SCRIPT；FINAL_PROMPT 仅作为后续模型编译输入，不参与画布拆卡。真实模型质量仍需集成环境验证。' })
      if (typeof onGenerated === 'function') onGenerated(refreshed)
    } catch (error) {
      setNotice({ type: 'error', message: error?.message || '剧本生成失败' })
    } finally {
      setWorking(false)
    }
  }

  const editOutput = (value) => {
    if (value !== editorOutput) setUndoStack((current) => [...current, editorOutput].slice(-30))
    setEditorOutput(value)
  }

  const undo = () => {
    setUndoStack((current) => {
      if (!current.length) return current
      setEditorOutput(current[current.length - 1])
      return current.slice(0, -1)
    })
  }

  const saveVersion = () => {
    const versions = saveWorkbenchVersion(window.localStorage, projectId, bookId, draftSnapshot)
    setLocalVersions(versions)
    setNotice({ type: 'success', message: '当前完整工作台快照仅保存到本机 localStorage。服务端 StageRun 历史目前只用于恢复可编辑输出；输入、人物/场景、模式和约束的完整服务端版本恢复仍需与 A 协调持久化接口。' })
  }

  const applyReplace = (all) => {
    if (!findText) return
    setUndoStack((current) => [...current, editorOutput].slice(-30))
    setEditorOutput(replaceInText(editorOutput, findText, replacement, all))
    if (!all) setReplaceOpen(false)
  }

  const exportText = () => {
    const artifact = workbenchExportArtifact({ editorOutput, sourceText, bookTitle, outputMode })
    if (typeof onExport === 'function') {
      onExport(artifact)
      return
    }
    const blob = new Blob([artifact.text], { type: 'text/plain;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const anchor = document.createElement('a')
    anchor.href = url
    anchor.download = artifact.filename
    document.body.appendChild(anchor)
    anchor.click()
    anchor.remove()
    URL.revokeObjectURL(url)
  }

  const shotDocument = useMemo(() => splitShotDocument(editorOutput), [editorOutput])
  const shotCards = shotDocument.cards
  const updateShot = (index, body) => {
    setUndoStack((current) => [...current, editorOutput].slice(-30))
    const cards = shotCards.map((card, cardIndex) => cardIndex === index ? { ...card, body } : card)
    setEditorOutput(joinShotDocument({ ...shotDocument, cards }))
  }

  const stageItems = STAGES.map((stage) => {
    const value = safeStage(summary, stage)
    return {
      key: stage,
      label: stage,
      children: (
        <div>
          <Space wrap style={{ marginBottom: 8 }}>
            <Tag>{value?.status || 'pending'}</Tag>
            {value?.promptKey && <Typography.Text type="secondary">{value.promptKey} v{value.promptVersion || '-'}</Typography.Text>}
          </Space>
          <pre className="script-workbench-stage-output">{value?.outputText || value?.errorMessage || '暂无 Stage 输出'}</pre>
        </div>
      ),
    }
  })

  const renderEntityCard = (entity, index, kind) => {
    const isCharacter = kind === 'character'
    const setter = isCharacter ? setCharacters : setScenes
    return (
      <Card
        size="small"
        key={entity.id || kind + '-' + String(index)}
        title={(isCharacter ? '人物 ' : '场景 ') + String(index + 1)}
        extra={<Button danger type="text" onClick={() => removeEntity(setter, index)}>删除</Button>}
        className="script-entity-card"
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          <Input
            aria-label={(isCharacter ? '人物名称' : '场景名称') + String(index + 1)}
            value={entity.name}
            placeholder={isCharacter ? '人物名称' : '场景名称'}
            onChange={(event) => updateEntity(setter, index, { name: event.target.value })}
          />
          <TextArea
            aria-label={(isCharacter ? '人物描述' : '场景描述') + String(index + 1)}
            autoSize={{ minRows: 2, maxRows: 5 }}
            value={entity.description}
            placeholder={isCharacter ? '身份、外形、关系与稳定辨识特征' : '空间、时代、陈设与稳定环境事实'}
            onChange={(event) => updateEntity(setter, index, { description: event.target.value })}
          />
          {isCharacter && (
            <>
              <Checkbox
                checked={entity.protagonist}
                onChange={(event) => updateEntity(setter, index, { protagonist: event.target.checked })}
              >
                设为主角
              </Checkbox>
              <TextArea
                aria-label={'人物参考图' + String(index + 1)}
                autoSize={{ minRows: 2, maxRows: 4 }}
                value={(entity.referenceImages || []).join('\n')}
                placeholder="每行一个参考图 URL；最多 8 张"
                onChange={(event) => addReferenceImage(index, event.target.value)}
              />
              <Typography.Text type="secondary">参考图 URL 仅作为文本元数据进入提示词；当前文本生成链不会读取图片像素。</Typography.Text>
              <Button onClick={() => void requestReferenceImage(index)}>
                {typeof onRequestReferenceImage === 'function' ? '通过共享图片服务补参考图' : '共享图片服务未注入'}
              </Button>
            </>
          )}
        </Space>
      </Card>
    )
  }

  if (loading) {
    return <Card loading title="剧本生成工作台" />
  }

  return (
    <div className="script-workbench">
      <Alert
        showIcon
        type="info"
        message="模块边界"
        description="文本生成复用现有 Provider。参考图 URL 当前仅作为提示词文本元数据，不代表图片理解；共享图片按钮尚未注入真实图片服务。配音/音频生成也未接入本工作台。本模块不维护供应商、账号或模型密钥，测试不会调用真实付费模型。"
        style={{ marginBottom: 16 }}
      />
      {notice && <Alert showIcon closable type={notice.type} message={notice.message} onClose={() => setNotice(null)} style={{ marginBottom: 16 }} />}

      <Row gutter={[16, 16]}>
        <Col xs={24} xl={9}>
          <Card title="1. 原文 / TXT" className="script-workbench-panel">
            <TextArea
              aria-label="剧本原文"
              value={sourceText}
              onChange={(event) => setSourceText(event.target.value)}
              autoSize={{ minRows: 14, maxRows: 28 }}
              placeholder="粘贴小说原文，或载入 TXT"
            />
            <Space wrap style={{ marginTop: 12 }}>
              <label className="ant-btn ant-btn-default script-file-button">
                载入 TXT
                <input type="file" accept=".txt,text/plain" onChange={addTextFile} />
              </label>
              <Button type="primary" loading={working} onClick={() => void extractEntities()}>提取人物与场景</Button>
            </Space>
          </Card>
        </Col>

        <Col xs={24} xl={15}>
          <Card title="2. 人物 / 场景 / 主角 / 参考图" className="script-workbench-panel">
            <Typography.Title level={5}>人物</Typography.Title>
            <div className="script-entity-grid">
              {characters.map((entity, index) => renderEntityCard(entity, index, 'character'))}
            </div>
            <Button onClick={() => setCharacters((current) => [...current, emptyEntity('character', current.length)])}>添加人物</Button>
            <Divider />
            <Typography.Title level={5}>场景</Typography.Title>
            <div className="script-entity-grid">
              {scenes.map((entity, index) => renderEntityCard(entity, index, 'scene'))}
            </div>
            <Button onClick={() => setScenes((current) => [...current, emptyEntity('scene', current.length)])}>添加场景</Button>
          </Card>
        </Col>
      </Row>

      <Card title="3. 生成设置" className="script-workbench-panel">
        <Row gutter={[16, 12]}>
          <Col xs={24} lg={10}>
            <Form.Item label="开头模式">
              <Radio.Group
                optionType="button"
                buttonStyle="solid"
                value={openingMode}
                options={OPENING_OPTIONS}
                onChange={(event) => setOpeningMode(event.target.value)}
              />
            </Form.Item>
          </Col>
          <Col xs={24} lg={7}>
            <Form.Item label="输出模式">
              <Select aria-label="输出模式" value={outputMode} options={OUTPUT_OPTIONS} onChange={setOutputMode} />
            </Form.Item>
          </Col>
          <Col xs={24} lg={7}>
            <Form.Item label="Director">
              <Radio.Group
                value={directorMode}
                options={[{ label: 'Normal', value: 'normal' }, { label: 'H3', value: 'h3' }]}
                onChange={(event) => setDirectorMode(event.target.value)}
              />
            </Form.Item>
          </Col>
        </Row>
        <Divider orientation="left">约束设置</Divider>
        <Row gutter={[12, 12]}>
          <Col xs={24} md={12}><TextArea aria-label="画面前缀" placeholder="画面前缀" value={constraints.visualPrefix} onChange={(event) => setConstraints((current) => ({ ...current, visualPrefix: event.target.value }))} /></Col>
          <Col xs={24} md={12}><TextArea aria-label="画质约束" placeholder="画质约束" value={constraints.quality} onChange={(event) => setConstraints((current) => ({ ...current, quality: event.target.value }))} /></Col>
          <Col xs={24} md={12}><TextArea aria-label="画面限制" placeholder="画面限制" value={constraints.pictureLimit} onChange={(event) => setConstraints((current) => ({ ...current, pictureLimit: event.target.value }))} /></Col>
          <Col xs={24} md={12}><TextArea aria-label="负面提示词" placeholder="负面提示词" value={constraints.negativePrompt} onChange={(event) => setConstraints((current) => ({ ...current, negativePrompt: event.target.value }))} /></Col>
        </Row>
        <Space wrap style={{ marginTop: 16 }}>
          <Button type="primary" loading={working} onClick={() => void generate()}>生成剧本与分镜</Button>
          <Button onClick={saveVersion}>保存当前版本</Button>
          <Button onClick={() => setHistoryOpen(true)}>历史恢复</Button>
        </Space>
      </Card>

      <Card
        title="4. 分镜 / 成品编辑"
        className="script-workbench-panel"
        extra={(
          <Space wrap>
            <Button disabled={!undoStack.length} onClick={undo}>撤销</Button>
            <Button disabled={!editorOutput} onClick={() => setReplaceOpen(true)}>查找替换</Button>
            <Button disabled={!editorOutput && !sourceText} onClick={exportText}>导出 TXT</Button>
          </Space>
        )}
      >
        {outputMode === 'canvas' && shotCards.length > 0 ? (
          <div className="script-canvas-grid">
            {shotCards.map((card, index) => (
              <Card size="small" title={card.title} key={card.title + '-' + String(index)} className="script-shot-card">
                <TextArea
                  aria-label={'分镜编辑' + String(index + 1)}
                  value={card.body}
                  autoSize={{ minRows: 8, maxRows: 24 }}
                  onChange={(event) => updateShot(index, event.target.value)}
                />
              </Card>
            ))}
          </div>
        ) : (
          <TextArea
            aria-label="生成结果编辑"
            value={editorOutput}
            autoSize={{ minRows: 18, maxRows: 40 }}
            placeholder="生成后可在这里继续编辑"
            onChange={(event) => editOutput(event.target.value)}
          />
        )}
        <Divider orientation="left">Stage 原始结果 / 编译输入</Divider>
        <Alert
          type="warning"
          showIcon
          message="编辑区与 FINAL_PROMPT 已分离"
          description="编辑区只使用 SCRIPT 可编辑成品；FINAL_PROMPT 是系统预设、SCRIPT、HOOK、DIRECTOR 和处理规则的确定性编译输入，仅用于后续模型链路，不参与分镜拆卡。"
          style={{ marginBottom: 12 }}
        />
        {compiledPromptFromSummary(summary) && (
          <Typography.Paragraph type="secondary" ellipsis={{ rows: 3, expandable: true, symbol: '展开编译输入' }}>
            {compiledPromptFromSummary(summary)}
          </Typography.Paragraph>
        )}
        <Tabs items={stageItems} />
      </Card>

      <Drawer title="历史恢复" width={720} open={historyOpen} onClose={() => setHistoryOpen(false)}>
        <Alert
          type="warning"
          showIcon
          message="历史恢复范围"
          description="本机手工版本保存在 localStorage，可恢复输入、实体、模式、约束和编辑稿；服务端生成历史当前只恢复可编辑输出。服务端完整版本恢复仍待与 A 协调持久化接口。"
          style={{ marginBottom: 16 }}
        />
        <Typography.Title level={5}>本机手工版本</Typography.Title>
        {localVersions.length ? localVersions.map((entry) => (
          <Card size="small" key={entry.id} style={{ marginBottom: 10 }}>
            <Space direction="vertical">
              <Typography.Text>{entry.label || '手工保存'} · {entry.savedAt || ''}</Typography.Text>
              <Button onClick={() => { restoreDraft(entry.draft); setHistoryOpen(false) }}>恢复此版本</Button>
            </Space>
          </Card>
        )) : <Typography.Text type="secondary">暂无本机手工版本</Typography.Text>}
        <Divider />
        <Typography.Title level={5}>服务端生成历史</Typography.Title>
        {(summary?.history || []).length ? summary.history.map((entry) => (
          <Card size="small" key={entry.run?.id} style={{ marginBottom: 10 }}>
            <Space direction="vertical" style={{ width: '100%' }}>
              <Typography.Text>Run #{entry.run?.id} · {entry.run?.status}</Typography.Text>
              <Typography.Paragraph ellipsis={{ rows: 3 }}>{historyOutput(entry) || '无输出'}</Typography.Paragraph>
              <Button disabled={!historyOutput(entry)} onClick={() => {
                setUndoStack((current) => editorOutput ? [...current, editorOutput].slice(-30) : current)
                setEditorOutput(historyOutput(entry))
                setHistoryOpen(false)
              }}>恢复生成结果</Button>
            </Space>
          </Card>
        )) : <Typography.Text type="secondary">暂无服务端生成历史</Typography.Text>}
      </Drawer>

      <Modal
        title="查找替换"
        open={replaceOpen}
        onCancel={() => setReplaceOpen(false)}
        footer={null}
      >
        <Form layout="vertical">
          <Form.Item label="查找内容"><Input value={findText} onChange={(event) => setFindText(event.target.value)} /></Form.Item>
          <Form.Item label="替换为"><Input value={replacement} onChange={(event) => setReplacement(event.target.value)} /></Form.Item>
          <Space>
            <Button disabled={!findText} onClick={() => applyReplace(false)}>替换当前</Button>
            <Button type="primary" disabled={!findText} onClick={() => applyReplace(true)}>全部替换</Button>
          </Space>
        </Form>
      </Modal>
    </div>
  )
}
