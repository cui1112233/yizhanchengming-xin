import { Alert, Button, Empty, Progress, Spin, Upload } from 'antd'
import { FileAudio, FileImage, FileVideo, UploadCloud } from 'lucide-react'
import { useCallback, useEffect, useState } from 'react'

import { apiRequest } from '../../shared/api/client.js'
import { materialErrorMessage, materialSummary } from './materialModel.js'

const { Dragger } = Upload

function MediaIcon({ type }) {
  if (type === 'audio') return <FileAudio size={18} />
  if (type === 'image') return <FileImage size={18} />
  return <FileVideo size={18} />
}

export function MaterialPanel({ projectID, onInsert }) {
  const [materials, setMaterials] = useState([])
  const [loading, setLoading] = useState(true)
  const [progress, setProgress] = useState(0)
  const [error, setError] = useState('')

  const reload = useCallback(() => {
    setLoading(true)
    return apiRequest(`/api/magic-cut/projects/${projectID}/materials`)
      .then(payload => setMaterials(payload.materials || []))
      .catch(requestError => setError(requestError.message || '素材列表读取失败'))
      .finally(() => setLoading(false))
  }, [projectID])

  useEffect(() => { reload() }, [reload])

  function upload({ file, onSuccess, onError, onProgress }) {
    setError('')
    setProgress(1)
    const form = new FormData()
    form.append('file', file)
    const request = new XMLHttpRequest()
    request.open('POST', `/api/magic-cut/projects/${projectID}/materials`)
    request.withCredentials = true
    request.upload.onprogress = event => {
      if (!event.lengthComputable) return
      const percent = Math.round((event.loaded / event.total) * 100)
      setProgress(percent)
      onProgress?.({ percent })
    }
    request.onload = () => {
      let payload = {}
      try { payload = JSON.parse(request.responseText || '{}') } catch { payload = {} }
      if (request.status >= 200 && request.status < 300) {
        setMaterials(current => [payload.material, ...current.filter(item => item.id !== payload.material?.id)].filter(Boolean))
        setProgress(0)
        onSuccess?.(payload)
        return
      }
      const message = materialErrorMessage(payload.code)
      setError(message)
      setProgress(0)
      onError?.(new Error(message))
    }
    request.onerror = () => {
      const message = '素材上传中断，请检查本机网络后重试'
      setError(message)
      setProgress(0)
      onError?.(new Error(message))
    }
    request.send(form)
  }

  return (
    <section className="editor-panel material-panel" aria-label="素材区">
      <header><b>素材</b><Button size="small" onClick={reload}>刷新</Button></header>
      <Dragger className="material-drop" multiple={false} showUploadList={false} customRequest={upload} accept="video/*,audio/*,image/*">
        <UploadCloud size={22} />
        <p>点击或拖入素材</p>
        <small>视频、音频、图片均可</small>
      </Dragger>
      {progress ? <Progress percent={progress} size="small" status="active" /> : null}
      {error ? <Alert type="error" showIcon message={error} /> : null}
      {loading ? <div className="panel-loading"><Spin /></div> : materials.length ? (
        <div className="material-list">
          {materials.map(material => (
            <button type="button" className="material-item" key={material.id} onDoubleClick={() => onInsert?.(material)} title="双击添加到时间线">
              <span className={`material-item__icon is-${material.mediaType}`}><MediaIcon type={material.mediaType} /></span>
              <span><b>{material.originalName}</b><small>{materialSummary(material)}</small></span>
              <em>{material.status === 'ready' ? '可用' : material.status}</em>
            </button>
          ))}
        </div>
      ) : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="还没有素材" />}
    </section>
  )
}
