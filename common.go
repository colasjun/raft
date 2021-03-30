package main

import (
	"math/rand"
	"time"
)

// 返回随机数
func RandInt64(min, max int64) int64 {
	rand.Seed(time.Now().UnixNano())
	return rand.Int63n(max - min) + min
}

// 返回毫秒数
func getMillSecond() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}