package batchfactoryv11

import "testing"

func TestSelectComicCollectionVideoIDsKeepsBookOrderAndRejectsUnknown(t *testing.T) {
	videos := []Video{{ID: "page-1"}, {ID: "page-2"}, {ID: "page-3"}, {ID: "page-4"}}
	selected, unknown := SelectComicCollectionVideoIDs(videos, []string{"page-4", "page-2"})
	if len(unknown) != 0 || len(selected) != 2 || selected[0] != "page-2" || selected[1] != "page-4" {
		t.Fatalf("selected=%v unknown=%v", selected, unknown)
	}
	selected, unknown = SelectComicCollectionVideoIDs(videos, []string{"page-2", "missing"})
	if len(selected) != 1 || selected[0] != "page-2" || len(unknown) != 1 || unknown[0] != "missing" {
		t.Fatalf("selected=%v unknown=%v", selected, unknown)
	}
}
