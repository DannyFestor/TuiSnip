package memoization

import (
	"crypto/sha256"
	"fmt"
)

type HInt int

func (h HInt) Hash() string {
	return fmt.Sprintf("%x", sha256.Sum256(fmt.Appendf(nil, "%d", h)))
}
