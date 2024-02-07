package common

import (
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

const ReadTimeOut = 60 * time.Second
const maxRangeHash = 100
const CommandSubscribe = 0
const CommandUnsubscribe = 1
const CommandChat = 2

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func hashString(id string) uint64 {
	hashBytes := []byte(id)
	hashValue := uint64(hashBytes[0])<<56 | uint64(hashBytes[1])<<48 |
		uint64(hashBytes[2])<<40 | uint64(hashBytes[3])<<32 |
		uint64(hashBytes[4])<<24 | uint64(hashBytes[5])<<16 |
		uint64(hashBytes[6])<<8 | uint64(hashBytes[7])
	return hashValue
}

func HashToRange(id string) int {
	hashValue := hashString(id)
	rangeValue := int(hashValue % maxRangeHash)
	return rangeValue
}
