package filestore

import (
	"crypto/md5"
	"encoding/hex"
)

func getEtag(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}
