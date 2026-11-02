package batchfactoryv11

// SelectComicCollectionVideoIDs returns the requested video IDs in the book's
// original page order. An omitted selection means the entire book.
func SelectComicCollectionVideoIDs(videos []Video, requested []string) ([]string, []string) {
	videoIDs := make([]string, 0, len(videos))
	for _, video := range videos {
		videoIDs = append(videoIDs, video.ID)
	}
	return SelectComicCollectionVideoIDsFromIDs(videoIDs, requested)
}

func SelectComicCollectionVideoIDsFromIDs(videoIDs, requested []string) ([]string, []string) {
	if len(requested) == 0 {
		return append([]string(nil), videoIDs...), nil
	}
	requestedSet := make(map[string]struct{}, len(requested))
	for _, id := range requested {
		if id != "" {
			requestedSet[id] = struct{}{}
		}
	}
	selected := make([]string, 0, len(requestedSet))
	unknown := make([]string, 0)
	known := make(map[string]struct{}, len(videoIDs))
	for _, videoID := range videoIDs {
		known[videoID] = struct{}{}
		if _, ok := requestedSet[videoID]; ok {
			selected = append(selected, videoID)
		}
	}
	for id := range requestedSet {
		if _, ok := known[id]; !ok {
			unknown = append(unknown, id)
		}
	}
	return selected, unknown
}

func sameComicCollectionItems(left, right []ComicCollectionItem) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].Ordinal != right[index].Ordinal ||
			left[index].VideoID != right[index].VideoID ||
			left[index].ComicPageID != right[index].ComicPageID ||
			left[index].ImageID != right[index].ImageID ||
			left[index].StorageRef != right[index].StorageRef ||
			left[index].MediaType != right[index].MediaType {
			return false
		}
	}
	return true
}

func latestDistinctComicCollections(collections []ComicCollection) []ComicCollection {
	result := make([]ComicCollection, 0, len(collections))
	for _, collection := range collections {
		matchingIndex := -1
		for index := range result {
			if sameComicCollectionItems(result[index].Items, collection.Items) {
				matchingIndex = index
				break
			}
		}
		if matchingIndex >= 0 {
			result[matchingIndex] = collection
			continue
		}
		result = append(result, collection)
	}
	return result
}
