// Package characterarchive reads and writes the observed Luoshui V2 character
// archive format. CBC is a compatibility requirement, not authenticated storage.
package characterarchive

import (
	"bytes"
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"unicode/utf8"
)

//go:embed format_v2.bin
var material []byte

const signature = "LUOSHUI_CHAR_EXPORT_V2_"
const DefaultMaxInputBytes int64 = 128 << 20
const DefaultMaxDecompressedBytes int64 = 256 << 20

// Limits bounds supplied JSON/archive bytes and the entire inflated payload,
// including the format signature. Both values must be positive.
type Limits struct {
	MaxInputBytes        int64
	MaxDecompressedBytes int64
}

func defaults() Limits { return Limits{DefaultMaxInputBytes, DefaultMaxDecompressedBytes} }
func (l Limits) validate() error {
	if l.MaxInputBytes <= 0 || l.MaxDecompressedBytes <= 0 || l.MaxDecompressedBytes == int64(^uint64(0)>>1) {
		return fmt.Errorf("character archive: limits must be positive and decompressed limit must allow one sentinel byte")
	}
	return nil
}

func EncodeJSON(data []byte) ([]byte, error) { return EncodeJSONWithLimits(data, defaults()) }
func DecodeJSON(data []byte) ([]byte, error) { return DecodeJSONWithLimits(data, defaults()) }

func validJSON(data []byte) bool { return utf8.Valid(data) && json.Valid(data) }

func EncodeJSONWithLimits(data []byte, l Limits) ([]byte, error) {
	if err := l.validate(); err != nil {
		return nil, err
	}
	if int64(len(data)) > l.MaxInputBytes {
		return nil, fmt.Errorf("character archive: input exceeds limit")
	}
	if !validJSON(data) {
		return nil, fmt.Errorf("character archive: invalid UTF-8 JSON")
	}
	if int64(len(data))+int64(len(signature)) > l.MaxDecompressedBytes {
		return nil, fmt.Errorf("character archive: decompressed payload exceeds limit")
	}
	var compressed bytes.Buffer
	zw := zlib.NewWriter(&compressed)
	if _, err := zw.Write([]byte(signature)); err != nil {
		return nil, err
	}
	if _, err := zw.Write(data); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	x := compressed.Bytes()
	xor(x)
	inner, err := encrypt(x, material[:32])
	if err != nil {
		return nil, err
	}
	outer, err := encrypt(inner, material[32:64])
	if err != nil {
		return nil, err
	}
	raw := make([]byte, 4, len(outer)+4)
	// MD5 prefix is required by the native file format; it provides no authentication.
	sum := md5.Sum(outer)
	copy(raw, sum[:4])
	raw = append(raw, outer...)
	result := make([]byte, base64.StdEncoding.EncodedLen(len(raw)))
	base64.StdEncoding.Encode(result, raw)
	return result, nil
}

func DecodeJSONWithLimits(data []byte, l Limits) ([]byte, error) {
	if err := l.validate(); err != nil {
		return nil, err
	}
	if int64(len(data)) > l.MaxInputBytes {
		return nil, fmt.Errorf("character archive: input exceeds limit")
	}
	// The native decoder tolerates ASCII space, tab, CR and LF. Strip only
	// those characters, after enforcing the raw input limit. Unlike the native
	// permissive Base64 decoder, reject other garbage and noncanonical pad bits.
	clean := make([]byte, 0, len(data))
	for _, c := range data {
		switch c {
		case ' ', '\t', '\r', '\n':
			continue
		}
		clean = append(clean, c)
	}
	if len(clean) == 0 {
		return nil, fmt.Errorf("character archive: invalid base64")
	}
	data = clean
	raw, err := base64.StdEncoding.Strict().DecodeString(string(data))
	if err != nil {
		return nil, fmt.Errorf("character archive: invalid base64: %w", err)
	}
	if len(raw) < 4+16+16 {
		return nil, fmt.Errorf("character archive: truncated outer envelope")
	}
	// Verify the format checksum before decrypting either CBC layer.
	sum := md5.Sum(raw[4:])
	if !bytes.Equal(raw[:4], sum[:4]) {
		return nil, fmt.Errorf("character archive: checksum mismatch")
	}
	inner, err := decrypt(raw[4:], material[32:64])
	if err != nil {
		return nil, fmt.Errorf("character archive: outer layer: %w", err)
	}
	compressed, err := decrypt(inner, material[:32])
	if err != nil {
		return nil, fmt.Errorf("character archive: inner layer: %w", err)
	}
	xor(compressed)
	source := bytes.NewReader(compressed)
	zr, err := zlib.NewReader(source)
	if err != nil {
		return nil, fmt.Errorf("character archive: zlib: %w", err)
	}
	plain, readErr := io.ReadAll(io.LimitReader(zr, l.MaxDecompressedBytes+1))
	closeErr := zr.Close()
	if int64(len(plain)) > l.MaxDecompressedBytes {
		return nil, fmt.Errorf("character archive: decompressed payload exceeds limit")
	}
	if readErr != nil {
		return nil, fmt.Errorf("character archive: zlib: %w", readErr)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("character archive: zlib close: %w", closeErr)
	}
	if source.Len() != 0 {
		return nil, fmt.Errorf("character archive: trailing compressed data")
	}
	if !bytes.HasPrefix(plain, []byte(signature)) {
		return nil, fmt.Errorf("character archive: invalid signature")
	}
	result := plain[len(signature):]
	if !validJSON(result) {
		return nil, fmt.Errorf("character archive: invalid UTF-8 JSON")
	}
	return result, nil
}

func xor(b []byte) {
	for i := range b {
		b[i] ^= material[64+i%32]
	}
}
func encrypt(plain, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	pad := aes.BlockSize - len(plain)%aes.BlockSize
	padded := make([]byte, len(plain)+pad)
	copy(padded, plain)
	for i := len(plain); i < len(padded); i++ {
		padded[i] = byte(pad)
	}
	result := make([]byte, aes.BlockSize+len(padded))
	if _, err := rand.Read(result[:aes.BlockSize]); err != nil {
		return nil, err
	}
	cipher.NewCBCEncrypter(block, result[:aes.BlockSize]).CryptBlocks(result[aes.BlockSize:], padded)
	return result, nil
}
func decrypt(raw, key []byte) ([]byte, error) {
	if len(raw) < 2*aes.BlockSize || (len(raw)-aes.BlockSize)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("invalid IV/ciphertext length")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	plain := make([]byte, len(raw)-aes.BlockSize)
	cipher.NewCBCDecrypter(block, raw[:aes.BlockSize]).CryptBlocks(plain, raw[aes.BlockSize:])
	pad := int(plain[len(plain)-1])
	if pad < 1 || pad > aes.BlockSize || pad > len(plain) {
		return nil, fmt.Errorf("invalid PKCS7 padding")
	}
	for _, v := range plain[len(plain)-pad:] {
		if int(v) != pad {
			return nil, fmt.Errorf("invalid PKCS7 padding")
		}
	}
	return plain[:len(plain)-pad], nil
}
