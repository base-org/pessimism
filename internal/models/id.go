package models

import (
	"crypto/sha256"
	"strings"
)

type ID = [16]byte

func NilID() ID {
	return [16]byte{0}
}

func Strings2ID(tokens ...string) ID {
	concatStr := strings.Join(tokens[:], "")
	bytes := sha256.New().Sum([]byte(concatStr))

	return *(*[16]byte)(bytes)
}
