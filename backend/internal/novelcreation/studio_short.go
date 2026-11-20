package novelcreation

func studioShortDirection(p StudioShortProposal) string {
	for _, v := range []string{p.Direction, p.Instruction, p.Summary, p.Title} {
		if v = jsTrim(v); v != "" {
			return v
		}
	}
	return ""
}

func normalizeStudioShort(p StudioShortProposal) (StudioShortProposal, error) {
	p.Title = jsTrim(p.Title)
	p.Direction, p.Instruction, p.Summary = jsTrim(p.Direction), jsTrim(p.Instruction), jsTrim(p.Summary)
	p.ReaderPreferences = jsTrim(p.ReaderPreferences)
	if studioShortDirection(p) == "" {
		return p, inputError("短篇建议缺少创作方向，请补充后重新讨论")
	}
	if jsLength(p.Title) > 255 || jsLength(p.Direction) > 50000 || jsLength(p.Instruction) > 50000 || jsLength(p.Summary) > 50000 || jsLength(p.ReaderPreferences) > 12000 {
		return p, inputError("短篇建议内容超出上限")
	}
	if p.TargetWordCount == 0 {
		p.TargetWordCount = 13000
	}
	if p.TargetChapters == 0 {
		p.TargetChapters = 1
	}
	if p.TargetWordCount < 1000 || p.TargetWordCount > 500000 || p.TargetChapters < 1 || p.TargetChapters > 50 {
		return p, inputError("短篇建议字数或篇章数量不合法")
	}
	return p, nil
}
