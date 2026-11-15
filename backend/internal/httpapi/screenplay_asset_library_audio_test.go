package httpapi

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"qiantie/backend/internal/screenplay"
	"testing"
)

func libraryReferenceWAV() []byte {
	b := make([]byte, 60)
	copy(b, "RIFF")
	binary.LittleEndian.PutUint32(b[4:], 52)
	copy(b[8:], "WAVEfmt ")
	binary.LittleEndian.PutUint32(b[16:], 16)
	binary.LittleEndian.PutUint16(b[20:], 1)
	binary.LittleEndian.PutUint16(b[22:], 1)
	binary.LittleEndian.PutUint32(b[24:], 8000)
	binary.LittleEndian.PutUint32(b[28:], 16000)
	binary.LittleEndian.PutUint16(b[32:], 2)
	binary.LittleEndian.PutUint16(b[34:], 16)
	copy(b[36:], "data")
	binary.LittleEndian.PutUint32(b[40:], 16)
	return b
}

func TestScreenplayAssetLibraryReferenceAudioCreateReplaceAndPreserve(t *testing.T) {
	s, _, objects, _, _ := assetLibraryServiceFixture(t)
	ctx := context.Background()
	in := libraryUploadFixture(t)
	in.AudioBytes = libraryReferenceWAV()
	in.AudioName = "参考声音.wav"
	entry, err := s.UploadEntry(ctx, "alice", in)
	if err != nil {
		t.Fatal(err)
	}
	public := screenplayLibraryPublic(entry)
	if public["audio"] == "" || public["audioFileName"] != in.AudioName {
		t.Fatal(public)
	}
	audio, err := libraryAudioObject(entry)
	if err != nil || audio.ObjectKey == "" {
		t.Fatal(audio, err)
	}
	reader, err := objects.Open(ctx, audio)
	if err != nil {
		t.Fatal(err)
	}
	reader.Close()
	raw, _ := json.Marshal(public)
	if bytes.Contains(raw, []byte(audio.ObjectKey)) {
		t.Fatal("private audio object leaked")
	}
	in.RequestKey = "audio-change-001"
	in.Bytes = nil
	in.AudioName = "更新音频.wav"
	in.AudioBytes = append([]byte(nil), in.AudioBytes...)
	in.AudioBytes[48] = 7
	changed, err := s.ReplaceReferenceAudio(ctx, "alice", entry.ID, entry.Revision, in)
	if err != nil {
		t.Fatal(err)
	}
	if changed.ID != entry.ID || changed.Revision != entry.Revision+1 || changed.ImageObject != entry.ImageObject || screenplayLibraryPublic(changed)["audioFileName"] != in.AudioName {
		t.Fatal(changed)
	}
	again, err := s.ReplaceReferenceAudio(ctx, "alice", entry.ID, entry.Revision, in)
	if err != nil || again.Revision != changed.Revision {
		t.Fatal(err, again)
	}
	invalid := in
	invalid.RequestKey = "audio-invalid-001"
	invalid.AudioBytes = []byte("fake audio")
	if _, err = s.ReplaceReferenceAudio(ctx, "alice", entry.ID, changed.Revision, invalid); !errors.Is(err, screenplay.ErrInvalidArgument) {
		t.Fatal(err)
	}
	in.RequestKey = "image-keep-audio"
	in.Bytes = libraryReplacementPNG(t)
	in.AudioBytes = nil
	in.AudioName = ""
	imageChanged, err := s.ReplaceImage(ctx, "alice", entry.ID, changed.Revision, in)
	if err != nil {
		t.Fatal(err)
	}
	if screenplayLibraryPublic(imageChanged)["audioFileName"] != "更新音频.wav" {
		t.Fatal("image-only update erased reference audio")
	}
	in.RequestKey = "image-and-audio"
	in.AudioName = "合并更新.wav"
	in.AudioBytes = libraryReferenceWAV()
	combined, err := s.ReplaceImage(ctx, "alice", entry.ID, imageChanged.Revision, in)
	if err != nil || combined.Revision != imageChanged.Revision+1 || screenplayLibraryPublic(combined)["audioFileName"] != in.AudioName {
		t.Fatal(err, combined)
	}
	current, _ := s.Repository.GetAssetLibraryEntry(ctx, "alice", entry.ID)
	if screenplayLibraryPublic(current)["audioFileName"] != in.AudioName {
		t.Fatal("reread lost audio")
	}
}
