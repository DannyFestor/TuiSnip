package memoization

import (
	"crypto/sha256"
	"fmt"
)

type HString string

func (h HString) Hash() string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(h)))
}
