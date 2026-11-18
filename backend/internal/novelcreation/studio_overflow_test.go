package novelcreation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestStudioOversizedReplyPreservesBoundedEvidenceWithoutRebilling(t *testing.T) {
	s, st, p, ss := studioFixture(t)
	p.reply = `{"message":"` + strings.Repeat("长😀", 100000) + `","action":null}`
	in := StudioSendInput{IdempotencyKey: "overflow", Content: "想法", ModelID: "own"}
	snap, e := s.SendStudioMessage(context.Background(), "alice", ss.ID, in)
	if e != nil || snap.Turn.State != "invalid_output" || !strings.Contains(snap.Turn.ErrorMessage, "预览") {
		t.Fatal("overflow not explicit", e)
	}
	raw, e := st.LoadStudio(context.Background(), "alice", ss.ID, false)
	if e != nil {
		t.Fatal(e)
	}
	var r studioRecord
	if e = json.Unmarshal(raw, &r); e != nil {
		t.Fatal(e)
	}
	turn := r.Turns[0]
	sum := sha256.Sum256([]byte(p.reply))
	if turn.RawReply == "" || len(turn.RawReply) > 64*1024 || !utf8.ValidString(turn.RawReply) || !turn.RawReplyTruncated || turn.RawReplyBytes != len(p.reply) || turn.RawReplySHA256 != hex.EncodeToString(sum[:]) {
		t.Fatal("overflow audit evidence missing")
	}
	if _, e = s.SendStudioMessage(context.Background(), "alice", ss.ID, in); e != nil || p.calls != 1 {
		t.Fatal("overflow replay rebilled", e)
	}
}
