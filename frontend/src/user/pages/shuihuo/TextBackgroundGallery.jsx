import { Button, Empty, Modal, Spin } from 'antd';
import { useState } from 'react';
import AuthenticatedTextPageImage from './AuthenticatedTextPageImage.jsx';

export default function TextBackgroundGallery({value,items=[],loading=false,disabled=false,onChange}) {
  const [open,setOpen]=useState(false);
  const [selected,setSelected]=useState('');
  const image=Boolean(value?.id);
  return <>
    <Button disabled={disabled} onClick={()=>{setSelected(value?.id || '');setOpen(true);}}>{image?'更换图库背景':'打开图库'}</Button>
    {image && <Button disabled={disabled} onClick={()=>onChange({kind:'color',color:value?.color || '#111318'})}>使用背景色</Button>}
    <Modal title="我的背景图库" open={open} className="batch-factory-background-gallery" onCancel={()=>setOpen(false)}
      okText={selected?'使用此背景':'使用背景色'} cancelText="取消" onOk={()=>{
        onChange(selected?{kind:'library',id:selected,color:value?.color || '#111318'}:{kind:'color',color:value?.color || '#111318'});
        setOpen(false);
      }}>
      <Button type={selected?'default':'primary'} onClick={()=>setSelected('')}>背景颜色</Button>
      {loading?<Spin/>:items.length?<div className="batch-factory-background-gallery-grid">{items.map(item=><button type="button" key={item.id}
        aria-pressed={selected===item.id} className={selected===item.id?'is-selected':''} onClick={()=>setSelected(item.id)}>
        <AuthenticatedTextPageImage passive contentUrl={item.contentUrl} alt={item.name || '背景图'}/>
        <span>{item.name || '背景图'}</span>
      </button>)}</div>:<Empty description="图库暂无背景，请先上传背景图"/>}
    </Modal>
  </>;
}
