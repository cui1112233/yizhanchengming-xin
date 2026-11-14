import {Modal} from 'antd';
import {useCallback,useEffect,useRef} from 'react';

// Static Ant Design dialogs live outside their caller's React tree.
// Bind them to that caller so leaving a project cannot submit an old action.
export function useComicConfirmation(key) {
  const scope=useRef({key,version:0,active:true}),dialogs=useRef(new Set());
  if(scope.current.key!==key)scope.current={key,version:scope.current.version+1,active:true};
  const version=scope.current.version;
  useEffect(()=>{
    scope.current.active=true;
    return()=>{
      if(scope.current.version===version)scope.current.active=false;
      for(const entry of [...dialogs.current])if(entry.version===version){
        dialogs.current.delete(entry);
        if(!entry.settled){entry.settled=true;entry.cancel?.();}
        entry.dialog.destroy();
      }
    };
  },[key,version]);
  return useCallback(options=>{
    const valid=()=>scope.current.active&&scope.current.version===version&&scope.current.key===key;
    if(!valid()){options.onCancel?.();return null;}
    const entry={version,settled:false,cancel:options.onCancel,dialog:null};
    entry.dialog=Modal.confirm({...options,
      onOk:(...args)=>{if(!valid()||entry.settled)return;entry.settled=true;try{const result=options.onOk?.(...args);return result?.then?result.catch(error=>{if(valid())entry.settled=false;throw error;}):result;}catch(error){entry.settled=false;throw error;}},
      onCancel:(...args)=>{if(entry.settled)return;entry.settled=true;return options.onCancel?.(...args);},
      afterClose:()=>{dialogs.current.delete(entry);options.afterClose?.();},
    });
    dialogs.current.add(entry);
    return entry.dialog;
  },[key,version]);
}
