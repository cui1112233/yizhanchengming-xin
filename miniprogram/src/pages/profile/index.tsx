import { hideStudioNavigation } from "../../ui/studioNavigation";
import { Button, Image, ScrollView, Switch, Text, View } from "@tarojs/components";
import Taro, { useDidHide, useDidShow } from "@tarojs/taro";
import { useEffect, useRef, useState } from "react";
import BrandHeader from "../../components/BrandHeader";
import AppIcon from "../../components/AppIcon";
import { profileFacts, type PublicMemberCenter } from "../../config/profileData";
import { apiUrl } from "../../config/runtime";
import { apiRequest } from "../../services/api";
import { readMemberProfile } from "../../services/memberProfile";
import { profileQuota, profileTaskMetrics } from "../../services/profileOverview";
import { uploadProfileAvatar } from "../../services/profileEditor";
import { logoutAccount, hasAccountSession } from "../../services/session";
import { sessionRevision } from "../../services/sessionBoundary";
import { loadTaskSummary, type TaskSummary } from "../../services/taskSummary";
import { ThemedPage, useMobileTheme, type MobileThemePreference } from "../../theme";
import { GLASS_CHANGE, GLASS_PREFERENCE } from "../../ui/liquidGlass";
import { publishGlassLayout, publishGlassScroll } from "../../ui/glassScrollBridge";
import { classifyProfileFailure, type ProfileFailureCategory, type ProfilePhase } from "./profileDiagnostics";
import assetsIcon from "../../assets/profile/profile-assets-v1.webp";
import preferenceIcon from "../../assets/profile/profile-preference-v1.webp";
import membershipIcon from "../../assets/profile/profile-membership-v1.webp";
import securityIcon from "../../assets/profile/profile-security-v1.webp";
import helpIcon from "../../assets/profile/profile-help-v1.webp";
import transparencyIcon from "../../assets/profile/profile-transparency-v1.webp";
import petalIcon from "../../assets/profile/profile-petal-v1.webp";
import profileIdentityBackground from "../../assets/profile/profile-identity-sunset-v2.webp";
import "./parity.scss";

const PROFILE_PETALS = ["one", "two", "three", "four", "five", "six", "seven"] as const;

const failureLabels: Record<ProfileFailureCategory, string> = {
  none: "", "session-missing": "未找到可用登录会话", "session-expired": "登录状态已失效",
  timeout: "读取超时，请检查网络后重试", domain: "当前网络域名配置不可用",
  network: "网络连接失败，请稍后重试", "invalid-data": "服务端未返回完整账号资料", unknown: "账号资料暂时无法读取",
};

function hasMember(center: PublicMemberCenter | null) {
  return Boolean(center?.member && typeof center.member === "object" && !Array.isArray(center.member));
}

export default function ProfilePage() {
  const { preference, setTheme } = useMobileTheme();
  const [appearanceOpen, setAppearanceOpen] = useState(false);
  useDidHide(() => setAppearanceOpen(false));
  useEffect(() => {
    if (appearanceOpen) hideStudioNavigation(true);
    return () => { if (appearanceOpen) hideStudioNavigation(false); };
  }, [appearanceOpen]);

  const [reducedGlass, setReducedGlass] = useState(() => Taro.getStorageSync(GLASS_PREFERENCE) === true);
  useDidShow(() => setReducedGlass(Taro.getStorageSync(GLASS_PREFERENCE) === true));
  const [center, setCenter] = useState<PublicMemberCenter | null>(null);
  const [taskSummary, setTaskSummary] = useState<TaskSummary>();
  const [diagnostic, setDiagnostic] = useState<{ phase: ProfilePhase; failure: ProfileFailureCategory; elapsedMs: number }>({ phase: "session-check", failure: "none", elapsedMs: 0 });
  const [loading, setLoading] = useState(true);
  const [uploading, setUploading] = useState(false);
  const nextAttempt = useRef(0);
  const activeAttempt = useRef<number | null>(null);
  const watchdog = useRef<ReturnType<typeof setTimeout>>();
  const mounted = useRef(true);
  const avatarLock = useRef(false);

  useEffect(() => () => {
    clearTimeout(watchdog.current);
    mounted.current = false;
    activeAttempt.current = null;
    nextAttempt.current += 1;
  }, []);

  const load = () => {
    if (!mounted.current || activeAttempt.current !== null) return;
    const attempt = ++nextAttempt.current;
    const revision = sessionRevision();
    const startedAt = Date.now();
    setDiagnostic({ phase: "session-check", failure: "none", elapsedMs: 0 });
    if (!hasAccountSession()) {
      setCenter(null);
      setLoading(false);
      setDiagnostic({ phase: "failed", failure: "session-missing", elapsedMs: Date.now() - startedAt });
      void Taro.navigateTo({ url: "/pages/bind/index" });
      return;
    }
    activeAttempt.current = attempt;
    setLoading(true);
    setDiagnostic({ phase: "request-started", failure: "none", elapsedMs: 0 });
    const current = () => mounted.current && activeAttempt.current === attempt && sessionRevision() === revision;
    const taskCurrent = () => mounted.current && nextAttempt.current === attempt && sessionRevision() === revision;
    void loadTaskSummary(apiRequest).then(result => { if (taskCurrent()) setTaskSummary(result); }).catch(() => { if (taskCurrent()) setTaskSummary(undefined); });
    clearTimeout(watchdog.current);
    watchdog.current = setTimeout(() => {
      if (!current()) return;
      activeAttempt.current = null;
      setCenter(null);
      setLoading(false);
      setDiagnostic({ phase: "failed", failure: "timeout", elapsedMs: Date.now() - startedAt });
    }, 15000);
    void readMemberProfile()
      .then(result => {
        if (!current()) return;
        setDiagnostic({ phase: "response-received", failure: "none", elapsedMs: Date.now() - startedAt });
        if (!hasMember(result)) {
          setCenter(null);
          setDiagnostic({ phase: "failed", failure: "invalid-data", elapsedMs: Date.now() - startedAt });
          return;
        }
        setCenter(result);
        setDiagnostic({ phase: "rendered", failure: "none", elapsedMs: Date.now() - startedAt });
      })
      .catch(cause => {
        if (!current()) return;
        setDiagnostic({ phase: "failed", failure: classifyProfileFailure(cause), elapsedMs: Date.now() - startedAt });
      })
      .finally(() => {
        clearTimeout(watchdog.current);
        if (!mounted.current || activeAttempt.current !== attempt) return;
        activeAttempt.current = null;
        setLoading(false);
        if (sessionRevision() !== revision) {
          nextAttempt.current += 1;
          setCenter(null);
          setDiagnostic({ phase: "failed", failure: "session-expired", elapsedMs: Date.now() - startedAt });
        }
      });
  };

  useDidShow(load);
  useEffect(() => { if (nextAttempt.current === 0) load(); }, []);

  const cancelAndLogin = () => {
    clearTimeout(watchdog.current);
    activeAttempt.current = null;
    nextAttempt.current += 1;
    setCenter(null);
    setLoading(false);
    void logoutAccount().catch(() => undefined);
    void Taro.reLaunch({ url: "/pages/bind/index" });
  };

  const facts = profileFacts(center);
  const memberLoaded = hasMember(center);
  const displayName = facts.displayName || facts.username;
  const taskMetrics = profileTaskMetrics(taskSummary);
  const quota = profileQuota(center);
  const openDetail = (key: string) => Taro.navigateTo({ url: `/pages/profile-detail/index?key=${encodeURIComponent(key)}` });

  const chooseAvatar = async () => {
    if (!memberLoaded || avatarLock.current) return;
    avatarLock.current = true;
    const session = sessionRevision();
    try {
      const picked = await Taro.chooseImage({ count: 1, sizeType: ["compressed"], sourceType: ["album", "camera"] });
      const path = picked.tempFilePaths?.[0];
      if (!path) return;
      setUploading(true);
      const file = await new Promise<{ data: string }>((resolve, reject) => Taro.getFileSystemManager().readFile({ filePath: path, encoding: "base64", success: resolve as any, fail: reject }));
      if (session !== sessionRevision()) return;
      const type = path.toLowerCase().endsWith(".png") ? "image/png" : path.toLowerCase().endsWith(".webp") ? "image/webp" : "image/jpeg";
      const result = await uploadProfileAvatar(`data:${type};base64,${file.data}`, facts.username || "");
      if (session !== sessionRevision()) return;
      setCenter(result);
      Taro.showToast({ title: "头像已更新并核对", icon: "success" });
    } catch (cause: any) {
      if (!cause?.errMsg?.includes("cancel")) Taro.showToast({ title: cause?.message || "头像上传失败", icon: "none" });
    } finally {
      setUploading(false);
      avatarLock.current = false;
    }
  };

  const setReducedTransparency = (enabled: boolean) => {
    Taro.setStorageSync(GLASS_PREFERENCE, enabled);
    setReducedGlass(enabled);
    Taro.eventCenter?.trigger(GLASS_CHANGE);
  };

  useEffect(() => { Taro.nextTick?.(() => publishGlassLayout("profile")); }, [loading, memberLoaded, center, taskSummary]);

  const menuRows = [
    { key: "assets", title: "我的资产", description: "管理你的作品、素材与草稿", icon: assetsIcon, action: () => Taro.switchTab({ url: "/pages/workspace/index" }) },
    { key: "preference", title: "创作偏好", description: "设置你的风格、题材与创作习惯", icon: preferenceIcon, action: () => setAppearanceOpen(true) },
    { key: "membership", title: "会员与用量", description: "查看会员状态与创作额度", icon: membershipIcon, action: () => openDetail("memberCenter") },
    { key: "security", title: "账号与安全", description: "管理账号信息、登录与安全设置", icon: securityIcon, action: () => openDetail("security") },
    { key: "help", title: "帮助与反馈", description: "使用指南、常见问题与意见反馈", icon: helpIcon, action: () => Taro.showToast({ title: "帮助与反馈暂未开放", icon: "none" }) },
  ];

  return (
    <ThemedPage className="mp-page mp-profile-page profile-space-page">
      <BrandHeader eyebrow="" title="我的空间" primary />
      <ScrollView className="mp-profile-scroll" scrollY scrollWithAnimation={false} onScroll={event => publishGlassScroll("profile", event.detail.scrollTop)}>
        <View className="profile-space-heading"><Text className="profile-space-title">我的空间</Text><Text className="profile-space-subtitle">账号、资产与创作偏好</Text></View>
        {loading ? <View className="mp-card mp-profile-state"><Text className="mp-row-title">正在读取账号资料</Text><Text className="mp-row-description">正在连接服务端，最多等待 15 秒；失败后可重试。</Text><Button className="mp-secondary mp-profile-cancel" onClick={cancelAndLogin}>取消并返回登录</Button></View> : null}
        {!loading && !memberLoaded ? <View className="mp-card mp-profile-state"><Text className="mp-row-title">无法读取当前账号资料</Text><Text className="mp-error">{failureLabels[diagnostic.failure]}</Text><View className="mp-profile-actions"><Button className="mp-primary mp-profile-retry" onClick={load}>重新读取</Button><Button className="mp-secondary mp-profile-cancel" onClick={cancelAndLogin}>返回登录</Button></View></View> : null}
        {memberLoaded ? <>
          <View className={`profile-space-identity mp-profile-hero liquid-glass--primary${reducedGlass ? " is-reduced-motion" : ""}`} data-glass-role="primary">
            <Image className="profile-space-identity-background" src={profileIdentityBackground} mode="aspectFill" />
            <View className="profile-space-particles" ariaHidden>{PROFILE_PETALS.map(name => <Image key={name} className={`profile-space-petal is-${name}`} src={petalIcon} mode="aspectFit" />)}</View>
            <View className="profile-space-avatar" onClick={chooseAvatar}>{facts.avatarUrl ? <Image className="profile-space-avatar-image" src={apiUrl(facts.avatarUrl)} mode="aspectFill" /> : <Text>{displayName ? displayName.slice(0, 1).toUpperCase() : "?"}</Text>}<View className="profile-space-camera">{uploading ? <Text>上传中</Text> : <AppIcon name="camera" size={13} />}</View></View>
            <View className="profile-space-identity-copy">
              <View className="profile-space-name-line"><Text className="profile-space-name">{displayName || "未设置名称"}</Text><Text className="profile-space-badge">{facts.role || "创作者"}</Text></View>
              <Text className="profile-space-welcome">欢迎回来，继续把灵感变成作品吧！</Text>
              <Text className="profile-space-team">{facts.teamName || (facts.username ? `@${facts.username}` : "账号信息未完整返回")}</Text>
              <View className="profile-space-metrics"><View><AppIcon name="clapperboard" size={17} /><Text className="profile-space-metric-value">{taskMetrics.running ?? "—"}</Text><Text>进行中</Text></View><View className="profile-space-metric-divider" /><View><AppIcon name="shield-check" size={17} /><Text className="profile-space-metric-value">{taskMetrics.completed ?? "—"}</Text><Text>已完成</Text></View></View>
            </View>
          </View>
          <Button className="profile-space-quota liquid-glass--secondary" data-glass-role="secondary" onClick={() => openDetail("usage")}><Image className="profile-space-quota-icon" src={membershipIcon} mode="aspectFit" /><View className="profile-space-quota-copy"><Text className="profile-space-quota-title">创作额度</Text><Text className="profile-space-quota-value">{quota.value}</Text><Text className="profile-space-quota-caption">{quota.caption}</Text></View><View className="profile-space-quota-link"><Text>查看详情</Text><AppIcon name="chevron-right" size={16} /></View></Button>
          <View className="profile-space-list liquid-glass--list">{menuRows.map(row => <Button className={`profile-space-row profile-space-${row.key}${row.key === "preference" ? " studio-appearance-entry" : ""}`} key={row.key} onClick={row.action}><Image className="profile-space-row-icon" src={row.icon} mode="aspectFit" /><View className="profile-space-row-copy"><Text className="profile-space-row-title">{row.title}</Text><Text className="profile-space-row-description">{row.description}</Text></View><AppIcon name="chevron-right" size={18} className="profile-space-row-arrow" /></Button>)}</View>
          <View className="profile-space-transparency liquid-glass--list"><Image className="profile-space-transparency-icon" src={transparencyIcon} mode="aspectFit" /><View className="profile-space-transparency-copy"><Text className="profile-space-row-title">降低透明度</Text><Text className="profile-space-row-description">减少界面透明效果，获得更清晰的显示</Text></View><Switch ariaLabel="降低透明度" checked={reducedGlass} color="#50eadc" onChange={event => setReducedTransparency(event.detail.value)} /></View>
          {diagnostic.failure !== "none" ? <View className="mp-profile-inline-error"><Text className="mp-error">{failureLabels[diagnostic.failure]}</Text></View> : null}
        </> : null}
        {hasAccountSession() ? <Button className="mp-profile-logout" onClick={cancelAndLogin}>退出登录</Button> : null}
        <View className="mp-profile-bottom-space" />
      </ScrollView>
      {appearanceOpen ? <View className="studio-profile-mask" onClick={() => setAppearanceOpen(false)}><View className="studio-profile-sheet" role="dialog" ariaLabel="创作偏好" onClick={event => event.stopPropagation()}><View className="studio-profile-sheet-head"><Text>创作偏好</Text><Button onClick={() => setAppearanceOpen(false)}>完成</Button></View><View className="studio-theme-preferences"><Text className="mp-row-title">界面主题</Text><View className="studio-theme-options">{([["light", "浅色"], ["dark", "深色"], ["system", "跟随系统"]] as Array<[MobileThemePreference, string]>).map(([value, label]) => <Button key={value} className={preference === value ? "studio-theme-option is-active" : "studio-theme-option"} ariaLabel={`${label}${preference === value ? "，已选中" : ""}`} onClick={() => setTheme(value)}>{label}</Button>)}</View></View></View></View> : null}
    </ThemedPage>
  );
}
