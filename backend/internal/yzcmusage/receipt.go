package yzcmusage

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// A receipt contains only attribution metadata. API keys and response content
// remain in memory and must never enter this recovery directory.
type receipt struct {
	Request        Request   `json:"request"`
	CredentialHash string    `json:"credentialHash"`
	Status         int       `json:"httpStatus"`
	At             time.Time `json:"recordedAt"`
}

func (r receipt) valid() bool {
	hash, err := hex.DecodeString(r.CredentialHash)
	return err == nil && len(hash) == 32 && r.Request.ID != "" && len(r.Request.ID) <= 64 && r.Request.Owner != "" && len(r.Request.Owner) <= 32 && len(r.Request.Feature) <= 64 && !r.At.IsZero() && r.Status >= 100 && r.Status <= 599
}
func (t *Tracker) receiptDirectory(create bool) (string, error) {
	if strings.TrimSpace(t.ReceiptDir) == "" {
		return "", errors.New("usage receipt directory is not configured")
	}
	dir, err := filepath.Abs(t.ReceiptDir)
	if err != nil {
		return "", err
	}
	if create {
		if err = os.MkdirAll(dir, 0700); err != nil {
			return "", err
		}
	}
	info, err := os.Lstat(dir)
	if err != nil {
		return "", err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return "", errors.New("invalid usage receipt directory")
	}
	return dir, nil
}
func (t *Tracker) saveReceipt(r receipt) error {
	if !r.valid() {
		return errors.New("invalid usage receipt")
	}
	dir, err := t.receiptDirectory(true)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(r)
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, ".receipt-*.tmp")
	if err != nil {
		return err
	}
	tmp := file.Name()
	defer os.Remove(tmp)
	if err = file.Chmod(0600); err != nil {
		file.Close()
		return err
	}
	if _, err = file.Write(raw); err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	// Hard-link publication is atomic and never replaces an existing receipt.
	// The first recorded actor therefore survives duplicate deliveries/restarts.
	dest := filepath.Join(dir, "receipt-"+Fingerprint(r.Request.ID)+".json")
	if err = os.Link(tmp, dest); err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}
func (t *Tracker) recordReceipt(ctx context.Context, r Request, hash string, status int, at time.Time) {
	if err := t.Store.recordAt(ctx, r, hash, status, at); err != nil {
		if spoolErr := t.saveReceipt(receipt{Request: r, CredentialHash: hash, Status: status, At: at}); spoolErr != nil {
			log.Print("YZCM usage: request persistence and recovery receipt failed")
		} else {
			log.Print("YZCM usage: request deferred to durable recovery receipt")
		}
	}
}
func (t *Tracker) flushReceipts(parent context.Context) {
	dir, err := t.receiptDirectory(false)
	if err != nil {
		return
	}
	directory, err := os.Open(dir)
	if err != nil {
		return
	}
	defer directory.Close()
	// Bound both the number of receipts and database wait on each pass.
	entries, err := directory.ReadDir(100)
	if err != nil && err != io.EOF {
		return
	}
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	for _, entry := range entries {
		if ctx.Err() != nil {
			return
		}
		name := entry.Name()
		if entry.Type()&os.ModeSymlink != 0 || entry.IsDir() || !strings.HasPrefix(name, "receipt-") || !strings.HasSuffix(name, ".json") || len(name) != len("receipt-")+64+len(".json") {
			continue
		}
		path := filepath.Join(dir, name)
		file, err := os.Open(path)
		if err != nil {
			continue
		}
		raw, readErr := io.ReadAll(io.LimitReader(file, 16385))
		file.Close()
		if readErr != nil || len(raw) > 16384 {
			continue
		}
		var r receipt
		if json.Unmarshal(raw, &r) != nil || !r.valid() || name != "receipt-"+Fingerprint(r.Request.ID)+".json" {
			continue
		}
		if err = t.Store.recordAt(ctx, r.Request, r.CredentialHash, r.Status, r.At); err != nil {
			continue
		}
		if err = os.Remove(path); err != nil && !os.IsNotExist(err) {
			log.Print("YZCM usage: recovered receipt cleanup failed")
		}
	}
}
