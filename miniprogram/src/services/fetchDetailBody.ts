type Payload=Record<string,unknown>;
const text=(value:unknown)=>typeof value==='string'?value:'';
function count(value:unknown,body:string){const number=Number(value);return value!==undefined&&value!==null&&value!==''&&Number.isFinite(number)&&number>=0?number:body?Array.from(body).length:null;}
export function fetchDetailBody(payload:Payload){
 const processed=text(payload.original||payload.processedText||payload.processed_text||payload.body||payload.text);
 const raw=text(payload.originalRaw||payload.original_raw);
 return {processed,raw,processedCount:count(payload.processedChars??payload.processed_chars,processed),rawCount:count(payload.originalRawChars??payload.original_raw_chars,raw)};
}
