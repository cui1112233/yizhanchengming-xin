package characterarchive

import (
	"bytes"
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func wrapCompressed(t *testing.T, compressed []byte) []byte {
	t.Helper()
	x := append([]byte(nil), compressed...)
	xor(x)
	inner, err := encrypt(x, material[:32])
	if err != nil {
		t.Fatal(err)
	}
	return wrapInner(t, inner)
}
func wrapInner(t *testing.T, inner []byte) []byte {
	t.Helper()
	outer, err := encrypt(inner, material[32:64])
	if err != nil {
		t.Fatal(err)
	}
	return encodeEnvelopeForTest(outer)
}
func zipPayload(t *testing.T, p []byte) []byte {
	t.Helper()
	var b bytes.Buffer
	z := zlib.NewWriter(&b)
	_, err := z.Write(p)
	if err != nil {
		t.Fatal(err)
	}
	if err = z.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}
func TestRejectMalformedLayers(t *testing.T) {
	valid, err := EncodeJSON([]byte(`{"name":"test","unknown":123}`))
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := base64.StdEncoding.DecodeString(string(valid))
	inner, _ := decrypt(raw[4:], material[32:64])
	compressed, _ := decrypt(inner, material[:32])
	xor(compressed)
	// Encrypt an all-zero block with no valid padding to exercise deterministic rejection.
	invalidPadded := func(key []byte) []byte {
		b, _ := aes.NewCipher(key)
		out := make([]byte, 32)
		cipher.NewCBCEncrypter(b, out[:16]).CryptBlocks(out[16:], make([]byte, 16))
		return out
	}
	outerPadding := encodeEnvelopeForTest(invalidPadded(material[32:64]))
	tests := []struct {
		name     string
		data     []byte
		contains string
	}{
		{"empty", nil, "base64"}, {"bad base64", []byte("!"), "base64"}, {"noncanonical", []byte("AB=="), "base64"},
		{"short envelope", []byte("AAAA"), "truncated"},
		{"truncated block", encodeEnvelopeForTest(raw[4 : len(raw)-1]), "length"},
		{"outer padding", outerPadding, "outer layer"},
		{"inner length", wrapInner(t, []byte{1, 2, 3}), "inner layer"},
		{"inner padding", wrapInner(t, invalidPadded(material[:32])), "inner layer"},
		{"zlib", wrapCompressed(t, []byte("invalid zlib")), "zlib"},
		{"truncated zlib", wrapCompressed(t, compressed[:len(compressed)-2]), "zlib"},
		{"trailing zlib", wrapCompressed(t, append(append([]byte(nil), compressed...), 0)), "trailing"},
		{"signature", wrapCompressed(t, zipPayload(t, []byte("WRONG[]"))), "signature"},
		{"JSON", wrapCompressed(t, zipPayload(t, []byte(signature+"{"))), "JSON"},
		{"UTF8", wrapCompressed(t, zipPayload(t, append([]byte(signature+`"`), 0xff, '"'))), "JSON"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeJSON(tt.data)
			if err == nil || !strings.Contains(err.Error(), tt.contains) || got != nil {
				t.Fatalf("expected %s failure with no partial data; got %v", tt.contains, err)
			}
		})
	}
	// Prefix is a format checksum, not an authentication mechanism.
	raw[0] ^= 1
	if _, err := DecodeJSON([]byte(base64.StdEncoding.EncodeToString(raw))); err == nil || !strings.Contains(err.Error(), "checksum") {
		t.Fatal("modified checksum must be rejected")
	}
}

func TestLimitsAndInvalidJSON(t *testing.T) {
	for _, v := range [][]byte{nil, []byte("{"), {'"', 0xff, '"'}} {
		if _, err := EncodeJSON(v); err == nil {
			t.Fatal("accepted invalid JSON")
		}
	}
	data := []byte(`{"a":"` + strings.Repeat("x", 10000) + `"}`)
	encoded, err := EncodeJSON(data)
	if err != nil {
		t.Fatal(err)
	}
	exact := int64(len(signature) + len(data))
	for _, l := range []Limits{{0, 1}, {1, 0}, {-1, 100}, {100, -1}} {
		if _, err := EncodeJSONWithLimits([]byte(`[]`), l); err == nil {
			t.Fatal("bad limits accepted")
		}
		if _, err := DecodeJSONWithLimits(encoded, l); err == nil {
			t.Fatal("bad limits accepted")
		}
	}
	if _, err := EncodeJSONWithLimits(data, Limits{int64(len(data) - 1), exact}); err == nil {
		t.Fatal("encode input limit ignored")
	}
	if _, err := EncodeJSONWithLimits(data, Limits{int64(len(data)), exact - 1}); err == nil {
		t.Fatal("encode inflated limit ignored")
	}
	if _, err := DecodeJSONWithLimits(encoded, Limits{int64(len(encoded) - 1), exact}); err == nil {
		t.Fatal("archive input limit ignored")
	}
	if _, err := DecodeJSONWithLimits(encoded, Limits{int64(len(encoded)), exact - 1}); err == nil {
		t.Fatal("inflate limit ignored")
	}
	if _, err := DecodeJSONWithLimits(encoded, Limits{int64(len(encoded)), exact}); err != nil {
		t.Fatal(err)
	}
}

// Explicit opt-in writes only self-created fixture data for the external native helper.
func TestWriteGoInteropFixture(t *testing.T) {
	destination := os.Getenv("CHARACTER_ARCHIVE_INTEROP_DIR")
	if destination == "" {
		t.Skip("external interop output not requested")
	}
	source, err := os.ReadFile("testdata/native.json")
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := EncodeJSON(source)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(destination, "go-generated.lschar"), encoded, 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(destination, "go-generated-expected.json"), source, 0600); err != nil {
		t.Fatal(err)
	}
}

func encodeEnvelopeForTest(outer []byte) []byte {
	sum := md5.Sum(outer)
	return []byte(base64.StdEncoding.EncodeToString(append(sum[:4:4], outer...)))
}
func TestEncodeChecksum(t *testing.T) {
	encoded, err := EncodeJSON([]byte(`[]`))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := base64.StdEncoding.DecodeString(string(encoded))
	if err != nil {
		t.Fatal(err)
	}
	sum := md5.Sum(raw[4:])
	if !bytes.Equal(raw[:4], sum[:4]) {
		t.Fatal("encoded prefix must match outer envelope checksum")
	}
}

func TestNativeFixtureASCIIWhitespace(t *testing.T) {
	native, err := os.ReadFile("testdata/native.lschar")
	if err != nil {
		t.Fatal(err)
	}
	expected, err := DecodeJSON(native)
	if err != nil {
		t.Fatal(err)
	}
	spaced := []byte(" \t\r\n")
	for i, c := range native {
		spaced = append(spaced, c)
		if i%17 == 0 {
			spaced = append(spaced, []byte(" \t\r\n")...)
		}
	}
	spaced = append(spaced, []byte(" \t\r\n")...)
	got, err := DecodeJSON(spaced)
	if err != nil || !bytes.Equal(got, expected) {
		t.Fatalf("native archive with ASCII whitespace must decode: %v", err)
	}
	if _, err := DecodeJSONWithLimits(spaced, Limits{int64(len(native)), DefaultMaxDecompressedBytes}); err == nil {
		t.Fatal("raw input limit must precede whitespace removal")
	}
	garbage := append(append([]byte(nil), native...), []byte("!!!")...)
	if _, err := DecodeJSON(garbage); err == nil {
		t.Fatal("non-whitespace garbage must be rejected")
	}
}
