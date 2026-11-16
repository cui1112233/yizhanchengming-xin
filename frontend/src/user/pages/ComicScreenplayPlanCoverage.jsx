import {Alert} from 'antd';

export default function ComicScreenplayPlanCoverage({chapter}){
 const result=chapter?.planning_coverage;
 const rounds=Number.isInteger(chapter?.planning_rounds)&&chapter.planning_rounds>0?` · 生成${chapter.planning_rounds}轮`:'';
 if(!chapter?.planning_text?.trim())return null;
 if(!result)return <Alert type="info" showIcon message="已保存规划，尚未检查完整性" description="这是之前保存的规划，请核对六章内容与分节表。"/>;
 if(result.ok)return <Alert type="success" showIcon message="规划已通过结构检查" description={`${result.chars}字 · ${result.rows}节${rounds}。仍需核对剧情、台词和资产是否准确。`}/>;
 const missing=(result.missing||[]).map(n=>`第${n}章`).join('、');
 const wrong=(result.wrong||[]).map(item=>`第${item.heading}章应为「${item.expected}」`).join('；');
 return <Alert type="warning" showIcon message="规划可能仍不完整，草稿已保存" description={[`${result.chars}字 · 已识别${result.rows}节${rounds}`,missing&&`缺少${missing}`,wrong,chapter.planning_continuation_error&&`续写中断：${chapter.planning_continuation_error}`,'请打开「查看规划」，核对六章正文、分节表和结尾。'].filter(Boolean).join('；')}/>;
}
