import {Alert} from 'antd';
export default function ComicScreenplayPlanningProgress({progress,error}){
 if(error)return <Alert type="warning" showIcon message="暂时无法读取规划进度" description={`${error}${progress?` 最后确认：第 ${progress.round} / ${progress.total} 轮，已收到 ${progress.chars} 字。`:''}`}/>;
 if(!progress)return <Alert type="info" showIcon message="正在等待规划进度" description="正在核对本次请求，不会重复提交生成。"/>;
 const missing=progress.missing.map(n=>`第${n}章`).join('、');
 const title=progress.status==='cancelled'?'规划已取消':progress.status==='failed'?'规划生成失败':progress.status==='succeeded'?'规划已完整并保存':progress.status==='partial'?'规划草稿已保存':`正在生成分镜规划 · 第 ${progress.round} / ${progress.total} 轮`;
 return <Alert type={['failed','cancelled'].includes(progress.status)?'warning':'info'} showIcon message={title} description={[`本次已收到 ${progress.chars} 字`,missing&&`待补充：${missing}`,progress.status==='running'&&(progress.phase==='requesting'?'正在等待本轮模型返回':'本轮已返回，正在核对完整性')].filter(Boolean).join('；')}/>;
}
