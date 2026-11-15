const formats={png:'image/png',jpg:'image/jpeg',jpeg:'image/jpeg',webp:'image/webp',gif:'image/gif'};
const mimeTypes=new Set(Object.values(formats));

// CM adaptation: browser order, one root category, nested images flattened.
// The original native importer ordering and nested-directory rules are unknown.
export function buildCharacterDirectoryImport(files,createKey=()=>crypto.randomUUID()) {
  const selection=Array.from(files||[]);if(!selection.length)return null;
  const paths=selection.map(file=>{
    const path=String(file.webkitRelativePath||'');
    if(!path||path.startsWith('/')||path.includes('\\')||/^[A-Za-z]:/.test(path)||/[\x00-\x1f\x7f]/.test(path))throw new Error('文件夹相对路径无效，请重新选择整个图片文件夹');
    const parts=path.split('/');
    if(parts.length<2||parts.some(part=>!part||part==='.'||part==='..')||!parts[0].trim())throw new Error('文件夹相对路径无效，请重新选择整个图片文件夹');
    if(parts.at(-1)!==file.name)throw new Error('文件名与相对路径不一致，请重新选择文件夹');
    return {file,relativePath:path,root:parts[0]};
  });
  if(new Set(paths.map(item=>item.root)).size!==1)throw new Error('一次只能导入同一个根文件夹，请重新选择');
  const category=paths[0].root,items=[],skipped=[];
  for(const item of paths){
    const extension=item.file.name.split('.').at(-1).toLowerCase(),mime=formats[extension];
    let reason='';if(!mime||(item.file.type&&!mimeTypes.has(item.file.type.toLowerCase())))reason='不支持的图片格式，仅支持 PNG、JPEG、WEBP、GIF';
    else if(item.file.size>40*1024*1024)reason='图片超过40MB';
    if(reason){skipped.push({relativePath:item.relativePath,reason});continue;}
    const file=item.file.type?item.file:new File([item.file],item.file.name,{type:mime,lastModified:item.file.lastModified});
    items.push({file,relativePath:item.relativePath,type:'character',category,name:`${category}${items.length+1}`,requestKey:createKey(),status:'queued',error:''});
  }
  return {category,items,skipped};
}
