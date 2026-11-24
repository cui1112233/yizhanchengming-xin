export type NovelStorage={get:(key:string)=>any;set:(key:string,value:any)=>unknown;remove:(key:string)=>unknown};
export function createNovelJournal(storage:NovelStorage,scope:string){
  if(!scope.trim())throw Error('尚未确认账号，无法保存创作草稿');
  const key=(name:string)=>`yzcm.novel.v1:${encodeURIComponent(scope)}:${encodeURIComponent(name)}`;
  return {
    read(name:string){const value=storage.get(key(name));return value&&typeof value==='object'?value:null;},
    save(name:string,value:any){try{storage.set(key(name),JSON.parse(JSON.stringify(value)));}catch{throw Error('恢复记录无法保存，尚未发出新请求，请保留原文');}},
    clear(name:string){storage.remove(key(name));},
  };
}
export type NovelJournal=ReturnType<typeof createNovelJournal>;
