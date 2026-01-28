package sessions

import (
	"crypto/rand"
	"fmt"
)

func newUUID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate uuid: %w", err)
	}

	// Version 4 UUID.
	b[6] = (b[6] & 0x0f) | 0x40
	// Variant is 10xxxxxx.
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
