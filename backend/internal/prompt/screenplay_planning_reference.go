package prompt

import "fmt"

// ScreenplayPlanningReference uses the selected complete storyboard template as a
// reference. These independent planning instructions are not a recovered native planner.
func ScreenplayPlanningReference(id, name, body, density, style string) string {
	return ScreenplayStoryboardPlanning + "\n" + ScreenplayStoryboardDensityRule(density) + fmt.Sprintf(`
本次只输出规划；下面完整分镜模板用于理解后续镜头粒度、时长及字段要求，不在本次输出最终分镜。
所选分镜提示词：%s（%s）
<storyboard_reference>
%s
</storyboard_reference>
画风参考：%s`, name, id, body, style)
}
