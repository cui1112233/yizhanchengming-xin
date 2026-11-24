import Taro from '@tarojs/taro';
import {apiUrl,requireApiOrigin} from '../config/runtime';
import {getBusinessToken,getSessionCookie} from './session';
import {onSessionInvalidated,sessionRevision} from './sessionBoundary';
import {apiRequest} from './api';
const files=new Set<string>();
export function forgetPluginFile(path:string){if(!files.has(path))return;files.delete(path);try{Taro.getFileSystemManager().unlink({filePath:path,fail:()=>undefined});}catch{/* already cleared */}}
onSessionInvalidated(()=>{for(const path of [...files])forgetPluginFile(path);});
const current=(revision:number)=>{if(revision!==sessionRevision())throw Error('账号已切换，请重新打开插件');};
export async function choosePluginMedia(kind:'image'|'video'){
 const revision=sessionRevision();let picked:Awaited<ReturnType<typeof Taro.chooseMedia>>;
 try{picked=await Taro.chooseMedia({count:1,mediaType:[kind],sourceType:['album','camera']});}catch(e:any){if(/cancel/i.test(e.errMsg||''))return null;throw e;}
 current(revision);const file=picked.tempFiles[0];if(!file?.tempFilePath)return null;return{path:file.tempFilePath,size:file.size,kind};
}
export async function uploadPluginMedia(path:string,filePath:string,formData:Record<string,string>={}){
 if(!/^\/api\/(?:agent\/tasks\/[^/?#]+\/assets|ai-upscale\/jobs)$/.test(path))throw Error('上传目标无效');
 requireApiOrigin();const revision=sessionRevision(),token=getBusinessToken(),cookie=getSessionCookie();
 const response=await Taro.uploadFile({url:apiUrl(path),filePath,name:'file',timeout:600000,formData,header:{...(token?{Authorization:`Bearer ${token}`} :{}),...(cookie?{Cookie:cookie}:{})}});current(revision);
 let body:any;try{body=JSON.parse(response.data);}catch{throw Error('上传结果未确认，请刷新对应项目核对，勿重复上传');}
 if(response.statusCode<200||response.statusCode>=300)throw Object.assign(Error(body.error||'上传失败'),{statusCode:response.statusCode});return body;
}
export async function savePluginFile(data:string|ArrayBuffer,extension:string){
 if(!/^[a-z0-9]{1,6}$/.test(extension))throw Error('文件类型无效');const revision=sessionRevision();const path=`${Taro.env.USER_DATA_PATH}/plugin-${Date.now()}-${Math.random().toString(36).slice(2,7)}.${extension}`;
 await new Promise<void>((resolve,reject)=>Taro.getFileSystemManager().writeFile({filePath:path,data,...(typeof data==='string'?{encoding:'utf8' as const}:{}),success:()=>resolve(),fail:reject}));files.add(path);try{current(revision);}catch(e){forgetPluginFile(path);throw e;}return path;
}
function utf8(value:string){const matches=encodeURIComponent(value).match(/%[A-F0-9]{2}|./g)||[];return Uint8Array.from(matches.map(c=>c[0]==='%'?parseInt(c.slice(1),16):c.charCodeAt(0)));}
export function segmentMultipart(file:ArrayBuffer,episodes:any[],resolution='',fps=0){
 const boundary=`yzcm-${Date.now()}-${Math.random().toString(36).slice(2)}`;
 const fields=Object.entries({episodes:JSON.stringify(episodes),resolution,fps:String(fps)}).map(([name,value])=>`--${boundary}\r\nContent-Disposition: form-data; name="${name}"\r\n\r\n${value}\r\n`).join('');
 const head=utf8(`${fields}--${boundary}\r\nContent-Disposition: form-data; name="video"; filename="input.mp4"\r\nContent-Type: video/mp4\r\n\r\n`),tail=utf8(`\r\n--${boundary}--\r\n`),result=new Uint8Array(head.length+file.byteLength+tail.length);
 result.set(head);result.set(new Uint8Array(file),head.length);result.set(tail,head.length+file.byteLength);return{body:result.buffer,contentType:`multipart/form-data; boundary=${boundary}`};
}
export async function exportSegmentVideo(filePath:string,episodes:any[],resolution='',fps=0){
 const revision=sessionRevision();const info=await Taro.getFileInfo({filePath});if(!('size' in info))throw Error('无法读取原视频信息');if(info.size>32*1024*1024)throw Error('手机分段导出目前支持32MB以内视频，更大视频请使用电脑版导出');
 const data=await new Promise<ArrayBuffer>((resolve,reject)=>Taro.getFileSystemManager().readFile({filePath,success:r=>typeof r.data==='string'?reject(Error('视频读取格式无效')):resolve(r.data),fail:reject}));current(revision);
 const multipart=segmentMultipart(data,episodes,resolution,fps);const bytes=await apiRequest<ArrayBuffer>('/api/plugins/ai-segment/export',{method:'POST',data:multipart.body,header:{'content-type':multipart.contentType},responseType:'arraybuffer',timeout:600000});current(revision);
 if(!(bytes instanceof ArrayBuffer)||bytes.byteLength<4||new Uint8Array(bytes)[0]!==80||new Uint8Array(bytes)[1]!==75)throw Error('未收到有效导出文件');return savePluginFile(bytes,'zip');
}
