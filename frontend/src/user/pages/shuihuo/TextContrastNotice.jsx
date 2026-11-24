import {hasLowTextContrast,readableTextColor,textColorScheme} from './textLayoutTemplates.js';
import './publish-settings.css';

export default function TextContrastNotice({color,background,warning='',onChange}){
 if(!hasLowTextContrast(color,background) && !warning.includes('对比度较低'))return null;
 const image=Boolean(background?.id);
 return <div className="text-contrast-notice" role="alert"><span>文字与背景太接近，可能看不清。</span><button type="button" onClick={()=>onChange(image?textColorScheme('light'):{textColor:readableTextColor(background)})}>{image?'使用浅字深底':'一键调整文字颜色'}</button></div>;
}
