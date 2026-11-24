import { Button, Image, Text, View } from "@tarojs/components";
import Taro, { useDidHide, useDidShow } from "@tarojs/taro";
import { useEffect, useRef, useState } from "react";
import { API_ORIGIN_CONFIGURATION_ERROR, isApiOriginConfigured } from "../../config/runtime";
import {onSessionInvalidated} from "../../services/sessionBoundary";
import { apiRequest } from "../../services/api";
import { deriveTaskPageState } from "../../services/taskPageState";
import { filterTaskEntries, loadTaskSummary, taskCounts, taskDestination, type AggregatedTask, type TaskFilter, type TaskSummary } from "../../services/taskSummary";
import BrandHeader from "../../components/BrandHeader";
import AppIcon from "../../components/AppIcon";
import taskHeroArt from "../../assets/tasks/task-center-runtime-hero-v2.webp";
import taskRunningArt from "../../assets/functions/clapper-play-v2.webp";
import taskAttentionArt from "../../assets/tasks/task-attention-clock.webp";
import taskCompletedArt from "../../assets/tasks/task-completed-check.webp";
import { ThemedPage } from "../../theme";
import { taskVisualAsset } from "./taskVisuals";
import "./tasks-studio.scss";

const FILTERS: TaskFilter[] = ["全部", "进行中", "需重试", "已完成"];
function statusLabel(entry: AggregatedTask) {
  if (entry.status === "running") return "进行中";
  if (entry.status === "retry") return "需重试";
  if (entry.status === "completed") return entry.kind==='automation'?'流程已结束':'已完成';
  return entry.statusDetail || (entry.rawStatus ? `状态：${entry.rawStatus}` : "状态未返回");
}

function formatTimestamp(value: string | undefined) {
  if (!value) return "时间未返回";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString("zh-CN", { month: "numeric", day: "numeric", hour: "2-digit", minute: "2-digit" });
}

function validProgress(entry: AggregatedTask) {
  const completed = entry.progress?.completed;
  const total = entry.progress?.total;
  if (!Number.isFinite(completed) || !Number.isFinite(total) || Number(total) <= 0) return undefined;
  return { completed: Math.max(0, Number(completed)), total: Number(total), percent: Math.min(100, Math.max(0, Math.round((Number(completed) / Number(total)) * 100))) };
}

function aggregateProgress(entries: AggregatedTask[]) {
  const active=entries.filter(entry=>entry.status==='running');
  return active.length===1?validProgress(active[0])?.percent:undefined;
}

export default function TasksPage() {
  const [summary, setSummary] = useState<TaskSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [retryingId, setRetryingId] = useState("");
  const [configurationError, setConfigurationError] = useState("");
  const [filter, setFilter] = useState<TaskFilter>("全部");
  const [page,setPage]=useState(0);
  const requestVersion = useRef(0);
  useEffect(()=>{const off=onSessionInvalidated(()=>{requestVersion.current++;setSummary(null);setLoading(false);setRefreshing(false);setConfigurationError('账号已切换，请重新进入任务中心');});return()=>{off();requestVersion.current++;};},[]);

  const load = () => {
    const version = ++requestVersion.current;

    if (!isApiOriginConfigured()) { setConfigurationError(API_ORIGIN_CONFIGURATION_ERROR); setLoading(false); setRefreshing(false); return; }
    setPage(0);setConfigurationError(""); setLoading(!summary); setRefreshing(Boolean(summary));
    void loadTaskSummary(apiRequest,{shouldContinue:()=>version===requestVersion.current,onProgress:partial=>{
      if(version===requestVersion.current&&partial.sources.some(source=>source.state==='ready')){setSummary(partial);setLoading(false);setRefreshing(true);}
    }})
      .then((result) => {if (version === requestVersion.current) setSummary(result); })
      .catch(() => { if (version === requestVersion.current) setSummary(null); })
      .finally(() => { if (version === requestVersion.current) { setLoading(false); setRefreshing(false); } });
  };

  useDidShow(load);
  useDidHide(()=>{requestVersion.current++;setLoading(false);setRefreshing(false);});
  const pageState = deriveTaskPageState({ loading, summary, configurationError });
  const entries = summary?.entries || [];
  const overallProgress = aggregateProgress(entries);
  const filtered = filterTaskEntries(entries, filter);
  const counts = taskCounts(entries);
  const pages=Math.max(1,Math.ceil(filtered.length/30)),currentPage=Math.min(page,pages-1);
  const visibleEntries=filtered.slice(currentPage*30,currentPage*30+30);
  const failedSources = summary?.sources.filter((source) => source.state === "error") || [];
  const openTask = (entry: AggregatedTask) => { if (entry.id) void Taro.navigateTo({ url: taskDestination(entry) }); };
  const retryTask = async (entry: AggregatedTask, event?: { stopPropagation?: () => void }) => {
    event?.stopPropagation?.();
    if (!entry.retryable || entry.source !== "fetch" || !entry.id || retryingId) { openTask(entry); return; }
    setRetryingId(entry.id);
    try {
      await apiRequest("/api/batch-rewrite/tasks/batch-retry", { method: "POST", timeout: 180000, data: { mode: "selected", ids: [entry.id] } });
      Taro.showToast({ title: "已提交重试", icon: "success" });
      load();
    } catch (cause) {
      Taro.showToast({ title: cause instanceof Error ? cause.message : "重试失败", icon: "none" });
    } finally { setRetryingId(""); }
  };

  return <ThemedPage className="mp-page mp-primary-page mp-tasks-page task-studio-page" dynamicGlass={false}>
    <View className="task-studio-header-scene">
      <BrandHeader title="任务" primary />
      <View className="task-studio-hero">
        <View className="task-studio-hero-copy"><Text className="task-studio-title">任务中心</Text><Text className="task-studio-subtitle">追踪全部创作与生产记录</Text><Text className="task-studio-title-stroke" /></View>
        <Image className="task-studio-hero-art" src={taskHeroArt} mode="aspectFit" />
        {overallProgress !== undefined ? <View className="task-studio-hero-progress"><Text className="task-studio-hero-progress-value">{overallProgress}%</Text></View> : null}
        <Button className="task-studio-refresh" aria-label="刷新任务" loading={refreshing} disabled={loading} onClick={load}><AppIcon name="refresh-ccw" size={16} /></Button>
      </View>
    </View>

    {pageState.kind === "configuration-error" ? <View className="task-studio-state liquid-glass--secondary"><View><Text className="task-studio-state-title">服务端未配置</Text><Text>{pageState.message}</Text></View></View> : null}
    {pageState.kind === "loading" ? <View className="task-studio-state liquid-glass--secondary"><View><Text className="task-studio-state-title">正在读取任务</Text><Text>正在分别读取各个真实服务端来源。</Text></View></View> : null}
    {pageState.kind === "all-error" ? <View className="task-studio-state liquid-glass--secondary"><View><Text className="task-studio-state-title">暂时无法读取任务</Text><Text>所有任务来源均未返回可展示数据。</Text></View><Button className="task-studio-retry" onClick={load}>重新读取</Button></View> : null}

    {summary && pageState.kind !== "configuration-error" && pageState.kind !== "all-error" ? <View className="task-studio-body">
      {refreshing?<Text className="task-studio-sync-note">已返回的记录先显示，其余来源仍在同步，统计尚未完整。</Text>:null}
      <View className="task-studio-counts liquid-glass--secondary" data-glass-role="secondary">
        <View className="task-studio-count-item is-running"><Image className="task-studio-count-art" src={taskRunningArt} mode="aspectFit" /><View><Text className="task-studio-count-value">{counts.running}</Text><Text>进行中</Text></View></View><View className="task-studio-count-divider" />
        <View className="task-studio-count-item is-retry"><Image className="task-studio-count-art" src={taskAttentionArt} mode="aspectFit" /><View><Text className="task-studio-count-value">{counts.retry}</Text><Text>需重试</Text></View></View><View className="task-studio-count-divider" />
        <View className="task-studio-count-item is-completed"><Image className="task-studio-count-art" src={taskCompletedArt} mode="aspectFit" /><View><Text className="task-studio-count-value">{counts.completed}</Text><Text>已完成</Text></View></View>
      </View>
      <View className="task-studio-segments liquid-glass--list" data-glass-role="list">{FILTERS.map((item) => <Text key={item} className={filter === item ? "task-studio-filter is-active" : "task-studio-filter"} onClick={() => {setFilter(item);setPage(0);}}>{item}</Text>)}</View>
      {failedSources.map((source) => <View className="task-studio-source-error" key={source.source}><View><Text className="task-studio-state-title">{source.sourceLabel} 暂未同步</Text><Text>{source.error}</Text></View><Button className="task-studio-retry" onClick={load}>重试</Button></View>)}
      <View className="task-studio-list liquid-glass--list" data-glass-role="list">
        {visibleEntries.map((entry, index) => {
          const progress = validProgress(entry); const visual = taskVisualAsset(entry); const disabled = !entry.id;
          return <View className={`task-studio-row is-${entry.status}${disabled ? " is-disabled" : ""}${index === visibleEntries.length - 1 ? " is-last" : ""}`} key={`${entry.source}-${entry.id || index}`} onClick={() => openTask(entry)}>
            <Image className="task-studio-row-art" src={visual} mode="aspectFit" />
            <View className="task-studio-row-copy"><Text className="task-studio-row-title">{entry.title || "服务端未返回标题"}</Text><Text className="task-studio-row-source">{entry.sourceLabel} · {formatTimestamp(entry.createdAt)}</Text></View>
            <View className="task-studio-row-state">{progress ? <><Text className="task-studio-progress-value">{entry.kind==='automation'?`就绪 ${progress.completed}/${progress.total}`:`${progress.percent}%`}</Text><View className="task-studio-progress-track"><View style={{ width: `${progress.percent}%` }} /></View></> : null}<Text className={`task-studio-status mp-status-pill is-${entry.status}`}>{disabled ? "记录 ID 未返回" : statusLabel(entry)}</Text>{entry.status === "retry" && !disabled ? <Button className="task-studio-row-retry" loading={retryingId === entry.id} onClick={(event) => void retryTask(entry, event)}> {entry.retryable ? "重试" : "查看处理"}</Button> : null}</View>
            {!disabled ? <AppIcon className="task-studio-row-arrow" name="chevron-right" size={17} /> : null}
          </View>;
        })}
        {filtered.length>30?<View className="task-studio-pagination"><Button className="task-studio-prev" disabled={currentPage===0} onClick={()=>{setPage(currentPage-1);void Taro.pageScrollTo?.({scrollTop:0,duration:150});}}>上一页</Button><Text>第 {currentPage+1} / {pages} 页 · 共 {filtered.length} 条</Text><Button className="task-studio-next" disabled={currentPage>=pages-1} onClick={()=>{setPage(currentPage+1);void Taro.pageScrollTo?.({scrollTop:0,duration:150});}}>下一页</Button></View>:null}
        {pageState.kind === "empty" || (pageState.kind === "content" && !filtered.length) ? <View className="task-studio-empty"><AppIcon name="circle-dashed" size={28} /><Text className="task-studio-state-title">{pageState.kind === "empty" ? "暂无真实任务记录" : `暂无${filter}任务`}</Text><Text>服务端返回新任务后会显示在这里。</Text></View> : null}
      </View>
    </View> : null}
  </ThemedPage>;
}
