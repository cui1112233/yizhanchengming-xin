package batchfactoryv11

import (
	"context"
	"encoding/binary"
	"math"
)

// AudioDuration 描述一次媒体时长测量的结果。
// Estimated 为 true 表示没有拿到精确时长，数值来自字节数/码率估算，
// 调用方应把它当作“估算值”继续后续流程，而不是判定失败。
type AudioDuration struct {
	MS        int64
	Source    string
	Estimated bool
}

type audioDurationMeasurer interface {
	Measure(context.Context, []byte) (AudioDuration, error)
}

// MeasureAudioDuration 返回带来源与估算标记的时长。
// 内置解析器实现 audioDurationMeasurer；测试或自定义探针保持原有
// DurationMS 语义，来源标记为 custom。
func MeasureAudioDuration(ctx context.Context, probe H3AudioDurationProbe, data []byte) (AudioDuration, error) {
	if measurer, ok := probe.(audioDurationMeasurer); ok {
		return measurer.Measure(ctx, data)
	}
	ms, err := probe.DurationMS(ctx, data)
	if err != nil {
		return AudioDuration{}, err
	}
	return AudioDuration{MS: ms, Source: "custom"}, nil
}

// parseBuiltinAudioDuration 在纯 Go 内解析常见容器/编码的时长，不依赖 PATH。
// 目前支持 WAV、MP3、M4A/MP4、FLAC、AAC(ADTS)。
func parseBuiltinAudioDuration(data []byte) (AudioDuration, bool) {
	if len(data) < 12 {
		return AudioDuration{}, false
	}
	switch {
	case string(data[:4]) == "RIFF" && string(data[8:12]) == "WAVE":
		if ms, ok := parseWAVDuration(data); ok {
			return AudioDuration{MS: ms, Source: "builtin:wav"}, true
		}
	case string(data[:4]) == "fLaC":
		if ms, ok := parseFLACDuration(data); ok {
			return AudioDuration{MS: ms, Source: "builtin:flac"}, true
		}
	case isISOBaseMedia(data):
		if ms, ok := parseMP4Duration(data); ok {
			return AudioDuration{MS: ms, Source: "builtin:mp4"}, true
		}
	}
	if duration, ok := parseMP3Duration(data); ok {
		return duration, true
	}
	if duration, ok := parseADTSDuration(data); ok {
		return duration, true
	}
	return AudioDuration{}, false
}

// skipID3v2 跳过 MP3/ADTS 文件头部的 ID3v2 标签，返回首个音频帧的候选偏移。
func skipID3v2(data []byte) int {
	if len(data) < 10 || string(data[:3]) != "ID3" {
		return 0
	}
	size := int(data[6])<<21 | int(data[7])<<14 | int(data[8])<<7 | int(data[9])
	end := 10 + size
	if end < 10 || end > len(data) {
		return 0
	}
	return end
}

func parseWAVDuration(data []byte) (int64, bool) {
	if len(data) < 24 {
		return 0, false
	}
	byteRate := 0
	dataSize := -1
	for pos := 12; pos+8 <= len(data); {
		chunkID := string(data[pos : pos+4])
		chunkSize := int(binary.LittleEndian.Uint32(data[pos+4 : pos+8]))
		body := pos + 8
		switch chunkID {
		case "fmt ":
			if body+16 > len(data) {
				return 0, false
			}
			channels := int(binary.LittleEndian.Uint16(data[body+2 : body+4]))
			sampleRate := int(binary.LittleEndian.Uint32(data[body+4 : body+8]))
			byteRate = int(binary.LittleEndian.Uint32(data[body+8 : body+12]))
			bitsPerSample := int(binary.LittleEndian.Uint16(data[body+14 : body+16]))
			if byteRate <= 0 && sampleRate > 0 && channels > 0 && bitsPerSample > 0 {
				byteRate = sampleRate * channels * bitsPerSample / 8
			}
		case "data":
			available := len(data) - body
			if chunkSize > 0 && chunkSize <= available {
				dataSize = chunkSize
			} else {
				// 流式写入或截断的文件，data 块长度不可信时按剩余字节计算。
				dataSize = available
			}
			pos = len(data)
			continue
		}
		if chunkSize < 0 {
			return 0, false
		}
		pos = body + chunkSize + chunkSize&1
	}
	if byteRate <= 0 || dataSize <= 0 {
		return 0, false
	}
	ms := int64(math.Round(float64(dataSize) / float64(byteRate) * 1000))
	return ms, ms > 0
}

type mp3FrameHeader struct {
	frameLen    int
	sampleRate  int
	samples     int
	bitrateKbps int
}

var (
	mpeg1Layer1Bitrates = []int{0, 32, 64, 96, 128, 160, 192, 224, 256, 288, 320, 352, 384, 416, 448}
	mpeg1Layer2Bitrates = []int{0, 32, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, 384}
	mpeg1Layer3Bitrates = []int{0, 32, 40, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320}
	mpeg2Layer1Bitrates = []int{0, 32, 48, 56, 64, 80, 96, 112, 128, 144, 160, 176, 192, 224, 256}
	mpeg2Layer23Bitrate = []int{0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160}
)

func mp3BitrateKbps(versionBits, layerBits byte, index int) int {
	if index <= 0 || index >= len(mpeg1Layer1Bitrates) {
		return 0
	}
	switch {
	case versionBits == 3:
		switch layerBits {
		case 3:
			return mpeg1Layer1Bitrates[index]
		case 2:
			return mpeg1Layer2Bitrates[index]
		default:
			return mpeg1Layer3Bitrates[index]
		}
	default:
		if layerBits == 3 {
			return mpeg2Layer1Bitrates[index]
		}
		return mpeg2Layer23Bitrate[index]
	}
}

func mp3SampleRate(versionBits byte, index int) int {
	switch versionBits {
	case 3:
		return []int{44100, 48000, 32000}[index]
	case 2:
		return []int{22050, 24000, 16000}[index]
	default:
		return []int{11025, 12000, 8000}[index]
	}
}

func parseMP3FrameHeader(data []byte) (mp3FrameHeader, bool) {
	if len(data) < 4 || data[0] != 0xFF || data[1]&0xE0 != 0xE0 {
		return mp3FrameHeader{}, false
	}
	versionBits := (data[1] >> 3) & 0x03
	layerBits := (data[1] >> 1) & 0x03
	bitrateIndex := int((data[2] >> 4) & 0x0F)
	sampleRateIndex := int((data[2] >> 2) & 0x03)
	padding := int((data[2] >> 1) & 0x01)
	if versionBits == 1 || layerBits == 0 || sampleRateIndex == 3 {
		return mp3FrameHeader{}, false
	}
	bitrateKbps := mp3BitrateKbps(versionBits, layerBits, bitrateIndex)
	if bitrateKbps <= 0 {
		return mp3FrameHeader{}, false
	}
	sampleRate := mp3SampleRate(versionBits, sampleRateIndex)
	samples := 1152
	switch {
	case layerBits == 3:
		samples = 384
	case layerBits == 2:
		samples = 1152
	case versionBits != 3:
		samples = 576
	}
	frameLen := (samples/8)*bitrateKbps*1000/sampleRate + padding
	if frameLen < 4 {
		return mp3FrameHeader{}, false
	}
	return mp3FrameHeader{frameLen: frameLen, sampleRate: sampleRate, samples: samples, bitrateKbps: bitrateKbps}, true
}

func parseMP3Duration(data []byte) (AudioDuration, bool) {
	start := skipID3v2(data)
	header, pos, ok := findMP3Frame(data, start)
	if !ok {
		return AudioDuration{}, false
	}
	// Xing/Info 帧记录总帧数，并可通过 LAME 标签扣除编码器前后填充，
	// 结果与 ffprobe 的 format duration 一致（帧计数法会多出约 1-2 帧）。
	if xing, ok := parseMP3XingDuration(data, pos, header); ok && xing > 0 {
		return AudioDuration{MS: xing, Source: "builtin:mp3"}, true
	}
	totalSamples := 0
	frames := 0
	for cur := pos; cur+4 <= len(data); {
		frame, ok := parseMP3FrameHeader(data[cur : cur+4])
		if !ok || frame.frameLen <= 0 {
			break
		}
		totalSamples += frame.samples
		frames++
		cur += frame.frameLen
	}
	if frames > 0 && header.sampleRate > 0 {
		ms := int64(math.Round(float64(totalSamples) / float64(header.sampleRate) * 1000))
		if ms > 0 {
			return AudioDuration{MS: ms, Source: "builtin:mp3"}, true
		}
	}
	// 首帧头可读但帧链被截断/拼接时，用首帧码率按字节数估算。
	payload := len(data) - pos
	if payload > 0 && header.bitrateKbps > 0 {
		seconds := float64(payload) * 8 / float64(header.bitrateKbps*1000)
		ms := int64(math.Round(seconds * 1000))
		if ms > 0 {
			return AudioDuration{MS: ms, Source: "estimate:mp3-cbr", Estimated: true}, true
		}
	}
	return AudioDuration{}, false
}

// parseMP3XingDuration 解析首帧内的 Xing/Info 头以及 LAME 延迟/填充信息。
func parseMP3XingDuration(data []byte, frameStart int, header mp3FrameHeader) (int64, bool) {
	if frameStart+4 > len(data) || header.sampleRate <= 0 || header.samples <= 0 {
		return 0, false
	}
	versionBits := (data[frameStart+1] >> 3) & 0x03
	mono := (data[frameStart+3] >> 6) == 3
	sideInfo := 17
	switch {
	case versionBits == 3 && mono:
		sideInfo = 17
	case versionBits == 3:
		sideInfo = 32
	case mono:
		sideInfo = 9
	default:
		sideInfo = 17
	}
	crcLen := 0
	if data[frameStart+1]&0x01 == 0 {
		crcLen = 2
	}
	pos := frameStart + 4 + crcLen + sideInfo
	if pos+8 > len(data) {
		return 0, false
	}
	if string(data[pos:pos+4]) != "Xing" && string(data[pos:pos+4]) != "Info" {
		return 0, false
	}
	flags := binary.BigEndian.Uint32(data[pos+4 : pos+8])
	fieldPos := pos + 8
	var frameCount int
	if flags&0x01 != 0 {
		if fieldPos+4 > len(data) {
			return 0, false
		}
		frameCount = int(binary.BigEndian.Uint32(data[fieldPos : fieldPos+4]))
		fieldPos += 4
	}
	if flags&0x02 != 0 {
		fieldPos += 4
	}
	if flags&0x04 != 0 {
		fieldPos += 100
	}
	if flags&0x08 != 0 {
		fieldPos += 4
	}
	if frameCount <= 0 {
		return 0, false
	}
	delay, padding := 0, 0
	// LAME 标签：9 字节编码器版本 + 12 字节增益/低通/峰值等字段，
	// delay/padding 位于标签起始后第 21 字节（各 12 位）。
	if fieldPos+24 <= len(data) {
		tag := data[fieldPos+21 : fieldPos+24]
		delay = int(tag[0])<<4 | int(tag[1])>>4
		padding = (int(tag[1])&0x0F)<<8 | int(tag[2])
		if delay > header.sampleRate || padding > header.sampleRate {
			delay, padding = 0, 0
		}
	}
	totalSamples := frameCount*header.samples - delay - padding
	if totalSamples <= 0 {
		totalSamples = frameCount * header.samples
	}
	ms := int64(math.Round(float64(totalSamples) / float64(header.sampleRate) * 1000))
	if ms <= 0 {
		return 0, false
	}
	return ms, true
}

func findMP3Frame(data []byte, start int) (mp3FrameHeader, int, bool) {
	if start < 0 || start+4 > len(data) {
		return mp3FrameHeader{}, 0, false
	}
	limit := start + 64*1024
	if limit > len(data) {
		limit = len(data)
	}
	for i := start; i+4 <= limit; i++ {
		if data[i] != 0xFF || data[i+1]&0xE0 != 0xE0 {
			continue
		}
		if header, ok := parseMP3FrameHeader(data[i : i+4]); ok {
			return header, i, true
		}
	}
	return mp3FrameHeader{}, 0, false
}

func isISOBaseMedia(data []byte) bool {
	if len(data) < 12 {
		return false
	}
	switch string(data[4:8]) {
	case "ftyp", "moov", "mdat", "free", "skip", "wide":
		return true
	}
	return false
}

type mp4Box struct {
	bodyStart int
	bodyEnd   int
}

func findMP4Box(data []byte, boxType string, start, end int) (mp4Box, bool) {
	if start < 0 || end > len(data) {
		return mp4Box{}, false
	}
	for pos := start; pos+8 <= end; {
		size := int(binary.BigEndian.Uint32(data[pos : pos+4]))
		headerLen := 8
		if size == 1 {
			if pos+16 > end {
				return mp4Box{}, false
			}
			size = int(binary.BigEndian.Uint64(data[pos+8 : pos+16]))
			headerLen = 16
		} else if size == 0 {
			size = end - pos
		}
		if size < headerLen || pos+size > end {
			return mp4Box{}, false
		}
		if string(data[pos+4:pos+8]) == boxType {
			return mp4Box{bodyStart: pos + headerLen, bodyEnd: pos + size}, true
		}
		pos += size
	}
	return mp4Box{}, false
}

func parseMP4Duration(data []byte) (int64, bool) {
	moov, ok := findMP4Box(data, "moov", 0, len(data))
	if !ok {
		return 0, false
	}
	mvhd, ok := findMP4Box(data, "mvhd", moov.bodyStart, moov.bodyEnd)
	if !ok {
		return 0, false
	}
	version := data[mvhd.bodyStart]
	var timescale uint32
	var duration uint64
	switch version {
	case 0:
		if mvhd.bodyStart+20 > mvhd.bodyEnd {
			return 0, false
		}
		timescale = binary.BigEndian.Uint32(data[mvhd.bodyStart+12 : mvhd.bodyStart+16])
		duration = uint64(binary.BigEndian.Uint32(data[mvhd.bodyStart+16 : mvhd.bodyStart+20]))
	case 1:
		if mvhd.bodyStart+28 > mvhd.bodyEnd {
			return 0, false
		}
		timescale = binary.BigEndian.Uint32(data[mvhd.bodyStart+20 : mvhd.bodyStart+24])
		duration = binary.BigEndian.Uint64(data[mvhd.bodyStart+24 : mvhd.bodyStart+32])
	default:
		return 0, false
	}
	if timescale == 0 || duration == 0 {
		// 分片 MP4 的 mvhd duration 可能为 0，交给外部探针或估算路径。
		return 0, false
	}
	ms := int64(math.Round(float64(duration) / float64(timescale) * 1000))
	return ms, ms > 0
}

func parseFLACDuration(data []byte) (int64, bool) {
	if len(data) < 8+18 || string(data[:4]) != "fLaC" || data[4]&0x7F != 0 {
		return 0, false
	}
	streamInfo := data[8:]
	if len(streamInfo) < 18 {
		return 0, false
	}
	word := binary.BigEndian.Uint64(streamInfo[10:18])
	sampleRate := int(word >> 44)
	totalSamples := word & ((1 << 36) - 1)
	if sampleRate == 0 || totalSamples == 0 {
		return 0, false
	}
	ms := int64(math.Round(float64(totalSamples) / float64(sampleRate) * 1000))
	return ms, ms > 0
}

var adtsSampleRates = []int{96000, 88200, 64000, 48000, 44100, 32000, 24000, 22050, 16000, 12000, 11025, 8000, 7350}

func adtsFrameLen(data []byte) (int, bool) {
	if len(data) < 7 || data[0] != 0xFF || data[1]&0xF6 != 0xF0 {
		return 0, false
	}
	length := int(data[3]&0x03)<<11 | int(data[4])<<3 | int(data[5]>>5)
	if length < 7 {
		return 0, false
	}
	return length, true
}

func parseADTSDuration(data []byte) (AudioDuration, bool) {
	start := skipID3v2(data)
	pos := -1
	limit := start + 64*1024
	if limit > len(data) {
		limit = len(data)
	}
	for i := start; i+7 <= limit; i++ {
		if data[i] != 0xFF || data[i+1]&0xF6 != 0xF0 {
			continue
		}
		if _, ok := adtsFrameLen(data[i : i+7]); ok {
			pos = i
			break
		}
	}
	if pos < 0 {
		return AudioDuration{}, false
	}
	sampleRateIndex := int((data[pos+2] >> 2) & 0x0F)
	if sampleRateIndex >= len(adtsSampleRates) || adtsSampleRates[sampleRateIndex] <= 0 {
		return AudioDuration{}, false
	}
	sampleRate := adtsSampleRates[sampleRateIndex]
	frames := 0
	for cur := pos; cur+7 <= len(data); {
		length, ok := adtsFrameLen(data[cur : cur+7])
		if !ok {
			break
		}
		frames++
		cur += length
	}
	if frames == 0 {
		return AudioDuration{}, false
	}
	ms := int64(math.Round(float64(frames*1024) / float64(sampleRate) * 1000))
	if ms <= 0 {
		return AudioDuration{}, false
	}
	return AudioDuration{MS: ms, Source: "builtin:adts"}, true
}

// estimateAudioDurationByBytes 是最后的降级路径：解析不出精确时长时按
// 64kbps 默认码率估算字节时长，并明确标记为估算值，避免整条配音流程失败。
func estimateAudioDurationByBytes(data []byte) AudioDuration {
	const defaultBitrateKbps = 64
	seconds := float64(len(data)) * 8 / float64(defaultBitrateKbps*1000)
	ms := int64(math.Round(seconds * 1000))
	if ms < 1 {
		ms = 1
	}
	return AudioDuration{MS: ms, Source: "estimate:bytes@64kbps", Estimated: true}
}
