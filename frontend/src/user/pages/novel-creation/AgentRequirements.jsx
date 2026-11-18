import { Button, Checkbox, Form, Input, Select } from 'antd';

const categories = [{value:'long_fiction',label:'长篇正文'},{value:'short_fiction',label:'短篇小说'},{value:'imitation',label:'仿写创作'}];
export default function AgentRequirements({writing=false}) {
 const field=writing?'customWritePrompts':'reviewPrompts';
 const label=writing?'写作要求':'补充审稿';
 return <section className="novel-agent-section novel-agent-requirements">
  <Form.List name={field}>{(fields,{add,remove})=><>
   <div className="novel-agent-section-heading"><div><h3>{writing?'自定义写作提示词':'自定义补充审稿提示词'}</h3><p>{writing?'勾选后用于对应模式的正文写作与修订。':'勾选后用于章节审稿与修订，补充你关心的检查要求。'}</p></div>
   <Button disabled={fields.length>=50} onClick={()=>add({id:crypto.randomUUID(),kind:writing?undefined:'custom',category:writing?'long_fiction':undefined,title:writing?'自定义写作要求':'自定义审稿',content:'',enabled:true})}>新增{label}</Button></div>
   {!fields.length&&<div className="novel-agent-empty">{writing?'暂无自定义写作提示词。':'暂无自定义项，可点击「新增补充审稿」添加。'}</div>}
   {fields.map(({key,name,...rest},index)=><div key={key} className="novel-agent-prompt-card">
    <Form.Item {...rest} hidden name={[name,'id']}><Input /></Form.Item>
    {!writing&&<Form.Item {...rest} hidden name={[name,'kind']}><Input /></Form.Item>}
    <div className="novel-agent-prompt-heading">
     <Form.Item {...rest} name={[name,'enabled']} valuePropName="checked"><Checkbox aria-label={`启用${label}${index+1}`}>启用</Checkbox></Form.Item>
     <Form.Item {...rest} name={[name,'title']} rules={[{required:true,whitespace:true,message:'请填写标题'},{max:100,message:'标题最多100字'}]}><Input maxLength={100} aria-label={`${label}${index+1}标题`} placeholder="提示词标题" /></Form.Item>
     {writing&&<Form.Item {...rest} name={[name,'category']} rules={[{required:true,message:'请选择写作类别'}]}><Select aria-label={`${label}${index+1}类别`} options={categories} /></Form.Item>}
     <Button type="text" danger aria-label={`删除${label}${index+1}`} onClick={()=>remove(name)}>删除</Button>
    </div>
    <Form.Item {...rest} name={[name,'content']} dependencies={[[field,name,'enabled']]} rules={[{max:8000,message:'每条要求最多8000字'},({getFieldValue})=>({validator(_,value){return getFieldValue([field,name,'enabled'])&&!String(value||'').trim()?Promise.reject(new Error('启用前请填写要求内容')):Promise.resolve();}})]}>
     <Input.TextArea aria-label={`${label}${index+1}内容`} autoSize={{minRows:writing?3:2,maxRows:10}} maxLength={8000} placeholder={writing?'写作要求说明……':'审稿要求说明……'} />
    </Form.Item>
   </div>)}
  </>}</Form.List>
 </section>;
}
