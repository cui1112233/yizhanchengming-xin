import { Button, Input, Picker, Switch, Text, Textarea, View } from "@tarojs/components";
import { useEffect, useRef, useState } from "react";
import { apiRequest as rawApiRequest } from "../../services/api";
import { cleanFetchWebSettings, writeFetchMaintenance } from "../../services/fetchMaintenance";

const root = "/api/batch-rewrite";
export default function FetchMaintenancePanel({ panel, request:apiRequest=rawApiRequest }: { panel: string;request?:typeof rawApiRequest }) {
  const [config, setConfig] = useState<any>(null);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [notice, setNotice] = useState("");
  const [password, setPassword] = useState("");
  const [sample, setSample] = useState("");
  const [preview, setPreview] = useState("");
  const [diagnostics, setDiagnostics] = useState<any>(null);
  const [perVersion, setPerVersion] = useState(false);
  const epoch = useRef(0), lock = useRef(false), uncertain = useRef(false);
  const accept = (result: any) => {
    if (!result || typeof result !== "object" || !result.app_config) throw new Error("配置响应格式无效");
    const web = cleanFetchWebSettings(result.web_submit || result.app_config.web_submit || {});
    setPerVersion(Object.values(web.profile_bindings || {}).some(Boolean));
    setConfig({ ...result, app_config: { ...result.app_config, ...(result.app_config.web_submit ? { web_submit: cleanFetchWebSettings(result.app_config.web_submit) } : {}) }, web_submit: web });
  };
  const load = async () => {
    const version = ++epoch.current;
    setError("");
    try {
      const result = await apiRequest<any>(`${root}/config`);
      if (version !== epoch.current) return;
      accept(result); uncertain.current = false;
      if (panel === "web") {
        const environment = await apiRequest(`${root}/web-submit/environment`);
        if (version === epoch.current) setDiagnostics(environment);
      }
      if (panel === "logs") {
        const [logs, issues] = await Promise.all([apiRequest(`${root}/logs`), apiRequest(`${root}/issues`)]);
        if (version === epoch.current) setDiagnostics({ logs, issues });
      }
    } catch (cause: any) { if (version === epoch.current) setError(cause.message || "读取失败"); }
  };
  useEffect(() => { void load(); return () => { epoch.current++; }; }, [panel]);
  const patch = (section: string, key: string, value: any) => {
    epoch.current++;
    setConfig((current: any) => ({ ...current, app_config: { ...current.app_config, [section]: { ...current.app_config[section], [key]: value } } }));
  };
  const patchWeb = (key: string, value: any) => { epoch.current++; setConfig((current: any) => ({ ...current, web_submit: { ...current.web_submit, [key]: value } })); };
  const action = async (path: string, data: any, readPath = `${root}/config`) => {
    if (lock.current) return;
    if (uncertain.current) { setError("上次操作结果未知，请刷新核对后再操作"); return; }
    lock.current = true; setBusy(true); setError(""); setNotice(""); const version = ++epoch.current;
    try {
      const result = await writeFetchMaintenance(apiRequest, path, "POST", data, readPath);
      if (version === epoch.current) {
        accept(result); setPassword(""); setNotice("操作已返回，已读取服务端最新配置");
        if (panel === "web") {
          const environment = await apiRequest(`${root}/web-submit/environment`);
          if (version === epoch.current) setDiagnostics(environment);
        }
      }
    } catch (cause: any) { if (String(cause.message).includes("结果未知")) uncertain.current = true; if (version === epoch.current) setError(cause.message || "操作失败"); }
    finally { lock.current = false; setBusy(false); }
  };
  const app = config?.app_config || {}, web = config?.web_submit || {};
  const textField = (label: string, value: string, update: (value: string) => void) => <View key={label}><Text className="mp-field-label">{label}</Text><Textarea maxlength={-1} disabled={busy} className="mp-textarea" value={value || ""} onInput={e => update(e.detail.value)} /></View>;
  const toggle = (label: string, checked: boolean, update: (value: boolean) => void) => <View className="mp-fetch-toggle-row" key={label}><Text>{label}</Text><Switch disabled={busy} checked={checked} onChange={e => update(e.detail.value)} /></View>;
  const profiles = [{ id: "", name: "跟随默认配置档" }, ...(Array.isArray(web.upload_profiles) ? web.upload_profiles.filter((p: any) => p.enabled !== false) : [])];
  const profilePicker = (label: string, value: string, update: (v: string) => void) => <View key={label}><Text>{label}</Text><Picker disabled={busy} range={profiles.map((p: any) => p.name || p.id)} value={Math.max(0, profiles.findIndex((p: any) => String(p.id || p.name || "") === value))} onChange={e => { const p = profiles[Number(e.detail.value)]; update(String(p?.id || (Number(e.detail.value) ? p?.name : "") || "")); }}><View className="mp-input">{profiles.find((p: any) => String(p.id || p.name || "") === value)?.name || value || "跟随默认配置档"}</View></Picker></View>;
  return <View className="mp-card mp-fetch-panel">
    <Text className="mp-row-title">{({ config: "流程与抓取配置", prompts: "知识库 / 提示词 / 规则", web: "网站提交配置", logs: "日志与问题记录" } as any)[panel]}</Text>
    {error ? <Text className="mp-error">{error}</Text> : null}{notice ? <Text className="mp-row-description">{notice}</Text> : null}
    <Button disabled={busy} onClick={load}>刷新核对配置</Button>
    {!config ? <Text>配置尚未读取成功</Text> : <>
      {panel === "config" ? <>
        {[["auto_classify_missing", "自动补齐风格 / 男女频"], ["auto_fetch_original", "自动抓取原文"], ["auto_rewrite_after_fetch", "自动生成 AI 版本"]].map(([key, label]) => toggle(label, app.workflow?.[key] !== false, v => patch("workflow", key, v)))}
        {[["default_max_txt", "默认截取字数", 4000, 100, 100000], ["concurrency", "抓取并发", 1, 1, 20], ["retries", "失败重试", 1, 0, 10]].map(([key, label, fallback, min, max]) => <View key={String(key)}><Text>{label}</Text><Input disabled={busy} maxlength={6} type="number" className="mp-input" value={String(app.fetch?.[String(key)] ?? fallback)} onInput={e => patch("fetch", String(key), Math.min(Number(max), Math.max(Number(min), Number(e.detail.value))))} /></View>)}
        <Button disabled={busy} loading={busy} onClick={() => action(`${root}/config`, { app_config: app })}>保存配置并读回</Button>
      </> : null}
      {panel === "prompts" ? <>
        <Text className="mp-row-description">知识库统计：{JSON.stringify(config.knowledge_summary || {})}</Text>
        {textField("改文提示词", app.rewrite?.prompt, v => patch("rewrite", "prompt", v))}
        {textField("处理规则提示词", app.rewrite?.processing_rule_prompt, v => patch("rewrite", "processing_rule_prompt", v))}
        {textField("知识库使用提示词", config.knowledge?.usage_prompt, v => { epoch.current++; setConfig((c: any) => ({ ...c, knowledge: { ...c.knowledge, usage_prompt: v } })); })}
        {toggle("敏感词 AI 修复", app.sensitive_ai?.enabled === true, v => patch("sensitive_ai", "enabled", v))}
        <Button disabled={busy} loading={busy} onClick={() => action(`${root}/config`, { app_config: app, knowledge: config.knowledge || {} })}>保存提示词与规则并读回</Button>
        {textField("规则测试文本（按已保存的服务端规则）", sample, v => { epoch.current++; setSample(v); setPreview(""); })}
        <Button disabled={busy || !sample.trim()} onClick={async () => { if (lock.current || !sample.trim()) return; lock.current = true; setBusy(true); setError(""); const version = ++epoch.current; try { const result = await apiRequest<any>(`${root}/rules/preview`, { method: "POST", data: { text: sample, scope: "original" } }); if (version === epoch.current) setPreview(String(result.processed ?? "")); } catch (e: any) { if (version === epoch.current) setError(e.message || "预览失败"); } finally { lock.current = false; setBusy(false); } }}>预览服务端规则</Button>
        {preview ? <Textarea maxlength={-1} disabled className="mp-result" value={preview} /> : null}
      </> : null}
      {panel === "web" ? <>
        <Text className="mp-field-label">用户名</Text><Input disabled={busy} maxlength={256} className="mp-input" value={web.username || ""} onInput={e => patchWeb("username", e.detail.value)} />
        <Text className="mp-field-label">替换密码（留空保留已保存密码）</Text><Input disabled={busy} maxlength={1024} password value={password} onInput={e => setPassword(e.detail.value)} />
        {profilePicker("默认配置档", web.selected_profile || "", v => patchWeb("selected_profile", v))}
        {toggle("按版本单独配置", perVersion, enabled => { setPerVersion(enabled); if (!enabled) patchWeb("profile_bindings", {}); })}
        {perVersion ? ["original", "ai1", "ai2", "ai3", "ai4", "ai5"].map(version => profilePicker(`${version} 配置档`, web.profile_bindings?.[version] || "", v => patchWeb("profile_bindings", { ...web.profile_bindings, [version]: v }))) : null}
        <Text>配置档 {profiles.length - 1} 个 · 风格 {Array.isArray(web.style_catalog) ? web.style_catalog.length : 0} 个</Text>
        <Button disabled={busy} loading={busy} onClick={() => action(`${root}/web-submit/config`, { settings: { ...cleanFetchWebSettings(web), ...(password ? { password } : {}) } })}>保存提交配置并读回</Button>
        <Button disabled={busy} onClick={() => action(`${root}/web-submit/sync-configs`, {})}>同步配置档并读回</Button><Button disabled={busy} onClick={() => action(`${root}/web-submit/sync-styles`, {})}>同步风格并读回</Button>
        {diagnostics ? <Text className="mp-row-description">{diagnostics.ok === true ? "环境检查通过" : "环境检查未通过"} · {(diagnostics.checks || []).map((c: any) => `${c.name || ""} ${c.detail || ""}`).join("；")}</Text> : null}
      </> : null}
      {panel === "logs" ? <>{diagnostics ? <><Text className="mp-field-label">运行日志</Text><Textarea maxlength={-1} disabled className="mp-result" value={JSON.stringify(diagnostics.logs, null, 2)} /><Text className="mp-field-label">问题记录</Text><Textarea maxlength={-1} disabled className="mp-result" value={JSON.stringify(diagnostics.issues, null, 2)} /></> : <Text>日志与问题记录尚未读取成功</Text>}</> : null}
    </>}
  </View>;
}
