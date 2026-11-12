import React, { useEffect, useRef, useState } from 'react';
import { previewDimensions } from './playerView.js';
export default function PreviewViewport({ ratio, zoom, children }) {
 const ref=useRef(null);
 const [size,setSize]=useState({width:1,height:1});
 useEffect(()=>{
   const observer=new ResizeObserver(entries=>{const {width,height}=entries[0].contentRect;setSize({width,height});});
   observer.observe(ref.current);return ()=>observer.disconnect();
 },[]);
 const dimensions=previewDimensions(ratio,size.width,size.height,zoom);
 return <div className="aicut-preview-viewport" ref={ref} aria-label="预览画布视口" data-preview-scale={dimensions.scale} style={{'--preview-width':`${dimensions.width}px`,'--preview-height':`${dimensions.height}px`}}>{children}</div>;
}
