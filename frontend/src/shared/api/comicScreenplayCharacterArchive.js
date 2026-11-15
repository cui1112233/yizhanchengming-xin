import { apiRequest } from './client.js';

const path='/api/comic-screenplays/asset-library/characters';
export function validateCharacterArchiveFiles(files) {
  if(!files?.length||!files.some(file=>['characters.lschar','characters.json'].includes(file.name)))throw new Error('请选择包含 characters.lschar 或 characters.json 的文件夹或归档');
}
export function importCharacterArchive({files,requestKey},{signal}={},request=apiRequest) {
  validateCharacterArchiveFiles(files);
  if(!requestKey)throw new Error('导入请求编号缺失');
  const form=new FormData();form.append('requestKey',requestKey);
  for(const file of files)form.append('files',file,file.webkitRelativePath||file.name);
  return request(`${path}/import`,{method:'POST',body:form,signal,suppressGlobalError:true}).then(data=>{
    if(!Number.isSafeInteger(data?.imported_count)||data.imported_count<1||!Array.isArray(data.entries)||data.entries.length!==data.imported_count)throw new Error('导入结果未确认，请使用本次请求重试核对');return data;
  });
}
export function clearCharacterArchive({requestKey,expected},{signal}={},request=apiRequest) {
  if(!requestKey||!Array.isArray(expected)||expected.some(entry=>!entry.id||!Number.isSafeInteger(entry.revision)||entry.revision<1))throw new Error('角色版本无效，请重新读取');
  return request(`${path}/clear`,{method:'POST',body:JSON.stringify({requestKey,expected}),signal,suppressGlobalError:true}).then(data=>{
    if(data?.cleared!==true||!Number.isSafeInteger(data.cleared_count)||data.cleared_count<0)throw new Error('清空结果未确认，请使用本次请求重试核对');return data;
  });
}
export function exportCharacterArchive({signal}={},request=apiRequest) {
  return request(`${path}/export`,{signal,responseType:'blob',suppressGlobalError:true}).then(blob=>{if(!(blob instanceof Blob)||!blob.size)throw new Error('导出文件为空，请重新读取后重试');return blob;});
}
