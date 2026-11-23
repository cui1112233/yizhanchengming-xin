package renderfont

import (
	"crypto/sha256"
	"fmt"
	"golang.org/x/image/font/sfnt"
	"strings"
	"testing"
)

func TestLicensedFontsMatchUpstream(t *testing.T) {
	for _, tc := range []struct {
		e           entry
		data        []byte
		sha, sample string
		missing     rune
	}{
		{noto(), notoBytes, "faa6c9df652116dde789d351359f3d7e5d2285a2b2a1f04a2d7244df706d5ea9", "你好龘㐀A", '🍎'},
		{ma(), maBytes, "6d2546bb189c732a8ca29af9e22457b152387d158aa459e4ac2ce1e51788b7fb", "你好书楷A", '龘'},
	} {
		info := tc.e.info
		if !info.Available || tc.e.err != nil {
			t.Fatal(tc.e.err)
		}
		family, err := tc.e.parsed.Name(nil, sfnt.NameIDFamily)
		if err != nil || family != info.Name {
			t.Fatalf("embedded family %q vs catalog %q", family, info.Name)
		}
		if info.Copyright == "" || info.License != "OFL-1.1" || !strings.Contains(info.LicenseText, "SIL OPEN FONT LICENSE Version 1.1") {
			t.Fatal(info.Name, " missing attribution or license")
		}
		if got := fmt.Sprintf("%x", sha256.Sum256(tc.data)); got != tc.sha {
			t.Fatal("upstream font digest mismatch", info.Name, got)
		}
		for _, r := range tc.sample {
			idx, err := tc.e.parsed.GlyphIndex(nil, r)
			if err != nil || idx == 0 {
				t.Fatalf("%s missing %q", family, r)
			}
		}
		if idx, err := tc.e.parsed.GlyphIndex(nil, tc.missing); err != nil || idx != 0 {
			t.Fatalf("fallback fixture did not exercise missing glyph %q", tc.missing)
		}
	}
}
