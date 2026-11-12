import React from 'react';
import { Dropdown } from 'antd';
import { Maximize2, Minimize2, Scan } from 'lucide-react';
const zooms=['fit',25,50,75,100,150,200];
export default function PlayerViewControls({quality,onQuality,zoom,onZoom,ratio,onRatio,fullscreen,onFullscreen,safeGuide,onSafeGuide}) {
 const popup=trigger=>trigger.parentElement;
 return <div className="aicut-player-options">
  <Dropdown trigger={['click']} getPopupContainer={popup} menu={{selectedKeys:[quality],items:[
   {key:'original',label:'原画 · 使用原始素材',onClick:()=>onQuality('original')},
   {key:'proxy',label:'流畅 · 优先代理素材',onClick:()=>onQuality('proxy')}
  ]}}><button type="button" aria-label="预览画质" title="预览画质；不改变导出质量">{quality==='original'?'原画':'流畅'}</button></Dropdown>
  <Dropdown trigger={['click']} getPopupContainer={popup} menu={{selectedKeys:[String(zoom)],items:[...zooms.map(value=>({key:String(value),label:value==='fit'?'适应窗口':`${value}%`,onClick:()=>onZoom(value)})),{type:'divider'},{key:'guide',label:safeGuide?'隐藏安全框':'显示安全框',onClick:onSafeGuide}]}}><button type="button" aria-label="预览缩放" title={zoom==='fit'?'预览缩放 · 适应窗口':`预览缩放 · ${zoom}%`}><Scan size={16}/>{zoom==='fit'?null:<small>{zoom}%</small>}</button></Dropdown>
  <Dropdown trigger={['click']} getPopupContainer={popup} menu={{selectedKeys:[ratio],items:['9:16','16:9','1:1'].map(value=>({key:value,label:`${value}${value==='9:16'?' · 竖屏':value==='16:9'?' · 横屏':' · 方形'}`,onClick:()=>onRatio(value)}))}}><button type="button" aria-label="画面比例" title="修改工程画面比例">{ratio}</button></Dropdown>
  <button type="button" aria-label={fullscreen?'退出全屏预览':'全屏预览'} title={fullscreen?'退出全屏预览 · Esc':'全屏预览'} onClick={onFullscreen}>{fullscreen?<Minimize2 size={16}/>:<Maximize2 size={16}/>}</button>
 </div>;
}
