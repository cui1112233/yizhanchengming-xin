package novelcreation

import promptcatalog "qiantie/backend/internal/prompt"

// Exact original literals, from fixed-target reviser_prompts constants.
// Pattern membership proven by module initializer 0x1801cb170, not guessed.
var NativeRevisionLocalPatterns = []string{
	"Paragraph uniformity|段落等长",       // constant 146
	"Hedge density|套话密度",              // constant 148
	"Formulaic transitions|公式化转折",     // constant 149
	"List-like structure|列表式结构",       // constant 150
	"Cross-chapter repetition|跨章重复",   // constant 151
	"AI-tell word density",            // constant 152
	"Fatigue word|高疲劳词",               // constant 153
	"Information Boundary Check|信息越界", // constant 154
	"Knowledge Base Pollution|知识库污染",  // constant 155
}
var NativeRevisionStructuralPatterns = []string{
	"OOC|人设|Character Fidelity|Character Matrix|Character.*Consistency", // constant 156
	"Mainline.*Drift|主线偏离|Outline Drift|大纲偏离|Chapter Memo Drift|章节备忘偏离", // constant 157
	"Conflict|冲突乏力|Payoff Dilution|爽点虚化",                                // constant 158
	"Timeline|时间线", // constant 159
	"Hook Check|伏笔检查|Hook.*Debt|伏笔.*债|未兑现",      // constant 160
	"Power Scaling|战力崩坏|金手指",                    // constant 161
	"Pacing|节奏",                                 // constant 162
	"POV Consistency|视角",                        // constant 163
	"Subplot Stagnation|支线停滞|Arc Flatline|弧线平坦", // constant 164
	"Relationship Dynamics|关系动态|情感表达",           // constant 165
	"Incentive Chain|利益链",                       // constant 166
	"Canon Event|正典|Mainline Canon",             // constant 167
}

const nativeRewriteOnlyInstruction = "\n\n分流指令：reviewer 报告的阻塞问题属于结构/语义错（人设崩、主线偏、爽点缺、时间线错、伏笔未收、memo 偏离等）。你必须输出 REVISED_CONTENT——禁止输出 PATCHES，这类问题不能靠补丁修复。如果无法安全重写，在 FIXED_ISSUES 里说明并留空 REVISED_CONTENT。"
const nativePatchOnlyInstruction = "\n\n分流指令：reviewer 报告的阻塞问题属于局部错（措辞、段落形状、疲劳词、信息越界、知识污染）。你必须只输出 PATCHES——不要整章改写。如果做不出补丁，留空 PATCHES。"

// RoutedNativeChapterRevision is a CM adapter for the recovered original branch.
// Neither instruction is sent for allow-full; the two exclusive branches never mix.
func RoutedNativeChapterRevision(mode string) promptcatalog.Definition {
	d := NativeChapterRevision
	d.Version = 3
	switch mode {
	case "rewrite-only":
		d.SystemPrompt += nativeRewriteOnlyInstruction
	case "patch-only":
		d.SystemPrompt += nativePatchOnlyInstruction
	}
	return d
}
