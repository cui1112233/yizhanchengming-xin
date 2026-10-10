package localartifact

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	ErrInvalidMP4    = errors.New("invalid mp4 artifact")
	ErrInvalidImage  = errors.New("invalid image artifact")
	ErrTooLarge      = errors.New("artifact exceeds size limit")
	ErrInvalidID     = errors.New("invalid artifact id")
	ErrAlreadyExists = errors.New("artifact already exists")
)

var safeArtifactID = regexp.MustCompile(`^[A-Za-z0-9_-]{3,96}$`)

type Artifact struct {
	ID         string
	MediaType  string
	ByteSize   int64
	SHA256     string
	StorageRef string
}

type Store struct {
	root     string
	maxBytes int64
}

func NewStore(root string, maxBytes int64) *Store {
	if maxBytes <= 0 {
		maxBytes = 1 << 30
	}
	return &Store{root: filepath.Clean(root), maxBytes: maxBytes}
}

// Root returns the private directory that is mounted for local artifacts.
func (s *Store) Root() string {
	if s == nil {
		return ""
	}
	return s.root
}

func (s *Store) SaveMP4(id string, src io.Reader) (artifact Artifact, err error) {
	if s == nil || !safeArtifactID.MatchString(id) || src == nil {
		return Artifact{}, ErrInvalidID
	}
	if err := os.MkdirAll(s.root, 0o750); err != nil {
		return Artifact{}, err
	}
	finalName := id + ".mp4"
	finalPath := filepath.Join(s.root, finalName)
	if _, statErr := os.Stat(finalPath); statErr == nil {
		return Artifact{}, ErrAlreadyExists
	} else if !os.IsNotExist(statErr) {
		return Artifact{}, statErr
	}

	tmp, err := os.CreateTemp(s.root, ".upload-*.tmp")
	if err != nil {
		return Artifact{}, err
	}
	tmpName := tmp.Name()
	committed := false
	defer func() {
		if !committed {
			_ = tmp.Close()
			_ = os.Remove(tmpName)
		}
	}()

	first := make([]byte, 24)
	nFirst, readErr := io.ReadFull(src, first)
	if readErr != nil {
		return Artifact{}, ErrInvalidMP4
	}
	first = first[:nFirst]
	if len(first) < 12 || string(first[4:8]) != "ftyp" {
		return Artifact{}, ErrInvalidMP4
	}
	if int64(len(first)) > s.maxBytes {
		return Artifact{}, ErrTooLarge
	}

	hash := sha256.New()
	writer := io.MultiWriter(tmp, hash)
	if _, err := writer.Write(first); err != nil {
		return Artifact{}, err
	}
	remainingLimit := s.maxBytes - int64(len(first)) + 1
	copied, err := io.Copy(writer, io.LimitReader(src, remainingLimit))
	if err != nil {
		return Artifact{}, err
	}
	total := int64(len(first)) + copied
	if total > s.maxBytes {
		return Artifact{}, ErrTooLarge
	}
	if err := tmp.Sync(); err != nil {
		return Artifact{}, err
	}
	if err := tmp.Close(); err != nil {
		return Artifact{}, err
	}
	if err := os.Rename(tmpName, finalPath); err != nil {
		return Artifact{}, fmt.Errorf("commit artifact: %w", err)
	}
	committed = true
	return Artifact{
		ID: id, MediaType: "video/mp4", ByteSize: total,
		SHA256: hex.EncodeToString(hash.Sum(nil)), StorageRef: finalName,
	}, nil
}

// SaveImage writes a verified image artifact into the same candidate store as
// local-executor media. The caller supplies the declared media type; its file
// signature must agree so a PNG cannot be persisted as a JPEG (or vice versa).
func (s *Store) SaveImage(id, mediaType string, src io.Reader) (artifact Artifact, err error) {
	mediaType = strings.ToLower(strings.TrimSpace(mediaType))
	extension, err := imageExtension(mediaType)
	if err != nil {
		return Artifact{}, err
	}
	if s == nil || !safeArtifactID.MatchString(id) || src == nil {
		return Artifact{}, ErrInvalidID
	}
	if err := os.MkdirAll(s.root, 0o750); err != nil {
		return Artifact{}, err
	}
	finalName := id + extension
	finalPath := filepath.Join(s.root, finalName)
	if _, statErr := os.Stat(finalPath); statErr == nil {
		return Artifact{}, ErrAlreadyExists
	} else if !os.IsNotExist(statErr) {
		return Artifact{}, statErr
	}

	tmp, err := os.CreateTemp(s.root, ".image-upload-*.tmp")
	if err != nil {
		return Artifact{}, err
	}
	tmpName := tmp.Name()
	committed := false
	defer func() {
		if !committed {
			_ = tmp.Close()
			_ = os.Remove(tmpName)
		}
	}()

	first := make([]byte, 16)
	nFirst, readErr := io.ReadFull(src, first)
	if readErr != nil {
		return Artifact{}, ErrInvalidImage
	}
	first = first[:nFirst]
	if !matchesImageMediaType(first, mediaType) {
		return Artifact{}, ErrInvalidImage
	}
	if int64(len(first)) > s.maxBytes {
		return Artifact{}, ErrTooLarge
	}

	hash := sha256.New()
	writer := io.MultiWriter(tmp, hash)
	if _, err := writer.Write(first); err != nil {
		return Artifact{}, err
	}
	remainingLimit := s.maxBytes - int64(len(first)) + 1
	copied, err := io.Copy(writer, io.LimitReader(src, remainingLimit))
	if err != nil {
		return Artifact{}, err
	}
	total := int64(len(first)) + copied
	if total > s.maxBytes {
		return Artifact{}, ErrTooLarge
	}
	if err := tmp.Sync(); err != nil {
		return Artifact{}, err
	}
	if err := tmp.Close(); err != nil {
		return Artifact{}, err
	}
	if err := os.Rename(tmpName, finalPath); err != nil {
		return Artifact{}, fmt.Errorf("commit image artifact: %w", err)
	}
	committed = true
	return Artifact{ID: id, MediaType: mediaType, ByteSize: total, SHA256: hex.EncodeToString(hash.Sum(nil)), StorageRef: finalName}, nil
}

// SaveMedia stores a browser-uploaded audio/video file after the caller has
// validated its type. Unlike SaveImage/SaveMP4 this method intentionally does
// not inspect a codec signature; browser uploads can be WAV, MP3, M4A, WebM or
// other formats that are only meaningful to the media pipeline.
func (s *Store) SaveMedia(id, mediaType, extension string, src io.Reader) (artifact Artifact, err error) {
	if s == nil || !safeArtifactID.MatchString(id) || src == nil {
		return Artifact{}, ErrInvalidID
	}
	mediaType = strings.ToLower(strings.TrimSpace(mediaType))
	extension = strings.ToLower(strings.TrimSpace(extension))
	extension = strings.TrimPrefix(extension, ".")
	if !strings.HasPrefix(mediaType, "audio/") && !strings.HasPrefix(mediaType, "video/") {
		return Artifact{}, ErrInvalidMP4
	}
	if extension == "" || len(extension) > 8 || strings.ContainsAny(extension, `/\\.`) {
		return Artifact{}, ErrInvalidID
	}
	if err := os.MkdirAll(s.root, 0o750); err != nil {
		return Artifact{}, err
	}
	finalName := id + "." + extension
	finalPath := filepath.Join(s.root, finalName)
	if _, statErr := os.Stat(finalPath); statErr == nil {
		return Artifact{}, ErrAlreadyExists
	} else if !os.IsNotExist(statErr) {
		return Artifact{}, statErr
	}
	tmp, err := os.CreateTemp(s.root, ".media-upload-*.tmp")
	if err != nil {
		return Artifact{}, err
	}
	tmpName := tmp.Name()
	committed := false
	defer func() {
		if !committed {
			_ = tmp.Close()
			_ = os.Remove(tmpName)
		}
	}()
	hash := sha256.New()
	writer := io.MultiWriter(tmp, hash)
	count, err := io.Copy(writer, io.LimitReader(src, s.maxBytes+1))
	if err != nil {
		return Artifact{}, err
	}
	if count == 0 {
		return Artifact{}, ErrInvalidMP4
	}
	if count > s.maxBytes {
		return Artifact{}, ErrTooLarge
	}
	if err := tmp.Sync(); err != nil {
		return Artifact{}, err
	}
	if err := tmp.Close(); err != nil {
		return Artifact{}, err
	}
	if err := os.Rename(tmpName, finalPath); err != nil {
		return Artifact{}, fmt.Errorf("commit media artifact: %w", err)
	}
	committed = true
	return Artifact{ID: id, MediaType: mediaType, ByteSize: count, SHA256: hex.EncodeToString(hash.Sum(nil)), StorageRef: finalName}, nil
}

func imageExtension(mediaType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(mediaType)) {
	case "image/png":
		return ".png", nil
	case "image/jpeg":
		return ".jpg", nil
	case "image/webp":
		return ".webp", nil
	default:
		return "", ErrInvalidImage
	}
}

func matchesImageMediaType(first []byte, mediaType string) bool {
	switch mediaType {
	case "image/png":
		return len(first) >= 8 && string(first[:8]) == "\x89PNG\r\n\x1a\n"
	case "image/jpeg":
		return len(first) >= 3 && first[0] == 0xff && first[1] == 0xd8 && first[2] == 0xff
	case "image/webp":
		return len(first) >= 12 && string(first[:4]) == "RIFF" && string(first[8:12]) == "WEBP"
	default:
		return false
	}
}

func (s *Store) Remove(storageRef string) error {
	if s == nil || filepath.Base(storageRef) != storageRef {
		return ErrInvalidID
	}
	err := os.Remove(filepath.Join(s.root, storageRef))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (s *Store) Open(storageRef string) (*os.File, error) {
	if s == nil || filepath.Base(storageRef) != storageRef {
		return nil, ErrInvalidID
	}
	return os.Open(filepath.Join(s.root, storageRef))
}

// Path returns a verified, root-confined artifact path for an uploader. It
// never accepts traversal or a missing object, so callers cannot upload an
// arbitrary server file by crafting a storage reference.
func (s *Store) Path(storageRef string) (string, error) {
	if s == nil || filepath.Base(storageRef) != storageRef {
		return "", ErrInvalidID
	}
	path := filepath.Join(s.root, storageRef)
	if _, err := os.Stat(path); err != nil {
		return "", err
	}
	return path, nil
}
