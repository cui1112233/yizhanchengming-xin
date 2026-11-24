# 任务中心来源与作品定位补齐

延续用户批准的全部功能对齐。本轮只修任务来源和打开目标，不重画主页。CM改读/api/agent/tasks及各项目generations/video-jobs，停止把Studio当CM；小说读取既有作品会话当前run和定时/连续任务中可查到的run；AI-CUT读取实际export-jobs。所有仅GET，不自动生成/重试。

保留项目ID与任务ID区别：批量视频返回原batch/book，CM任务返回对应project的任务页，私人定制historyId正确传递；小说返回project/session/run，AI-CUT返回project/options.timelineId的导出页。只有服务端给出真实进度才显示百分比。各类型调用匹配图标，跨来源同编号不得相互去重。

读取使用有限并发，单项目读取失败保留其他真实记录，并显示部分来源未同步；空列表与读取失败分开。相同CM视频任务在generations和video-jobs重复时去重；同小说run去重，定时计划清楚标识计划，不能冒充多个收费请求。作品总览用列表模式避免逐项目加载重数据。

手机只显示当前接口可查询的任务，不把voice同步结果编造成云任务；当前小说API没有全局历史列表，明确显示“近期写作任务”，保留既有会话当前任务及可查询的自动写作历史。现有主任务筛选保留，状态补齐awaiting_confirmation/queued/submitting等真实异步状态。

实施：先写真实来源/目标/部分读取失败测试；拆分taskFeeds客户端，扩展taskSummary元信息；各入口useLoad按本人作品目录定位，拒绝错项目；补原生只读验收、构建和全量测试，保存本地v88。
