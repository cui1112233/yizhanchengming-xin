import Taro, { useDidShow,useDidHide } from "@tarojs/taro";
import { Button, Checkbox, CheckboxGroup, Input, Picker, ScrollView, Text, Textarea, View } from "@tarojs/components";
import { useEffect, useRef, useState, type ReactNode } from "react";
import BrandHeader from "../../components/BrandHeader";
import AppIcon from "../../components/AppIcon";
import { apiRequest as rawApiRequest } from "../../services/api";
import { sessionConfirmation } from "../../services/sessionConfirmation";
import { createFetchRequestGuard } from "../../services/fetchWorkspaceRequests";
import { previewFetchSubmission, submitFetchSubmission } from "../../services/fetchSubmission";
import {
  buildFetchProcessPayload,
  fetchTaskOverallStatus,
  fetchTaskStages,
  fetchWorkspaceState,
  normalizeFetchTasks,
  taskDateKey,
  type FetchPlatform,
  type FetchTask,
  type FetchTextModel,
} from "../../services/fetchWorkspace";
import { withThemedPage } from "../../theme";
import FetchMaintenancePanel from "./FetchMaintenancePanel";
import {useBooktestJob} from "./useBooktestJob";
import {booktestInput,BOOKTEST_METHODS,fetchVersionTexts} from "../../services/booktestJobs";
import { fetchDetailBody } from "../../services/fetchDetailBody";
import { writeFetchMaintenance } from "../../services/fetchMaintenance";
import "./script-studio.scss";
import "./fetch-studio.scss";
import FetchAccountScope from "./FetchAccountScope";
import BooktestManagement from "./BooktestManagement";
import type {FetchRequest} from "../../services/booktestManagement";

type Schedule = { status?:string; id?: string; runAt?: string; run_at?: string; enabled?: boolean; autoSubmit?: boolean; auto_submit?: boolean };
type SettingsPanel = "queue" | "knowledge" | "batch" | "" | "parameters" | "versions" | "schedule" | "schedules" | "web" | "logs" | "config" | "prompts";

const VERSION_OPTIONS = ["original", "ai1", "ai2", "ai3", "ai4", "ai5"];

type TaskDetail = {
  task: FetchTask;
  loading: boolean;
  payload: Record<string, unknown> | null;
};

function todayDateKey() {
  const date = new Date();
  const local = new Date(date.getTime() - date.getTimezoneOffset() * 60_000);
  return local.toISOString().slice(0, 10);
}

function taskTitle(task: FetchTask) {
  return task.bookName || task.bookId || "未命名任务";
}

function object(value: unknown): Record<string, unknown> {
  return value && typeof value === "object" && !Array.isArray(value) ? value as Record<string, unknown> : {};
}

function valueText(value: unknown) {
  return typeof value === "string" ? value : value === undefined || value === null ? "" : String(value);
}

function taskDetailVersions(payload: Record<string, unknown>) {return fetchVersionTexts(payload);}

function displaySchedule(schedule: Schedule) {
  const value = String(schedule.runAt || schedule.run_at || "");
  if (!value) return "服务端未返回时间";
  const date = new Date(value);
  return Number.isFinite(date.getTime()) ? date.toLocaleString("zh-CN", { hour12: false }) : value;
}

export function FetchWorkspaceBody({title="小说获取",compact=false,initialTaskId="",request:apiRequest=rawApiRequest,scopeKey="self",isSelf=true,accountControl}:{title?:string;compact?:boolean;initialTaskId?:string;request?:FetchRequest;scopeKey?:string;isSelf?:boolean;accountControl?:ReactNode}) {
  const [inputExpanded, setInputExpanded] = useState(true);
  const [tab, setTab] = useState(initialTaskId ? "任务" : "处理");
  const [input, setInput] = useState("");
  const [platforms, setPlatforms] = useState<FetchPlatform[]>([]);
  const [platformId, setPlatformId] = useState("");
  const [models, setModels] = useState<FetchTextModel[]>([]);
  const [textModelId, setTextModelId] = useState("");
  const [maxTxt, setMaxTxt] = useState(4000);
  const [selectedVersions, setSelectedVersions] = useState<string[]>(["original", "ai1"]);
  const [aiSlotMethods, setAiSlotMethods] = useState<Record<string, string>>({});
  const [sensitiveAiEnabled, setSensitiveAiEnabled] = useState(false);
  const [tasks, setTasks] = useState<FetchTask[]>([]);
  const [taskDate, setTaskDate] = useState("");
  const [selectedTaskId, setSelectedTaskId] = useState("");
  const [taskDetail, setTaskDetail] = useState<TaskDetail | null>(null);
  const [taskActionId, setTaskActionId] = useState("");
  const [schedules, setSchedules] = useState<Schedule[]>([]);
  const [scheduleAt, setScheduleAt] = useState("");
  const [editingSchedule, setEditingSchedule] = useState<Schedule | null>(null);
  const [scheduleAutoSubmit, setScheduleAutoSubmit] = useState(false);
  const scheduleBusy = useRef(false), scheduleUnknown = useRef(false);
  const [settingsPanel, setSettingsPanel] = useState<SettingsPanel>("");
  const [processResult, setProcessResult] = useState<Record<string, unknown> | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const requestGuard = useRef(createFetchRequestGuard()).current;
  const { config: configVersion, tasks: taskVersion, schedules: scheduleVersion, detail: detailVersion } = requestGuard.versions;
  const settingsVersion = useRef(0), refreshVersion = useRef(0);
  const [modelSaving, setModelSaving] = useState(false);
  const unknownOutcome = useRef(false);
  const formUnknown = useRef(false);
  const operationBusy = useRef(false);
  const closeTaskDetail = () => { requestGuard.next("detail"); setTaskDetail(null); setSelectedTaskId(""); setSubmitResult(null); };
  const editConfig = (edit: () => void) => { requestGuard.next("config"); edit(); };
  const [submitResult, setSubmitResult] = useState<{ bookId: string; payload: Record<string, any> } | null>(null);
  useEffect(() => () => { configVersion.current += 1; taskVersion.current += 1; scheduleVersion.current += 1; detailVersion.current += 1; settingsVersion.current += 1; refreshVersion.current += 1; }, []);

  const [toolsOpen,setToolsOpen]=useState(!compact);
  const [tasksRead,setTasksRead]=useState<"loading"|"ready"|"error">("loading");
  const focusedTask=useRef("");
  useEffect(()=>{if(!initialTaskId||tasksRead!=="ready"||focusedTask.current===initialTaskId)return;const found=tasks.find(task=>task.bookId===initialTaskId);if(!found){setError("未找到指定任务，请核对账号或刷新任务列表");return;}focusedTask.current=initialTaskId;setTab("任务");setSelectedTaskId(found.bookId);void openTaskDetail(found);},[tasks,initialTaskId,tasksRead]);

  const loadTasks = (silent=false) => {
    if(!silent)setTasksRead("loading");
    const version = ++taskVersion.current;
    void apiRequest<{ tasks?: unknown }>("/api/batch-rewrite/tasks")
      .then((payload) => { if (version !== taskVersion.current) return; if (!Array.isArray(payload?.tasks)) throw new Error("任务列表响应格式无效"); setTasks(normalizeFetchTasks(payload.tasks)); setTasksRead("ready"); })
      .catch((cause: Error) => { if (version === taskVersion.current) {setTasksRead("error");setError(cause.message || "任务读取失败");} });
  };

  const job=useBooktestJob(compact,()=>loadTasks(true),apiRequest,scopeKey);

  const loadSchedules = () => {
    const version = ++scheduleVersion.current;
    void apiRequest<{ schedules?: Schedule[] }>("/api/batch-rewrite/schedules")
      .then((payload) => { if (version !== scheduleVersion.current) return; if (!Array.isArray(payload?.schedules)) throw new Error("定时任务响应格式无效"); setSchedules(payload.schedules); scheduleUnknown.current = false; })
      .catch((cause: Error) => { if (version === scheduleVersion.current) setError(cause.message || "定时任务读取失败"); });
  };

  const refresh = () => {
    const snapshot = requestGuard.beginRefresh();
    const version = snapshot.config;
    const refreshId = ++refreshVersion.current;
    setRefreshing(true);
    setTasksRead("loading");
    setError("");
    Promise.allSettled([
      apiRequest<unknown>("/api/batch-rewrite/config"),
      apiRequest<unknown>("/api/models?kind=text"),
      apiRequest<{ tasks?: unknown }>("/api/batch-rewrite/tasks"),
      apiRequest<{ schedules?: Schedule[] }>("/api/batch-rewrite/schedules"),
    ]).then(([configResult, modelResult, taskResult, scheduleResult]) => {
      const failures: string[] = [];
      const failure = (label: string, reason: unknown) => failures.push(`${label}：${reason instanceof Error ? reason.message : "读取失败"}`);
      if (configResult.status === "rejected") failure("配置", configResult.reason);
      if (modelResult.status === "rejected") failure("模型", modelResult.reason);
      if (requestGuard.current("config", version) && configResult.status === "fulfilled") {
      const modelPayload = modelResult.status === "fulfilled" ? modelResult.value : models;
      const modelsPayload = Array.isArray(modelPayload) ? modelPayload : (modelPayload as { models?: unknown[] } | null)?.models || [];
      const state = fetchWorkspaceState(configResult.value, modelsPayload);
      setPlatforms(state.platforms);
      setPlatformId((current) => state.platforms.some((platform) => platform.id === current) ? current : state.platformId);
      setModels(state.models);
      setTextModelId(state.textModelId);
      setMaxTxt(state.maxTxt);
      setSelectedVersions(state.selectedVersions);
      setAiSlotMethods(state.aiSlotMethods);
      setSensitiveAiEnabled(state.sensitiveAiEnabled);
      formUnknown.current = false;
      }
      if (requestGuard.current("tasks", snapshot.tasks)) {
        if (taskResult.status === "fulfilled" && Array.isArray(taskResult.value?.tasks)) { setTasks(normalizeFetchTasks(taskResult.value.tasks)); setTasksRead("ready"); unknownOutcome.current = false; }
        else {setTasksRead("error");failure("任务", taskResult.status === "rejected" ? taskResult.reason : new Error("响应格式无效"));}
      }
      if (requestGuard.current("schedules", snapshot.schedules)) {
        if (scheduleResult.status === "fulfilled" && Array.isArray(scheduleResult.value?.schedules)) setSchedules(scheduleResult.value.schedules);
        else failure("定时任务", scheduleResult.status === "rejected" ? scheduleResult.reason : new Error("响应格式无效"));
      }
      if (refreshId === refreshVersion.current) setError(failures.join("；"));
    }).catch((cause: Error) => { if (refreshId === refreshVersion.current) setError(cause.message || "工作台刷新失败"); })
      .finally(() => { if (refreshId === refreshVersion.current) setRefreshing(false); });
  };

  const visibleCycle=useRef(false);
  useEffect(()=>{if(!visibleCycle.current){visibleCycle.current=true;refresh();}},[]);
  useDidShow(()=>{if(!visibleCycle.current){visibleCycle.current=true;refresh();}});
  useDidHide(()=>{visibleCycle.current=false;});

  const saveTextModel = async (modelId: string) => {
    if (!requestGuard.beginModelSave()) return;
    editConfig(() => {});
    setModelSaving(true);
    setTextModelId(modelId);
    try {
      await apiRequest("/api/batch-rewrite/config", { method: "POST", data: { app_config: { text_model_id: modelId } } });
      const confirmed = await apiRequest<any>("/api/batch-rewrite/config");
      const state = fetchWorkspaceState(confirmed, models);
      const persistedId = confirmed?.app_config?.text_model_id || confirmed?.app_config?.textModelId;
      if (persistedId !== modelId) throw new Error("文本模型回读不一致，请刷新核对后重试");
      setTextModelId(state.textModelId);
      Taro.showToast({ title: "文本模型已保存", icon: "success" });
    } catch (cause: any) {
      setError(cause?.message || "文本模型保存失败");
      try {
        const current = await apiRequest<any>("/api/batch-rewrite/config");
        setTextModelId(String(current?.app_config?.text_model_id || current?.app_config?.textModelId || ""));
      } catch {
        setTextModelId("");
        setError("文本模型保存结果未知，回读失败；请刷新核对后再处理");
      }
    } finally {
      configVersion.current += 1;
      requestGuard.endModelSave();
      setModelSaving(false);
    }
  };

  const saveWorkForm = async () => {
    if (operationBusy.current || loading) return;
    if (formUnknown.current) { setError("处理规则保存结果未知，请刷新配置核对后再保存"); return; }
    if (!selectedVersions.length) { setError("请至少选择一个文案版本"); return; }
    operationBusy.current = true;
    setLoading(true);
    setError("");
    try {
      await writeFetchMaintenance(apiRequest, "/api/batch-rewrite/work-form", "POST", {
          state: {
            selected_versions: selectedVersions,
            ai_slot_methods: aiSlotMethods,
            sensitive_ai_enabled: sensitiveAiEnabled,
          },
        }, "/api/batch-rewrite/config");
      Taro.showToast({ title: "处理规则已读回", icon: "success" });
      setSettingsPanel("");
    } catch (cause: any) {
      formUnknown.current = String(cause?.message).includes("结果未知");
      setError(cause?.message || "处理规则保存失败");
    } finally {
      operationBusy.current = false;
      setLoading(false);
    }
  };

  const process = async () => {
    if (operationBusy.current || loading || modelSaving) return;
    if (unknownOutcome.current) { setError("上次操作结果未知，请先刷新任务核对，避免重复处理"); return; }
    if (!platforms.some((item) => item.id === platformId)) { setError("请选择服务端提供的书城平台"); return; }
    const payload = buildFetchProcessPayload({ input, platformId, maxTxt, selectedVersions, aiSlotMethods, sensitiveAiEnabled });
    if (!payload.input_text) {
      setError("请先粘贴小说或书单");
      return;
    }
    if(compact){
      operationBusy.current=true;setLoading(true);setError("");
      try{const inputPayload=booktestInput({input,platformId,maxTxt,selectedVersions,aiSlotMethods,sensitiveAiEnabled},textModelId);
        const confirmation=await sessionConfirmation(()=>Taro.showModal({title:"开始批量测书？",content:"将获取原文、按选中版本调用文本模型改文，并按线上流程自动提交视频管理系统，可能产生费用。是否继续？"}));
        if(!confirmation.confirm)return;
        if(await job.start(inputPayload)){setTab("任务");loadTasks();}
      }catch(cause:any){setError(cause.message||"启动测书失败");}finally{operationBusy.current=false;setLoading(false);}
      return;
    }
    setError("");
    operationBusy.current = true;
    taskVersion.current += 1;
    setLoading(true);
    try {
      const result = await apiRequest<Record<string, unknown> & { tasks?: unknown }>("/api/batch-rewrite/process", { method: "POST", data: payload, timeout: 180000 });
      if (!Array.isArray(result?.tasks) || result.ok === false || result.error) throw new Error(String(result?.message || result?.error || "处理接口未返回有效任务，请刷新核对"));
      setProcessResult(result);
      setTasks(normalizeFetchTasks(result.tasks));
      setTasksRead("ready");
      setTab("任务");
      Taro.showToast({ title: "小说处理已返回", icon: "success" });
    } catch (cause: any) {
      unknownOutcome.current = true;
      setError(`${cause?.message || "小说处理未确认"}。结果未知，服务端可能仍在处理；请先刷新任务核对，勿直接重复提交。`);
    } finally {
      operationBusy.current = false;
      setLoading(false);
    }
  };

  const retryBatch = async (ids?: string[]) => {
    if (operationBusy.current || unknownOutcome.current) { setError("请先刷新任务核对上次操作结果，再决定是否重试"); return; }
    operationBusy.current = true;
    taskVersion.current += 1;
    setLoading(true);
    setError("");
    try {
      const result = await apiRequest<any>("/api/batch-rewrite/tasks/batch-retry", { method: "POST", timeout: 180000, data: ids ? { mode: "selected", ids } : { mode: "failed" } });
      if (!result || result.ok === false || result.error || !Array.isArray(result.tasks) || ![result.retried, result.failed, result.skipped].every((count) => typeof count === "number" && Number.isFinite(count))) throw new Error("重试接口未返回有效结果");
      setTasks(normalizeFetchTasks(result.tasks));
      setTasksRead("ready");
      const summary = `重试已返回：完成 ${result.retried}，失败 ${result.failed}，跳过 ${result.skipped}`;
      if (result.failed > 0) setError(summary);
      else Taro.showToast({ title: summary, icon: "none" });
      loadTasks();
    } catch (cause: any) {
      unknownOutcome.current = true;
      setError(`${cause?.message || "重试未确认"}。结果未知，服务端可能仍在处理；请先刷新任务核对，勿重复重试。`);
    } finally {
      operationBusy.current = false;
      setLoading(false);
    }
  };
  const retryFailed = () => retryBatch();

  const openTaskDetail = async (task: FetchTask) => {
    const version = ++detailVersion.current;
    setError("");
    setTaskDetail({ task, loading: true, payload: null });
    try {
      const payload = await apiRequest<Record<string, unknown>>(`/api/batch-rewrite/tasks/${encodeURIComponent(task.bookId)}`);
      if (version !== detailVersion.current) return;
      setTaskDetail({ task, loading: false, payload });
    } catch (cause: any) {
      if (version !== detailVersion.current) return;
      setTaskDetail(null);
      setError(cause?.message || "任务详情读取失败");
    }
  };

  const runTaskAction = async (task: FetchTask, successTitle: string, request: () => Promise<unknown>) => {
    if (operationBusy.current || unknownOutcome.current) { setError("请先刷新任务核对上次操作结果，再决定是否重试"); return false; }
    operationBusy.current = true;
    taskVersion.current += 1;
    setTaskActionId(task.bookId);
    setError("");
    try {
      const result = await request() as any;
      if (!result || typeof result !== "object" || result.ok === false || result.error) throw new Error(result?.message || result?.error || "服务端未确认操作，请刷新核对");
      Taro.showToast({ title: successTitle, icon: "success" });
      loadTasks();
      if (taskDetail?.task.bookId === task.bookId) void openTaskDetail(task);
      return true;
    } catch (cause: any) {
      unknownOutcome.current = true;
      setError(`${cause?.message || "操作未确认"}。结果未知，请先刷新任务核对，避免重复操作。`);
      return false;
    } finally {
      setTaskActionId("");
      operationBusy.current = false;
    }
  };

  const refetchOriginal = (task: FetchTask) => runTaskAction(task, "原文抓取已返回", () =>
    apiRequest(`/api/batch-rewrite/tasks/${encodeURIComponent(task.bookId)}/fetch`, {
      method: "POST",
      timeout: 180000,
      data: { max_txt: maxTxt, sensitive_ai_enabled: sensitiveAiEnabled },
    }),
  );

  const generateAi = (task: FetchTask) => runTaskAction(task, "AI 生成已返回", () =>
    apiRequest(`/api/batch-rewrite/tasks/${encodeURIComponent(task.bookId)}/generate-ai`, {
      method: "POST",
      timeout: 180000,
      data: { selected_versions: selectedVersions, ai_slot_methods: aiSlotMethods },
    }),
  );

  const retryTask = (task: FetchTask) => retryBatch([task.bookId]);

  const deleteTask = async (task: FetchTask) => {
    const result = await sessionConfirmation(() => Taro.showModal({
      title: "删除任务",
      content: `确认删除《${taskTitle(task)}》？原文、AI 版本与日志会一起删除。`,
      confirmText: "删除",
      confirmColor: "#d45555",
    }));
    if (!result.confirm) return;
    const deleted = await runTaskAction(task, "任务已删除", () =>
      apiRequest("/api/batch-rewrite/tasks/batch-delete", { method: "POST", data: { mode: "selected", ids: [task.bookId] } }),
    );
    if (deleted && taskDetail?.task.bookId === task.bookId) closeTaskDetail();
  };

  const mutateSchedule = async (path: string, method: "POST" | "PATCH" | "DELETE", data?: unknown) => {
    if (scheduleBusy.current) return false;
    if (scheduleUnknown.current) { setError("定时操作结果未知，请先刷新定时任务核对"); return false; }
    scheduleBusy.current = true; setLoading(true); setError("");
    const version = ++scheduleVersion.current;
    try {
      const result = await writeFetchMaintenance(apiRequest, path, method, data, "/api/batch-rewrite/schedules");
      if (!Array.isArray(result?.schedules)) throw new Error("定时操作结果未知：回读格式无效，请刷新核对");
      if (version === scheduleVersion.current) setSchedules(result.schedules);
      return true;
    } catch (cause: any) { scheduleUnknown.current = String(cause.message).includes("结果未知"); setError(cause.message || "定时任务操作失败"); return false; }
    finally { scheduleBusy.current = false; setLoading(false); }
  };
  const createSchedule = async () => {
    if (loading || scheduleBusy.current) return;
    if (!editingSchedule && !platforms.some((item) => item.id === platformId)) { setError("请选择服务端提供的书城平台"); return; }
    if (!editingSchedule && !input.trim()) {
      setError("请先填写小说或书单输入");
      return;
    }
    const runAt = new Date(scheduleAt);
    if (!Number.isFinite(runAt.getTime()) || runAt.getTime() <= Date.now()) {
      setError("请输入晚于当前时间的执行时间");
      return;
    }
    let scheduleInput:Record<string,unknown>|null=null;
    if(compact&&!editingSchedule){try{scheduleInput=booktestInput({input,platformId,maxTxt,selectedVersions,aiSlotMethods,sensitiveAiEnabled},textModelId);}catch(cause:any){setError(cause.message);return;}}
    if(compact){const confirmation=await sessionConfirmation(()=>Taro.showModal({title:editingSchedule?"保存定时修改？":"创建定时测书？",content:`将在所选时间调用当前处理配置${scheduleAutoSubmit?"并自动提交视频管理系统":"，不自动提交网站"}，可能产生模型费用。`}));if(!confirmation.confirm)return;}
    const saved = await mutateSchedule(editingSchedule ? `/api/batch-rewrite/schedules/${encodeURIComponent(String(editingSchedule.id))}` : "/api/batch-rewrite/schedules", editingSchedule ? "PATCH" : "POST", editingSchedule ? (compact?{run_at:runAt.toISOString(),auto_submit:scheduleAutoSubmit}:{ runAt: runAt.toISOString(), autoSubmit: scheduleAutoSubmit }) : compact?{...scheduleInput,run_at:runAt.toISOString(),auto_submit:scheduleAutoSubmit}: {
          runAt: runAt.toISOString(),
          enabled: true,
          autoSubmit: scheduleAutoSubmit,
          inputSnapshot: buildFetchProcessPayload({ input, platformId, maxTxt, selectedVersions, aiSlotMethods, sensitiveAiEnabled }),
      });
    if (saved) {
      setScheduleAt("");
      setEditingSchedule(null); setSettingsPanel("schedules");
      Taro.showToast({ title: "定时任务已读回", icon: "success" });
    }
  };

  const openSettings = (panel: SettingsPanel) => {
    ++settingsVersion.current;
    setSettingsPanel(panel);
    if (panel === "schedule") { setEditingSchedule(null); setScheduleAt(""); setScheduleAutoSubmit(false); }
    if (panel === "schedules") loadSchedules();
  };

  const currentPlatformIndex = Math.max(0, platforms.findIndex((platform) => platform.id === platformId));
  const previewSubmit = async (task: FetchTask) => {
    if (loading) return;
    setLoading(true); setError(""); setSubmitResult(null);
    try { const payload = await previewFetchSubmission(apiRequest, task.bookId); setSubmitResult({ bookId: task.bookId, payload }); }
    catch (cause: any) { setError(cause.message || "提交预览失败"); }
    finally { setLoading(false); }
  };
  const confirmSubmit = async () => {
    if (!submitResult || loading || !submitResult.payload.groups?.length) return;
    const confirmation = await sessionConfirmation(() => Taro.showModal({ title: "开始网站提交", content: "将当前小说按服务端版本配置提交到目标网站，确认执行？" }));
    if (!confirmation.confirm) return;
    setLoading(true); setError("");
    try { const payload = await submitFetchSubmission(apiRequest, submitResult.bookId); setSubmitResult({ ...submitResult, payload }); loadTasks(); }
    catch (cause: any) { setError(cause.message || "提交结果未确认，请刷新任务状态"); }
    finally { setLoading(false); }
  };
  const currentModelIndex = Math.max(0, models.findIndex((model) => model.id === textModelId));
  const versionSummary = selectedVersions.map((version) => version === "original" ? "原文" : version.toUpperCase()).join("、") || "未选择版本";
  const visibleTasks = taskDate ? tasks.filter((task) => taskDateKey(task) === taskDate) : tasks;
  const selectedTask = tasks.find(task => task.bookId === selectedTaskId);
  const selectedStages = selectedTask ? fetchTaskStages(selectedTask) : [];

  return <View className={`mp-page mp-detail-page mp-fetch-page script-studio fetch-studio${compact?" fetch-booktest":""}`}><BrandHeader title={title} showBack />{accountControl}
    <View className="mp-fetch-hero studio-titlebar">
      <View><Text className="studio-eyebrow">STORY COLLECTION</Text><Text className="studio-title">{compact?"从书单到可用文案":"收集下一段好故事"}</Text><View className="mp-fetch-meta"><Text>{platforms.length} 个平台</Text><Text>{tasksRead==="ready"?tasks.length:"—"} 个任务</Text></View></View>
      <Button className="studio-history" loading={refreshing} disabled={loading || refreshing} onClick={refresh} aria-label={`刷新${title}`}><AppIcon name="refresh-ccw" size={19} /></Button>
    </View>
    <BooktestManagement request={apiRequest} scopeKey={scopeKey} isSelf={isSelf} date={taskDate} />
    {job.recovery?<View className="mp-card"><Text className="mp-error">上次启动结果待核对</Text><Text className="studio-caption">{job.recovery.at} · {String(job.recovery.input?.inputText||'').slice(0,120)}</Text><Button onClick={job.refresh}>读取最近作业</Button>{job.recent.filter(item=>!job.recovery.baselineIds?.includes(item.id)).map(item=><Button key={item.id} onClick={async()=>{const result=await sessionConfirmation(()=>Taro.showModal({title:'关联已有作业',content:`确认 ${item.id} 是上次提交对应的作业？不会重新启动。`}));if(result.confirm)await job.adopt(item.id);}}>{String(item.id)} · {String(item.status||'未知')} · 核对关联</Button>)}<Button onClick={async()=>{const result=await sessionConfirmation(()=>Taro.showModal({title:'解除等待记录',content:'请先核对服务端最近作业与任务正文；仅在确认上次没有启动作业后解除，否则再次处理可能重复产生费用。'}));if(result.confirm)job.release();}}>已核对未启动，解除等待</Button></View>:null}
    <View className="studio-tabs"><Button id="booktest-input-tab" className={`studio-tab ${tab === "处理" ? "is-active" : ""}`} onClick={() => setTab("处理")}>处理</Button><Button id="booktest-task-tab" className={`studio-tab ${tab === "任务" ? "is-active" : ""}`} onClick={() => { setTab("任务"); loadTasks(); }}>任务</Button></View>
    {error ? <View className="mp-fetch-alert"><Text>{error}</Text><Button className="mp-fetch-alert-action" onClick={refresh}>刷新核对</Button></View> : null}
    {compact&&(tab==="任务"||job.job||job.error)?<View className="booktest-job"><View className="studio-section-head"><Text>后台作业</Text><Button className="studio-link" onClick={()=>void job.refresh()}>核对最近作业</Button></View>{job.error?<Text className="mp-error">{job.error}</Text>:null}{job.job?<><Text className="mp-row-title">{({queued:"排队中",running:"正在处理",done:"处理完成",failed:"处理失败",stopped:"已停止",cancelled:"已取消"} as Record<string,string>)[job.job.status]||"状态待确认"}</Text><Text className="studio-caption">{job.job.id}</Text>{job.job.error?<Text className="mp-error">{String(job.job.error)}</Text>:null}{(Array.isArray(job.job.progress)?job.job.progress:[]).slice(-3).map((item:any,index:number)=><Text className="studio-caption" key={index}>{typeof item==="string"?item:String(item.message||item.detail||item.status||"")}</Text>)}{job.running?<Button className="studio-link" disabled={job.busy} onClick={async()=>{const confirmation=await sessionConfirmation(()=>Taro.showModal({title:"停止本次测书？",content:"将请求停止当前作业；已经完成的原文和文案保留，不改变全自动队列开关。"}));if(confirmation.confirm)try{await job.stop();}catch(cause:any){setError(cause.message||"停止请求结果未知，请刷新核对");}}}>停止本次任务</Button>:null}</>:<Text className="studio-caption">最近作业将在这里显示，进入页面不会自动执行。</Text>}</View>:null}
    {tab === "处理" ? <View className="mp-fetch-content">
      <View className="studio-paper"><View className="studio-paper-head"><Text>小说 / 书单输入</Text><Text className="studio-paper-tag">一行一本</Text></View>{inputExpanded ? <Textarea maxlength={-1} className="studio-source-input" value={input} onInput={(event) => setInput(event.detail.value)} placeholder={"粘贴书籍 ID、链接或现有书单\n书籍ID  书名  推荐理由  男女频  标签  评级"} /> : <Text className="studio-source-excerpt">{input || "还没有书单，展开后粘贴内容"}</Text>}<View className="studio-paper-foot"><Text>已输入 {input.split(/\r?\n/).filter(line => line.trim()).length} 行</Text><Button className="studio-link" onClick={() => setInputExpanded(!inputExpanded)}>{inputExpanded ? "收起书单" : "展开书单"}</Button></View></View>
      <View className="studio-section-head"><Text>处理设置</Text><Text className="studio-caption">按本次书单配置</Text></View>
      <View className="fetch-core-settings">
        <Picker mode="selector" range={platforms.map((platform) => platform.name)} value={currentPlatformIndex} onChange={(event) => editConfig(() => setPlatformId(platforms[Number(event.detail.value)]?.id || platformId))}><View className="mp-fetch-setting"><Text>目标平台</Text><View><Text>{platforms.find((platform) => platform.id === platformId)?.name || "默认平台"}</Text><AppIcon name="chevron-right" size={14} /></View></View></Picker>
        <Picker mode="selector" disabled={modelSaving} range={models.map((model) => model.name)} value={currentModelIndex} onChange={(event) => { const next = models[Number(event.detail.value)]?.id; if (next) void saveTextModel(next); }}><View className="mp-fetch-setting"><Text>文本模型</Text><View><Text>{models.find((model) => model.id === textModelId)?.name || "暂无可用模型"}</Text><AppIcon name="chevron-right" size={14} /></View></View></Picker>
        <Button id="booktest-parameters" className="mp-fetch-setting" onClick={() => openSettings("parameters")}><Text>处理参数</Text><View><Text>{maxTxt} 字</Text><AppIcon name="chevron-right" size={14} /></View></Button>
      </View>
      <View className="fetch-version-entry">
        <Button id="booktest-versions" className="mp-fetch-setting" onClick={() => openSettings("versions")}><Text>版本配置 / 处理规则</Text><View><Text>{versionSummary}</Text><AppIcon name="chevron-right" size={14} /></View></Button>
      </View>
      {compact?<Button id="booktest-tools" className="studio-link booktest-tools-toggle" onClick={()=>setToolsOpen(!toolsOpen)}>{toolsOpen?"收起更多工具":"更多工具 · 定时、规则与提交"}</Button>:null}
      {toolsOpen?<><View className="studio-section-head"><Text>工作工具</Text><Text className="studio-caption">定时 · 配置 · 提交</Text></View>
      <View className="fetch-tool-grid">
        <Button id="booktest-tool-config" className="mp-fetch-setting" onClick={() => openSettings("config")}><AppIcon name="sliders" size={19} /><Text className="fetch-tool-title">流程与抓取配置</Text><Text className="studio-caption">抓取与处理流程</Text></Button>
        <Button id="booktest-tool-prompts" className="mp-fetch-setting" onClick={() => openSettings("prompts")}><AppIcon name="notebook-tabs" size={19} /><Text className="fetch-tool-title">知识库 / 提示词 / 规则预览</Text><Text className="studio-caption">查看与维护规则</Text></Button>
        <Button id="booktest-tool-schedule" className="mp-fetch-setting" onClick={() => openSettings("schedule")}><Text>开始定时</Text><View><Text>按当前配置创建</Text><AppIcon name="chevron-right" size={14} /></View></Button>
        <Button id="booktest-tool-schedules" className="mp-fetch-setting" onClick={() => openSettings("schedules")}><Text>定时任务</Text><View><Text>{schedules.length ? `${schedules.length} 个` : "暂无"}</Text><AppIcon name="chevron-right" size={14} /></View></Button>
        <Button id="booktest-tool-web" className="mp-fetch-setting" onClick={() => openSettings("web")}><Text>网站提交</Text><View><Text>视频管理系统</Text><AppIcon name="chevron-right" size={14} /></View></Button>
        <Button id="booktest-tool-logs" className="mp-fetch-setting" onClick={() => openSettings("logs")}><Text>日志</Text><View><Text>运行记录</Text><AppIcon name="chevron-right" size={14} /></View></Button>
      </View>
      </>:null}
      {settingsPanel ? <View className="studio-overlay"><View className="studio-backdrop" onClick={() => setSettingsPanel("")} /><View className="studio-sheet"><View className="studio-sheet-handle" /><View className="studio-sheet-header"><Text>工作设置</Text><Button className="studio-link" onClick={() => setSettingsPanel("")}>收起面板</Button></View><ScrollView scrollY className="studio-sheet-scroll"><View className="studio-sheet-body">
      {settingsPanel === "parameters" ? <View className="mp-card mp-fetch-panel"><Text className="mp-row-title">处理参数</Text><Text className="mp-row-description">只截取用于处理的部分，完整原文独立保留。</Text><Input className="mp-input" type="number" value={String(maxTxt)} onInput={(event) => editConfig(() => setMaxTxt(Number(event.detail.value) || 4000))} /><Button className="mp-secondary" onClick={() => setSettingsPanel("")}>完成</Button></View> : null}
      {settingsPanel === "versions" ? <View className="mp-card mp-fetch-panel"><Text className="mp-row-title">版本配置 / 处理规则</Text><CheckboxGroup onChange={(event) => editConfig(() => setSelectedVersions(event.detail.value))}><View className="mp-fetch-version-list">{VERSION_OPTIONS.map((version) => <Checkbox key={version} value={version} checked={selectedVersions.includes(version)}><Text>{version === "original" ? "原文" : version.toUpperCase()}</Text></Checkbox>)}</View></CheckboxGroup>{selectedVersions.filter(version=>version!=="original").map(version=><Picker key={version} range={BOOKTEST_METHODS.map(method=>method.name)} value={Math.max(0,BOOKTEST_METHODS.findIndex(method=>method.id===(aiSlotMethods[version]||"auto")))} onChange={event=>editConfig(()=>setAiSlotMethods(current=>({...current,[version]:BOOKTEST_METHODS[Number(event.detail.value)]?.id||"auto"})))}><View className="mp-fetch-setting booktest-method"><Text>{version.toUpperCase()} 改文方式</Text><View><Text>{BOOKTEST_METHODS.find(method=>method.id===(aiSlotMethods[version]||"auto"))?.name||"请选择"}</Text><AppIcon name="chevron-right" size={14}/></View></View></Picker>)}<View className="mp-fetch-toggle-row"><View><Text className="mp-row-title">敏感内容处理</Text><Text className="mp-row-description">按服务端规则处理敏感文本。</Text></View><Text className={sensitiveAiEnabled ? "mp-switch mp-switch-on" : "mp-switch"} onClick={() => editConfig(() => setSensitiveAiEnabled(!sensitiveAiEnabled))}>{sensitiveAiEnabled ? "开" : "关"}</Text></View><Button className="mp-primary" loading={loading} onClick={saveWorkForm}>保存处理规则</Button></View> : null}
      {settingsPanel === "schedule" ? <View className="mp-card mp-fetch-panel"><Text className="mp-row-title">{editingSchedule ? "编辑定时任务" : "创建定时任务"}</Text><Text className="mp-row-description">本地时间，例如 2026-10-17T09:30。编辑保留原任务输入快照。</Text>{compact?<View className="booktest-date"><Picker mode="date" value={scheduleAt.slice(0,10)} onChange={event=>setScheduleAt(`${event.detail.value}T${scheduleAt.slice(11,16)||"09:00"}`)}><View className="mp-input">{scheduleAt.slice(0,10)||"选择日期"}</View></Picker><Picker mode="time" value={scheduleAt.slice(11,16)||"09:00"} onChange={event=>setScheduleAt(`${scheduleAt.slice(0,10)||new Date().toLocaleDateString('sv-SE')}T${event.detail.value}`)}><View className="mp-input">{scheduleAt.slice(11,16)||"选择时间"}</View></Picker></View>:<Input maxlength={32} disabled={loading} className="mp-input" value={scheduleAt} onInput={e => setScheduleAt(e.detail.value)} placeholder="2026-10-17T09:30" />}<CheckboxGroup onChange={e => setScheduleAutoSubmit(e.detail.value.includes("autoSubmit"))}><Checkbox value="autoSubmit" checked={scheduleAutoSubmit}>执行后自动提交网站（可产生外部提交）</Checkbox></CheckboxGroup><Button disabled={loading} loading={loading} className="mp-primary" onClick={createSchedule}>{editingSchedule ? "保存定时修改" : "创建定时任务"}</Button></View> : null}
      {settingsPanel === "schedules" ? <View className="mp-card mp-fetch-panel"><Text className="mp-row-title">定时任务</Text><Button disabled={loading} onClick={loadSchedules}>刷新定时任务</Button>{schedules.map((schedule, index) => <View className="mp-fetch-schedule" key={schedule.id || index}><View><Text className="mp-row-title">{displaySchedule(schedule)}</Text><Text className="mp-row-description">{compact?({preparing:"准备中",scheduled:"待执行",paused:"已暂停",cancelled:"已取消",completed:"已完成",failed:"失败"} as Record<string,string>)[schedule.status||""]||schedule.status||"状态待确认":schedule.enabled === false ? "已停用" : "已启用"} · {schedule.autoSubmit === true || schedule.auto_submit === true ? "自动提交" : "不自动提交"}</Text><Button disabled={loading || !schedule.id} onClick={() => { setEditingSchedule(schedule); const date = new Date(String(schedule.runAt || schedule.run_at || "")); setScheduleAt(Number.isFinite(date.getTime()) ? new Date(date.getTime() - date.getTimezoneOffset() * 60000).toISOString().slice(0, 16) : ""); setScheduleAutoSubmit(schedule.autoSubmit === true || schedule.auto_submit === true); setSettingsPanel("schedule"); }}>编辑</Button>{compact?<Button disabled={loading||!schedule.id||['cancelled','canceled','completed','done','failed','stopped'].includes(schedule.status||'')} onClick={async()=>{const confirmation=await sessionConfirmation(()=>Taro.showModal({title:"取消这个定时任务？",content:"将停止尚未执行的处理；服务端可能清理此定时创建且尚未使用的书目，已保存正文保留。"}));if(confirmation.confirm)await mutateSchedule(`/api/batch-rewrite/schedules/${encodeURIComponent(String(schedule.id))}/cancel`,"POST",{});}}>取消定时</Button>:<Button disabled={loading || !schedule.id} onClick={() => mutateSchedule(`/api/batch-rewrite/schedules/${encodeURIComponent(String(schedule.id))}`, "PATCH", { enabled: schedule.enabled === false })}>{schedule.enabled === false ? "启用" : "停用"}</Button>}<Button disabled={loading || !schedule.id || (compact&&!['cancelled','canceled','completed','done','failed','stopped','success','skipped_existing'].includes(schedule.status||""))} onClick={async () => { const result = await sessionConfirmation(() => Taro.showModal({ title: "删除定时任务", content: "删除后将不再按此时间执行，是否继续？" })); if (result.confirm) await mutateSchedule(`/api/batch-rewrite/schedules/${encodeURIComponent(String(schedule.id))}`, "DELETE"); }}>删除定时</Button></View></View>)}{!schedules.length ? <Text>当前没有服务端定时任务。</Text> : null}</View> : null}
      {["config", "prompts", "web", "logs"].includes(settingsPanel) ? <FetchMaintenancePanel key={settingsPanel} panel={settingsPanel} request={apiRequest} /> : null}
      </View></ScrollView></View></View> : null}
      {processResult ? <View className="mp-card mp-fetch-summary"><View className="mp-fetch-summary-head"><Text className="mp-row-title">本次处理</Text><Text className="mp-status-pill">{Number(processResult.fetch_failed ?? processResult.fetchFailed ?? 0) ? "存在抓取失败" : "处理已返回"}</Text></View><View className="mp-fetch-summary-grid">{[["解析", processResult.parsed ?? processResult.parsed_count ?? "—"], ["有效", processResult.unique_tasks ?? processResult.uniqueTasks ?? "—"], ["原文", processResult.fetched ?? "—"], ["AI 版本", processResult.generated_ai_files ?? processResult.generatedAiFiles ?? "—"]].map(([label, value]) => <View key={String(label)}><Text>{String(value)}</Text><Text>{String(label)}</Text></View>)}</View></View> : null}
      <View className="studio-dock"><View className="studio-dock-copy"><Text>{loading ? "正在处理，请稍候" : "书单准备好了吗"}</Text><Text className="studio-caption">{versionSummary} · 最多 {maxTxt} 字</Text></View><Button className="studio-primary" disabled={!input.trim() || !platformId || loading || refreshing || modelSaving || (compact&&(job.busy||job.running||Boolean(job.recovery)))} loading={loading} onClick={process}><AppIcon name="sparkles" size={18} />开始处理</Button></View>
    </View> : <View className="mp-fetch-content">
      <View className="mp-fetch-task-toolbar">
        <View className="booktest-task-heading"><Text>测书任务</Text><Text className="mp-row-description">按日期查看处理进度与已保存版本</Text></View>
        <View><Button className="mp-fetch-toolbar-button" loading={refreshing} onClick={refresh}><AppIcon name="refresh-ccw" size={15} />刷新</Button><Button className="mp-fetch-toolbar-button mp-fetch-retry" loading={loading} onClick={retryFailed}>重试失败</Button></View>
      </View>
      <View className="mp-fetch-task-filter">
        <Text>查看日期</Text>
        <Picker mode="date" value={taskDate} onChange={(event) => setTaskDate(event.detail.value)}><View className="mp-fetch-task-date"><Text>{taskDate || "全部日期"}</Text><AppIcon name="chevron-right" size={14} /></View></Picker>
        <Button className="mp-fetch-filter-button" onClick={() => setTaskDate(todayDateKey())}>今天</Button>
        <Button className="mp-fetch-filter-button" onClick={() => setTaskDate("")}>全部</Button>
      </View>
      <View className="mp-fetch-task-day-summary"><Text>{taskDate || "全部日期"}</Text><Text>{visibleTasks.length} 个任务</Text></View>
      {taskDate && tasks.length > visibleTasks.length ? <View className="mp-fetch-alert"><Text>已有 {tasks.length - visibleTasks.length} 个任务被日期筛选隐藏</Text><Button className="mp-fetch-alert-action" onClick={() => setTaskDate("")}>查看全部任务</Button></View> : null}
      {visibleTasks.map((task) => {
        const stages = fetchTaskStages(task);
        return <Button className="fetch-book-row" key={task.bookId} onClick={() => { setSelectedTaskId(task.bookId); setSubmitResult(null); void openTaskDetail(task); }} aria-label={`查看${taskTitle(task)}详情`}>
          <View className="mp-fetch-task-head">
            <View><Text className="mp-row-title">{taskTitle(task)}</Text><Text className="mp-row-description">书城：{task.platformName || "未标记"} · {task.updatedAt || task.createdAt || "服务端任务"}</Text></View>
            <Text className="mp-status-pill">{fetchTaskOverallStatus(task)}</Text>
          </View>
          <View className="mp-fetch-mini-progress">{stages.map((stage) => <View className={`mp-fetch-stage-dot is-${stage.state}`} key={stage.key} />)}</View>
        </Button>;
      })}
      {tasksRead==="loading"?<Text className="studio-caption">正在读取任务…</Text>:tasksRead==="error"?<Text className="studio-caption">任务读取失败，请刷新重试；已有记录暂时保留。</Text>:null}
      {tasksRead==="ready"&&!visibleTasks.length ? <View className="mp-card"><Text className="mp-row-title">{taskDate ? "该日期暂无任务" : "暂无测书任务"}</Text><Text className="mp-row-description">先在“处理”页提交一个书号或链接，或切换查看日期。</Text></View> : null}
    </View>}
    {selectedTask ? <View className="studio-overlay"><View className="studio-backdrop" onClick={closeTaskDetail} /><View className="studio-sheet"><View className="studio-sheet-handle" /><View className="studio-sheet-header"><Text>单本信息</Text><Button className="studio-link" onClick={closeTaskDetail}>关闭</Button></View><ScrollView scrollY className="studio-sheet-scroll"><View className="studio-sheet-body">
      <Text className="fetch-detail-title">{taskTitle(selectedTask)}</Text><Text className="studio-caption">书号 {selectedTask.bookId} · {selectedTask.platformName || "未标记平台"}</Text>
      {error ? <Text className="mp-error">{error}</Text> : null}
      <View className="mp-fetch-stage-graph">{selectedStages.map((stage, index) => <View className={`mp-fetch-stage-item is-${stage.state}`} key={stage.key}><View className="mp-fetch-stage-node" /><Text>{stage.label}</Text>{index < selectedStages.length - 1 ? <View className="mp-fetch-stage-line" /> : null}</View>)}</View>
      <View className="mp-fetch-task-states"><Text>原文：{selectedTask.originalStatus || "未抓取"}</Text><Text>AI：{selectedTask.aiStatus || (selectedTask.aiGeneratedCount ? `${selectedTask.aiGeneratedCount} 版` : "未生成")}</Text><Text>更新时间：{selectedTask.updatedAt || selectedTask.createdAt || "—"}</Text></View>
      <View className="mp-fetch-task-actions"><Button className="mp-secondary" loading={taskActionId === selectedTask.bookId} onClick={() => void openTaskDetail(selectedTask)}>查看</Button><Button className="mp-secondary" loading={taskActionId === selectedTask.bookId} onClick={() => void refetchOriginal(selectedTask)}>抓取原文</Button><Button className="mp-secondary" loading={taskActionId === selectedTask.bookId} onClick={() => void generateAi(selectedTask)}>生成 AI</Button><Button className="mp-secondary" loading={taskActionId === selectedTask.bookId} onClick={() => void retryTask(selectedTask)}>重试</Button><Button className="mp-fetch-delete" loading={taskActionId === selectedTask.bookId} onClick={() => void deleteTask(selectedTask)}>删除</Button></View>
      <Button className="mp-secondary fetch-detail-submit" disabled={loading} onClick={() => previewSubmit(selectedTask)}>预览网站提交</Button>
      {submitResult?.bookId === selectedTask.bookId ? <View className="mp-card"><Text className="mp-row-title">网站提交 · {submitResult.bookId}</Text><Text className="mp-row-description">提交组 {submitResult.payload.groups?.length ?? "未返回"} · 跳过 {submitResult.payload.skipped?.length ?? "未返回"}</Text>{typeof submitResult.payload.success_groups === "number" ? <Text className="mp-row-description">已确认 {submitResult.payload.success_groups} 组 · 目标站待确认 {submitResult.payload.accepted_groups} 组 · 失败 {submitResult.payload.failed_groups} 组</Text> : <Button disabled={loading || !submitResult.payload.groups?.length} className="mp-primary" onClick={confirmSubmit}>确认开始提交</Button>}{(submitResult.payload.skipped || []).map((item: any, index: number) => <Text key={index} className="mp-row-description">{item.version || ""} · {item.error || item.status || "已跳过"}</Text>)}<Button onClick={() => setSubmitResult(null)}>关闭提交信息</Button></View> : null}
      {taskDetail?.task.bookId === selectedTask.bookId ? <View className="mp-fetch-detail-panel"><Text className="mp-row-title">任务详情</Text>{taskDetail.loading ? <Text className="mp-row-description">正在读取服务端任务详情…</Text> : <>{fetchDetailBody(taskDetail.payload || {}).processed ? <><Text className="mp-field-label">处理后原文</Text><Text className="studio-caption">处理字数 / 完整原文字数：{fetchDetailBody(taskDetail.payload||{}).processedCount??"—"} / {fetchDetailBody(taskDetail.payload||{}).rawCount??"—"}</Text><Textarea maxlength={-1} className="mp-result mp-fetch-detail-text" value={fetchDetailBody(taskDetail.payload || {}).processed} disabled /></> : <Text className="mp-row-description">服务端未返回原文。</Text>}{fetchDetailBody(taskDetail.payload||{}).raw?<View><Text className="mp-field-label">完整原文（未截取）</Text><Textarea maxlength={-1} className="mp-result mp-fetch-detail-text" value={fetchDetailBody(taskDetail.payload||{}).raw} disabled /></View>:null}{taskDetailVersions(taskDetail.payload || {}).map(version => <View className="mp-fetch-detail-version" key={version.key}><Text className="mp-field-label">{version.label}</Text><Textarea maxlength={-1} className="mp-result mp-fetch-detail-text" value={version.content} disabled /></View>)}</>}</View> : null}
    </View></ScrollView></View></View> : null}
  </View>;
}

function FetchWorkspace(props:{title?:string;compact?:boolean;initialTaskId?:string}){return <FetchAccountScope>{(scope)=><FetchWorkspaceBody key={scope.key} {...props} request={scope.request} scopeKey={scope.key} isSelf={scope.self} accountControl={scope.control} />}</FetchAccountScope>;}
export default withThemedPage(FetchWorkspace);
