package material

import (
	"crypto/rand"
	"encoding/hex"
)

func NewID() string     { return "mcm_" + randomHex(16) }
func NewTaskID() string { return "mct_" + randomHex(16) }

func randomHex(size int) string {
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		panic(err)
	}
	return hex.EncodeToString(buffer)
}
