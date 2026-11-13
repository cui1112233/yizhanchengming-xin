export function createBatchCompletionAudio({
  AudioContextClass=globalThis.AudioContext || globalThis.webkitAudioContext,
  load=globalThis.fetch?.bind(globalThis),url='/assets/sounds/batch-complete.wav'
}={}){
  let context=null,bufferPromise=null,disposed=false;
  const buffer=()=>{
    if(!bufferPromise)bufferPromise=load(url,{cache:'force-cache'}).then(response=>{
      if(!response.ok)throw Error('completion audio unavailable');return response.arrayBuffer();
    }).then(data=>context.decodeAudioData(data)).catch(error=>{bufferPromise=null;throw error;});
    return bufferPromise;
  };
  return {
    async prepare(){
      if(disposed || !AudioContextClass || !load)return false;
      try{
        if(!context || context.state==='closed'){context=new AudioContextClass();bufferPromise=null;}
        // resume() must be invoked before awaiting anything, in the trusted gesture.
        const resumed=context.state==='suspended'?context.resume():Promise.resolve();
        await Promise.all([resumed,buffer()]);return context.state==='running';
      }catch{return false;}
    },
    async play(preferences={}, {isCurrent=()=>true,getPreferences=()=>preferences}={}){
      const {soundEnabled=true,soundVolume=60}=preferences;
      let volume=Number.isFinite(Number(soundVolume))?Math.min(100,Math.max(0,Number(soundVolume))):60;
      if(disposed || !soundEnabled || volume===0 || !context || context.state!=='running')return false;
      try{
        const sample=await buffer();if(disposed || context.state!=='running')return false;
        if(!isCurrent())return false;
        const current=getPreferences() || {};
        if(current.soundEnabled===false)return false;
        volume=Number.isFinite(Number(current.soundVolume))?Math.min(100,Math.max(0,Number(current.soundVolume))):60;
        if(volume===0)return false;
        const source=context.createBufferSource(),gain=context.createGain();
        source.buffer=sample;gain.gain.setValueAtTime(volume/100,context.currentTime);
        source.connect(gain);gain.connect(context.destination);
        source.onended=()=>{source.disconnect();gain.disconnect();};
        source.start();return true;
      }catch{return false;}
    },
    dispose(){disposed=true;context?.close()?.catch?.(()=>{});}
  };
}
