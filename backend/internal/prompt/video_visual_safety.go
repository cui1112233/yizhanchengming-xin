package prompt

import "strings"

// videoVisualSafetyReplacer is the backend-owned last-resort rewrite used only
// after a video provider rejects a prompt and the text rewrite channel is
// temporarily unavailable. Keep structural labels, names and timing intact;
// replace only visual actions/details that commonly trigger moderation.
var videoVisualSafetyReplacer = strings.NewReplacer(
	"掐住脖子", "逼近对峙",
	"边缘破损不齐", "边缘保持整齐",
	"证件整体被撕毁", "证件已被收起",
	"破损证件", "被收起的证件",
	"撕毁动作", "收起动作",
	"当众撕毁", "当众收起",
	"撕碎", "收起",
	"撕毁", "收起",
	"鲜血喷射", "紧张气氛",
	"血液喷射", "紧张气氛",
	"血流成河", "事后凌乱环境",
	"鲜血直流", "踉跄后退",
	"血腥", "紧张",
	"鲜血", "红色痕迹",
	"流血", "神情紧张",
	"开放性伤口", "紧张神情",
	"伤口", "紧张神情",
	"尸体", "无人空镜",
	"举刀", "握紧手中物件",
	"挥刀", "挥动手中物件",
	"刀具", "普通物件",
	"手枪", "普通物件",
	"枪支", "普通物件",
	"殴打", "激烈对峙",
	"群殴", "多人激烈对峙",
	"掐颈", "逼近对峙",
	"自残", "情绪低落",
	"昏迷", "安静坐下",
	"死亡特写", "凝重空镜",
	"裸露", "衣着完整",
	"性暗示", "日常交流",
	"亲吻", "保持距离交谈",
	"拥抱", "远处点头致意",
)

// VideoVisualSafetyFallback returns a deterministic, structurally stable copy
// of a rejected video prompt. It does not call a model and therefore remains
// available during a transient text-model outage.
func VideoVisualSafetyFallback(text string) string {
	return strings.TrimSpace(videoVisualSafetyReplacer.Replace(strings.TrimSpace(text)))
}
