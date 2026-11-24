import Taro,{useLoad} from '@tarojs/taro';
import {Button,Text} from '@tarojs/components';
import {useRef,useState} from 'react';
import {legacyModuleDetailDestination} from '../../config/modules';
import {ThemedPage} from '../../theme';
import './redirect.scss';

// Keep old task/history links valid without pulling workspace code into main.
export default function LegacyModuleDetailEntry(){
 const destination=useRef('');const [error,setError]=useState('');
 const open=()=>{setError('');void Taro.redirectTo({url:destination.current}).catch(()=>setError('功能页面暂时未能打开，请重试'));};
 useLoad(params=>{destination.current=legacyModuleDetailDestination(params||{});open();});
 return <ThemedPage className="mp-page mp-module-redirect"><Text>{error||'正在打开功能…'}</Text>{error?<Button onClick={open}>重新打开</Button>:null}</ThemedPage>;
}
