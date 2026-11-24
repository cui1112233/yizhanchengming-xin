import BrandHeader from '../../../components/BrandHeader';
import {ThemedPage} from '../../../theme';
import {Button,Text,View} from '@tarojs/components';
import {apiRequest} from '../../../services/api';
import {onSessionInvalidated} from '../../../services/sessionBoundary';
import { useEffect,useRef,useState } from "react";
import { readComicActiveRun, clearComicActiveRun,prepareComicRecovery,configureComicRecoveryScope } from "../../services/comicRecovery";
import ManualComicPrivatePage from "./ManualComicPrivatePage";
import LegacyComicPrivatePage from "./LegacyComicPrivatePage";
export { LegacyComicPrivatePage };
/** Only previously submitted automatic jobs retain their original recovery path. */
export function ComicEntryContent() {
    const [initialSource, setInitialSource] = useState("");
    const [legacy, setLegacy] = useState(() => { const run = readComicActiveRun(); return Boolean(run?.submitted && run.mode !== "manual"); });
    return legacy ? <LegacyComicPrivatePage onNew={source => { clearComicActiveRun(); setInitialSource(source || ""); setLegacy(false); }}/> : <ManualComicPrivatePage initialSource={initialSource}/>;
}

export default function ComicEntry(){
 const [ready,setReady]=useState(false),[error,setError]=useState('');const epoch=useRef(0);
 const load=async()=>{const ticket=++epoch.current;setReady(false);setError('');configureComicRecoveryScope('');try{const response:any=await apiRequest('/api/member/me');if(ticket!==epoch.current)return;if(!response?.member?.username)throw Error('漫画账号尚未确认');await prepareComicRecovery(apiRequest,response.member.username);if(ticket===epoch.current)setReady(true);}catch(c:any){if(ticket===epoch.current)setError(c.message||'恢复记录核对失败');}};
 useEffect(()=>{void load();const off=onSessionInvalidated(()=>{epoch.current++;configureComicRecoveryScope('');setReady(false);setError('账号已切换，请重新核对漫画账号');});return()=>{epoch.current++;off();};},[]);
 return ready?<ComicEntryContent/>:<ThemedPage className="mp-page"><BrandHeader title="漫画生成" showBack/><Text>{error||'正在核对漫画账号与恢复记录…'}</Text>{error?<Button className="mp-secondary" onClick={load}>重新读取漫画账号</Button>:null}</ThemedPage>;
}
