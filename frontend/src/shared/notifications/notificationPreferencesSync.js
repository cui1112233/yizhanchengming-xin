const normalize=value=>({soundEnabled:value?.soundEnabled!==false,soundVolume:Number.isFinite(value?.soundVolume)?Math.min(100,Math.max(0,value.soundVolume)):60});
export function installNotificationPreferencesSync({target,storage,getOwner,onUpdate}){
  const key=()=>`qiantie:notification-preferences:${getOwner()}`;
  const local=event=>{
    if(!getOwner())return;
    const preferences=normalize(event.detail);onUpdate(preferences);
    try{storage.setItem(key(),JSON.stringify({...preferences,at:Date.now()}));}catch{/* Same-tab settings remain effective. */}
  };
  const remote=event=>{
    if(!getOwner() || event.key!==key() || !event.newValue)return;
    try{const value=JSON.parse(event.newValue);if(typeof value?.soundEnabled==='boolean')onUpdate(normalize(value));}catch{/* Ignore invalid or unrelated storage. */}
  };
  target.addEventListener('qiantie:notifications-updated',local);target.addEventListener('storage',remote);
  return()=>{target.removeEventListener('qiantie:notifications-updated',local);target.removeEventListener('storage',remote);};
}
