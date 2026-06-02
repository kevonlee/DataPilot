package handler

import (
	"crypto/rand"
	"math/big"
	"strconv"
	"time"
)

func timeNow() time.Time {
	return time.Now()
}

func itoa(i int) string {
	return strconv.Itoa(i)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		b[i] = letters[idx.Int64()]
	}
	return string(b)
}
