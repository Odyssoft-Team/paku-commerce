package id

import (
	"crypto/rand"
	"encoding/hex"
)

func NewRequestID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
