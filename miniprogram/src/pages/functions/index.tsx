import Taro, { useDidShow } from "@tarojs/taro";
import { Button, Image, Text, View } from "@tarojs/components";
import { useState } from "react";
import AppIcon from "../../components/AppIcon";
import { getModule, moduleRoute, type CatalogModuleKey } from "../../config/modules";
import { applyMobileTheme, readMobileTheme, ThemedPage } from "../../theme";
import { FUNCTION_VISUALS } from "./functionVisuals";
import heroArtwork from "../../assets/functions/functions-hero-reference-hd.jpg";
import "./gallery.scss";

const ACTIVE_MODULE_KEY = "yzcm.mobile.activeModule";
const CONTENT_KEYS: CatalogModuleKey[] = ["script", "novel", "booktest"];
const PRODUCTION_KEYS: CatalogModuleKey[] = ["shuihuo", "comic", "aicut"];
const TOOL_KEYS: CatalogModuleKey[] = ["agent", "voice", "plugins"];

export default function FunctionsPage() {
  const [activeKey, setActiveKey] = useState<CatalogModuleKey>(() => getModule(String(Taro.getStorageSync(ACTIVE_MODULE_KEY) || "script")).key);

  useDidShow(() => {
    applyMobileTheme(readMobileTheme());
    setActiveKey(getModule(String(Taro.getStorageSync(ACTIVE_MODULE_KEY) || "script")).key);
  });

  const selectModule = (key: CatalogModuleKey) => {
    Taro.setStorageSync(ACTIVE_MODULE_KEY, key);
    setActiveKey(key);
    void Taro.navigateTo({ url: moduleRoute(key) });
  };

  const cardClass = (key: CatalogModuleKey, className: string) =>
    `${className} tools-entry tools-entry-${key}${activeKey === key ? " is-recent" : ""}`;

  const compactCard = (key: CatalogModuleKey, slot = "") => {
    const visual = FUNCTION_VISUALS[key];
    return (
      <Button className={`${cardClass(key, "gallery-compact")} ${slot}`.trim()} key={key} hoverClass="tools-entry-pressed" onClick={() => selectModule(key)}>
        <Image className="gallery-compact-art" src={visual.asset} mode="aspectFit" />
        <View className="gallery-compact-copy">
          <Text className="gallery-compact-title">{visual.label}</Text>
          <Text className="gallery-compact-note">{visual.description}</Text>
        </View>
        <AppIcon name="chevron-right" size={17} />
      </Button>
    );
  };

  const featureCard = (key: CatalogModuleKey, slot = "") => {
    const visual = FUNCTION_VISUALS[key];
    return (
      <Button className={`${cardClass(key, "gallery-feature")} ${slot}`.trim()} hoverClass="tools-entry-pressed" onClick={() => selectModule(key)}>
        <View className="gallery-feature-copy"><Text className="gallery-feature-name">{visual.label}</Text><Text className="gallery-feature-description">{visual.description}</Text></View>
        <Image className="gallery-feature-art" src={visual.asset} mode="aspectFit" />
        <View className="gallery-feature-arrow"><AppIcon name="chevron-right" size={14} /></View>
      </Button>
    );
  };

  const compactStack = (first: CatalogModuleKey, second: CatalogModuleKey) => (
    <View className="gallery-compact-stack">
      {compactCard(first, "is-compact-top")}
      {compactCard(second, "is-compact-bottom")}
    </View>
  );

  const activeVisual = FUNCTION_VISUALS[activeKey];
  return (
    <ThemedPage className="mp-page mp-primary-page mp-functions-page functions-studio functions-glass function-gallery" dynamicGlass={false}>
      <View className="gallery-hero-shell">
        <Image className="gallery-hero-art" src={heroArtwork} mode="widthFix" />
        <Text className="gallery-visually-hidden">我的创作空间</Text>
        <Text className="gallery-visually-hidden">继续创作</Text>
        <Button className="gallery-resume" hoverClass="tools-entry-pressed" onClick={() => selectModule(activeKey)}>
          <Text className="gallery-visually-hidden">最近使用 {activeVisual.label}</Text>
        </Button>
      </View>

      <View className="gallery-body">
        <View className="gallery-section gallery-content-section">
          <View className="gallery-section-heading"><Text className="gallery-section-title">内容创作</Text><Text className="gallery-section-caption">从灵感到内容，打造你的专属作品</Text></View>
          <View className="gallery-feature-layout gallery-mosaic">
            {featureCard(CONTENT_KEYS[0], "is-feature")}
            {compactStack(CONTENT_KEYS[1], CONTENT_KEYS[2])}
          </View>
        </View>

        <View className="gallery-section gallery-production-section">
          <View className="gallery-section-heading"><Text className="gallery-section-title">生产成片</Text><Text className="gallery-section-caption">从文字到视频，多形态一键成片</Text></View>
          <View className="gallery-feature-layout gallery-mosaic is-reversed">
            {compactStack(PRODUCTION_KEYS[0], PRODUCTION_KEYS[2])}
            {featureCard(PRODUCTION_KEYS[1], "is-feature")}
          </View>
        </View>

        <View className="gallery-section gallery-assistant-section">
          <View className="gallery-section-heading"><Text className="gallery-section-title">AI 与工具</Text><Text className="gallery-section-caption">拓展创作边界，提升创作效率</Text></View>
          <View className="gallery-feature-layout gallery-mosaic">
            {featureCard(TOOL_KEYS[0], "is-feature")}
            {compactStack(TOOL_KEYS[1], TOOL_KEYS[2])}
          </View>
        </View>
      </View>
    </ThemedPage>
  );
}
