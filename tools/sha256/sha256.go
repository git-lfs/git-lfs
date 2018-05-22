package sha256

import (
	"hash"

	"github.com/minio/sha256-simd"
)

func New() hash.Hash {
	return sha256.New()
}
