import {Select} from 'antd';
export default function TextHighlightRuleSelect({value='',legacyKeywords=false,onChange,disabled=false}){
 const rules=value==='off'?[]:value?value.split(','):legacyKeywords?['keywords']:[];
 return <Select aria-label="标红规则" mode="multiple" allowClear disabled={disabled} placeholder="关闭标红" value={rules} options={[{value:'dialogue',label:'对话'},{value:'lines',label:'指定行',title:'按所选行的第一个字定位，标红完整句子；引号内标红完整对话'},{value:'keywords',label:'关键词'}]} onChange={next=>onChange(next.length?next.join(','):'off')} getPopupContainer={trigger=>trigger.parentElement}/>;
}
