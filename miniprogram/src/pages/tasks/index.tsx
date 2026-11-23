import { Button, Image, Text, View } from "@tarojs/components";
import Taro, { useDidShow } from "@tarojs/taro";
import { useEffect, useRef, useState } from "react";
import { API_ORIGIN_CONFIGURATION_ERROR, isApiOriginConfigured } from "../../config/runtime";
import { apiRequest } from "../../services/api";
import { deriveTaskPageState } from "../../services/taskPageState";
import { filterTaskEntries, loadTaskSummary, taskCounts, taskDestination, type AggregatedTask, type TaskFilter, type TaskSummary } from "../../services/taskSummary";
import BrandHeader from "../../components/BrandHeader";
import AppIcon from "../../components/AppIcon";
import { FUNCTION_VISUALS } from "../functions/functionVisuals";
import { ThemedPage } from "../../theme";
import "./tasks-studio.scss";

const FILTERS: TaskFilter[] = ["全部", "进行中", "待确认", "已完成"];
const SOURCE_VISUAL = { script: FUNCTION_VISUALS.script, agent: FUNCTION_VISUALS.agent, fetch: FUNCTION_VISUALS.fetch, batch: FUNCTION_VISUALS.batch, shuihuo: FUNCTION_VISUALS.shuihuo };

function statusLabel(entry: AggregatedTask) {
  if (entry.status === "running") return "进行中";
  if (entry.status === "attention") return "待确认";
  if (entry.status === "failed") return "失败";
  if (entry.status === "completed") return "已完成";
  return entry.rawStatus ? `状态：${entry.rawStatus}` : "状态同步中";
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

export default function TasksPage() {
  const [summary, setSummary] = useState<TaskSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [configurationError, setConfigurationError] = useState("");
  const [filter, setFilter] = useState<TaskFilter>("全部");
  const requestVersion = useRef(0);
  useEffect(() => () => { requestVersion.current += 1; }, []);

  const load = () => {
    const version = ++requestVersion.current;
    if (!isApiOriginConfigured()) { setConfigurationError(API_ORIGIN_CONFIGURATION_ERROR); setLoading(false); setRefreshing(false); return; }
    setConfigurationError(""); setLoading(!summary); setRefreshing(Boolean(summary));
    void loadTaskSummary(apiRequest)
      .then((result) => { if (version === requestVersion.current) setSummary(result); })
      .catch(() => { if (version === requestVersion.current) setSummary(null); })
      .finally(() => { if (version === requestVersion.current) { setLoading(false); setRefreshing(false); } });
  };

  useDidShow(load);
  const pageState = deriveTaskPageState({ loading, summary, configurationError });
  const entries = summary?.entries || [];
  const filtered = filterTaskEntries(entries, filter);
  const counts = taskCounts(entries);
  const failedSources = summary?.sources.filter((source) => source.state === "error") || [];
  const openTask = (entry: AggregatedTask) => { if (entry.id) void Taro.navigateTo({ url: taskDestination(entry) }); };

  return <ThemedPage className="mp-page mp-primary-page mp-tasks-page task-studio-page">
    <BrandHeader title="任务" primary />
    <View className="task-studio-hero">
      <View><Text className="task-studio-kicker">一战晟铭 · AI 创意套件</Text><Text className="task-studio-title">任务中心</Text><Text className="task-studio-subtitle">追踪全部创作与生产记录</Text></View>
      <Button className="task-studio-refresh" aria-label="刷新任务" loading={refreshing} disabled={loading} onClick={load}><AppIcon name="refresh-ccw" size={18} /></Button>
    </View>

    {pageState.kind === "configuration-error" ? <View className="task-studio-state liquid-glass--secondary"><View><Text className="task-studio-state-title">服务端未配置</Text><Text>{pageState.message}</Text></View></View> : null}
    {pageState.kind === "loading" ? <View className="task-studio-state liquid-glass--secondary"><View><Text className="task-studio-state-title">正在读取任务</Text><Text>正在分别读取各个真实服务端来源。</Text></View></View> : null}
    {pageState.kind === "all-error" ? <View className="task-studio-state liquid-glass--secondary"><View><Text className="task-studio-state-title">暂时无法读取任务</Text><Text>所有任务来源均未返回可展示数据。</Text></View><Button className="task-studio-retry" onClick={load}>重新读取</Button></View> : null}

    {summary && pageState.kind !== "configuration-error" && pageState.kind !== "all-error" ? <View className="task-studio-body">
      <View className="task-studio-counts liquid-glass--secondary" data-glass-role="secondary">
        <View className="task-studio-count-item is-running"><AppIcon name="clapperboard" size={24} /><View><Text className="task-studio-count-value">{counts.running}</Text><Text>进行中</Text></View></View><View className="task-studio-count-divider" />
        <View className="task-studio-count-item is-attention"><AppIcon name="timer" size={24} /><View><Text className="task-studio-count-value">{counts.attention}</Text><Text>待确认</Text></View></View><View className="task-studio-count-divider" />
        <View className="task-studio-count-item is-completed"><AppIcon name="shield-check" size={24} /><View><Text className="task-studio-count-value">{counts.completed}</Text><Text>已完成</Text></View></View>
      </View>
      <View className="task-studio-segments liquid-glass--list" data-glass-role="list">{FILTERS.map((item) => <Text key={item} className={filter === item ? "task-studio-filter is-active" : "task-studio-filter"} onClick={() => setFilter(item)}>{item}</Text>)}</View>
      {failedSources.map((source) => <View className="task-studio-source-error" key={source.source}><View><Text className="task-studio-state-title">{source.sourceLabel} 暂未同步</Text><Text>{source.error}</Text></View><Button className="task-studio-retry" onClick={load}>重试</Button></View>)}
      <View className="task-studio-list liquid-glass--list" data-glass-role="list">
        {filtered.map((entry, index) => {
          const progress = validProgress(entry); const visual = SOURCE_VISUAL[entry.source]; const disabled = !entry.id;
          return <View className={`task-studio-row is-${entry.status}${disabled ? " is-disabled" : ""}`} key={`${entry.source}-${entry.id || index}`} onClick={() => openTask(entry)}>
            <Image className="task-studio-row-art" src={visual.asset} mode="aspectFit" />
            <View className="task-studio-row-copy"><Text className="task-studio-row-title">{entry.title || "服务端未返回标题"}</Text><Text className="task-studio-row-source">{entry.sourceLabel} · {formatTimestamp(entry.createdAt)}</Text></View>
            <View className="task-studio-row-state">{progress ? <><Text className="task-studio-progress-value">{progress.percent}%</Text><View className="task-studio-progress-track"><View style={{ width: `${progress.percent}%` }} /></View></> : null}<Text className={`task-studio-status is-${entry.status}`}>{disabled ? "记录 ID 未返回" : statusLabel(entry)}</Text></View>
            {!disabled ? <AppIcon className="task-studio-row-arrow" name="chevron-right" size={17} /> : null}
          </View>;
        })}
        {pageState.kind === "empty" || (pageState.kind === "content" && !filtered.length) ? <View className="task-studio-empty"><AppIcon name="circle-dashed" size={28} /><Text className="task-studio-state-title">{pageState.kind === "empty" ? "暂无真实任务记录" : `暂无${filter}任务`}</Text><Text>服务端返回新任务后会显示在这里。</Text></View> : null}
      </View>
    </View> : null}
  </ThemedPage>;
}
