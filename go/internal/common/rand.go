package common

import (
	"math/rand"
	"sync"
	"time"
)

var (
	defaultRand      = rand.New(rand.NewSource(time.Now().UnixNano()))
	defaultRandMutex = &sync.Mutex{}
)

// RandIntn returns a random int.

func RandIntn(n int) (v int) {
	defaultRandMutex.Lock()
	v = defaultRand.Intn(n)
	defaultRandMutex.Unlock()
	return
}
