package batchfactoryv11

import (
	"context"
	"encoding/binary"
	"testing"
)

func buildTestWAV(sampleRate, channels, bitsPerSample, milliseconds int) []byte {
	byteRate := sampleRate * channels * bitsPerSample / 8
	dataSize := byteRate * milliseconds / 1000
	buf := make([]byte, 44+dataSize)
	copy(buf[0:4], "RIFF")
	binary.LittleEndian.PutUint32(buf[4:8], uint32(36+dataSize))
	copy(buf[8:12], "WAVE")
	copy(buf[12:16], "fmt ")
	binary.LittleEndian.PutUint32(buf[16:20], 16)
	binary.LittleEndian.PutUint16(buf[20:22], 1)
	binary.LittleEndian.PutUint16(buf[22:24], uint16(channels))
	binary.LittleEndian.PutUint32(buf[24:28], uint32(sampleRate))
	binary.LittleEndian.PutUint32(buf[28:32], uint32(byteRate))
	binary.LittleEndian.PutUint16(buf[32:34], uint16(channels*bitsPerSample/8))
	binary.LittleEndian.PutUint16(buf[34:36], uint16(bitsPerSample))
	copy(buf[36:40], "data")
	binary.LittleEndian.PutUint32(buf[40:44], uint32(dataSize))
	return buf
}

func buildTestMP3Frames(frames int) []byte {
	// MPEG1 Layer III / 128kbps / 44100Hz / mono -> 帧长 417 字节，1152 采样/帧。
	const frameLen = 417
	buf := make([]byte, frames*frameLen)
	for i := 0; i < frames; i++ {
		offset := i * frameLen
		buf[offset] = 0xFF
		buf[offset+1] = 0xFB
		buf[offset+2] = 0x90
		buf[offset+3] = 0xC0
	}
	return buf
}

func buildTestMP4(timescale, duration uint32) []byte {
	ftyp := make([]byte, 8)
	binary.BigEndian.PutUint32(ftyp[0:4], 8)
	copy(ftyp[4:8], "ftyp")

	mvhd := make([]byte, 28)
	binary.BigEndian.PutUint32(mvhd[0:4], 28)
	copy(mvhd[4:8], "mvhd")
	binary.BigEndian.PutUint32(mvhd[8+12:8+16], timescale)
	binary.BigEndian.PutUint32(mvhd[8+16:8+20], duration)

	moov := make([]byte, 8+len(mvhd))
	binary.BigEndian.PutUint32(moov[0:4], uint32(len(moov)))
	copy(moov[4:8], "moov")
	copy(moov[8:], mvhd)
	return append(ftyp, moov...)
}

func TestParseBuiltinAudioDurationWAV(t *testing.T) {
	data := buildTestWAV(8000, 1, 16, 1500)
	got, ok := parseBuiltinAudioDuration(data)
	if !ok || got.Source != "builtin:wav" || got.Estimated || got.MS != 1500 {
		t.Fatalf("wav parse = %#v ok=%v, want 1500ms builtin:wav", got, ok)
	}
}

func TestParseBuiltinAudioDurationMP3(t *testing.T) {
	// 100 帧 × 1152 采样 / 44100Hz = 2.612244...秒 -> 2612ms
	data := buildTestMP3Frames(100)
	got, ok := parseBuiltinAudioDuration(data)
	if !ok || got.Source != "builtin:mp3" || got.Estimated {
		t.Fatalf("mp3 parse = %#v ok=%v, want builtin:mp3", got, ok)
	}
	if got.MS != 2612 {
		t.Fatalf("mp3 duration = %dms, want 2612ms", got.MS)
	}
}

func TestParseBuiltinAudioDurationM4A(t *testing.T) {
	data := buildTestMP4(1000, 2500)
	got, ok := parseBuiltinAudioDuration(data)
	if !ok || got.Source != "builtin:mp4" || got.MS != 2500 {
		t.Fatalf("mp4 parse = %#v ok=%v, want 2500ms builtin:mp4", got, ok)
	}
}

func TestFFprobeH3AudioDurationUsesBuiltinWithoutExternalBinary(t *testing.T) {
	// PATH 置空可证明没有调用系统 ffprobe/ffmpeg。
	t.Setenv("PATH", "")
	t.Setenv("QIANTIE_FFPROBE_PATH", "")
	t.Setenv("QIANTIE_FFMPEG_PATH", "")

	audio := buildTestMP3Frames(100)
	duration, err := (FFprobeH3AudioDuration{}).Measure(context.Background(), audio)
	if err != nil {
		t.Fatalf("measure without ffprobe failed: %v", err)
	}
	if duration.Source != "builtin:mp3" || duration.MS != 2612 || duration.Estimated {
		t.Fatalf("measure = %#v, want builtin:mp3 2612ms", duration)
	}

	ms, err := (FFprobeH3AudioDuration{}).DurationMS(context.Background(), audio)
	if err != nil || ms != 2612 {
		t.Fatalf("DurationMS = %d, %v, want 2612, nil", ms, err)
	}
}

func TestFFprobeH3AudioDurationEstimatesUnknownFormatInsteadOfFailing(t *testing.T) {
	t.Setenv("PATH", "")
	t.Setenv("QIANTIE_FFPROBE_PATH", "")
	t.Setenv("QIANTIE_FFMPEG_PATH", "")

	// 无法识别的容器：30000 字节按默认 64kbps 估算为 3.75 秒。
	audio := make([]byte, 30000)
	audio[0] = 0x00
	duration, err := (FFprobeH3AudioDuration{}).Measure(context.Background(), audio)
	if err != nil {
		t.Fatalf("unknown format must fall back to estimate, got error: %v", err)
	}
	if !duration.Estimated || duration.MS != 3750 || duration.Source != "estimate:bytes@64kbps" {
		t.Fatalf("estimate = %#v, want 3750ms estimated", duration)
	}
}

func TestFFprobeH3AudioDurationExactRejectsUnknownFormat(t *testing.T) {
	t.Setenv("PATH", "")
	t.Setenv("QIANTIE_FFPROBE_PATH", "")
	t.Setenv("QIANTIE_FFMPEG_PATH", "")

	audio := make([]byte, 30000)
	if _, err := (FFprobeH3AudioDuration{}).DurationMSExact(context.Background(), audio); err == nil {
		t.Fatal("DurationMSExact must not use byte estimation")
	}
}
