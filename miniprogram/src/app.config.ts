import { uiPalette } from "./ui/tokens";
const palette = uiPalette("dark");

export default {
  darkmode: true,
  entryPagePath: "pages/functions/index",
  usingComponents: {},
  pages: [
    "pages/bind/index",
    "pages/workspace/index",
    "pages/tasks/index",
    "pages/functions/index",
    "pages/profile/index",
    "pages/profile-detail/index",
    "pages/module-detail/index",
    "pages/batch/index",
    "pages/glass-lab/index",
  ],
  // 主包已接近 2048 KB 上限，新增功能页统一进分包。
  subPackages: [
    {
      root: "comic",
      pages: [
        "pages/index/index",
        "pages/aicut/index",
        "pages/novel/index",
        "pages/member-center/index",
        "pages/usage-stats/index",
        "pages/tts/index",
        "pages/comic-new/index",
        "pages/comic-config/index",
        "pages/comic-task/index",
        "pages/comic-result/index",
      ],
    },
  ],
  window: {
    navigationStyle: "custom",
    navigationBarTitleText: "一战晟铭",
    navigationBarBackgroundColor: palette.canvas,
    navigationBarTextStyle: "white",
    backgroundColor: palette.canvas,
    backgroundTextStyle: "dark",
  },
  tabBar: {
    custom: true,
    color: palette.textMuted,
    selectedColor: palette.accent,
    backgroundColor: palette.surface,
    borderStyle: "black",
    list: [
      { pagePath: "pages/functions/index", text: "功能", iconPath: "assets/icons/tab-functions-normal-dark.png", selectedIconPath: "assets/icons/tab-functions-selected-dark.png" },
      { pagePath: "pages/workspace/index", text: "工作台", iconPath: "assets/icons/tab-workspace-normal-dark.png", selectedIconPath: "assets/icons/tab-workspace-selected-dark.png" },
      { pagePath: "pages/tasks/index", text: "任务", iconPath: "assets/icons/tab-tasks-normal-dark.png", selectedIconPath: "assets/icons/tab-tasks-selected-dark.png" },
      { pagePath: "pages/profile/index", text: "我的", iconPath: "assets/icons/tab-profile-normal-dark.png", selectedIconPath: "assets/icons/tab-profile-selected-dark.png" },
    ],
  },
};
