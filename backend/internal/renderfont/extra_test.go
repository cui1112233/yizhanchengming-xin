package renderfont

import "testing"

func TestAllCatalogFontsRenderChinese(t *testing.T) {
	for _, item := range Catalog() {
		if !item.Available || item.LicenseText == "" {
			t.Fatalf("unavailable font: %s: %v", item.ID, extraFonts()[item.ID].err)
		}
		face, err := Open(item.ID, 44)
		if err != nil {
			t.Fatalf("%s: %v", item.ID, err)
		}
		for _, r := range "春风吹过山川，今天读一本好书。" {
			if _, _, ok := face.GlyphBounds(r); !ok {
				t.Fatalf("%s: missing %c", item.ID, r)
			}
		}
		face.Close()
	}
}
