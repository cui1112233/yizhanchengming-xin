package project

import (
	"crypto/rand"
	"encoding/hex"
)

func NewID() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		panic("secure random project ID unavailable: " + err.Error())
	}
	return "mcp_" + hex.EncodeToString(value[:])
}
