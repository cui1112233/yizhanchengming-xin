import Taro, { useDidShow } from "@tarojs/taro";
import { Button, Image, Text, View } from "@tarojs/components";
import { useEffect, useRef, useState } from "react";
import { readAccountSession, hasAccountSession } from "../../services/session";
import { createSessionGate, getSessionErrorMessage, resolveSessionGate } from "../../services/sessionGate";
import {onSessionInvalidated} from "../../services/sessionBoundary";
import { apiRequest } from "../../services/api";
import { loadTaskSummary, taskDestination, type AggregatedTask } from "../../services/taskSummary";
import BrandHeader from "../../components/BrandHeader";
import AppIcon from "../../components/AppIcon";
import { getModule, moduleRoute } from "../../config/modules";
import { FUNCTION_VISUALS } from "../functions/functionVisuals";
import { ThemedPage } from "../../theme";
import { isLocalAccountLogin } from "../../config/localAuth";
import { selectCurrentTask, workbenchProgress, workbenchStatus } from "./workbenchState";
import "./studio.scss";

const ACTIVE_MODULE_KEY = "yzcm.mobile.activeModule";

function projectTitle(task: AggregatedTask) {
  return task.title || "未命名项目";
}

export default function WorkspacePage() {
  const localLogin = isLocalAccountLogin();
  const initialGate = createSessionGate(false);
  const [status, setStatus] = useState(initialGate.status);
  const [error, setError] = useState(initialGate.error);
  const [activeModule, setActiveModule] = useState(() => getModule(String(Taro.getStorageSync(ACTIVE_MODULE_KEY) || "script")));
  const [records, setRecords] = useState<AggregatedTask[]>([]);
  const [summaryLoading, setSummaryLoading] = useState(false);
  const [summaryError, setSummaryError] = useState("");
  const [summaryWarning, setSummaryWarning] = useState("");
  const readVersion = useRef(0);
  useEffect(()=>{const off=onSessionInvalidated(()=>{readVersion.current++;setRecords([]);setSummaryLoading(false);setSummaryError('账号已切换，请重新连接');setStatus('binding-required');});return()=>{off();readVersion.current++;};},[]);

  const currentTask = selectCurrentTask(activeModule.key, records);
  const currentStatus = workbenchStatus(currentTask);
  const progress = workbenchProgress(currentTask);
  const visual = FUNCTION_VISUALS[activeModule.key];

  const loadSummary = () => {
    const version = readVersion.current;
    setSummaryLoading(true);
    setSummaryError("");
    setSummaryWarning("");
    void loadTaskSummary(apiRequest)
      .then((summary) => {
        if (version !== readVersion.current) return;
        const unavailable = summary.sources.filter((source) => source.state === "error");
        if (unavailable.length === summary.sources.length) throw new Error("最近记录暂不可读取");
        setRecords(summary.entries.filter((entry) => entry.id));
        setSummaryWarning(unavailable.length ? `${unavailable.map((source) => source.sourceLabel).join("、")}暂未同步` : "");
      })
      .catch((cause) => {
        if (version !== readVersion.current) return;
        setRecords([]);
        setSummaryError(cause?.message || "执行信息读取失败");
      })
      .finally(() => { if (version === readVersion.current) setSummaryLoading(false); });
  };

  const connectSession = () => {
    const version = ++readVersion.current;
    const selected = getModule(String(Taro.getStorageSync(ACTIVE_MODULE_KEY) || "script"));
    setActiveModule(selected);
    setError("");
    setStatus("checking");
    setRecords([]);
    if (!hasAccountSession()) {
      setStatus("binding-required");
      return;
    }
    void readAccountSession().then((session) => {
      if (version !== readVersion.current) return;
      const gate = resolveSessionGate(session);
      setStatus(gate.status);
      if (gate.status === "ready") loadSummary();
    }).catch((cause) => {
      if (version !== readVersion.current) return;
      const next = resolveSessionGate(cause);
      setError(getSessionErrorMessage(cause));
      setStatus(next.status);
    });
  };

  useDidShow(connectSession);

  const openCurrent = () => {
    if (currentTask) {
      void Taro.navigateTo({ url: taskDestination(currentTask) });
      return;
    }
    void Taro.navigateTo({ url: moduleRoute(activeModule.key) });
  };

  return <ThemedPage className="mp-page mp-primary-page mp-workspace-home studio-home" animateTheme>
    <BrandHeader title="工作台" primary />
    <View className="workbench-hero">
      <Text className="workbench-kicker">一战晟铭 · AI 创意套件</Text>
      <Text className="workbench-title">工作台</Text>
      <Text className="workbench-subtitle">创作执行中心 · 立即生成作品</Text>
    </View>

    {status === "checking" ? <View className="mp-inline-state"><Text>正在同步你的创作工作台…</Text></View> : null}
    {status === "error" ? <View className="mp-card mp-session-state"><Text className="mp-row-title">暂时无法验证当前账号</Text><Text className="mp-error">{error || "登录验证失败，请重试"}</Text><Button className="mp-primary" onClick={connectSession}>重新连接</Button><Button className="mp-secondary" onClick={() => void Taro.navigateTo({ url: "/pages/bind/index" })}>账号登录</Button></View> : null}
    {status === "binding-required" ? <View className="mp-card mp-session-state"><Text className="mp-row-title">{localLogin ? "请登录本地账号" : "登录已有创作账号"}</Text><Text className="mp-row-description">{localLogin ? "登录后读取本地服务的创作内容。" : "使用网页版账号密码，同步作品、配置和任务。"}</Text><Button className="mp-primary" onClick={() => void Taro.navigateTo({ url: "/pages/bind/index" })}>账号登录</Button></View> : null}

    {status === "ready" ? <View className="workbench-body">
      <View className="workbench-current-card liquid-glass--secondary" data-glass-role="secondary">
        <Image className="workbench-current-art" src={visual.asset} mode="aspectFit" />
        <View className="workbench-current-copy">
          <Text className="workbench-current-label">当前工作区</Text>
          <Text className="workbench-current-title">{visual.label}</Text>
          <View className={`workbench-status is-${currentStatus.tone}`}><Text className="workbench-status-dot" /><Text>{currentStatus.label}</Text></View>
        </View>
        <Button className="workbench-primary-action" hoverClass="workbench-pressed" onClick={openCurrent}><Text>{currentTask ? "继续执行" : "开始配置"}</Text><AppIcon name="chevron-right" size={17} /></Button>
      </View>

      {summaryLoading ? <View className="workbench-loading" ariaLabel="正在读取当前执行"><View /><View /><View /></View> : null}
      {!summaryLoading && summaryError ? <View className="workbench-error liquid-glass--secondary"><View><Text className="workbench-section-title">执行信息暂不可读取</Text><Text className="workbench-note">{summaryError}</Text></View><Button className="workbench-retry" onClick={loadSummary}>重新读取</Button></View> : null}

      {!summaryLoading && !summaryError && currentTask ? <>
        <View className="workbench-project-card liquid-glass--secondary" data-glass-role="secondary">
          <View className="workbench-project-copy">
            <Text className="workbench-project-title">《{projectTitle(currentTask)}》</Text>
            {progress ? <><Text className="workbench-progress-value">{progress.percent}%</Text><View className="workbench-progress-track"><View style={{ width: `${progress.percent}%` }} /></View><Text className="workbench-progress-note">已完成 {progress.completed} / {progress.total}</Text></> : <Text className="workbench-project-state">{currentStatus.label}</Text>}
          </View>
          <Image className="workbench-project-art" src={visual.asset} mode="aspectFit" />
        </View>

        <View className="workbench-section-heading"><Text className="workbench-section-mark" /><Text>本次执行</Text><View /></View>
        <View className="workbench-detail-card liquid-glass--secondary" data-glass-role="secondary">
          <View className="workbench-detail-row"><AppIcon name="sparkles" size={19} /><Text>当前功能</Text><Text>{visual.label}</Text></View>
          <View className="workbench-detail-row"><AppIcon name="folder-clock" size={19} /><Text>执行状态</Text><Text>{currentStatus.label}</Text></View>
          {progress ? <View className="workbench-detail-row"><AppIcon name="notebook-tabs" size={19} /><Text>完成数量</Text><Text>{progress.completed} / {progress.total}</Text></View> : null}
          {currentTask.createdAt ? <View className="workbench-detail-row"><AppIcon name="timer" size={19} /><Text>最近更新</Text><Text>{currentTask.createdAt}</Text></View> : null}
        </View>
      </> : null}

      {!summaryLoading && !summaryError && !currentTask ? <View className="workbench-empty liquid-glass--secondary" data-glass-role="secondary">
        <Image className="workbench-empty-art" src={visual.asset} mode="aspectFit" />
        <Text className="workbench-section-title">{activeModule.key==='novel'?'小说创作进度保存在对应作品内':'当前功能还没有可继续的任务'}</Text>
        <Text className="workbench-note">{activeModule.key==='novel'?'进入小说创作，打开作品查看讨论、写作阶段和已保存章节。':`点击“开始配置”进入 ${visual.label}，真正提交后任务会同步到这里。`}</Text>
      </View> : null}

      {summaryWarning ? <View className="workbench-warning"><Text>{summaryWarning}</Text><Button className="workbench-retry" onClick={loadSummary}>重试</Button></View> : null}
    </View> : null}
  </ThemedPage>;
}
