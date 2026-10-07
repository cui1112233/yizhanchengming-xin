import React, { useCallback, useEffect, useState } from 'react'
import { Alert, Button, Card, Empty, Input, Layout, List, Select, Spin, Typography, Upload } from 'antd'
import { agentAttachmentContentURL, continueAgentProject, deleteAgentAttachment, getAgentCanvas, listAgentAttachments, listAgentCanvasVersions, listAgentExecutions, listAgentMessages, listAgentSkills, restoreAgentCanvas, saveAgentCanvas, uploadAgentAttachment } from './api.js'
import './agent-studio.css'

function attachmentError(error, fallback) {
  if (error?.status === 403) return '无权限查看或操作该项目附件。'
  return error?.message || fallback
}

function canvasText(document) {
  try { return JSON.stringify(document || { nodes: [], edges: [] }, null, 2) } catch { return '{\n  "nodes": [],\n  "edges": []\n}' }
}

export default function AgentCanvasPage() {
  const id = Number(new URLSearchParams(window.location.search).get('projectId'))
  const [text, setText] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)
  const [attachments, setAttachments] = useState([])
  const [attachmentsLoading, setAttachmentsLoading] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [canvas, setCanvas] = useState({ loading: true, revision: 0, document: '{\n  "nodes": [],\n  "edges": []\n}', saving: false })
  const [versions, setVersions] = useState([])
  const [skills, setSkills] = useState([])
  const [skillIDs, setSkillIDs] = useState([])
  const [messages,setMessages]=useState([])
  const [executions,setExecutions]=useState([])

  const refreshAttachments = useCallback(async () => {
    if (!id) return
    setAttachmentsLoading(true)
    try {
      const result = await listAgentAttachments(id)
      setAttachments(result.attachments || [])
    } catch (cause) {
      setError(attachmentError(cause, '读取附件失败'))
    } finally {
      setAttachmentsLoading(false)
    }
  }, [id])

  const refreshCanvas = useCallback(async () => {
    if (!id) return
    setCanvas((value) => ({ ...value, loading: true }))
    try {
      const [current, history] = await Promise.all([getAgentCanvas(id), listAgentCanvasVersions(id)])
      if (!current?.canvas) throw new Error('画布响应无效')
      setCanvas((value) => ({ ...value, loading: false, revision: current.canvas.revision, document: canvasText(current.canvas.document) }))
      setVersions(history?.versions || [])
    } catch (cause) {
      setCanvas((value) => ({ ...value, loading: false }))
      setError(attachmentError(cause, '读取画布失败'))
    }
  }, [id])

  const refreshConversation=useCallback(async()=>{try{const [m,e]=await Promise.all([listAgentMessages(id),listAgentExecutions(id)]);setMessages(m.messages||[]);setExecutions(e.executions||[])}catch(cause){setError(attachmentError(cause,'读取会话失败'))}},[id])
  useEffect(() => { void refreshAttachments(); void refreshCanvas(); void refreshConversation(); void listAgentSkills().then((x) => setSkills(x.skills || []) ).catch(() => {}) }, [refreshAttachments, refreshCanvas, refreshConversation])

  const send = async () => {
    setBusy(true)
    setError('')
    try {
      await continueAgentProject(id, { content: text, skillIds: skillIDs })
      setText('')
      await refreshConversation()
    } catch (cause) {
      setError(cause.message || '执行失败')
    } finally {
      setBusy(false)
    }
  }

  const saveCanvas = async () => {
    let document
    try { document = JSON.parse(canvas.document) } catch { setError('画布 JSON 格式无效，未保存。'); return }
    setCanvas((value) => ({ ...value, saving: true }))
    setError('')
    try {
      const result = await saveAgentCanvas(id, { revision: canvas.revision, document })
      setCanvas((value) => ({ ...value, revision: result.canvas.revision, document: canvasText(result.canvas.document) }))
      const history = await listAgentCanvasVersions(id)
      setVersions(history.versions || [])
    } catch (cause) {
      setError(cause?.code === 'canvas_revision_conflict' ? '画布已被其他会话更新，请刷新后再保存。' : attachmentError(cause, '保存画布失败'))
    } finally {
      setCanvas((value) => ({ ...value, saving: false }))
    }
  }

  const restoreCanvas = async (revision) => {
    setError('')
    try {
      const result = await restoreAgentCanvas(id, revision)
      setCanvas((value) => ({ ...value, revision: result.canvas.revision, document: canvasText(result.canvas.document) }))
      const history = await listAgentCanvasVersions(id)
      setVersions(history.versions || [])
    } catch (cause) {
      setError(attachmentError(cause, '恢复画布失败'))
    }
  }

  const upload = async ({ file }) => {
    setUploading(true)
    setError('')
    try {
      await uploadAgentAttachment(id, file)
      await refreshAttachments()
    } catch (cause) {
      setError(attachmentError(cause, '上传失败'))
    } finally {
      setUploading(false)
    }
  }

  const remove = async (attachmentID) => {
    setError('')
    try {
      await deleteAgentAttachment(id, attachmentID)
      await refreshAttachments()
    } catch (cause) {
      setError(attachmentError(cause, '删除失败'))
    }
  }

  if (!id) return <Alert type="warning" message="项目链接无效" />

  return <Layout className="agent-canvas">
    <Layout.Sider width={260}><Typography.Title level={5}>画布历史</Typography.Title>
      <List size="small" locale={{ emptyText: '暂无版本' }} dataSource={versions} renderItem={(version) => <List.Item actions={[<Button key="restore" size="small" onClick={() => void restoreCanvas(version.revision)}>恢复版本 {version.revision}</Button>]}><span>版本 {version.revision}</span></List.Item>} />
    </Layout.Sider>
    <Layout.Content>
      <Card title={`创作画布${canvas.loading ? '（加载中）' : ` · 画布版本 ${canvas.revision}`}`} extra={<Button loading={canvas.saving} onClick={() => void saveCanvas()}>保存画布</Button>}>
        {error && <Alert type="error" message={error} showIcon />}
        <Typography.Paragraph>画布保存在服务端；保存使用 revision 乐观锁，恢复历史会生成一个新的版本。</Typography.Paragraph>
        {canvas.loading ? <Spin aria-label="画布加载中" /> : <Input.TextArea aria-label="画布 JSON" value={canvas.document} onChange={(event) => setCanvas((value) => ({ ...value, document: event.target.value }))} autoSize={{ minRows: 16 }} />}
      </Card>
    </Layout.Content>
    <Layout.Sider width={360}>
      <Card title="项目对话">
        <List size="small" locale={{emptyText:'暂无消息'}} dataSource={messages} renderItem={m=><List.Item><Typography.Text strong>{m.role||m.Role}：</Typography.Text>{m.content||m.Content}</List.Item>}/>
        <List size="small" locale={{emptyText:null}} dataSource={executions.slice(0,3)} renderItem={e=><List.Item><Typography.Text type={e.status==='executor_unavailable'?'danger':undefined}>执行 {e.status||e.Status}{(e.errorMessage||e.ErrorMessage)?`：${e.errorMessage||e.ErrorMessage}`:''}</Typography.Text></List.Item>}/>
        <section className="agent-attachments" aria-label="项目附件">
          <Typography.Text strong>附件</Typography.Text>
          <Upload maxCount={1} showUploadList={false} customRequest={upload} disabled={uploading}>
            <Button loading={uploading}>上传附件</Button>
          </Upload>
          {attachmentsLoading ? <Spin size="small" aria-label="附件加载中" /> : attachments.length === 0 ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无附件" /> : <div className="agent-attachment-list">
            {attachments.map((attachment) => {
              const attachmentID = attachment.ID || attachment.id
              const filename = attachment.Filename || attachment.filename
              return <div className="agent-attachment-row" key={attachmentID}>
                <a href={agentAttachmentContentURL(id, attachmentID)}>{filename}</a>
                <Button size="small" onClick={() => void remove(attachmentID)}>删除</Button>
              </div>
            })}
          </div>}
        </section>
        <Select mode="multiple" value={skillIDs} onChange={setSkillIDs} options={skills.map((skill) => ({ value: skill.id || skill.ID, label: `${skill.name || skill.Name} v${skill.version || skill.Version}` }))} placeholder="选择用户技能" />
        <Input.TextArea value={text} onChange={(event) => setText(event.target.value)} placeholder="继续创作…" />
        <Button type="primary" loading={busy} disabled={!text.trim()} onClick={() => void send()}>发送</Button>
      </Card>
    </Layout.Sider>
  </Layout>
}
