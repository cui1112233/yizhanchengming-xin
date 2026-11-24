import {useMemo,useState} from 'react';
import {Button} from 'antd';
import {listTextPages} from '../../../shared/api/batchFactoryText.js';
import {layoutFields} from './textMediaActions.js';
import {useDraftPreview} from './TextPageEditor.jsx';

export default function TextSettingsPreview({batchId,bookId,value}){
 const [source,setSource]=useState(null),[error,setError]=useState(''),[busy,setBusy]=useState(false);
 const session=useMemo(()=>source?{id:source.page.id,rev:source.page.revision,docRev:source.document.revision,cover:false,draft:{text:source.page.text,layout:Object.fromEntries(Object.entries(layoutFields).map(([key,path])=>[path,value[key]])),background:value.textBackground}}:null,[source,value]);
 const preview=useDraftPreview(batchId,bookId,session);
 async function load(){setBusy(true);setError('');try{const data=await listTextPages(batchId,bookId);if(!data.document || !data.pages?.length)throw Error('请先完成文字分页，再预览排版');setSource({page:data.pages[0],document:data.document});}catch(e){setError(e.message);}finally{setBusy(false);}}
 return <div className="text-settings-preview"><Button loading={busy} disabled={!batchId || !bookId} onClick={load}>预览当前排版</Button>{source && <Button onClick={()=>setSource(null)}>收起预览</Button>}{preview.url && <img alt="当前排版预览" src={preview.url}/>}<span role={error || preview.error?'alert':undefined}>{error || preview.error || preview.warning || ''}</span></div>;
}
