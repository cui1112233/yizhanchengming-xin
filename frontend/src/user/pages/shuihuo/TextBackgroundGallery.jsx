import { Alert, Button, ConfigProvider, Empty, Modal, Spin, Upload } from 'antd';
import { useState } from 'react';
import AuthenticatedTextPageImage from './AuthenticatedTextPageImage.jsx';

export default function TextBackgroundGallery({value,items=[],loading=false,disabled=false,onChange,onUpload}) {
  const [open,setOpen]=useState(false);
  const [selected,setSelected]=useState('');
  const [uploading,setUploading]=useState(false);
  const [error,setError]=useState('');
  async function upload(file){
    if(uploading || disabled)return;
    setUploading(true);setError('');
    try{const item=await onUpload(file);if(!item?.id)throw Error('背景上传失败');setSelected(item.id);}
    catch(e){setError(e.message || '背景上传失败，请重试');}
    finally{setUploading(false);}
  }
  return <ConfigProvider theme={{token:{colorBgElevated:'var(--sh-panel)',colorBgContainer:'var(--sh-panel-strong)',colorText:'var(--sh-text)',colorTextSecondary:'var(--sh-muted)',colorBorder:'var(--sh-line)',colorPrimary:'var(--sh-accent)'}}}>
    <Button disabled={disabled} onClick={()=>{setSelected(value?.id || '');setError('');setOpen(true);}}>打开图库</Button>
    <Modal title="我的背景图库" open={open} className="batch-factory-background-gallery" onCancel={()=>{if(!uploading)setOpen(false);}}
      confirmLoading={uploading} cancelButtonProps={{disabled:uploading}} closable={!uploading} maskClosable={!uploading} keyboard={!uploading} okButtonProps={{disabled:disabled || loading || uploading}}
      okText={selected?'使用此背景':'使用背景色'} cancelText="取消" onOk={()=>{
        if(disabled || uploading || loading)return;
        onChange(selected?{kind:'library',id:selected,color:value?.color || '#111318'}:{kind:'color',color:value?.color || '#111318'});
        setOpen(false);
      }}>
      <div className="text-background-gallery-toolbar"><Button disabled={uploading} type={selected?'default':'primary'} onClick={()=>setSelected('')}>背景颜色</Button>
      {onUpload && <Upload accept="image/png,image/jpeg,image/webp" showUploadList={false} disabled={disabled || uploading} beforeUpload={file=>{void upload(file);return false;}}><Button loading={uploading} disabled={disabled}>上传背景图</Button></Upload>}</div>
      {error && <Alert type="error" showIcon message={error}/>}
      {loading?<Spin/>:items.length?<div className="batch-factory-background-gallery-grid">{items.map(item=><button type="button" key={item.id}
        disabled={uploading} aria-pressed={selected===item.id} className={selected===item.id?'is-selected':''} onClick={()=>setSelected(item.id)}>
        <AuthenticatedTextPageImage passive contentUrl={item.contentUrl} alt={item.name || '背景图'}/>
        <span>{item.name || '背景图'}</span>
      </button>)}</div>:<Empty description="图库暂无背景，请先上传背景图"/>}
    </Modal>
  </ConfigProvider>;
}
