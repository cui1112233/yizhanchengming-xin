package novelcreation

func LongImitationSummary(c LongImitationCreation) map[string]any {
	in := c.Input
	input := map[string]any{"idempotencyKey": in.IdempotencyKey, "modelId": in.ModelID, "thinkingLevel": in.ThinkingLevel, "title": in.Title, "storyIdea": in.StoryIdea, "selectedDirection": in.SelectedDirection, "targetWordCount": in.TargetWordCount, "targetChapters": in.TargetChapters, "readerPreferences": in.ReaderPreferences, "supplementaryInfo": in.SupplementaryInfo, "knowledgeBookIds": in.KnowledgeBookIDs, "memoReferences": in.MemoReferences}
	styles := []map[string]string{}
	for _, s := range c.Styles {
		styles = append(styles, map[string]string{"id": s.Profile.ID, "title": s.Profile.Title})
	}
	memos := []ImitationMemoChoice{}
	for _, n := range c.Memos {
		memos = append(memos, ImitationMemoChoice{ProjectID: n.ProjectID, ID: n.ID, Title: n.Title})
	}
	return map[string]any{"requirements": c.Requirements, "input": input, "statistics": c.Statistics, "styleDimensions": c.StyleDimensions, "analysis": c.Analysis, "hasReference": c.Input.ReferenceText != "", "referenceCharacters": jsNonspaceLength(c.Input.ReferenceText), "chunkCount": len(c.ReferenceChunks), "chunkingMethod": c.ChunkingMethod, "styles": styles, "memos": memos, "autoTitle": c.AutoTitle, "createdAt": c.CreatedAt}
}
