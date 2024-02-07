package redis

import "fmt"

const PubSubChannel = "pub:sub"

func XStreamListMessage(room string) string {
	return fmt.Sprintf("r:%s", room)
}
