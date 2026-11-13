package prompt

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"
)

func TestCreativeOriginalBodiesAndPurposes(t *testing.T) {
	expected := map[string]string{CreativeScript144: "c102e5bfb4bca00bcf23b62d07a183c8dbf7cd66910e61fcf4d1fc113da7a250", CreativeCharacter213: "8de48b78e374a9b0e3211139dcbe55fc200099598d280d2955ece1ba7c4198ad", CreativeScene289: "bd843acd53b61e1603e35d7f740be6b128a818b0b673611725165c9e2a9feb11", CreativeStoryboard307: "b29a9eb293340ea33ba80092e25c756f423c41ece29f9f60763f123ee79eb382"}
	for id, want := range expected {
		p, err := ResolveCreativePrompt(id, creativePromptTypes[id])
		if err != nil {
			t.Fatal(err)
		}
		if got := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.ReplaceAll(p.Body, "\r\n", "\n")))); got != want {
			t.Fatalf("%s digest=%s", id, got)
		}
		if _, err := ResolveCreativePrompt(id, "wrong-purpose"); err == nil {
			t.Fatal("wrong purpose accepted")
		}
	}
}
