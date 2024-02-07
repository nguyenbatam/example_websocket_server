package services

import (
	"context"
	"github.com/nguyenbatam/example_websocket_server/models/protobuf"
	"github.com/nguyenbatam/example_websocket_server/redis"
	"google.golang.org/protobuf/proto"
	"strconv"
	"strings"
	"time"
)

const MaxLengthStream = 100

func AddMessageToRoomChat(msg *protobuf.MessageChat) error {
	msg.Timestamp = uint64(time.Now().UnixMilli())
	bin, _ := proto.Marshal(msg)
	data := map[string]interface{}{
		"data": bin,
	}
	err := redis.AddItemToXStream(context.Background(), redis.XStreamListMessage(msg.Room), data, MaxLengthStream)
	if err != nil {
		return err
	}
	err = redis.PublishMsg(context.Background(), redis.PubSubChannel, string(bin))
	return err
}

func GetMessageByRoomChat(ctx context.Context, room string, start int64, limit int64) ([]protobuf.MessageChat, uint64, error) {
	messages, err := redis.GetItemsByXStream(ctx, redis.XStreamListMessage(room), start, limit)
	if err != nil {
		return nil, 0, err
	}
	if messages == nil {
		return nil, 0, nil
	}
	nextId := "0-0"
	var list []protobuf.MessageChat
	for _, message := range messages {
		bin := message.Values["data"].(string)
		var m protobuf.MessageChat
		err := proto.Unmarshal([]byte(bin), &m)
		if err != nil {
			continue
		}
		nextId = message.ID
		list = append(list, m)
	}
	next := strings.Split(nextId, "-")[0]
	nextLast, _ := strconv.ParseUint(next, 10, 64)
	return list, nextLast, nil
}
