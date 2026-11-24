import Taro, { useLoad } from "@tarojs/taro";
import { Button, Text, View } from "@tarojs/components";
import { useState } from "react";
import AmbientStarfield from "../../components/AmbientStarfield";
import BrandHeader from "../../components/BrandHeader";
import { isModuleKey, moduleRoute, type ModuleKey } from "../../config/modules";
import ScriptWorkspace from "./ScriptWorkspace";
import FetchWorkspace from "./FetchWorkspace";
import AgentWorkspace from "./AgentWorkspace";
import PanelWorkspace from "./PanelWorkspace";
import ShuihuoWorkspace from "./ShuihuoWorkspace";
import { ThemedPage } from "../../theme";

type DetailModuleKey = Exclude<ModuleKey, "batch" | "toolkit" | "comic">;

function isDetailModuleKey(key: string): key is DetailModuleKey {
  return isModuleKey(key) && key !== "batch" && key !== "toolkit" && key !== "comic";
}

export default function ModuleDetailPage() {
  const [key, setKey] = useState<DetailModuleKey | "invalid" | "loading">("loading");
  const [historyId, setHistoryId] = useState("");
  const [taskId, setTaskId] = useState("");
  const [projectId, setProjectId] = useState("");

  useLoad((params) => {
    const requestedKey = String(params?.key || "");
    setHistoryId(String(params?.historyId || ""));
    setTaskId(String(params?.taskId || ""));
    setProjectId(String(params?.projectId || ""));
    if (requestedKey === "batch" || requestedKey === "comic" || requestedKey === "aicut") {
      void Taro.redirectTo({ url: moduleRoute(requestedKey) });
      return;
    }
    setKey(isDetailModuleKey(requestedKey) ? requestedKey : "invalid");
  });

  if (key === "script") return <ScriptWorkspace historyId={historyId} />;
  if (key === "fetch") return <FetchWorkspace />;
  if (key === "agent") return <AgentWorkspace taskId={taskId} projectId={projectId} />;
  if (key === "panel") return <PanelWorkspace />;
  if (key === "shuihuo") return <ShuihuoWorkspace projectId={projectId} />;
  if (key === "loading") return <ThemedPage className="mp-page"><Text>正在打开工作区…</Text></ThemedPage>;

  return <ThemedPage className="mp-page mp-detail-page"><AmbientStarfield /><BrandHeader eyebrow="创作能力" title="功能入口" showBack /><View className="mp-detail-content"><View className="mp-card"><Text className="mp-row-title">该功能入口已失效</Text><Text className="mp-row-description">请从“功能”页选择一个当前可用的工作区。</Text><Button className="mp-primary" onClick={() => { void Taro.switchTab({ url: "/pages/functions/index" }); }}>返回功能页</Button></View></View></ThemedPage>;
}
